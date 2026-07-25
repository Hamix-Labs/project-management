import { ConfirmDialog } from "@/components/feedback/ConfirmDialog";

export type TaskBulkCloseRow = {
  id: string;
  title: string;
  number?: number | null;
};

type Props = {
  tasks: readonly TaskBulkCloseRow[];
  busy: boolean;
  error?: string | null;
  onCancel: () => void;
  onConfirm: () => void;
};

/**
 * Confirms `POST /tasks/{id}/close` for the current selection. Close is
 * reversible (see docs/api.md — `/reopen`), so the copy is deliberately
 * reassuring compared to the retired bulk-delete confirm.
 */
export function TaskBulkCloseConfirmModal({
  tasks,
  busy,
  error = null,
  onCancel,
  onConfirm,
}: Props) {
  const count = tasks.length;
  const noun = count === 1 ? "task" : "tasks";

  return (
    <ConfirmDialog
      title={`Close ${count} ${noun}?`}
      description={
        count === 1 ? (
          <>
            <strong>{tasks[0]?.title ?? "This task"}</strong> will stop
            execution and be marked closed.
          </>
        ) : (
          <>The {count} selected tasks will stop execution and be marked closed.</>
        )
      }
      footnote="Stops execution for each task. You can reopen later."
      confirmLabel={busy ? "Closing…" : `Close ${count}`}
      confirmVariant="danger"
      busy={busy}
      busyLabel="Closing tasks…"
      cancelDisabled={busy}
      confirmDisabled={busy}
      error={error}
      onCancel={onCancel}
      onConfirm={onConfirm}
      titleId="task-bulk-close-title"
      descriptionId="task-bulk-close-description"
      confirmTestId="task-bulk-close-confirm"
    />
  );
}
