import { useEffect, useRef, type RefObject } from "react";
import {
  decidePromptCanvasClick,
  isChildlessEmptyParagraph,
  resolvePromptCanvasClickRegion,
  type PromptCanvasClickBlock,
} from "./decidePromptCanvasClick";
import {
  applyPromptCanvasClick,
  type PromptCanvasClickEditor,
} from "./applyPromptCanvasClick";

const INTERACTIVE_SELECTOR = [
  "a[href]",
  "button",
  "input",
  "textarea",
  "select",
  "summary",
  '[role="button"]',
  '[role="menuitem"]',
  '[role="option"]',
  '[role="combobox"]',
  ".bn-side-menu",
].join(",");

const HEADER_CHROME_SELECTOR = [
  ".prompt-editor-doc-header",
  ".prompt-editor-session-alert",
  ".prompt-editor-topbar",
].join(",");

const MOVE_THRESHOLD_PX = 4;

function isInteractiveTarget(target: EventTarget | null): boolean {
  return target instanceof Element && Boolean(target.closest(INTERACTIVE_SELECTOR));
}

function isSelectingText(): boolean {
  const selection = window.getSelection();
  return Boolean(selection && !selection.isCollapsed);
}

function blockDom(
  editorHost: HTMLElement,
  blockId: string | undefined,
): Element | null {
  if (!blockId) {
    return null;
  }
  return editorHost.querySelector(
    `[data-node-type="blockContainer"][data-id="${CSS.escape(blockId)}"]`,
  );
}

type UsePromptCanvasClickFocusArgs = {
  canvasRef: RefObject<HTMLElement | null>;
  editorHost: HTMLElement | null;
  editor: PromptCanvasClickEditor & {
    document: Array<PromptCanvasClickBlock & { id?: string }>;
  };
  disabled?: boolean;
};

/**
 * Notion-style click-to-focus on the prompt canvas padding and header→body gap.
 *
 * Listens on the canvas element (siblings include the doc header). Decision
 * policy lives in {@link decidePromptCanvasClick}; this hook only measures,
 * classifies, and applies.
 */
export function usePromptCanvasClickFocus({
  canvasRef,
  editorHost,
  editor,
  disabled = false,
}: UsePromptCanvasClickFocusArgs): void {
  const pointerRef = useRef<{ x: number; y: number } | null>(null);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas || disabled || !editorHost) {
      return;
    }

    const onPointerDown = (event: PointerEvent) => {
      if (event.button !== 0) {
        return;
      }
      pointerRef.current = { x: event.clientX, y: event.clientY };
    };

    const onClick = (event: MouseEvent) => {
      if (event.button !== 0) {
        return;
      }

      const origin = pointerRef.current;
      pointerRef.current = null;
      if (origin) {
        const dx = event.clientX - origin.x;
        const dy = event.clientY - origin.y;
        if (dx * dx + dy * dy > MOVE_THRESHOLD_PX * MOVE_THRESHOLD_PX) {
          return;
        }
      }

      if (isSelectingText()) {
        return;
      }

      const target = event.target;
      if (!(target instanceof Node) || !canvas.contains(target)) {
        return;
      }

      const blocks = editor.document;
      const first = blocks[0] as (PromptCanvasClickBlock & { id?: string }) | undefined;
      const last = blocks[blocks.length - 1] as
        | (PromptCanvasClickBlock & { id?: string })
        | undefined;
      const firstEl = blockDom(editorHost, first?.id);
      const lastEl = blockDom(editorHost, last?.id);
      const meta = canvas.querySelector(".prompt-editor-doc-header__meta");

      const region = resolvePromptCanvasClickRegion({
        clientY: event.clientY,
        headerBottom: meta?.getBoundingClientRect().bottom ?? 0,
        firstBlockTop: firstEl?.getBoundingClientRect().top ?? null,
        lastBlockBottom: lastEl?.getBoundingClientRect().bottom ?? null,
        targetInHeaderChrome: Boolean(
          target instanceof Element && target.closest(HEADER_CHROME_SELECTOR),
        ),
        targetInBlock: Boolean(
          target instanceof Element &&
            target.closest('[data-node-type="blockContainer"]'),
        ),
      });

      const decision = decidePromptCanvasClick({
        region,
        hitInteractive: isInteractiveTarget(target),
        isSelectingText: false,
        hasBlocks: blocks.length > 0,
        lastBlockIsEmptyParagraph: isChildlessEmptyParagraph(last),
      });

      if (decision === "ignore") {
        return;
      }

      event.preventDefault();
      applyPromptCanvasClick(editor, decision);
    };

    canvas.addEventListener("pointerdown", onPointerDown);
    canvas.addEventListener("click", onClick);
    return () => {
      canvas.removeEventListener("pointerdown", onPointerDown);
      canvas.removeEventListener("click", onClick);
    };
  }, [canvasRef, disabled, editor, editorHost]);
}
