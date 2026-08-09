# ADR-0100: Prompt editor owns the drag-handle menu open signal for block highlight

## Context

Clicking the six-dot drag handle opens BlockNote's Delete / Colors menu with no
visual cue for which block the menu will act on (issue #158). The menu is
portaled outside the block's DOM, so no CSS selector from the menu can reach the
block, and BlockNote never sets a ProseMirror `NodeSelection` or stamps a class
on the target.

[ADR-0097](./ADR-0097-prompt-editor-side-menu-anchoring.md) already gives the
prompt editor `{ show, blockId }` from the side-menu extension and a live
block-id → DOM lookup. That is not enough on its own: `show` is also true on
plain hover, and the `menuFrozen` flag that distinguishes menu-open from hover
lives on the plugin view and is not part of the emitted extension state.

Sibling issue #156 ("Turn into") must extend the same drag-handle menu. Both
issues need a prompt-owned menu seam; neither should fork BlockNote's
`DragHandleButton` if a thinner cut works.

## Decision

1. **Stamp `data-prompt-block-active` on live block containers**, resolved via
   `findPromptBlockElement` (shared with the side-menu anchor). CSS paints the
   fill. Do **not** introduce ProseMirror `NodeSelection` management — the repo
   has none today, and it would add focus/caret risk for no gain over an
   attribute the stylesheet already owns.
2. **Own the drag-handle menu as `PromptEditorDragHandleMenu`**, passed to
   stock `DragHandleButton` via `dragHandleMenu`. Default items stay Delete and
   Colors; #156 adds "Turn into" in this file without rewriting highlight wiring.
3. **Observe menu-open via a mount beacon inside the dropdown.** Ariakit's
   menu uses `unmountOnHide`, so a child effect that reports open on mount and
   close on unmount is the open signal. Report through
   `PromptDragHandleMenuOpenProvider` because `dragHandleMenu` is typed as
   `FC` with no props.
4. **Decide which ids to highlight in a pure function**
   (`decideActiveBlockIds`): menu-open **or** drag-in-flight, never hover
   alone; when the handle's block is in `editor.getSelection()?.blocks`,
   highlight every selected block (matching Delete).
5. **Re-stamp on editor change** while active, because ProseMirror replaces
   block DOM nodes and would otherwise drop the attribute.

## Consequences

- Hovering a block no longer looks like selection; only menu-open and drag do.
- #156 rebases by editing `PromptEditorDragHandleMenu` item list; recommended
  merge order is this PR first, then #156.
- On every BlockNote upgrade, re-verify `DragHandleButton`'s `dragHandleMenu`
  prop, Ariakit `unmountOnHide` on `Menu.Dropdown`, and
  `data-node-type="blockContainer"` / `data-id` on block DOM.
- Nested children are not covered by a parent's fill — only the block's own
  `.bn-block-content` row. Multi-select stamps each selected container.

## Alternatives considered

- **ProseMirror `NodeSelection` + `.ProseMirror-selectednode`** — activates
  BlockNote's hardcoded outline, but introduces selection/focus management the
  repo otherwise avoids; rejected.
- **Bind highlight to side-menu `show`** — lights every hovered block; rejected.
- **Fork `DragHandleButton` for `onOpenChange`** — works, but duplicates the
  grip button and is unnecessary once the menu dropdown's unmount is the signal;
  rejected for now.
- **CSS-only from the menu** — impossible; the menu is portaled out of the
  block tree.

Extends [ADR-0097](./ADR-0097-prompt-editor-side-menu-anchoring.md).
