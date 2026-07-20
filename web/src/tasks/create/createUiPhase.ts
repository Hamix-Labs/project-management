/**
 * Explicit create/edit modal UI phase (F-06-03).
 * Shell booleans are derived; transitions go through `reduceCreateUiPhase`.
 */

import type { ComposeOperation, ComposeTarget } from "./types";

export type { ComposeOperation, ComposeTarget };

export type CreateUiPhase =
  | { kind: "closed" }
  | { kind: "draftPicker" }
  | { kind: "repositorySetup" }
  | {
      kind: "compose";
      target: ComposeTarget;
      operation: ComposeOperation;
      editingTaskId: string | null;
      editingTemplateId: string | null;
    };

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
    };

export function initialCreateUiPhase(): CreateUiPhase {
  return { kind: "closed" };
}

export function reduceCreateUiPhase(
  _phase: CreateUiPhase,
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
    default: {
      const _exhaustive: never = event;
      return _exhaustive;
    }
  }
}

export function deriveCreateUiFlags(phase: CreateUiPhase) {
  return {
    createModalOpen: phase.kind === "compose",
    draftPickerOpen: phase.kind === "draftPicker",
    repositorySetupPromptOpen: phase.kind === "repositorySetup",
    editingTaskId: phase.kind === "compose" ? phase.editingTaskId : null,
    editingTemplateId: phase.kind === "compose" ? phase.editingTemplateId : null,
    composeTarget: phase.kind === "compose" ? phase.target : ("task" as const),
    composeOperation:
      phase.kind === "compose" ? phase.operation : ("create" as const),
  };
}
