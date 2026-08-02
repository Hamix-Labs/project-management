import { ConfirmDialog } from "@/components/feedback/ConfirmDialog";

type Props = {
  taskTitle: string;
  saving: boolean;
  pending: boolean;
  error?: string | null;
  /** Stack layer approve from review (non-root) vs mark-done from pr_ready. */
  mode?: "mark_done" | "stack_layer";
  onCancel: () => void;
  onConfirm: () => void;
};

export function TaskApproveConfirmDialog({
  taskTitle,
  saving,
  pending,
  error = null,
  mode = "mark_done",
  onCancel,
  onConfirm,
}: Props) {
  const layer = mode === "stack_layer";
  return (
    <ConfirmDialog
      title={layer ? "Approve this stack layer?" : "Mark this task done?"}
      description={<strong>{taskTitle}</strong>}
      footnote={
        layer
          ? "Accepts this layer's work. The worktree root publishes the GitHub stack with Approve & Open PR."
          : "Marks the task Done after the pull request is open. Dependents waiting on this task can proceed."
      }
      confirmLabel={layer ? "Approve" : "Mark done"}
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
