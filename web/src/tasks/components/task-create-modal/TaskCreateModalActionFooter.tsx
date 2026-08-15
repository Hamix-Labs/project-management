import type { ChecklistItemDraft, PriorityChoice } from "@/types";
import { TaskCreateModalEditFooterActions } from "./fields/TaskCreateModalEditFooterActions";
import { TaskCreateModalFooterActions } from "./fields/TaskCreateModalFooterActions";
import type { TaskCreateModalPresentation } from "./taskCreateModalPresentation";

type Props = {
  presentation: TaskCreateModalPresentation;
  title: string;
  priority: PriorityChoice;
  checklistItems: ChecklistItemDraft[];
  repositoryId: string;
  draftSaving: boolean;
  /** Associates submit with a form when the footer is portaled outside it. */
  form?: string;
  onClose: () => void;
  onSaveDraft: () => void;
};

export function TaskCreateModalActionFooter({
  presentation,
  title,
  priority,
  checklistItems,
  repositoryId,
  draftSaving,
  form,
  onClose,
  onSaveDraft,
}: Props) {
  if (presentation.isEdit) {
    return (
      <TaskCreateModalEditFooterActions
        disabled={presentation.disabled}
        saveDisabled={!title.trim()}
        form={form}
        onClose={onClose}
      />
    );
  }

  return (
    <TaskCreateModalFooterActions
      variant={presentation.isTemplateMode ? "template" : "task-create"}
      disabled={presentation.disabled}
      draftSaving={draftSaving}
      title={title}
      priority={priority}
      checklistItems={checklistItems}
      repositoryId={repositoryId}
      requireGitBinding
      form={form}
      onClose={onClose}
      onSaveDraft={presentation.isTemplateMode ? undefined : onSaveDraft}
    />
  );
}
