import type { ChecklistItemDraft, PriorityChoice } from "@/types";
import { TaskCreateModalEditFooterActions } from "./fields/TaskCreateModalEditFooterActions";
import { TaskCreateModalFooterActions } from "./fields/TaskCreateModalFooterActions";
import type { TaskCreateModalPresentation } from "./taskCreateModalPresentation";

type Props = {
  presentation: TaskCreateModalPresentation;
  title: string;
  priority: PriorityChoice;
  checklistItems: ChecklistItemDraft[];
  worktreeId: string;
  draftSaving: boolean;
  onClose: () => void;
  onSaveDraft: () => void;
};

export function TaskCreateModalActionFooter({
  presentation,
  title,
  priority,
  checklistItems,
  worktreeId,
  draftSaving,
  onClose,
  onSaveDraft,
}: Props) {
  if (presentation.isTaskEdit) {
    return (
      <TaskCreateModalEditFooterActions
        disabled={presentation.disabled}
        saveDisabled={!title.trim()}
        onClose={onClose}
      />
    );
  }

  if (presentation.isTemplateMode && presentation.isEdit) {
    return (
      <TaskCreateModalEditFooterActions
        disabled={presentation.disabled}
        saveDisabled={!title.trim()}
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
      worktreeId={worktreeId}
      requireGitBinding
      onClose={onClose}
      onSaveDraft={presentation.isTemplateMode ? undefined : onSaveDraft}
    />
  );
}
