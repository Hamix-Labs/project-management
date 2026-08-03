import { useEffect, useRef, type RefObject } from "react";
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
};

/**
 * BlockNote's code block uses content:"plain", which createReactBlockSpec cannot
 * host. We keep createCodeBlockSpec (highlighter + shortcuts) and replace the
 * native <select> chrome with a Notion-like searchable toolbar.
 *
 * This is a finite DOM bridge: mark each block once, ignore mutations under our
 * toolbar roots, and unmount only when the block leaves the document.
 * Longer-term exit: a first-class React code block when BlockNote can host
 * plain-content blocks (or an upstream extension API).
 */
export function useEnhanceCodeBlockToolbars(
  containerRef: RefObject<HTMLElement | null>,
  disabled: boolean,
) {
  const mountsRef = useRef(new Map<Element, MountRecord>());
  const languagesRef = useRef<PromptCodeLanguage[]>(promptCodeLanguages());
  const disabledRef = useRef(disabled);
  disabledRef.current = disabled;

  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    const mounts = mountsRef.current;
    let sweeping = false;
    let debounceTimer: number | null = null;

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
      if (wrap.dataset.promptCodeToolbar === "1") {
        return;
      }

      const select = wrap.querySelector("select");
      if (!(select instanceof HTMLSelectElement)) return;

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

      const root = createRoot(mount);
      const onSelectChange = () => {
        const record = mounts.get(wrap);
        if (record) renderToolbar(wrap, record);
      };
      select.addEventListener("change", onSelectChange);
      const record: MountRecord = { root, select, onSelectChange };
      mounts.set(wrap, record);
      renderToolbar(wrap, record);
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
          if (live.has(wrap)) continue;
          if (!wrap.isConnected) {
            record.select.removeEventListener("change", record.onSelectChange);
            record.root.unmount();
            mounts.delete(wrap);
          }
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
      if (debounceTimer != null) window.clearTimeout(debounceTimer);
      for (const [, record] of mounts) {
        record.select.removeEventListener("change", record.onSelectChange);
        record.root.unmount();
      }
      mounts.clear();
    };
  }, [containerRef]);

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
