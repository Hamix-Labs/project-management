/**
 * Canonical "Turn into" / block-type catalog for the Prompt IDE.
 *
 * Drawn from the schema's convertible inline/plain types and the slash-menu
 * keys we treat as conversions (not inserts). Excludes contentless blocks
 * (`repoFileEmbed`, media, table, divider) — converting those would destroy
 * props or structure. Heading levels stop at 3 to match the primary slash
 * entries; H1/H3 still render with BlockNote defaults until #159 styles them.
 */

export type PromptBlockContentKind = "inline" | "plain";

export type PromptBlockTypeTarget = {
  /** Stable id shared by the side menu, toolbar, and slash-menu adapters. */
  key: string;
  type: string;
  props?: Record<string, boolean | number | string>;
  contentKind: PromptBlockContentKind;
  /** Human label — BlockNote dictionary titles are applied in the UI layer when available. */
  label: string;
};

/** Default language applied when converting *into* a code block. Matches `promptCodeBlockOptions`. */
export const PROMPT_CODE_BLOCK_DEFAULT_LANGUAGE = "javascript";

export const PROMPT_BLOCK_TYPE_TARGETS: readonly PromptBlockTypeTarget[] = [
  {
    key: "paragraph",
    type: "paragraph",
    contentKind: "inline",
    label: "Paragraph",
  },
  {
    key: "heading",
    type: "heading",
    props: { level: 1, isToggleable: false },
    contentKind: "inline",
    label: "Heading 1",
  },
  {
    key: "heading_2",
    type: "heading",
    props: { level: 2, isToggleable: false },
    contentKind: "inline",
    label: "Heading 2",
  },
  {
    key: "heading_3",
    type: "heading",
    props: { level: 3, isToggleable: false },
    contentKind: "inline",
    label: "Heading 3",
  },
  {
    key: "quote",
    type: "quote",
    contentKind: "inline",
    label: "Quote",
  },
  {
    key: "bullet_list",
    type: "bulletListItem",
    contentKind: "inline",
    label: "Bullet List",
  },
  {
    key: "numbered_list",
    type: "numberedListItem",
    contentKind: "inline",
    label: "Numbered List",
  },
  {
    key: "check_list",
    type: "checkListItem",
    contentKind: "inline",
    label: "Check List",
  },
  {
    key: "toggle_list",
    type: "toggleListItem",
    contentKind: "inline",
    label: "Toggle List",
  },
  {
    key: "code_block",
    type: "codeBlock",
    props: { language: PROMPT_CODE_BLOCK_DEFAULT_LANGUAGE },
    contentKind: "plain",
    label: "Code Block",
  },
] as const;

/** Slash-menu keys that insert a new structure rather than convert a block. */
export const PROMPT_SLASH_INSERT_ONLY_KEYS = new Set([
  "table",
  "image",
  "video",
  "audio",
  "file",
  "divider",
  "emoji",
]);

/** Block types that must never be a conversion source or target. */
export const PROMPT_NON_CONVERTIBLE_BLOCK_TYPES = new Set([
  "repoFileEmbed",
  "image",
  "video",
  "audio",
  "file",
  "table",
  "divider",
]);

export function promptBlockTypeTargetByKey(
  key: string,
): PromptBlockTypeTarget | undefined {
  return PROMPT_BLOCK_TYPE_TARGETS.find((target) => target.key === key);
}
