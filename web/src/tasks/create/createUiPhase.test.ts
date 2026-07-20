import { describe, expect, it } from "vitest";
import {
  deriveCreateUiFlags,
  initialCreateUiPhase,
  reduceCreateUiPhase,
} from "./createUiPhase";

describe("createUiPhase", () => {
  it("starts closed", () => {
    expect(initialCreateUiPhase()).toEqual({ kind: "closed" });
    expect(deriveCreateUiFlags(initialCreateUiPhase())).toMatchObject({
      createModalOpen: false,
      draftPickerOpen: false,
      repositorySetupPromptOpen: false,
      editingTaskId: null,
    });
  });

  it("opens compose for task edit and derives flags", () => {
    const phase = reduceCreateUiPhase(initialCreateUiPhase(), {
      type: "openCompose",
      target: "task",
      operation: "edit",
      editingTaskId: "t1",
    });
    expect(phase).toEqual({
      kind: "compose",
      target: "task",
      operation: "edit",
      editingTaskId: "t1",
      editingTemplateId: null,
    });
    expect(deriveCreateUiFlags(phase)).toMatchObject({
      createModalOpen: true,
      draftPickerOpen: false,
      editingTaskId: "t1",
      composeOperation: "edit",
    });
  });

  it("transitions picker → closed → repo setup", () => {
    const picker = reduceCreateUiPhase(initialCreateUiPhase(), {
      type: "showDraftPicker",
    });
    expect(deriveCreateUiFlags(picker).draftPickerOpen).toBe(true);
    const closed = reduceCreateUiPhase(picker, { type: "close" });
    expect(closed.kind).toBe("closed");
    const repo = reduceCreateUiPhase(closed, { type: "showRepositorySetup" });
    expect(deriveCreateUiFlags(repo).repositorySetupPromptOpen).toBe(true);
  });
});
