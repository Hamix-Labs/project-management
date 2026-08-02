# ADR-0096: Full-page Prompt Editor (BlockNote)

## Context

Task create/edit and polish previously embedded TipTap inside modals. Long
implementation briefs felt like form fields, and the always-visible formatting
toolbar conflicted with a Notion-like authoring experience. `@` repo-file
mentions (path + optional line range) must keep working against the existing
HTML wire format validated by the API.

## Decision

1. **Full-page route** `/prompt/:sourceKind/:sourceId` is the only place users
   author prompts. Compose modals and polish show an **Open Prompt Editor** CTA
   plus a text preview.
2. **BlockNote** (`@blocknote/core` + `@blocknote/react` + `@blocknote/ariakit`)
   is the editor engine. TipTap is removed from the SPA.
3. **Wire format stays HTML** in `initial_prompt` / draft `payload.initial_prompt`.
   BlockNote document JSON is editor-internal only (no new DB column).
4. **Document adapters** load/save by source kind:
   - `draft` → `/task-drafts`
   - `task` → `GET`/`PATCH` `/tasks/{id}`
   - `template` → template get/patch (or ephemeral for `new`)
   - `ephemeral` → `sessionStorage` (polish instructions)
5. **Compose suspend**: leaving for the editor sets UI phase
   `promptEditorSuspended` without `resetNewTaskForm`, so metadata survives.
   Done writes HTML back and resumes compose.
6. **Repo file mentions** remain HTML chips
   `span[data-repo-file][data-path]…`, implemented as BlockNote custom inline
   content with `parse` / `toExternalHTML`.

## Consequences

- Prompt authoring is a first-class product surface; create modal stays metadata-focused.
- Refresh/share of create prompts continues to use drafts; polish uses ephemeral sessions.
- Future templates, AI assist, media, and versioning can hang off
  `PromptDocumentAdapter` / `PromptDocumentRef` without changing HTTP shapes.
- Detail sanitizer allowlists repo-file chip attributes on `span`.

## Alternatives considered

- Overlay editor above the modal — weaker refresh/share; rejected.
- BlockNote JSON as canonical storage — requires API/DB migration; rejected for v1.
- Keep TipTap + bubble menu — rejected; product chose BlockNote.
