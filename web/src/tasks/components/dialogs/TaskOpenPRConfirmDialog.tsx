import { ConfirmDialog } from "@/components/feedback/ConfirmDialog";

type Props = {
  taskTitle: string;
  saving: boolean;
  pending: boolean;
  error?: string | null;
  onCancel: () => void;
  onConfirm: () => void;
};

export function TaskOpenPRConfirmDialog({
  taskTitle,
  saving,
  pending,
  error = null,
  onCancel,
  onConfirm,
}: Props) {
  return (
    <ConfirmDialog
      title="Approve and open a pull request?"
      description={<strong>{taskTitle}</strong>}
      footnote="Approves the agent’s work and resumes the same conversation to push the branch and open a PR. The task moves to PR ready when the pull request exists."
      confirmLabel="Approve & open PR"
      confirmVariant="primary"
      busy={pending}
      cancelDisabled={saving}
      confirmDisabled={saving}
      error={error}
      onCancel={onCancel}
      onConfirm={onConfirm}
      titleId="task-open-pr-dialog-title"
      descriptionId="task-open-pr-dialog-description"
    />
  );
}
