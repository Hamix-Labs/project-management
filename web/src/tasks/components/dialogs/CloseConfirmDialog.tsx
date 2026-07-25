import { ConfirmDialog } from "@/components/feedback/ConfirmDialog";
import { taskDisplayRef } from "@/lib/taskShortId";

type Props = {
  taskTitle: string;
  /** Task id — powers the `#N` / short-UUID reference in the copy. */
  taskId: string;
  /** Optional per-project sequential number (`#N`). */
  taskNumber?: number | null;
  saving: boolean;
  closePending: boolean;
  error?: string | null;
  onCancel: () => void;
  onConfirm: () => void;
};

/**
 * Confirms a `POST /tasks/{id}/close`. The wording is deliberately
 * reassuring — closing is reversible via `/reopen` (see docs/api.md),
 * unlike the retired hard-delete flow which permanently dropped the row.
 */
export function CloseConfirmDialog({
  taskTitle,
  taskId,
  taskNumber,
  saving,
  closePending,
  error = null,
  onCancel,
  onConfirm,
}: Props) {
  const ref = taskDisplayRef({ id: taskId, number: taskNumber });
  const shownTitle = taskTitle.trim() || ref;
  return (
    <ConfirmDialog
      title="Close this task?"
      description={
        <>
          <strong>{shownTitle}</strong> will stop execution and be marked closed.
        </>
      }
      footnote={`Stops execution and closes ${ref}. You can reopen later.`}
      confirmLabel="Close task"
      confirmVariant="danger"
      titleId="close-dialog-title"
      descriptionId="close-dialog-description"
      busy={closePending}
      cancelDisabled={saving}
      confirmDisabled={saving}
      error={error}
      onCancel={onCancel}
      onConfirm={onConfirm}
    />
  );
}
