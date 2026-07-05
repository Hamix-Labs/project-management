import type { Status } from "@/types";
import type { TestScenario } from "@/tasks/test-scenarios";
import { isUiFeatureOmitted } from "@/launch/omittedFeatures";

export type TaskCreateModalPresentation = {
  isTaskEdit: boolean;
  isTemplateMode: boolean;
  isEdit: boolean;
  disabled: boolean;
  tagsAndDependenciesUiEnabled: boolean;
  scheduleUiEnabled: boolean;
  modalBusy: boolean;
  modalTitle: string;
  modalTitleId: string;
  modalDescribedBy: string | undefined;
  idsPrefix: string;
  status: Status;
  showTestScenarios: boolean;
  showDraftStatus: boolean;
};

type TaskCreateModalMode =
  | "task-edit"
  | "template-edit"
  | "template-create"
  | "task-create";

type PresentationPreset = {
  modalTitle: string;
  modalTitleId: string;
  idsPrefix: string;
  modalDescribedBy: string | undefined;
  modalBusy: (input: ResolvePresentationInput) => boolean;
};

type ResolvePresentationInput = {
  editingTaskId: string | null;
  composeTarget: "task" | "template";
  composeOperation: "create" | "edit";
  composeStatus: Status | undefined;
  pending: boolean;
  saving: boolean;
  patchPending: boolean;
  draftSaveLabel: string | null;
  onApplyTestScenario?: (scenario: TestScenario) => void;
};

const PRESENTATION_PRESETS: Record<TaskCreateModalMode, PresentationPreset> = {
  "task-edit": {
    modalTitle: "Edit task",
    modalTitleId: "task-edit-modal-title",
    idsPrefix: "task-edit",
    modalDescribedBy: "task-edit-modal-description",
    modalBusy: (input) => input.patchPending,
  },
  "template-edit": {
    modalTitle: "Edit template",
    modalTitleId: "task-template-edit-modal-title",
    idsPrefix: "task-template-edit",
    modalDescribedBy: undefined,
    modalBusy: (input) => input.pending || input.saving,
  },
  "template-create": {
    modalTitle: "New template",
    modalTitleId: "task-template-create-modal-title",
    idsPrefix: "task-template-new",
    modalDescribedBy: undefined,
    modalBusy: (input) => input.pending || input.saving,
  },
  "task-create": {
    modalTitle: "New task",
    modalTitleId: "task-create-modal-title",
    idsPrefix: "task-new",
    modalDescribedBy: undefined,
    modalBusy: (input) => input.pending,
  },
};

function resolveTaskCreateModalMode(input: ResolvePresentationInput): TaskCreateModalMode {
  const isTaskEdit = input.editingTaskId != null;
  if (isTaskEdit) {
    return "task-edit";
  }
  if (input.composeTarget === "template") {
    return input.composeOperation === "edit" ? "template-edit" : "template-create";
  }
  return "task-create";
}

export function resolveTaskCreateModalPresentation(
  input: ResolvePresentationInput,
): TaskCreateModalPresentation {
  const isTaskEdit = input.editingTaskId != null;
  const isTemplateMode = input.composeTarget === "template";
  const isEdit = isTaskEdit || (isTemplateMode && input.composeOperation === "edit");
  const mode = resolveTaskCreateModalMode(input);
  const preset = PRESENTATION_PRESETS[mode];
  const disabled = input.pending || input.saving;
  const tagsAndDependenciesUiEnabled = !isUiFeatureOmitted("tagsAndDependencies");
  const scheduleUiEnabled = !isUiFeatureOmitted("schedule");
  const status = input.composeStatus ?? "ready";
  const showTestScenarios = !isEdit && Boolean(input.onApplyTestScenario);
  const showDraftStatus = !isEdit && !isTemplateMode && Boolean(input.draftSaveLabel);

  return {
    isTaskEdit,
    isTemplateMode,
    isEdit,
    disabled,
    tagsAndDependenciesUiEnabled,
    scheduleUiEnabled,
    modalBusy: preset.modalBusy(input),
    modalTitle: preset.modalTitle,
    modalTitleId: preset.modalTitleId,
    modalDescribedBy: preset.modalDescribedBy,
    idsPrefix: preset.idsPrefix,
    status,
    showTestScenarios,
    showDraftStatus,
  };
}
