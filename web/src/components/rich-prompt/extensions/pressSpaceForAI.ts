import { Extension } from "@tiptap/core";
import type { EditorState } from "@tiptap/pm/state";

export type PressSpaceForAIOptions = {
  /** Called when Space is pressed while the caret is at the start of an empty block. */
  onAiTrigger?: (msg: string) => void;
};

/**
 * True when the selection is a collapsed caret sitting at the start (offset 0)
 * of an otherwise-empty text block (paragraph / heading / list item). Callers
 * gate Space-for-AI and the `/` slash menu on this condition — mid-line Space
 * must fall through as a regular space and stray `/` must type as `/`.
 */
export function isCaretAtEmptyBlockStart(state: EditorState): boolean {
  const { selection } = state;
  if (!selection.empty) return false;
  const { $from } = selection;
  if ($from.parentOffset !== 0) return false;
  const parent = $from.parent;
  if (!parent.isTextblock) return false;
  return parent.content.size === 0;
}

export const PressSpaceForAI = Extension.create<PressSpaceForAIOptions>({
  name: "pressSpaceForAI",

  addOptions() {
    return {
      onAiTrigger: undefined,
    };
  },

  addKeyboardShortcuts() {
    return {
      Space: () => {
        const { state } = this.editor;
        if (!isCaretAtEmptyBlockStart(state)) return false;
        this.options.onAiTrigger?.("");
        return true;
      },
    };
  },
});
