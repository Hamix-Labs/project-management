import { describe, it } from "vitest";

describe("create flow invariants", () => {
  it("documents I1–I8 enforcement locations", () => {
    // I1 — useTaskCreateDraftAutosave skips when editingTaskId set
    // I2 — useTaskCreateMutations saveDraft onSuccess + useTasksApp stale-save tests
    // I3 — useTaskCreateMutations create onSuccess draft_id ref check
    // I4 — resumeDraftByID requestedResumeRef + useTasksApp resume race tests
    // I6 — validateCreateForm.test / useTaskCreateFlow.test default project
    // I7 — decideCreateEntry.test + useTaskCreateEntryActions openCreateModal
    // I8 — useTaskCreateDraftAutosave: debounced effect keeps the dirty gate,
    //      saveDraftNow / saveDraftNowAsync drop it. Pinned by
    //      draftAutosave.test.tsx (explicit save when clean) and
    //      draftAutosave.test.tsx (untouched fresh draft writes nothing).
  });
});
