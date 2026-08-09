# ADR-0097: Prompt editor owns its BlockNote side-menu anchoring

## Context

The prompt IDE side menu (add-block and drag-handle buttons) stayed painted at a
block's old position after the block was dragged elsewhere (issue #138). Two
causes, both structural rather than cosmetic.

**Anchoring.** BlockNote's `SideMenuController` renders a `BlockPopover`, which
memoises `prosemirrorView.domAtPos(...)` keyed on the hovered block id, and
`GenericPopover` freezes the last measured rect once that node detaches. A block
keeps its id when it moves, so nothing recomputes the node and FloatingUI keeps
positioning against the detached original. `SideMenuView.updateStateFromMousePos`
compounds it: it returns early while the hovered block id is unchanged, so no
extension state update fires either. Any doc mutation that shifts the hovered
block hits this — drag and drop, undo, a delete above it — not just drag.

**Visibility.** Our stylesheet hid `.bn-side-menu` with `opacity: 0` and lifted
it via `.bn-block-outer:hover > .bn-side-menu`. Since BlockNote 0.52 the side
menu is portaled into `bn-container`, so that selector can never match, leaving
visibility to `.bn-side-menu:hover`. Browsers freeze `:hover` for the duration of
an HTML5 drag, which is why the stale buttons stayed visible mid-drag.

An earlier attempt (PR #146) forced a fresh `domAtPos` by changing
`SideMenuController`'s React `key` on every `dragover` animation frame. That
destroyed and recreated the portal roughly per frame, restarting both
FloatingUI's entry transition and the CSS opacity transition — visible colour
flicker — and tore down the drag-handle DOM while it was the active drag source.
BlockNote 0.53.0 contains no upstream fix, so upgrading was not an option.

## Decision

The prompt editor owns its side menu, built on BlockNote's public API
(`GenericPopover`, `SideMenu`, `useExtensionState`, `useEditorDOMElement`,
`useEditorChange`), with `sideMenu={false}` on `BlockNoteView`.

1. **The anchor is derived from block identity at measurement time.**
   `promptSideMenuAnchorRect(editorDom, blockId)` queries
   `[data-node-type="blockContainer"][data-id=...]` on every
   `getBoundingClientRect()` call and is passed as a virtual FloatingUI
   reference with `cacheMountedBoundingClientRect: false`. Position correctness
   therefore never depends on a React re-render or a remount. The stable
   `.bn-block-group` remains the `contextElement` so FloatingUI's auto-update
   observers have a connected scroll ancestor.
2. **Repositioning replaces remounting.** The component captures FloatingUI's
   `update` via `whileElementsMounted` and calls it from `useEditorChange`.
3. **Visibility is state, not CSS hover.** `SideMenuExtension.show` plus a block
   id drive `open`; `usePromptBlockDragState` adds a `prompt-side-menu--dragging`
   class while a drag is in flight. The menu stays **mounted** during the drag —
   it is only visually suppressed — so the drag handle survives as the drag
   source and still emits `dragend`.
4. **Drag completion is read from bubble-phase `drop` and `dragend`**, after
   ProseMirror's capture-phase handlers have applied the transaction, so the
   block is already at its new position when the menu is re-measured.

Upstream's "hide the side menu on ancestor scroll" behaviour is intentionally
dropped: our portal scrolls with the page, so repositioning is the better result.

## Consequences

- Issue #138 is fixed for every cause of block movement, not only drag and drop.
- No flicker: nothing unmounts during a drag, and the stylesheet is the single
  owner of the fade (FloatingUI's transition duration is set to `0`).
- We depend on five additional `@blocknote/react` exports. **On every BlockNote
  upgrade**, re-verify that `GenericPopover`, `SideMenu`, `useExtensionState`,
  `useEditorDOMElement`, and `useEditorChange` are still exported with
  compatible signatures, and that block DOM still exposes
  `data-node-type="blockContainer"` with `data-id`.
- If upstream fixes `BlockPopover` to resolve its reference lazily, delete
  `PromptEditorSideMenu` and return to `SideMenuController`.
- The dead `.bn-block-outer:hover` rule is gone, so the buttons now appear as
  soon as a block is hovered rather than only when the cursor is on them.

## Alternatives considered

- **Remount the controller to force re-measurement** (PR #146) — fixes position
  but causes the flicker and breaks the drag source; rejected.
- **Patch `@blocknote/react` locally** via patch-package — fixes all editors but
  adds a vendored patch to carry across upgrades; rejected for now.
- **Hide the side menu for the whole drag by unmounting it** — simplest, but
  removing the drag source mid-drag is not safe, and it leaves the underlying
  staleness for non-drag block moves; rejected.

Extends [ADR-0096](./ADR-0096-prompt-editor-blocknote.md).
