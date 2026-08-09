# ADR-0100: Prompt editor owns block-type conversion ("Turn into")

## Context

Users could not change an existing Prompt IDE block's type from the block
itself (issue #156). The workaround was insert-via-`+` / slash menu, copy
content, delete the old block. BlockNote 0.52 ships no drag-handle "Turn into"
item — `DragHandleMenu` defaults are Delete, Colors, and table-header toggles
only. Conversion *did* exist behind `BlockTypeSelect` on the selection toolbar,
but only after a mouse text selection, and that control returned `null` inside
code blocks because `codeBlock` was missing from `blockTypeSelectItems`. The
default slash menu uses `insertOrUpdateBlock`, which converts only empty blocks
and otherwise inserts below — exactly the reported behaviour.

Side-menu ownership already moved in-repo
([ADR-0097](./ADR-0097-prompt-editor-side-menu-anchoring.md),
[ADR-0098](./ADR-0098-prompt-editor-add-block-slot.md)). The drag-handle menu
seam and open-state beacon landed with the active-block highlight
([ADR-0099](./ADR-0099-prompt-editor-active-block-highlight.md) / PR #166):
`PromptEditorDragHandleMenu` is passed to stock `DragHandleButton` via
`dragHandleMenu`, and an Ariakit mount beacon reports open/close through
`PromptDragHandleMenuOpenProvider`. This ADR must extend that menu, not fork a
second one.

## Decision

1. **Conversion policy is a pure function.**
   `decideBlockTypeConversion(source, target)` returns `noop` /
   `unavailable` / `convert` with an `update` payload. It is React-free and
   BlockNote-free so it is unit-tested without mounting the editor. The thin
   adapter `applyBlockTypeConversion` runs matching updates inside one
   `editor.transact` (single undo) and skips the transaction when every block
   is already the target — honouring ADR-0098's idempotence invariant.
2. **One catalog, three entry points.** `PROMPT_BLOCK_TYPE_TARGETS` feeds the
   drag-handle Turn into submenu, the selection-toolbar select, and the
   in-place slash converters. `repoFileEmbed` and other contentless / structural
   types are never sources or targets.
3. **Content crossing `inline` ↔ `plain` is explicit.** BlockNote clears the
   node when `inline*` and `text*` differ unless `content` is passed. The
   policy flattens to unstyled text when entering or leaving a code block;
   same-kind converts omit `content` so marks, links, and mention chips survive.
4. **Extend `PromptEditorDragHandleMenu` only.** Turn into sits between Delete
   and Colors inside the #166 menu. The open beacon, context provider, and
   highlight hook are left alone — `PromptEditorSideMenu` already passes
   `dragHandleMenu={PromptEditorDragHandleMenu}`.
5. **Slash menu is prompt-owned.** `slashMenu={false}` plus a `/`
   `SuggestionMenuController` whose items call the same adapter for type keys;
   insert-only keys (table, media, divider, emoji) keep upstream behaviour.
6. **Toolbar select is prompt-owned.** `PromptEditorBlockTypeSelect` always
   renders while editable (including inside code blocks) and uses the catalog.

Markdown input rules (`## `, `- `, …) remain BlockNote's: they still only
apply while the block is empty, complementary to Turn into.

## Consequences

- Turn into / slash / toolbar agree on the convertible set.
- Converting to the current type does not dirty the document or autosave.
- Multi-block selections convert in one undo step.
- H1/H3 (and quote / check list) still use BlockNote default typography until
  heading-style follow-ups; this ADR does not add those CSS rules.
- **On every BlockNote upgrade**, re-verify `DragHandleButton` /
  `Generic.Menu` submenu keyboard behaviour, `updateBlock` plain/inline
  handling, and default slash item keys.

## Alternatives considered

- **Rely on selection-toolbar `BlockTypeSelect` only** — fails the discoverability
  report and still vanished in code blocks; rejected.
- **Fork a second drag-handle button/menu for open-state** — duplicates #166's
  beacon seam and fights the highlight PR; rejected.
- **Patch `@blocknote/react` via patch-package** — carries a vendored patch
  across upgrades for product-specific policy; rejected.

Extends ADR-0097, ADR-0098, and ADR-0099.
