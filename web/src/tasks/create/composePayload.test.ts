import { describe, expect, it } from "vitest";
import {
  buildComposePayloadFromForm,
  hydrateFormFromComposePayload,
  parseTagsFromCsv,
} from "./composePayload";
import type { TaskCreateFormFields } from "./types";

const baseFields: TaskCreateFormFields = {
  newTitle: "Ship feature",
  newPrompt: "<p>Do the thing</p>",
  newPriority: "medium",
  newTaskRunner: "cursor",
  newTaskCursorModel: "gpt-4",
  newTaskVerifyChatMode: "",
  newProjectID: "proj-1",
  newRepositoryID: "repo-1",
  newProjectContextItemIDs: ["ctx-1"],
  newSchedule: "2030-01-01T12:00:00Z",
  newAutonomyEnabled: true,
  newTagsCsv: "alpha, beta",
  newMilestone: "m1",
  newDependsOn: ["dep-1"],
  newWorktreeID: "wt-1",
  newChecklistItems: [{ text: "Criterion one" }],
  newDraftID: "draft-1",
};

describe("composePayload", () => {
  it("round-trips form fields through hydrate", () => {
    const payload = buildComposePayloadFromForm(baseFields);
    expect(payload.repository_id).toBe("repo-1");
    expect(payload.worktree_id).toBeUndefined();
    const hydrated = hydrateFormFromComposePayload(payload, undefined);
    expect(hydrated.title).toBe("Ship feature");
    expect(hydrated.prompt).toBe("<p>Do the thing</p>");
    expect(hydrated.priority).toBe("medium");
    expect(hydrated.runner).toBe("cursor");
    expect(hydrated.projectID).toBe("proj-1");
    expect(hydrated.repositoryID).toBe("repo-1");
    expect(hydrated.worktreeID).toBe("");
    expect(hydrated.schedule).toBe("2030-01-01T12:00:00Z");
    expect(hydrated.tagsCsv).toBe("alpha, beta");
    expect(hydrated.dependsOn).toEqual(["dep-1"]);
    expect(hydrated.checklistItems).toEqual([{ text: "Criterion one" }]);
  });

  it("lowercases and dedupes tags from Title Case CSV", () => {
    expect(parseTagsFromCsv("Backend, API, backend")).toEqual(["backend", "api"]);
    const payload = buildComposePayloadFromForm({
      ...baseFields,
      newTagsCsv: "Backend",
    });
    expect(payload.tags).toEqual(["backend"]);
  });
});
