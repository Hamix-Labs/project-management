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

/**
 * BlockNote's code block uses content:"plain", which createReactBlockSpec cannot
 * host. We keep createCodeBlockSpec (highlighter + shortcuts) and replace the
 * native <select> chrome with a Notion-like searchable toolbar.
 *
 * The toolbar is portaled onto the editor host (outside ProseMirror). Mutating
 * inside the code-block node view is wiped or ingested as document text.
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
      if (block instanceof HTMLElement) {
        delete block.dataset.promptCodeToolbar;
      }
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

    const enhanceBlock = (block: Element) => {
      if (!(block instanceof HTMLElement)) return;

      const select = block.querySelector(":scope > div select, :scope > div > select");
      if (!(select instanceof HTMLSelectElement)) return;

      const existing = mounts.get(block);
      if (existing && block.dataset.promptCodeToolbar === "1") {
        positionMount(container, block, existing.mountEl);
        return;
      }
      if (block.dataset.promptCodeToolbar === "1" && !existing) {
        container
          .querySelector(
            `.prompt-code-toolbar-root[data-for-block="${CSS.escape(block.dataset.promptCodeToolbarId ?? "")}"]`,
          )
          ?.remove();
        delete block.dataset.promptCodeToolbar;
        delete block.dataset.promptCodeToolbarId;
      }

      const toolbarId =
        block.dataset.promptCodeToolbarId ??
        `code-${Math.random().toString(36).slice(2, 10)}`;
      block.dataset.promptCodeToolbarId = toolbarId;
      block.dataset.promptCodeToolbar = "1";
      select.hidden = true;
      select.setAttribute("aria-hidden", "true");
      select.tabIndex = -1;

      const pre = block.querySelector("pre");
      const code = block.querySelector("code");
      pre?.setAttribute("spellcheck", "false");
      code?.setAttribute("spellcheck", "false");

      let mount = container.querySelector(
        `:scope > .prompt-code-toolbar-root[data-for-block="${CSS.escape(toolbarId)}"]`,
      );
      if (!(mount instanceof HTMLElement)) {
        mount = document.createElement("div");
        mount.className = "prompt-code-toolbar-root";
        mount.dataset.forBlock = toolbarId;
        container.appendChild(mount);
      }
      positionMount(container, block, mount);

      try {
        const root = resolveCreateRoot()(mount);
        const onSelectChange = () => {
          const record = mounts.get(block);
          if (record) renderToolbar(record);
        };
        select.addEventListener("change", onSelectChange);
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
        delete block.dataset.promptCodeToolbar;
        delete block.dataset.promptCodeToolbarId;
        mount.remove();
        select.hidden = false;
        select.removeAttribute("aria-hidden");
        select.tabIndex = 0;
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
          if (live.has(block) && block.isConnected) {
            positionMount(container, record.block, record.mountEl);
            continue;
          }
          teardownMount(block, record);
        }
      } finally {
        sweeping = false;
      }
    };

    const scheduleSweep = () => {
      if (debounceTimer != null) window.clearTimeout(debounceTimer);
      debounceTimer = window.setTimeout(() => {
        debounceTimer = null;
        sweep();
      }, 16);
    };

    sweep();
    for (const ms of [0, 50, 200, 500]) {
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

    const onScrollOrResize = () => scheduleSweep();
    container.addEventListener("scroll", onScrollOrResize, { passive: true });
    window.addEventListener("resize", onScrollOrResize);

    return () => {
      mo.disconnect();
      container.removeEventListener("scroll", onScrollOrResize);
      window.removeEventListener("resize", onScrollOrResize);
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
