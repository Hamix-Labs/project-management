# ADR-0099: Prompt editor canvas click-to-focus

## Context

The Prompt IDE document canvas had large vertical gaps (header→first block, and
`padding-bottom` below the last block) that did nothing on click. The caret could
only be placed by clicking the title input or an existing block, which made the
page feel broken rather than merely roomy (issue #154).

A naive "click empty space → always insert a paragraph" policy would reintroduce
the #153 failure mode: stray clicks mutate the document, fire `onChange`, and
autosave empty paragraphs into stored `initial_prompt` HTML.

## Decision

1. **The decision is a pure function.** `decidePromptCanvasClick` (plus
   `resolvePromptCanvasClickRegion` for geometry) returns `ignore` /
   `focusFirst` / `focusLast` / `appendAndFocus` from classified facts. It is
   React-free and BlockNote-free so the policy is unit-tested without mounting
   the editor — the same pattern as ADR-0098's `decideAddBlockSlot`.
2. **`usePromptCanvasClickFocus` is a thin adapter** on the document canvas
   element. It measures block rects, classifies the region, and calls
   `applyPromptCanvasClick`. It does not invent policy.
3. **Zones.** Clicks in the gap between the meta divider and the first block
   focus the first block. Clicks below the last block's bottom (including canvas
   `padding-bottom`) focus the end of the document.
4. **Append is conditional and idempotent.** An empty paragraph is inserted only
   when the last top-level block is *not* already a childless empty paragraph.
   When it is, the click only moves the selection — no change event, no autosave.
   Emptiness matches the add-block reusable-slot definition (ADR-0098).
5. **Do not steal real editing gestures.** Header chrome (title, meta, alerts),
   in-block targets, interactive controls (links, buttons, side menu, inputs),
   and click-drags that select text are ignored.

**Invariant this establishes:** empty-canvas clicks never dirty an already-empty
trailing writing slot. Together with ADR-0098, prompt-editor pointer affordances
that "give me somewhere to type" must be idempotent with respect to autosave.

## Consequences

- Header spacing can stay intentionally roomy without feeling dead.
- Repeat clicks below the document do not stack empty paragraphs.
- We own more pointer behaviour outside BlockNote's default contenteditable hit
  testing. **On every BlockNote upgrade**, re-verify that top-level blocks still
  expose `[data-node-type="blockContainer"][data-id=…]` and that
  `setTextCursorPosition` / `insertBlocks` signatures used by
  `applyPromptCanvasClick` remain compatible.

## Alternatives considered

- **Only tighten CSS spacing** — still leaves the bottom padding and remaining
  header gap unclickable; rejected as incomplete for #154.
- **Always append on below-last clicks** — dirties autosave on stray clicks;
  rejected (#153 class).
- **Put the policy in the React click handler** — untestable here because
  BlockNote is mocked wholesale in component tests; rejected in favour of the
  pure-function split.

Extends [ADR-0096](./ADR-0096-prompt-editor-blocknote.md) and
[ADR-0098](./ADR-0098-prompt-editor-add-block-slot.md).
