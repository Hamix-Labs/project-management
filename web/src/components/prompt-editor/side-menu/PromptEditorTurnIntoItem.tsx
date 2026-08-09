import { SideMenuExtension } from "@blocknote/core/extensions";
import {
  useBlockNoteEditor,
  useComponentsContext,
  useExtensionState,
} from "@blocknote/react";
import { applyBlockTypeConversion } from "./applyBlockTypeConversion";
import {
  offeredBlockTypeTargets,
  targetMatchesBlock,
  type ConversionBlock,
} from "./decideBlockTypeConversion";
import type { PromptBlockTypeTarget } from "./promptBlockTypeTargets";

/**
 * Drag-handle submenu: Turn into → supported block types.
 *
 * Uses BlockNote's Generic.Menu submenu primitives (same pattern as
 * `BlockColorsItem`) so Ariakit owns keyboard navigation — arrows, Enter,
 * Escape, and focus return to the drag handle.
 */
export function PromptEditorTurnIntoItem() {
  const Components = useComponentsContext()!;
  const editor = useBlockNoteEditor();
  const block = useExtensionState(SideMenuExtension, {
    editor,
    selector: (state) => state?.block,
  }) as (ConversionBlock & { id: string }) | undefined;

  const targets = offeredBlockTypeTargets(block);
  if (block === undefined || targets.length === 0) {
    return null;
  }

  const onSelect = (target: PromptBlockTypeTarget) => {
    applyBlockTypeConversion(editor, target, block);
  };

  return (
    <Components.Generic.Menu.Root position={"right"} sub={true}>
      <Components.Generic.Menu.Trigger sub={true}>
        <Components.Generic.Menu.Item
          className={"bn-menu-item"}
          subTrigger={true}
        >
          Turn into
        </Components.Generic.Menu.Item>
      </Components.Generic.Menu.Trigger>
      <Components.Generic.Menu.Dropdown
        sub={true}
        className={"bn-menu-dropdown bn-drag-handle-menu prompt-turn-into-menu"}
      >
        {targets.map((target) => (
          <Components.Generic.Menu.Item
            key={target.key}
            className={"bn-menu-item"}
            checked={targetMatchesBlock(block, target)}
            onClick={() => onSelect(target)}
          >
            {target.label}
          </Components.Generic.Menu.Item>
        ))}
      </Components.Generic.Menu.Dropdown>
    </Components.Generic.Menu.Root>
  );
}
