import { describe, expect, it } from "vitest";
import { FACTORY_REPO_DEFAULT_PROJECT_ID } from "@/test/factories/project";
import { CREATE_CHECKLIST_REQUIRED_MSG } from "../task-compose/checklistRequirement";
import { validateCreateFormChecklist } from "./validateCreateForm";

describe("validateCreateFormChecklist", () => {
  it("returns null when title or priority missing", () => {
    expect(validateCreateFormChecklist("", "medium", [])).toBeNull();
    expect(validateCreateFormChecklist("Title", "", [])).toBeNull();
  });

  it("requires at least one checklist row when title and priority set", () => {
    expect(validateCreateFormChecklist("Title", "medium", [])).toBe(
      CREATE_CHECKLIST_REQUIRED_MSG,
    );
    expect(validateCreateFormChecklist("Title", "medium", [{ text: "Ship" }])).toBeNull();
  });
});

describe("buildCreateTaskMutationInput", () => {
  it("includes default project id when unchanged", async () => {
    const { buildCreateTaskMutationInput } = await import("./buildCreateMutationInput");
    const input = buildCreateTaskMutationInput({
      newTitle: "Fresh task",
      newPrompt: "",
      newPriority: "medium",
      newTaskRunner: "cursor",
      newTaskCursorModel: "",
      newProjectID: FACTORY_REPO_DEFAULT_PROJECT_ID,
      newRepositoryID: "",
      newProjectContextItemIDs: [],
      newWorktreeID: "wt-1",
      newSchedule: null,
      newAutonomyEnabled: true,
      newTagsCsv: "",
      newMilestone: "",
      newDependsOn: [],
      newChecklistItems: [{ text: "Criterion" }],
      newDraftID: "draft-1",
    });
    expect(input.project_id).toBe(FACTORY_REPO_DEFAULT_PROJECT_ID);
  });
});
