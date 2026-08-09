# ADR-0098: Prompt editor add-block resolves a slot instead of always inserting

**Status:** Superseded by [ADR-0099](./ADR-0099-prompt-authoring-inline-tiptap.md)

## Context

Clicking the side menu's add-block (`+`) button next to a paragraph, dismissing
the slash menu without choosing anything, and clicking `+` again stacked a
second empty paragraph. Every further click added another (issue #153).

BlockNote's `AddBlockButton` runs `editor.insertBlocks([{ type: "paragraph" }],
block, "after")` unconditionally whenever the hovered block is non-empty. It
models the click as "create a block" and never looks at what already follows,
so the action is not idempotent. Two consequences beyond the visual noise:

- Each stray click mutates the document, which fires `onChange` and autosaves an
  empty paragraph into the stored `initial_prompt` HTML.
- Users read `+` as "give me somewhere to write below this", so the second click
  reads as a bug rather than as a second insert.

The same component binds `onClick` to the plus glyph rather than to the button.
The button is therefore not keyboard-operable at all, and the expanded hit area
in `app-prompt-editor-blocknote.css` (sized for WCAG 2.5.8) is inert for pointer
input outside the 18px icon.

## Decision

The `+` **resolves a writing slot**; insertion is one outcome of that
resolution, not the action itself.

1. **The decision is a pure function.** `decideAddBlockSlot(hovered, next)`
   returns `focusHovered`, `focusNext`, `insertAfter`, or `undefined`. It takes
   structural block snapshots, so it is testable without mounting BlockNote —
   which matters because BlockNote cannot be exercised in jsdom here and every
   editor test in the repo mocks it wholesale. Logic left in the component would
   in practice be untested.
2. **`next` is the hovered block's next sibling** (`editor.getNextBlock`), which
   is exactly where `insertBlocks(..., "after")` lands. Reuse is therefore the
   precise inverse of the insert it replaces.
3. **Only a childless empty paragraph is reusable.** The stock insert produces a
   paragraph, so reusing an empty heading or list item would hand the user a
   different block type than the one they asked for. An empty paragraph with
   children is a container for the blocks nested under it, not a free slot.
4. **`PromptEditorAddBlockButton` is a thin adapter** that applies the decision
   and opens the slash menu, rendered by `PromptEditorSideMenu` alongside
   BlockNote's own `DragHandleButton`.
5. **`onClick` moves to the button**, restoring keyboard activation and the
   stylesheet's hit area.

**Invariant this establishes:** side-menu affordances are idempotent — repeating
one with no intervening user input must not change the document. New side-menu
buttons are expected to hold this.

## Consequences

- Repeat clicks focus the existing empty block. At most one empty paragraph is
  ever created, and the reuse path touches only the selection, so it produces no
  change event and no autosave.
- The `+` is reachable by keyboard and its full hit area is live.
- We now own one more upstream component. **On every BlockNote upgrade**,
  re-verify `useComponentsContext`, `useDictionary`, `useExtension`,
  `useExtensionState`, and `editor.getNextBlock`, and diff upstream's
  `AddBlockButton` for behaviour worth adopting. The chrome itself is still
  BlockNote's: only the click policy is ours.
- The plus glyph is now an inline stroked SVG matching the rest of the prompt
  editor chrome, because `react-icons` is a BlockNote dev dependency and not one
  of ours.
- Documents saved before this change still contain stacked empty paragraphs.
  They are deliberately left alone — rewriting a user's document on open would
  mark it dirty without consent.

## Alternatives considered

- **Sweep trailing empty paragraphs on save or on hydrate** — treats the symptom,
  silently mutates user documents, and still leaves the second click inserting;
  rejected.
- **Always reuse any empty block below, whatever its type** — surprising when an
  empty heading sits below, since the user asked for a fresh block; rejected.
- **Patch `@blocknote/react` via patch-package** — fixes every editor but adds a
  vendored patch to carry across upgrades, for a policy that is product-specific
  anyway; rejected.

Extends [ADR-0097](./ADR-0097-prompt-editor-side-menu-anchoring.md).
