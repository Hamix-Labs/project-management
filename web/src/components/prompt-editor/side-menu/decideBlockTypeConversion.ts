import {
  PROMPT_BLOCK_TYPE_TARGETS,
  PROMPT_CODE_BLOCK_DEFAULT_LANGUAGE,
  PROMPT_NON_CONVERTIBLE_BLOCK_TYPES,
  type PromptBlockContentKind,
  type PromptBlockTypeTarget,
} from "./promptBlockTypeTargets";

/** Structural subset of a BlockNote block — schema-agnostic for unit tests. */
export type ConversionBlock = {
  type: string;
  props?: Record<string, unknown>;
  content?: unknown;
  children?: unknown[];
};

export type ConversionUpdate = {
  type: string;
  props?: Record<string, boolean | number | string>;
  /** Set only when crossing `inline` ↔ `plain`; omitted so marks survive same-kind converts. */
  content?: Array<{ type: "text"; text: string; styles: Record<string, never> }>;
};

export type BlockTypeConversionDecision =
  | { action: "noop" }
  | { action: "unavailable" }
  | {
      action: "convert";
      update: ConversionUpdate;
      /**
       * When true, the adapter must move `block.children` to siblings after the
       * updated block. All catalog targets keep children on the container today;
       * the flag exists so a future non-nesting target cannot silently drop them.
       */
      liftChildren: boolean;
    };

function contentKindForType(type: string): PromptBlockContentKind | undefined {
  if (type === "codeBlock") {
    return "plain";
  }
  if (PROMPT_NON_CONVERTIBLE_BLOCK_TYPES.has(type)) {
    return undefined;
  }
  // Default BlockNote text blocks (paragraph, heading, lists, quote, …).
  return "inline";
}

/** True when the block can participate in Turn into / in-place slash conversion. */
export function isConvertibleBlock(block: ConversionBlock | undefined): boolean {
  if (block === undefined) {
    return false;
  }
  if (PROMPT_NON_CONVERTIBLE_BLOCK_TYPES.has(block.type)) {
    return false;
  }
  // Contentless blocks expose `content: undefined`; only array content converts.
  return Array.isArray(block.content);
}

/**
 * Whether `block` already is `target`. Code-block `language` is ignored so
 * re-picking "Code Block" on a Python block is a no-op (language is owned by
 * the code toolbar, not by Turn into).
 */
export function targetMatchesBlock(
  block: ConversionBlock,
  target: PromptBlockTypeTarget,
): boolean {
  if (block.type !== target.type) {
    return false;
  }
  if (target.type === "codeBlock") {
    return true;
  }
  for (const [key, value] of Object.entries(target.props ?? {})) {
    if (block.props?.[key] !== value) {
      return false;
    }
  }
  return true;
}

/** Targets offered for a given source — empty when the source cannot convert. */
export function offeredBlockTypeTargets(
  block: ConversionBlock | undefined,
): readonly PromptBlockTypeTarget[] {
  if (!isConvertibleBlock(block)) {
    return [];
  }
  return PROMPT_BLOCK_TYPE_TARGETS;
}

function inlineNodeText(node: unknown): string {
  if (node == null) {
    return "";
  }
  if (typeof node === "string") {
    return node;
  }
  if (typeof node !== "object") {
    return "";
  }
  const record = node as Record<string, unknown>;
  if (typeof record.text === "string") {
    return record.text;
  }
  if (record.type === "link" && Array.isArray(record.content)) {
    return record.content.map(inlineNodeText).join("");
  }
  if (record.type === "repoFileMention") {
    const props = (record.props ?? {}) as Record<string, unknown>;
    return typeof props.path === "string" ? props.path : "";
  }
  if (Array.isArray(record.content)) {
    return record.content.map(inlineNodeText).join("");
  }
  return "";
}

/** Flatten inline/plain content to a single unstyled text run (code-block payload). */
export function flattenBlockContentToPlainText(content: unknown): string {
  if (!Array.isArray(content)) {
    return "";
  }
  return content.map(inlineNodeText).join("");
}

function plainContentFromText(text: string): ConversionUpdate["content"] {
  if (text.length === 0) {
    return [];
  }
  return [{ type: "text", text, styles: {} }];
}

function propsForTarget(
  target: PromptBlockTypeTarget,
): Record<string, boolean | number | string> | undefined {
  if (target.type === "codeBlock") {
    return { language: PROMPT_CODE_BLOCK_DEFAULT_LANGUAGE };
  }
  if (target.props === undefined) {
    return undefined;
  }
  return { ...target.props };
}

/**
 * Decides how to turn `source` into `target`.
 *
 * Pure and BlockNote-free so the policy is unit-tested without mounting the
 * editor (same pattern as {@link decideAddBlockSlot}). The adapter applies the
 * returned update inside `editor.transact`.
 *
 * Contract:
 * - `noop` — already the target (including codeBlock regardless of language);
 *   the adapter must not open a transaction.
 * - `unavailable` — source cannot convert (e.g. `repoFileEmbed`).
 * - `convert` — `update` is passed to `editor.updateBlock`; `content` is set
 *   only when crossing `inline` ↔ `plain` so same-kind converts keep marks,
 *   links, and mention chips. Children stay on the block unless `liftChildren`.
 */
export function decideBlockTypeConversion(
  source: ConversionBlock | undefined,
  target: PromptBlockTypeTarget,
): BlockTypeConversionDecision {
  if (!isConvertibleBlock(source)) {
    return { action: "unavailable" };
  }

  // source is defined after the guard above.
  const block = source as ConversionBlock;

  if (targetMatchesBlock(block, target)) {
    return { action: "noop" };
  }

  const fromKind = contentKindForType(block.type);
  const toKind = target.contentKind;
  if (fromKind === undefined) {
    return { action: "unavailable" };
  }

  const update: ConversionUpdate = {
    type: target.type,
    props: propsForTarget(target),
  };

  if (fromKind !== toKind) {
    // Crossing plain/inline replaces content; without an explicit payload
    // BlockNote clears the node (`inline*` vs `text*`).
    update.content = plainContentFromText(
      flattenBlockContentToPlainText(block.content),
    );
  }

  // Every catalog target is a blockContainer that keeps nested children.
  // Lift only if we ever add a non-nesting target — never drop them.
  const liftChildren = false;

  return { action: "convert", update, liftChildren };
}
