import type { PromptBlockTypeTarget } from "./promptBlockTypeTargets";
import {
  decideBlockTypeConversion,
  type ConversionBlock,
} from "./decideBlockTypeConversion";

type IdBlock = ConversionBlock & { id: string };

/**
 * Minimal editor surface the conversion adapter needs. Kept structural (not
 * BlockNote-generic) so callers with a custom schema cast cleanly.
 */
export type BlockTypeConversionEditor = {
  getSelection: () => { blocks: IdBlock[] } | undefined;
  transact: (fn: () => void) => void;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any -- schema-agnostic update payload
  updateBlock: (block: { id: string }, update: any) => void;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any -- schema-agnostic
  insertBlocks: (...args: any[]) => any;
};

function blocksToConvert(
  editor: BlockTypeConversionEditor,
  anchor: IdBlock,
): IdBlock[] {
  const selected = editor.getSelection()?.blocks;
  if (selected && selected.some((block) => block.id === anchor.id)) {
    return selected;
  }
  return [anchor];
}

/**
 * Applies {@link decideBlockTypeConversion} to the anchor block, or to every
 * selected block when the selection includes the anchor — matching
 * BlockNote's `BlockTypeSelect` / `RemoveBlockItem` multi-block behaviour.
 *
 * Opens a single `transact` so one undo reverts the whole conversion. Skips
 * the transaction entirely when every block is already the target (idempotent
 * side-menu invariant from ADR-0098).
 */
export function applyBlockTypeConversion(
  // eslint-disable-next-line @typescript-eslint/no-explicit-any -- accepts prompt BlockNoteEditor
  editor: any,
  target: PromptBlockTypeTarget,
  anchor: IdBlock,
): void {
  const typed = editor as BlockTypeConversionEditor;
  const blocks = blocksToConvert(typed, anchor);
  const planned = blocks.map((block) => ({
    block,
    decision: decideBlockTypeConversion(block, target),
  }));

  const conversions = planned.filter(
    (entry) => entry.decision.action === "convert",
  );
  if (conversions.length === 0) {
    return;
  }

  typed.transact(() => {
    for (const { block, decision } of conversions) {
      if (decision.action !== "convert") {
        continue;
      }
      const children =
        decision.liftChildren && Array.isArray(block.children)
          ? block.children
          : undefined;

      typed.updateBlock(block, {
        type: decision.update.type,
        props: decision.update.props,
        ...(decision.update.content !== undefined
          ? { content: decision.update.content }
          : {}),
        ...(decision.liftChildren ? { children: [] } : {}),
      });

      if (children && children.length > 0) {
        typed.insertBlocks(children, block, "after");
      }
    }
  });
}
