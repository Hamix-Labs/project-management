import { useCallback } from "react";
import { SideMenuExtension, SuggestionMenu } from "@blocknote/core/extensions";
import {
  useBlockNoteEditor,
  useComponentsContext,
  useDictionary,
  useExtension,
  useExtensionState,
} from "@blocknote/react";
import {
  decideAddBlockSlot,
  type AddBlockSlotBlock,
} from "./decideAddBlockSlot";

/** Stroked to match the prompt editor's chrome; the stylesheet clamps it to 18px. */
function PlusIcon() {
  return (
    <svg
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      aria-hidden="true"
      data-test="dragHandleAdd"
    >
      <path d="M12 5v14M5 12h14" />
    </svg>
  );
}

/**
 * Prompt IDE replacement for BlockNote's `AddBlockButton`.
 *
 * Two divergences from the stock button, both behavioural:
 *
 * 1. Insertion is decided by {@link decideAddBlockSlot} rather than run
 *    unconditionally, so repeat clicks reuse the empty block below instead of
 *    stacking new ones (and stop autosaving a paragraph per stray click).
 * 2. `onClick` sits on the button rather than the plus glyph. Upstream binds it
 *    to the icon, which leaves the button unreachable by keyboard and makes the
 *    expanded hit area in the stylesheet inert.
 */
export function PromptEditorAddBlockButton() {
  const Components = useComponentsContext()!;
  const dict = useDictionary();
  const editor = useBlockNoteEditor();
  const suggestionMenu = useExtension(SuggestionMenu);
  const block = useExtensionState(SideMenuExtension, {
    editor,
    selector: (state) => state?.block,
  });

  const onClick = useCallback(() => {
    if (block === undefined) {
      return;
    }

    const next = editor.getNextBlock(block);
    const slot = decideAddBlockSlot(
      block as AddBlockSlotBlock,
      next as AddBlockSlotBlock | undefined,
    );

    if (slot === undefined) {
      return;
    }

    if (slot === "insertAfter") {
      editor.setTextCursorPosition(
        editor.insertBlocks([{ type: "paragraph" }], block, "after")[0],
      );
    } else {
      // Reuse leaves the document untouched, so no change event and no autosave.
      editor.setTextCursorPosition(slot === "focusNext" ? next! : block);
    }

    suggestionMenu.openSuggestionMenu("/");
  }, [block, editor, suggestionMenu]);

  if (block === undefined) {
    return null;
  }

  return (
    <Components.SideMenu.Button
      className={"bn-button"}
      label={dict.side_menu.add_block_label}
      onClick={onClick}
      icon={<PlusIcon />}
    />
  );
}
