import { describe, expect, it } from "vitest";
import type { AppSettings } from "@/api/settings";
import type { TaskDraftDetail, TaskDraftPayload } from "@/types";
import { computeDraftAutosaveSignature } from "./draftPayload";
import { freshDraftFields, resumedDraftFields } from "./resumedDraftFields";

const settings = { runner: "codex", cursor_model: "gpt-5" } as AppSettings;

function draft(payload: Partial<TaskDraftPayload>): TaskDraftDetail {
  return {
    id: "draft-1",
    name: "Untitled draft",
    created_at: "2026-08-01T00:00:00Z",
    updated_at: "2026-08-01T00:00:00Z",
    payload: {
      title: "Title",
      initial_prompt: "<p>Body</p>",
      priority: "medium",
      checklist_items: [],
      ...payload,
    },
  };
}

describe("resumedDraftFields", () => {
  it("restores the git and project bindings the draft was saved with", () => {
    const fields = resumedDraftFields(
      draft({ project_id: "project-1", repository_id: "repo-1", worktree_id: "wt-1" }),
      settings,
    );
    expect(fields.newProjectID).toBe("project-1");
    expect(fields.newRepositoryID).toBe("repo-1");
    expect(fields.newWorktreeID).toBe("wt-1");
  });

  it("leaves bindings empty for legacy drafts that omit them", () => {
    const fields = resumedDraftFields(draft({}), settings);
    expect(fields.newProjectID).toBe("");
    expect(fields.newRepositoryID).toBe("");
    expect(fields.newWorktreeID).toBe("");
  });

  it("falls back to configured defaults when runner and model are absent", () => {
    const fields = resumedDraftFields(draft({}), settings);
    expect(fields.newTaskRunner).toBe("codex");
    expect(fields.newTaskCursorModel).toBe("gpt-5");
  });

  it("keeps an explicitly empty cursor model instead of re-defaulting it", () => {
    const fields = resumedDraftFields(draft({ cursor_model: "" }), settings);
    expect(fields.newTaskCursorModel).toBe("");
  });

  it("prefers the draft's own runner over the default", () => {
    const fields = resumedDraftFields(draft({ runner: " cursor " }), settings);
    expect(fields.newTaskRunner).toBe("cursor");
  });

  it("maps checklist items and keeps their verify commands", () => {
    const fields = resumedDraftFields(
      draft({
        checklist_items: [
          { text: "One", verify_commands: [{ command: "go test ./...", expected_outcome: "pass" }] },
          { text: "Two" },
        ],
      }),
      settings,
    );
    expect(fields.newChecklistItems).toEqual([
      { text: "One", verify_commands: [{ command: "go test ./...", expected_outcome: "pass" }] },
      { text: "Two" },
    ]);
  });

  // The whole point of the mapper: form state and baseline come from one value,
  // so an untouched resumed draft cannot look dirty and autosave immediately.
  it("produces a baseline that matches its own restored fields", () => {
    const resumed = draft({
      project_id: "project-1",
      repository_id: "repo-1",
      worktree_id: "wt-1",
      checklist_items: [{ text: "One" }],
    });
    const fields = resumedDraftFields(resumed, settings);
    expect(computeDraftAutosaveSignature(fields)).toBe(
      computeDraftAutosaveSignature(resumedDraftFields(resumed, settings)),
    );
  });
});

describe("freshDraftFields", () => {
  it("starts unbound with the operator's configured defaults", () => {
    const fields = freshDraftFields(settings, "generated-1");
    expect(fields).toEqual({
      newDraftID: "generated-1",
      newTitle: "",
      newPrompt: "",
      newPriority: "",
      newTaskRunner: "codex",
      newTaskCursorModel: "gpt-5",
      newProjectID: "",
      newRepositoryID: "",
      newWorktreeID: "",
      newChecklistItems: [],
    });
  });
});
