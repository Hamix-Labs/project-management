import { useEffect, useRef } from "react";
import { createRoot, type Root } from "react-dom/client";
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
};

/**
 * BlockNote's code block uses content:"plain", which createReactBlockSpec cannot
 * host. We keep createCodeBlockSpec (highlighter + shortcuts) and replace the
 * native <select> chrome with a Notion-like searchable toolbar.
 *
 * Pass the host element from a callback ref / useState — a plain useRef can be
 * null on the first effect tick and leave the MutationObserver never attached.
 *
 * StrictMode re-runs must clear the enhance marker and detach the mount node
 * on cleanup, or the second pass skips remount while CSS hides the native select.
 */
export function useEnhanceCodeBlockToolbars(
  container: HTMLElement | null,
  disabled: boolean,
) {
  const mountsRef = useRef(new Map<Element, MountRecord>());
  const languagesRef = useRef<PromptCodeLanguage[]>(promptCodeLanguages());
  const disabledRef = useRef(disabled);
  disabledRef.current = disabled;

  useEffect(() => {
    if (!container) return;

    const mounts = mountsRef.current;
    let sweeping = false;
    let debounceTimer: number | null = null;
    let raf = 0;
    let lateTimer: number | null = null;

    const teardownMount = (wrap: Element, record: MountRecord) => {
      record.select.removeEventListener("change", record.onSelectChange);
      if (wrap instanceof HTMLElement) {
        delete wrap.dataset.promptCodeToolbar;
      }
      mounts.delete(wrap);
      // Detach immediately so a StrictMode remount cannot reuse this node
      // while the previous createRoot is still attached.
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

    const renderToolbar = (wrap: Element, record: MountRecord) => {
      const { root, select } = record;
      const codeEl = wrap.parentElement?.querySelector("code");
      root.render(
        <CodeLanguageToolbar
          languages={languagesRef.current}
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
      const wrap = block.querySelector(":scope > div");
      if (!(wrap instanceof HTMLElement)) return;

      const select = wrap.querySelector("select");
      if (!(select instanceof HTMLSelectElement)) return;

      const existing = mounts.get(wrap);
      if (existing && wrap.dataset.promptCodeToolbar === "1") {
        return;
      }
      // Stale mark from a prior effect pass (StrictMode) or detached root.
      if (wrap.dataset.promptCodeToolbar === "1" && !existing) {
        wrap.querySelector(".prompt-code-toolbar-root")?.remove();
        delete wrap.dataset.promptCodeToolbar;
      }

      wrap.dataset.promptCodeToolbar = "1";
      select.hidden = true;
      select.setAttribute("aria-hidden", "true");
      select.tabIndex = -1;

      const pre = block.querySelector("pre");
      const code = block.querySelector("code");
      pre?.setAttribute("spellcheck", "false");
      code?.setAttribute("spellcheck", "false");

      let mount = wrap.querySelector(".prompt-code-toolbar-root");
      if (!(mount instanceof HTMLElement)) {
        mount = document.createElement("div");
        mount.className = "prompt-code-toolbar-root";
        wrap.appendChild(mount);
      }

      try {
        const root = createRoot(mount);
        const onSelectChange = () => {
          const record = mounts.get(wrap);
          if (record) renderToolbar(wrap, record);
        };
        select.addEventListener("change", onSelectChange);
        const record: MountRecord = {
          root,
          select,
          onSelectChange,
          mountEl: mount,
        };
        mounts.set(wrap, record);
        renderToolbar(wrap, record);
      } catch (err) {
        delete wrap.dataset.promptCodeToolbar;
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
            const wrap = block.querySelector(":scope > div");
            if (wrap) live.add(wrap);
          });

        for (const [wrap, record] of mounts) {
          if (live.has(wrap) && wrap.isConnected) continue;
          teardownMount(wrap, record);
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
    // BlockNote may paint ProseMirror after the first effect tick.
    raf = window.requestAnimationFrame(() => {
      sweep();
      lateTimer = window.setTimeout(sweep, 50);
    });

    const mo = new MutationObserver((mutations) => {
      for (const m of mutations) {
        const t = m.target;
        if (t instanceof Element && t.closest(".prompt-code-toolbar-root")) {
          continue;
        }
        if (
          m.type === "childList" &&
          [...m.addedNodes, ...m.removedNodes].every(
            (n) =>
              n instanceof Element &&
              (n.classList?.contains("prompt-code-toolbar-root") ||
                n.closest?.(".prompt-code-toolbar-root")),
          )
        ) {
          continue;
        }
        scheduleSweep();
        return;
      }
    });
    mo.observe(container, { childList: true, subtree: true });

    return () => {
      mo.disconnect();
      window.cancelAnimationFrame(raf);
      if (debounceTimer != null) window.clearTimeout(debounceTimer);
      if (lateTimer != null) window.clearTimeout(lateTimer);
      for (const [wrap, record] of [...mounts.entries()]) {
        teardownMount(wrap, record);
      }
      mounts.clear();
    };
  }, [container]);

  useEffect(() => {
    for (const [wrap, record] of mountsRef.current) {
      const codeEl = wrap.parentElement?.querySelector("code");
      record.root.render(
        <CodeLanguageToolbar
          languages={languagesRef.current}
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
