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

  it("suspends compose for prompt editor without clearing edit identity", () => {
    const compose = reduceCreateUiPhase(initialCreateUiPhase(), {
      type: "openCompose",
      target: "task",
      operation: "create",
      editingTaskId: null,
    });
    const suspended = reduceCreateUiPhase(compose, {
      type: "suspendForPromptEditor",
      target: "task",
      operation: "create",
    });
    expect(suspended.kind).toBe("promptEditorSuspended");
    expect(deriveCreateUiFlags(suspended)).toMatchObject({
      createModalOpen: false,
      promptEditorSuspended: true,
      composeTarget: "task",
    });
    const resumed = reduceCreateUiPhase(suspended, {
      type: "resumeComposeFromPromptEditor",
    });
    expect(resumed.kind).toBe("compose");
    expect(deriveCreateUiFlags(resumed).createModalOpen).toBe(true);
  });
});
