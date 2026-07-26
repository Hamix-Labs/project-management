import { ConfirmDialog } from "@/components/feedback/ConfirmDialog";

/**
 * `enable` is true when the operator is moving the task from `on_hold`
 * to `ready` (handing control back to the agent), false when they are
 * moving from `ready` to `on_hold` (parking the task).
 */
type Props = {
  enable: boolean;
  taskTitle: string;
  saving: boolean;
  pending: boolean;
  error?: string | null;
  onCancel: () => void;
  onConfirm: () => void;
};

export function AutonomyConfirmDialog({
  enable,
  taskTitle,
  saving,
  pending,
  error = null,
  onCancel,
  onConfirm,
}: Props) {
  const title = enable
    ? "Resume autonomous execution?"
    : "Pause this task?";
  const body = enable
    ? "The agent will pick this task up when no other task is running."
    : "The agent will stop considering this task. You can resume it any time from this page.";
  const confirmLabel = enable ? "Resume" : "Pause";

  return (
    <ConfirmDialog
      title={title}
      description={<strong>{taskTitle}</strong>}
      footnote={body}
      confirmLabel={confirmLabel}
      confirmVariant="primary"
      busy={pending}
      cancelDisabled={saving}
      confirmDisabled={saving}
      error={error}
      onCancel={onCancel}
      onConfirm={onConfirm}
      titleId="autonomy-dialog-title"
      descriptionId="autonomy-dialog-description"
    />
  );
}
