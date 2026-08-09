import type { BlockNoteEditor } from "@blocknote/core";
import {
  filterSuggestionItems,
  insertOrUpdateBlockForSlashMenu,
} from "@blocknote/core/extensions";
import {
  getDefaultReactSlashMenuItems,
  type DefaultReactSuggestionItem,
} from "@blocknote/react";
import { applyBlockTypeConversion } from "./applyBlockTypeConversion";
import type { ConversionBlock } from "./decideBlockTypeConversion";
import {
  PROMPT_SLASH_INSERT_ONLY_KEYS,
  promptBlockTypeTargetByKey,
  type PromptBlockTypeTarget,
} from "./promptBlockTypeTargets";

/**
 * `DefaultReactSuggestionItem` omits `key`, but `getDefaultReactSlashMenuItems`
 * spreads the core item so `key` is present at runtime.
 */
type SlashMenuItem = DefaultReactSuggestionItem & { key: string };

/**
 * Extra slash keys that convert a block but are not in the Turn into catalog
 * (toggle headings, H4–H6). Still convert in place on non-empty blocks.
 */
export function slashKeyToBlockTypeTarget(
  key: string,
): PromptBlockTypeTarget | undefined {
  const catalog = promptBlockTypeTargetByKey(key);
  if (catalog) {
    return catalog;
  }

  const toggle = /^toggle_heading(?:_([2-3]))?$/.exec(key);
  if (toggle) {
    const level = toggle[1] ? Number(toggle[1]) : 1;
    return {
      key,
      type: "heading",
      props: { level, isToggleable: true },
      contentKind: "inline",
      label: key,
    };
  }

  const heading = /^heading_([4-6])$/.exec(key);
  if (heading) {
    const level = Number(heading[1]);
    return {
      key,
      type: "heading",
      props: { level, isToggleable: false },
      contentKind: "inline",
      label: key,
    };
  }

  return undefined;
}

function convertFromSlashKey(
  // eslint-disable-next-line @typescript-eslint/no-explicit-any -- prompt schema
  editor: BlockNoteEditor<any, any, any>,
  key: string,
  fallback: () => void,
): void {
  if (PROMPT_SLASH_INSERT_ONLY_KEYS.has(key)) {
    fallback();
    return;
  }

  const target = slashKeyToBlockTypeTarget(key);
  if (!target) {
    fallback();
    return;
  }

  const block = editor.getTextCursorPosition().block as ConversionBlock & {
    id: string;
  };
  if (!Array.isArray(block.content)) {
    fallback();
    return;
  }

  applyBlockTypeConversion(editor, target, block);
  editor.setTextCursorPosition(block, "end");
}

/**
 * Slash-menu items that convert in place whenever the current block has array
 * content. Empty blocks still convert (after the menu clears `/`); non-empty
 * blocks no longer insert a sibling via `insertOrUpdateBlock`.
 *
 * Insert-only keys (table, media, divider, emoji) keep BlockNote's default.
 */
export function getPromptSlashMenuItems(
  // eslint-disable-next-line @typescript-eslint/no-explicit-any -- prompt schema
  editor: BlockNoteEditor<any, any, any>,
): DefaultReactSuggestionItem[] {
  return (getDefaultReactSlashMenuItems(editor) as SlashMenuItem[]).map(
    (item) => {
      if (PROMPT_SLASH_INSERT_ONLY_KEYS.has(item.key)) {
        return item;
      }

      const target = slashKeyToBlockTypeTarget(item.key);
      if (!target) {
        return item;
      }

      return {
        ...item,
        onItemClick: () => {
          convertFromSlashKey(editor, item.key, () => {
            insertOrUpdateBlockForSlashMenu(editor, {
              type: target.type,
              props: target.props,
            } as never);
          });
        },
      };
    },
  );
}

export function filterPromptSlashMenuItems(
  // eslint-disable-next-line @typescript-eslint/no-explicit-any -- prompt schema
  editor: BlockNoteEditor<any, any, any>,
  query: string,
): DefaultReactSuggestionItem[] {
  return filterSuggestionItems(getPromptSlashMenuItems(editor), query);
}
