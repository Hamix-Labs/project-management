import { useLayoutEffect, useEffect, useRef } from "react";
import ReactDOM from "react-dom/client";
import type { Root } from "react-dom/client";
import { CodeLanguageToolbar } from "./CodeLanguageToolbar";
import {
  promptCodeLanguages,
  type PromptCodeLanguage,
} from "./promptCodeBlockOptions";

type MountRecord = {
  root: Root;
  select: HTMLSelectElement;
  onSelectChange: () => void;
  mountEl: HTMLElement;
  block: HTMLElement;
};

function resolveCreateRoot(): typeof ReactDOM.createRoot {
  const mod = ReactDOM as typeof ReactDOM & {
    default?: { createRoot?: typeof ReactDOM.createRoot };
  };
  const fn = mod.createRoot ?? mod.default?.createRoot;
  if (typeof fn !== "function") {
    throw new Error("react-dom/client createRoot is unavailable");
  }
  return fn;
}

function hideNativeSelect(select: HTMLSelectElement) {
  select.hidden = true;
  select.setAttribute("aria-hidden", "true");
  select.tabIndex = -1;
  select.style.setProperty("display", "none", "important");
  const wrap = select.parentElement;
  if (wrap instanceof HTMLElement) {
    wrap.style.cssText =
      "position:absolute;width:0;height:0;overflow:hidden;opacity:0;pointer-events:none;border:none;margin:0;padding:0;";
  }
}

function positionMount(
  host: HTMLElement,
  block: HTMLElement,
  mountEl: HTMLElement,
) {
  const hostRect = host.getBoundingClientRect();
  const blockRect = block.getBoundingClientRect();
  const top = blockRect.top - hostRect.top + host.scrollTop + 8;
  const right = hostRect.right - blockRect.right + host.scrollLeft + 10;
  mountEl.style.top = `${Math.max(0, top)}px`;
  mountEl.style.right = `${Math.max(0, right)}px`;
  mountEl.style.left = "auto";
}

function findSelect(block: Element): HTMLSelectElement | null {
  const select = block.querySelector(":scope > div select");
  return select instanceof HTMLSelectElement ? select : null;
}

/**
 * BlockNote's code block uses content:"plain", which createReactBlockSpec cannot
 * host. We keep createCodeBlockSpec (highlighter + shortcuts) and replace the
 * native <select> chrome with a Notion-like searchable toolbar.
 *
 * The toolbar is portaled onto the editor host (outside ProseMirror). Do not key
 * mount lifetime off data-* on the node view — ProseMirror strips those and that
 * previously remounted a second toolbar on top of the native select every sweep.
 */
