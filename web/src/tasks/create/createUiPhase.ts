/**
 * Explicit create/edit modal UI phase (F-06-03).
 * Shell booleans are derived; transitions go through `reduceCreateUiPhase`.
 */

import type { ComposeOperation, ComposeTarget } from "./types";

export type { ComposeOperation, ComposeTarget };

export type ComposePhaseFields = {
  target: ComposeTarget;
  operation: ComposeOperation;
  editingTaskId: string | null;
  editingTemplateId: string | null;
};

export type CreateUiPhase =
  | { kind: "closed" }
  | { kind: "draftPicker" }
  | { kind: "repositorySetup" }
  | ({ kind: "compose" } & ComposePhaseFields)
  | ({ kind: "promptEditorSuspended" } & ComposePhaseFields);

export type CreateUiPhaseEvent =
  | { type: "close" }
  | { type: "showDraftPicker" }
  | { type: "showRepositorySetup" }
  | {
      type: "openCompose";
      target: ComposeTarget;
      operation: ComposeOperation;
      editingTaskId?: string | null;
      editingTemplateId?: string | null;
    }
  | {
      type: "suspendForPromptEditor";
      target: ComposeTarget;
      operation: ComposeOperation;
      editingTaskId?: string | null;
      editingTemplateId?: string | null;
    }
  | { type: "resumeComposeFromPromptEditor" };

export function initialCreateUiPhase(): CreateUiPhase {
  return { kind: "closed" };
}

export function reduceCreateUiPhase(
  phase: CreateUiPhase,
  event: CreateUiPhaseEvent,
): CreateUiPhase {
  switch (event.type) {
    case "close":
      return { kind: "closed" };
    case "showDraftPicker":
      return { kind: "draftPicker" };
    case "showRepositorySetup":
      return { kind: "repositorySetup" };
    case "openCompose":
      return {
        kind: "compose",
        target: event.target,
        operation: event.operation,
        editingTaskId: event.editingTaskId ?? null,
        editingTemplateId: event.editingTemplateId ?? null,
      };
    case "suspendForPromptEditor":
      return {
        kind: "promptEditorSuspended",
        target: event.target,
        operation: event.operation,
        editingTaskId: event.editingTaskId ?? null,
        editingTemplateId: event.editingTemplateId ?? null,
      };
    case "resumeComposeFromPromptEditor": {
      if (phase.kind !== "promptEditorSuspended") {
        return phase;
      }
      return {
        kind: "compose",
        target: phase.target,
        operation: phase.operation,
        editingTaskId: phase.editingTaskId,
        editingTemplateId: phase.editingTemplateId,
      };
    }
    default: {
      const _exhaustive: never = event;
      return _exhaustive;
    }
  }
}

function composeFieldsFromPhase(phase: CreateUiPhase): ComposePhaseFields {
  if (phase.kind === "compose" || phase.kind === "promptEditorSuspended") {
    return {
      target: phase.target,
      operation: phase.operation,
      editingTaskId: phase.editingTaskId,
      editingTemplateId: phase.editingTemplateId,
    };
  }
  return {
    target: "task",
    operation: "create",
    editingTaskId: null,
    editingTemplateId: null,
  };
}

export function deriveCreateUiFlags(phase: CreateUiPhase) {
  const fields = composeFieldsFromPhase(phase);
  return {
    createModalOpen: phase.kind === "compose",
    promptEditorSuspended: phase.kind === "promptEditorSuspended",
    draftPickerOpen: phase.kind === "draftPicker",
    repositorySetupPromptOpen: phase.kind === "repositorySetup",
    editingTaskId: fields.editingTaskId,
    editingTemplateId: fields.editingTemplateId,
    composeTarget: fields.target,
    composeOperation: fields.operation,
  };
}
