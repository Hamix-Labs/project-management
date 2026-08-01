import { TaskListEditGlyph } from "@/shared/ListRowActionGlyphs";

type Props = {
  saving: boolean;
  /** When false, edit is disabled (e.g. task is running). Close stays available. */
  canEdit?: boolean;
  onEdit: () => void;
  /**
   * Click handler for the Close action (POST /tasks/{id}/close). Rendered
   * for every non-closed task; on `closed` tasks the toolbar swaps in
   * `onReopen` instead.
   */
  onClose?: () => void;
  closePending?: boolean;
  /**
   * Click handler for the Reopen action (POST /tasks/{id}/reopen).
   * Rendered only when the current task's status is `closed` so the
   * user can reverse a close without leaving the detail page.
   */
  onReopen?: () => void;
  reopenPending?: boolean;
  /**
   * When set, shows the "Model configuration" action which opens the
   * model-configuration modal (consolidates the failure-recovery hint
   * that used to live inline below the action row).
   */
  onConfigureModel?: () => void;
  /**
   * Gates whether the "Model configuration" action is offered at all.
   * Today it is offered after a failed run; older copy referred to this
   * as `failedRunnerHint`.
   */
  showModelConfig?: boolean;
  /**
   * Autonomous-execution mode for this task. `"hidden"` suppresses the
   * toggle entirely (e.g. running, done, failed — the autonomy concept
   * does not apply). `"ready"` shows a "Pause" action; `"on_hold"`
   * shows a "Resume" action. Both actions go through a confirm dialog
   * upstream of `onToggleAutonomy`.
   */
  autonomyMode?: "hidden" | "ready" | "on_hold";
  onToggleAutonomy?: () => void;
  autonomyPending?: boolean;
  /** When set, shows Mark done for a task in pr_ready (POST /approve). */
  onApprove?: () => void;
  approvePending?: boolean;
  /** When set, shows Approve & Open PR for a task in review (POST /open-pr). */
  onOpenPr?: () => void;
  openPrPending?: boolean;
  /** When set, shows Polish for a task in review (POST /polish). */
  onPolish?: () => void;
  polishPending?: boolean;
};

export function TaskDetailToolbarActions({
  saving,
  canEdit = true,
  onEdit,
  onClose,
  closePending = false,
  onReopen,
  reopenPending = false,
  onConfigureModel,
  showModelConfig,
  autonomyMode = "hidden",
  onToggleAutonomy,
  autonomyPending = false,
  onApprove,
  approvePending = false,
  onOpenPr,
  openPrPending = false,
  onPolish,
  polishPending = false,
}: Props) {
  const showAutonomy =
    autonomyMode !== "hidden" && typeof onToggleAutonomy === "function";
  const autonomyLabel =
    autonomyMode === "on_hold" ? "Resume" : "Pause";
  const autonomyPendingLabel =
    autonomyMode === "on_hold" ? "Resuming…" : "Pausing…";
  const reviewActionsBusy = openPrPending || polishPending;

  return (
    <div className="task-detail-actions">
      {onOpenPr ? (
        <button
          type="button"
          className="task-detail-btn-approve"
          onClick={onOpenPr}
          disabled={saving || reviewActionsBusy || approvePending}
        >
          {openPrPending ? "Opening PR…" : "Approve & Open PR"}
        </button>
      ) : null}
      {onApprove ? (
        <button
          type="button"
          className="task-detail-btn-approve"
          onClick={onApprove}
          disabled={saving || approvePending || reviewActionsBusy}
        >
          {approvePending ? "Marking done…" : "Mark done"}
        </button>
      ) : null}
      {onPolish ? (
        <button
          type="button"
          className="task-detail-btn-polish"
          onClick={onPolish}
          disabled={saving || reviewActionsBusy || approvePending}
        >
          Polish
        </button>
      ) : null}
      {showAutonomy ? (
        <button
          type="button"
          className="task-detail-btn-autonomy"
          onClick={onToggleAutonomy}
          disabled={saving || autonomyPending}
          data-autonomy-mode={autonomyMode}
        >
          {autonomyPending ? autonomyPendingLabel : autonomyLabel}
        </button>
      ) : null}
      <button
        type="button"
        className="task-detail-btn-edit"
        onClick={onEdit}
        disabled={saving || !canEdit}
        title={
          canEdit ? undefined : "Cannot edit while the task is in progress"
        }
      >
        <TaskListEditGlyph />
        Edit task
      </button>
      {showModelConfig && onConfigureModel ? (
        <button
          type="button"
          className="task-detail-btn-model-config"
          onClick={onConfigureModel}
          disabled={saving}
        >
          Model configuration
        </button>
      ) : null}
      {onReopen ? (
        <button
          type="button"
          className="task-detail-btn-reopen"
          onClick={onReopen}
          disabled={saving || reopenPending}
        >
          {reopenPending ? "Reopening…" : "Reopen"}
        </button>
      ) : onClose ? (
        <button
          type="button"
          className="task-detail-btn-close"
          onClick={onClose}
          disabled={saving || closePending}
        >
          {closePending ? "Closing…" : "Close"}
        </button>
      ) : null}
    </div>
  );
}
