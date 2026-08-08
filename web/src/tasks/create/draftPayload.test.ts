import { describe, expect, it } from "vitest";
import {
  buildDraftSavePayload,
  computeDraftAutosaveSignature,
} from "./draftPayload";
import type { TaskCreateFormFields } from "./types";

const baseFields: TaskCreateFormFields = {
  newTitle: "Title",
  newPrompt: "Prompt",
  newPriority: "medium",
  newTaskRunner: "cursor",
  newTaskCursorModel: "",
  newProjectID: "default",
  newRepositoryID: "",
  newSchedule: null,
  newAutonomyEnabled: true,
  newTagsCsv: "",
  newMilestone: "",
  newDependsOn: [],
  newWorktreeID: "",
  newChecklistItems: [{ text: "One" }],
  newDraftID: "draft-a",
  newFunctionInputs: [],
};

describe("computeDraftAutosaveSignature", () => {
  it("changes when draft id changes", () => {
    const a = computeDraftAutosaveSignature(baseFields);
    const b = computeDraftAutosaveSignature({ ...baseFields, newDraftID: "draft-b" });
    expect(a).not.toBe(b);
  });

  it("is stable for identical fields", () => {
    const a = computeDraftAutosaveSignature(baseFields);
    const b = computeDraftAutosaveSignature({ ...baseFields });
    expect(a).toBe(b);
  });

  it("changes when repository changes", () => {
    const a = computeDraftAutosaveSignature(baseFields);
    const b = computeDraftAutosaveSignature({
      ...baseFields,
      newRepositoryID: "repo-1",
    });
    expect(a).not.toBe(b);
  });
});

describe("buildDraftSavePayload", () => {
  it("includes repository_id and worktree_id", () => {
    const payload = buildDraftSavePayload({
      ...baseFields,
      newRepositoryID: "repo-1",
      newWorktreeID: "wt-1",
    });
    expect(payload.payload.repository_id).toBe("repo-1");
    expect(payload.payload.worktree_id).toBe("wt-1");
    expect(payload.payload.project_id).toBe("default");
  });
});
