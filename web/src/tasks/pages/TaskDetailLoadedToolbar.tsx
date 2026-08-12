import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { StatusBadge } from "@/components/task-status";
import {
  TaskDetailSchedule,
  TaskDetailToolbarActions,
} from "../components/task-detail";
import { TokenUsageChip } from "../components/task-detail/TokenUsageChip";
import { ViewPullRequestLink } from "../components/task-detail/layout/ViewPullRequestLink";
import { useTaskCycles } from "../hooks/useTaskCycles";
import { canEditTask } from "../task-display/canEditTask";
import { taskEditPath, tasksNewPath } from "../composeRoutes";
import {
  CREATING_PR_STATUS_LABEL,
  isOpenPrRunKind,
  openPrSessionClearedByStatus,
  shouldShowCreatingPrLabel,
} from "../task-display/openPrRunDisplay";
import { statusNeedsUserInput } from "../task-display";
import type { TaskDetailLoadedViewProps } from "./TaskDetailLoadedView";

type TaskDetailLoadedToolbarProps = Pick<
  TaskDetailLoadedViewProps,
  | "task"
  | "saving"
  | "modals"
  | "autonomyMode"
  | "setAutonomyConfirmOpen"
  | "setApproveConfirmOpen"
  | "setOpenPrConfirmOpen"
  | "setPolishDialogOpen"
  | "setModelConfigOpen"
  | "approveMutation"
  | "openPrMutation"
  | "polishMutation"
  | "autonomyMutation"
>;

export function TaskDetailLoadedToolbar({
  task,
  saving,
  modals,
  autonomyMode,
  setAutonomyConfirmOpen,
  setApproveConfirmOpen,
  setOpenPrConfirmOpen,
  setPolishDialogOpen,
  setModelConfigOpen,
  approveMutation,
  openPrMutation,
  polishMutation,
  autonomyMutation,
}: TaskDetailLoadedToolbarProps) {
  const navigate = useNavigate();
  const inReview = task.status === "review";
  const inPrReady = task.status === "pr_ready";
  const isClosed = task.status === "closed";
  const needsUser = statusNeedsUserInput(task.status);
  const cyclesQuery = useTaskCycles(task.id, {
    enabled: task.status === "running",
  });
  const runningOpenPr = (cyclesQuery.data?.cycles ?? []).some(
    (c) => c.status === "running" && isOpenPrRunKind(c.meta),
  );
  const [openPrSession, setOpenPrSession] = useState(false);
  useEffect(() => {
    if (openPrMutation.isPending) {
      setOpenPrSession(true);
    }
  }, [openPrMutation.isPending]);
  useEffect(() => {
    if (openPrSessionClearedByStatus(task.status)) {
      setOpenPrSession(false);
    }
  }, [task.status]);
  const creatingPr = shouldShowCreatingPrLabel({
    mutationPending: openPrMutation.isPending,
    sessionActive: openPrSession,
    hasRunningOpenPrCycle: runningOpenPr,
  });
  const statusLabel = creatingPr ? CREATING_PR_STATUS_LABEL : undefined;

  return (
    <div className="task-detail-toolbar">
      <div className="task-detail-toolbar-left">
        <div
          className="task-detail-execution-bar"
          data-testid="task-detail-execution-bar"
        >
          <TokenUsageChip taskId={task.id} />
          <div
            className="task-detail-execution-bar__status"
            data-testid="task-detail-status"
          >
            <StatusBadge
              status={creatingPr ? "running" : task.status}
              label={statusLabel}
              className="task-detail-status-badge"
              data-needs-user={needsUser ? "true" : undefined}
            />
          </div>
          <ViewPullRequestLink url={task.pull_request_url ?? ""} />
        </div>
        <TaskDetailSchedule task={task} />
      </div>
      <TaskDetailToolbarActions
        saving={saving}
        canEdit={canEditTask(task.status) && !isClosed}
        onEdit={() => navigate(taskEditPath(task.id))}
        onClose={
          isClosed
            ? undefined
            : () =>
                modals.requestClose({
                  id: task.id,
                  title: task.title,
                  number: task.number,
                })
        }
        closePending={modals.closePending}
        onReopen={isClosed ? () => modals.reopen(task.id) : undefined}
        reopenPending={modals.reopenPending}
        onOpenPr={inReview ? () => setOpenPrConfirmOpen(true) : undefined}
        openPrPending={openPrMutation.isPending}
        onApprove={inPrReady ? () => setApproveConfirmOpen(true) : undefined}
        approvePending={approveMutation.isPending}
        onPolish={inReview ? () => setPolishDialogOpen(true) : undefined}
        polishPending={polishMutation.isPending}
        onEnqueue={
          !isClosed
            ? () => {
                const projectID = task.project_id?.trim() ?? "";
                const repositoryID = task.repository_id?.trim() ?? "";
                const worktreeID = task.worktree_id?.trim() ?? "";
                if (!projectID || !repositoryID || !worktreeID) return;
                navigate(
                  tasksNewPath({
                    project: projectID,
                    repository: repositoryID,
                    worktree: worktreeID,
                    lockGit: true,
                    lockProject: true,
                  }),
                );
              }
            : undefined
        }
        enqueueDisabledReason={
          !isClosed && !(task.worktree_id ?? "").trim()
            ? "Worktree is still provisioning"
            : undefined
        }
        onConfigureModel={() => setModelConfigOpen(true)}
        showModelConfig={task.status === "failed"}
        autonomyMode={autonomyMode}
        onToggleAutonomy={
          autonomyMode !== "hidden"
            ? () => setAutonomyConfirmOpen(true)
            : undefined
        }
        autonomyPending={autonomyMutation.isPending}
      />
    </div>
  );
}
