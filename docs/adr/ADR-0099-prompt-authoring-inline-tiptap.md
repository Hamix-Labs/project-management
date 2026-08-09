# ADR-0099: Prompt authoring returns to inline TipTap

**Date:** 2026-08-09  
**Status:** Accepted  
**Deciders:** Product / frontend maintainers

## Context

The full-page Prompt IDE ([ADR-0096](./ADR-0096-prompt-editor-blocknote.md)) moved create/edit/polish authoring onto `/prompt/:sourceKind/:sourceId` with BlockNote, compose suspend/resume, and `ImmersiveShell`. That surface grew chrome, side-menu, and embed complexity without enough product payoff, and it left compose/polish as CTAs into a separate route.

We are abandoning the Prompt IDE and restoring inline TipTap in the create/edit modal and polish dialog.

## Decision

1. **Inline TipTap** (`web/src/components/rich-prompt/`) is again the only SPA prompt authoring surface for compose and polish.
2. **Delete the full-page Prompt IDE**: `/prompt` route, `ImmersiveShell`, BlockNote packages, prompt-editor CSS tokens, and sessionStorage launch/suspend/resume.
3. **Wire format stays HTML** in `initial_prompt` / draft `payload.initial_prompt`. No DB migration.
4. **Accept imperfect round-trip** for prompts that only BlockNote could represent (custom embeds / nodes). TipTap loads HTML best-effort; operators may need to rewrite those briefs.
5. TipTap packages pin at **3.23.6** with `tippy.js` **^6.3.7**. All `@blocknote/*` and `@shikijs/*` dependencies are removed.

This ADR **supersedes** [ADR-0096](./ADR-0096-prompt-editor-blocknote.md), [ADR-0097](./ADR-0097-prompt-editor-side-menu-anchoring.md), and [ADR-0098](./ADR-0098-prompt-editor-add-block-slot.md).

## Consequences

### Positive

- Create/edit/polish keep an editable prompt field; no CTA that navigates away or 404s.
- Smaller SPA dependency graph and no immersive shell fork of app chrome.
- Resume existing TipTap mention/`@` repo-file chip behavior against the HTML wire format.

### Negative / accepted risks

- BlockNote-only HTML (e.g. file-embed blocks) may not edit cleanly in TipTap.
- Operators who bookmarked `/prompt/...` lose that route.

## Alternatives rejected

1. **Plain textarea** — loses structured Markdown, mentions, and formatting already validated by the API.
2. **BlockNote inside the create/polish modal** — keeps the heavy editor without the full-page product; worse cold-start than TipTap for the same wire format.
3. **Soft-deprecate the IDE** (leave route but hide CTAs) — orphans surface and docs; rejected in favor of a hard cut-over.
