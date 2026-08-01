import { ConfirmDialog } from "@/components/feedback/ConfirmDialog";

type Props = {
  saving: boolean;
  pending: boolean;
  error?: string | null;
  onCancel: () => void;
  onConfirm: () => void;
};

export function TaskOpenPRConfirmDialog({
  saving,
  pending,
  error = null,
  onCancel,
  onConfirm,
}: Props) {
  return (
    <ConfirmDialog
      title="Approve and Open a pull request?"
      description="Resumes the same conversation to push and open the PR."
      confirmLabel="Approve & Open PR"
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
