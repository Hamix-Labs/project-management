import { describe, expect, it } from "vitest";
import { draftPayloadFingerprint } from "../task-drafts";
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

  // The regression guard. Autosave compares fingerprints, so any persisted field
  // the fingerprint does not cover is a field the operator cannot save: the form
  // reads clean and both save paths return early. That is how repository_id was
  // lost. Deriving the fingerprint from the payload is what makes this hold, and
  // this test is what fails if someone hand-writes the field list again.
  it("fingerprints every field it persists", () => {
    const payload = buildDraftSavePayload({
      ...baseFields,
      newRepositoryID: "repo-1",
      newWorktreeID: "wt-1",
      newChecklistItems: [
        { text: "One", verify_commands: [{ command: "go test ./...", expected_outcome: "pass" }] },
      ],
    });
    const fingerprinted = JSON.parse(draftPayloadFingerprint(payload));

    expect(Object.keys(fingerprinted)).toEqual(Object.keys(payload));
    expect(Object.keys(fingerprinted.payload)).toEqual(Object.keys(payload.payload));
  });

  it("flips the fingerprint for every persisted payload field", () => {
    const baseline = computeDraftAutosaveSignature(baseFields);
    const edits: Array<Partial<TaskCreateFormFields>> = [
      { newDraftID: "other-draft" },
      { newTitle: "Other title" },
      { newPrompt: "Other prompt" },
      { newPriority: "high" },
      { newTaskRunner: "codex" },
      { newTaskCursorModel: "gpt-5" },
      { newProjectID: "other-project" },
      { newRepositoryID: "repo-2" },
      { newWorktreeID: "wt-2" },
      { newChecklistItems: [{ text: "Two" }] },
    ];
    for (const edit of edits) {
      expect(
        computeDraftAutosaveSignature({ ...baseFields, ...edit }),
        `editing ${Object.keys(edit)[0]} must dirty the draft`,
      ).not.toBe(baseline);
    }
  });
});