export function useEnhanceCodeBlockToolbars(
  container: HTMLElement | null,
  disabled: boolean,
) {
  const mountsRef = useRef(new Map<Element, MountRecord>());
  const languagesRef = useRef<PromptCodeLanguage[] | null>(null);
  if (languagesRef.current == null) {
    languagesRef.current = promptCodeLanguages();
  }
  const disabledRef = useRef(disabled);
  disabledRef.current = disabled;

  useLayoutEffect(() => {
    if (!container) return;

    const mounts = mountsRef.current;
    let sweeping = false;
    let debounceTimer: number | null = null;
    const pollTimers: number[] = [];

    const teardownMount = (block: Element, record: MountRecord) => {
      record.select.removeEventListener("change", record.onSelectChange);
      mounts.delete(block);
      record.mountEl.remove();
      const root = record.root;
      queueMicrotask(() => {
        try {
          root.unmount();
        } catch {
          // Root may already be gone after a fast remount.
        }
      });
    };

    const renderToolbar = (record: MountRecord) => {
      const { root, select, block } = record;
      const codeEl = block.querySelector("code");
      root.render(
        <CodeLanguageToolbar
          languages={languagesRef.current ?? []}
          value={select.value}
          disabled={disabledRef.current || select.disabled}
          onChange={(languageId) => {
            select.value = languageId;
            select.dispatchEvent(new Event("change", { bubbles: true }));
          }}
          onCopy={async () => {
            const text = codeEl?.textContent ?? "";
            await navigator.clipboard.writeText(text);
          }}
        />,
      );
    };

    const bindSelect = (
      record: MountRecord,
      select: HTMLSelectElement,
    ): MountRecord => {
      if (record.select === select) {
        hideNativeSelect(select);
        return record;
      }
      record.select.removeEventListener("change", record.onSelectChange);
      const onSelectChange = () => {
        const current = mounts.get(record.block);
        if (current) renderToolbar(current);
      };
      select.addEventListener("change", onSelectChange);
      const next: MountRecord = {
        ...record,
        select,
        onSelectChange,
      };
      mounts.set(record.block, next);
      hideNativeSelect(select);
      renderToolbar(next);
      return next;
    };

    const enhanceBlock = (block: Element) => {
      if (!(block instanceof HTMLElement)) return;

      const select = findSelect(block);
      if (!select) return;

      const existing = mounts.get(block);
      if (existing) {
        bindSelect(existing, select);
        positionMount(container, block, existing.mountEl);
        return;
      }

      // Re-associate if ProseMirror replaced the block element but left our portal.
      for (const [prevBlock, record] of mounts) {
        if (prevBlock === block) continue;
        if (prevBlock.isConnected) continue;
        if (!record.mountEl.isConnected) continue;
        mounts.delete(prevBlock);
        const moved: MountRecord = { ...record, block };
        mounts.set(block, moved);
        bindSelect(moved, select);
        positionMount(container, block, moved.mountEl);
        return;
      }

      const toolbarId = `code-${Math.random().toString(36).slice(2, 10)}`;
      const mount = document.createElement("div");
      mount.className = "prompt-code-toolbar-root";
      mount.dataset.forBlock = toolbarId;
      container.appendChild(mount);
      positionMount(container, block, mount);

      const pre = block.querySelector("pre");
      const code = block.querySelector("code");
      pre?.setAttribute("spellcheck", "false");
      code?.setAttribute("spellcheck", "false");

      try {
        const root = resolveCreateRoot()(mount);
        const onSelectChange = () => {
          const current = mounts.get(block);
          if (current) renderToolbar(current);
        };
        select.addEventListener("change", onSelectChange);
        hideNativeSelect(select);
        const record: MountRecord = {
          root,
          select,
          onSelectChange,
          mountEl: mount,
          block,
        };
        mounts.set(block, record);
        renderToolbar(record);
      } catch (err) {
        mount.remove();
        console.error("prompt-editor: code toolbar mount failed", err);
      }
    };

    const sweep = () => {
      if (sweeping) return;
      sweeping = true;
      try {
        const live = new Set<Element>();
        container
          .querySelectorAll('[data-content-type="codeBlock"]')
          .forEach((block) => {
            enhanceBlock(block);
            live.add(block);
          });

        for (const [block, record] of mounts) {
          if (live.has(block) && block.isConnected) continue;
          // Keep portal if we can re-associate on the next enhance pass.
          if (!block.isConnected && record.mountEl.isConnected) continue;
          teardownMount(block, record);
        }

        // Drop orphan portals left by remount bugs.
        container.querySelectorAll(":scope > .prompt-code-toolbar-root").forEach((el) => {
          const kept = [...mounts.values()].some((r) => r.mountEl === el);
          if (!kept) el.remove();
        });
      } finally {
        sweeping = false;
      }
    };

    const scheduleSweep = () => {
      if (debounceTimer != null) window.clearTimeout(debounceTimer);
      debounceTimer = window.setTimeout(() => {
        debounceTimer = null;
        sweep();
      }, 32);
    };

    const repositionAll = () => {
      for (const [, record] of mounts) {
        if (!record.block.isConnected) continue;
        positionMount(container, record.block, record.mountEl);
      }
    };

    sweep();
    for (const ms of [0, 50, 200]) {
      pollTimers.push(window.setTimeout(sweep, ms));
    }

    const mo = new MutationObserver((mutations) => {
      for (const m of mutations) {
        const t = m.target;
        if (t instanceof Element && t.closest(".prompt-code-toolbar-root")) {
          continue;
        }
        const nodes = [...m.addedNodes, ...m.removedNodes];
        if (
          m.type === "childList" &&
          nodes.length > 0 &&
          nodes.every(
            (n) =>
              n instanceof Element &&
              (n.classList.contains("prompt-code-toolbar-root") ||
                !!n.closest(".prompt-code-toolbar-root")),
          )
        ) {
          continue;
        }
        // Ignore decoration churn inside an already-enhanced code block.
        if (
          t instanceof Element &&
          t.closest('[data-content-type="codeBlock"]') &&
          !nodes.some(
            (n) =>
              n instanceof Element &&
              (n.matches?.('[data-content-type="codeBlock"]') ||
                !!n.querySelector?.('[data-content-type="codeBlock"]')),
          )
        ) {
          continue;
        }
        scheduleSweep();
        return;
      }
    });
    mo.observe(container, {
      childList: true,
      subtree: true,
      attributes: true,
      attributeFilter: ["data-content-type"],
    });

    container.addEventListener("scroll", repositionAll, { passive: true });
    window.addEventListener("resize", repositionAll);

    return () => {
      mo.disconnect();
      container.removeEventListener("scroll", repositionAll);
      window.removeEventListener("resize", repositionAll);
      if (debounceTimer != null) window.clearTimeout(debounceTimer);
      for (const id of pollTimers) window.clearTimeout(id);
      for (const [block, record] of [...mounts.entries()]) {
        teardownMount(block, record);
      }
      mounts.clear();
    };
  }, [container]);

  useEffect(() => {
    for (const [, record] of mountsRef.current) {
      const codeEl = record.block.querySelector("code");
      record.root.render(
        <CodeLanguageToolbar
          languages={languagesRef.current ?? []}
          value={record.select.value}
          disabled={disabled || record.select.disabled}
          onChange={(languageId) => {
            record.select.value = languageId;
            record.select.dispatchEvent(new Event("change", { bubbles: true }));
          }}
          onCopy={async () => {
            const text = codeEl?.textContent ?? "";
            await navigator.clipboard.writeText(text);
          }}
        />,
      );
    }
  }, [disabled]);
}
