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

export function resolveTaskCreateModalPresentation(input: {
  editingTaskId: string | null;
  composeTarget: "task" | "template";
  composeOperation: "create" | "edit";
  composeStatus: Status | undefined;
  pending: boolean;
  saving: boolean;
  patchPending: boolean;
  draftSaveLabel: string | null;
  onApplyTestScenario?: (scenario: TestScenario) => void;
}): TaskCreateModalPresentation {
  const isTaskEdit = input.editingTaskId != null;
  const isTemplateMode = input.composeTarget === "template";
  const isEdit = isTaskEdit || (isTemplateMode && input.composeOperation === "edit");
  const disabled = input.pending || input.saving;
  const tagsAndDependenciesUiEnabled = !isUiFeatureOmitted("tagsAndDependencies");
  const scheduleUiEnabled = !isUiFeatureOmitted("schedule");
  const modalBusy = isTaskEdit
    ? input.patchPending
    : input.pending || (isTemplateMode && input.saving);
  const modalTitle = isTaskEdit
    ? "Edit task"
    : isTemplateMode
      ? input.composeOperation === "edit"
        ? "Edit template"
        : "New template"
      : "New task";
  const modalTitleId = isEdit
    ? isTemplateMode
      ? "task-template-edit-modal-title"
      : "task-edit-modal-title"
    : isTemplateMode
      ? "task-template-create-modal-title"
      : "task-create-modal-title";
  const modalDescribedBy = isTaskEdit ? "task-edit-modal-description" : undefined;
  const idsPrefix = isEdit
    ? isTemplateMode
      ? "task-template-edit"
      : "task-edit"
    : isTemplateMode
      ? "task-template-new"
      : "task-new";
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
    modalBusy,
    modalTitle,
    modalTitleId,
    modalDescribedBy,
    idsPrefix,
    status,
    showTestScenarios,
    showDraftStatus,
  };
}
