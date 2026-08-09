import { useMemo } from "react";
import {
  useBlockNoteEditor,
  useComponentsContext,
  useEditorState,
} from "@blocknote/react";
import { applyBlockTypeConversion } from "../side-menu/applyBlockTypeConversion";
import {
  targetMatchesBlock,
  type ConversionBlock,
} from "../side-menu/decideBlockTypeConversion";
import { PROMPT_BLOCK_TYPE_TARGETS } from "../side-menu/promptBlockTypeTargets";

type IdBlock = ConversionBlock & { id: string };

/**
 * Block-type dropdown for the selection toolbar.
 *
 * Replaces BlockNote's `BlockTypeSelect`, which returns `null` unless the
 * current block matches an item — so the control vanished inside code blocks.
 * This always renders while editable, includes `codeBlock`, and uses the same
 * conversion policy as the drag-handle Turn into menu.
 */
export function PromptEditorBlockTypeSelect() {
  const Components = useComponentsContext()!;
  const editor = useBlockNoteEditor();

  const selectedBlocks = useEditorState({
    editor,
    selector: ({ editor: ed }) =>
      (ed.getSelection()?.blocks as IdBlock[] | undefined) || [
        ed.getTextCursorPosition().block as IdBlock,
      ],
  });
  const first = selectedBlocks[0];

  const items = useMemo(() => {
    return PROMPT_BLOCK_TYPE_TARGETS.map((target) => {
      const selected =
        first !== undefined && targetMatchesBlock(first, target);
      return {
        text: target.label,
        icon: <BlockTypeIcon label={target.label} />,
        isSelected: selected,
        onClick: () => {
          editor.focus();
          if (first === undefined) {
            return;
          }
          applyBlockTypeConversion(editor, target, first);
        },
      };
    });
  }, [editor, first]);

  if (!editor.isEditable) {
    return null;
  }

  return (
    <Components.FormattingToolbar.Select
      className={"bn-select"}
      items={items}
    />
  );
}

/** Minimal text glyph so we do not depend on react-icons. */
function BlockTypeIcon({ label }: { label: string }) {
  const letter = label.charAt(0).toUpperCase();
  return (
    <svg
      width="16"
      height="16"
      viewBox="0 0 16 16"
      aria-hidden="true"
      focusable="false"
    >
      <text
        x="8"
        y="12"
        textAnchor="middle"
        fontSize="10"
        fontFamily="var(--font-sans, sans-serif)"
        fill="currentColor"
      >
        {letter}
      </text>
    </svg>
  );
}
