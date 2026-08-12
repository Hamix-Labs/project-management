# ADR-0100: Task compose is a routed page

**Date:** 2026-08-12  
**Status:** Accepted  
**Deciders:** Product / frontend maintainers

## Context

Create, edit, and template compose have historically shared a single
52rem-wide modal (`TaskCreateModal` + `TaskCreateModalsLayer`) rendered
outside the routed outlet. That worked when the form was mostly title +
priority + a short prompt, but it now caps every future authoring surface:

- The rich-prompt editor (ADR-0099) is squeezed by the modal chrome — a
  ~9.5rem editor is not enough room for a real brief.
- Wave A of Task draft AI adds an on-page assist column that must sit
  beside the prompt while operators iterate. A modal cannot host that
  without turning into a full-screen dialog.
- Deep-linking a compose surface (`/tasks/new?project=…`, resume of a
  specific draft, template edit) is awkward from within a shell overlay:
  the URL is `/`, the "close" of the overlay is not a browser back, and
  refresh loses the form target.
- Create/edit and template compose are the same form. Routing them side
  by side keeps the entry points symmetric with the rest of the app.

## Decision

1. **Compose is a routed page**, not a modal. New routes registered
   *before* `tasks/:taskId` in [`App.tsx`](../../web/src/app/App.tsx):

   | Route | Purpose |
   | --- | --- |
   | `/tasks/new` | Create a task (accepts `?project=`, `?draft=`, git-assignment locks in query) |
   | `/tasks/:taskId/edit` | Edit an existing task |
   | `/templates/new` | Create a task template |
   | `/templates/:templateId/edit` | Edit an existing template |

2. **Page shell.** [`TaskComposePage`](../../web/src/tasks/pages/TaskComposePage.tsx)
   is thin: it reads route params, seeds `useTaskCreateFlow` on mount, and
   renders [`TaskComposeLayout`](../../web/src/tasks/components/task-compose/TaskComposeLayout.tsx)
   (back link + title + scrollable form body + sticky footer + reserved
   `task-compose-page__assist` slot for Plan 4). The form body itself is
   the existing `TaskCreateModalFormBody`; only the modal chrome is
   replaced.

3. **Modal shell deleted.** `TaskCreateModal`, `TaskCreateModalShell`,
   `TaskCreateModalHeader`, `TaskCreateModalActionFooter`, and
   `TaskCreateModalsLayer` are removed. The 52rem
   `.modal-shell--wide:has(.task-create-modal-sheet.task-create)` cap goes
   with them. Compose is full main-column width, capped at ~80rem
   (~90rem once the assist slot opens).

4. **Entry points navigate().** Every `openCreateModal`,
   `openTemplateCreateModal`, `openEdit`, `editTemplateByID`, resume-draft
   call becomes a `navigate(...)` to the appropriate compose route.
   `?create=1&project=…` on `/` redirects to `/tasks/new?project=…`.

5. **Draft picker + repo-setup move inline.** They previously lived in
   `TaskCreateModalsLayer` as their own overlays. Now they render as
   on-page phases of `/tasks/new` (repo-setup can still deep-link to
   `/repositories?register=1`). One overlay less to focus-trap.

6. **Compose lifecycle == route lifecycle.** The compose page owns the
   window in which autosave runs. Leaving the page (browser back, nav to
   another route, submit) unmounts the compose provider and behaves like
   the old modal `close`: abort in-flight draft save, keep server drafts.
   Invariants I1–I7 hold unchanged.

## Consequences

### Positive

- The prompt gets ~20rem of vertical room by default; the assist column
  has a home for Plan 4 without a second overlay.
- Compose surfaces are deep-linkable and refresh-safe.
- One less always-mounted overlay in `StandardShell`, one less
  focus-trap, one less lazy chunk gated on a boolean.

### Negative / accepted risks

- Integration helpers (`openNewTaskModal`, tests that look for
  `role="dialog"` on the compose form) require an update — the compose
  surface is now scoped to `role="form"` / the page landmark, not a
  dialog.
- Compose-in-progress on top of a task detail page is now a full route
  swap instead of an overlay. Operators editing from the task list lose
  the "close the modal and I'm back on the list" affordance; the back
  link on the compose page navigates back explicitly.

## Alternatives rejected

1. **Grow the modal to 80rem.** Cheap and preserves the current entry
   points, but a modal that fills the viewport is a page pretending to be
   a dialog — worse focus semantics, no deep link, still can't host the
   assist column comfortably.
2. **Nest a second overlay for assist.** Overlays stack poorly; adding
   another focus trap on top of the compose modal breaks keyboard flow.
3. **Split create and edit into different surfaces.** Rejected — the
   form is the same shape; two divergent implementations would drift.
