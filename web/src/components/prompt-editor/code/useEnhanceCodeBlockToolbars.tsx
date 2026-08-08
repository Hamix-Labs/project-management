import { useLayoutEffect, useEffect, useRef } from "react";
import { createBlockKeyFactory } from "./codeToolbarDom";
import {
  CodeToolbarMounts,
  type CodeBlockLanguageEditor,
} from "./codeToolbarMounts";
import {
  promptCodeLanguages,
  type PromptCodeLanguage,
} from "./promptCodeBlockOptions";

export type { CodeBlockLanguageEditor };

const CODE_BLOCK_SELECTOR = '[data-content-type="codeBlock"]';

/**
 * BlockNote's code block uses content:"plain", which createReactBlockSpec cannot
 * host. We keep createCodeBlockSpec (highlighter + shortcuts) and replace the
 * native <select> chrome with a Notion-like searchable toolbar.
 *
 * The toolbar is portaled onto the editor host, outside ProseMirror: mounting it
 * inside the node view made ProseMirror strip our markers and remount a second
 * toolbar on top of the native select on every sweep.
 */
export function useEnhanceCodeBlockToolbars(
  container: HTMLElement | null,
  disabled: boolean,
  editor?: CodeBlockLanguageEditor | null,
) {
  const languagesRef = useRef<PromptCodeLanguage[] | null>(null);
  if (languagesRef.current == null) {
    languagesRef.current = promptCodeLanguages();
  }
  const disabledRef = useRef(disabled);
  disabledRef.current = disabled;
  const mountsRef = useRef<CodeToolbarMounts | null>(null);

  useLayoutEffect(() => {
    if (!container) return;

    const mounts = new CodeToolbarMounts({
      host: container,
      editor,
      languages: () => languagesRef.current ?? [],
      isDisabled: () => disabledRef.current,
      keyOf: createBlockKeyFactory(),
    });
    mountsRef.current = mounts;

    let sweeping = false;
    let sweepTimer: number | null = null;
    let positionTimer: number | null = null;
    const pollTimers: number[] = [];

    const sweep = () => {
      if (sweeping) return;
      sweeping = true;
      try {
        const live = new Set<string>();
        container.querySelectorAll(CODE_BLOCK_SELECTOR).forEach((block) => {
          const key = mounts.ensure(block);
          if (key) live.add(key);
        });
        mounts.prune(live);
      } finally {
        sweeping = false;
      }
    };

    const scheduleSweep = () => {
      if (sweepTimer != null) window.clearTimeout(sweepTimer);
      sweepTimer = window.setTimeout(() => {
        sweepTimer = null;
        sweep();
      }, 32);
    };

    const reposition = () => mounts.repositionAll();
    const scheduleReposition = () => {
      if (positionTimer != null) return;
      positionTimer = window.setTimeout(() => {
        positionTimer = null;
        reposition();
      }, 32);
    };

    sweep();
    // BlockNote mounts its node views over a few frames after the editor appears.
    for (const ms of [0, 50, 200]) {
      pollTimers.push(window.setTimeout(sweep, ms));
    }

    const mo = new MutationObserver((mutations) => {
      let needsSweep = false;
      let needsReposition = false;
      for (const m of mutations) {
        const target = m.target;
        if (!(target instanceof Element)) {
          needsSweep = true;
          continue;
        }
        if (target.closest(".prompt-code-toolbar-root")) continue;
        // Highlighting rewrites spans inside <pre>, which can only shift blocks.
        if (target.closest("pre")) {
          needsReposition = true;
          continue;
        }
        needsSweep = true;
      }
      if (needsSweep) scheduleSweep();
      else if (needsReposition) scheduleReposition();
    });
    mo.observe(container, {
      childList: true,
      subtree: true,
      attributes: true,
      attributeFilter: ["data-content-type", "data-id"],
    });

    const unsubscribe = editor?.onChange(() => {
      mounts.renderAll();
      scheduleReposition();
    });

    container.addEventListener("scroll", reposition, { passive: true });
    window.addEventListener("resize", reposition);

    return () => {
      mo.disconnect();
      if (typeof unsubscribe === "function") unsubscribe();
      container.removeEventListener("scroll", reposition);
      window.removeEventListener("resize", reposition);
      if (sweepTimer != null) window.clearTimeout(sweepTimer);
      if (positionTimer != null) window.clearTimeout(positionTimer);
      for (const id of pollTimers) window.clearTimeout(id);
      mounts.clear();
      mountsRef.current = null;
    };
  }, [container, editor]);

  useEffect(() => {
    mountsRef.current?.renderAll();
  }, [disabled]);
}
