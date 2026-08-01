import { ConfirmDialog } from "@/components/feedback/ConfirmDialog";

type Props = {
  taskTitle: string;
  saving: boolean;
  pending: boolean;
  error?: string | null;
  onCancel: () => void;
  onConfirm: () => void;
};

export function TaskApproveConfirmDialog({
  taskTitle,
  saving,
  pending,
  error = null,
  onCancel,
  onConfirm,
}: Props) {
  return (
    <ConfirmDialog
      title="Mark this task done?"
      description={<strong>{taskTitle}</strong>}
      footnote="Marks the task Done after the pull request is open. Dependents waiting on this task can proceed."
      confirmLabel="Mark done"
      confirmVariant="primary"
      busy={pending}
      cancelDisabled={saving}
      confirmDisabled={saving}
      error={error}
      onCancel={onCancel}
      onConfirm={onConfirm}
      titleId="task-approve-dialog-title"
      descriptionId="task-approve-dialog-description"
    />
  );
}
