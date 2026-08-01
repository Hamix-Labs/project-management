import { Link, useParams } from "react-router-dom";
import { errorMessage } from "@/lib/errorMessage";
import { TaskDetailPageSkeleton } from "../components/skeletons";
import { useTaskDetailMutations } from "../hooks/useTaskDetailMutations";
import { useTaskDetailPageQueries } from "../hooks/useTaskDetailPageQueries";
import { useTaskDetailScheduling } from "../hooks/useTaskDetailScheduling";
import { useTasksAppMeta, useTasksAppModals } from "../app/TasksAppProvider";
import {
  resolveAutonomyMode,
  TaskDetailLoadedView,
} from "./TaskDetailLoadedView";

function renderMissingTaskId() {
  return (
    <p className="muted" role="status">
      Missing task id.
    </p>
  );
}

function renderTaskLoadError(error: unknown, onRetry: () => void) {
  return (
    <section className="panel task-detail-panel task-detail-content--enter">
      <div className="err" role="alert">
        <p>{errorMessage(error, "Could not load task.")}</p>
        <div className="task-detail-error-actions">
          <button
            type="button"
            className="secondary"
            onClick={() => void onRetry()}
          >
            Try again
          </button>
          <Link to="/" className="pd__back project-context-back-link">
            <span aria-hidden="true">&#8249;</span>
            All tasks
          </Link>
        </div>
      </div>
    </section>
  );
}

export function TaskDetailPage() {
  const modals = useTasksAppModals();
  const { saving } = useTasksAppMeta();
  const { taskId = "" } = useParams<{ taskId: string }>();
  const { taskQuery, dependencySummaries } = useTaskDetailPageQueries(taskId);
  const scheduling = useTaskDetailScheduling(taskId);
  const {
    modelConfigOpen,
    setModelConfigOpen,
    autonomyConfirmOpen,
    setAutonomyConfirmOpen,
    approveConfirmOpen,
    setApproveConfirmOpen,
    openPrConfirmOpen,
    setOpenPrConfirmOpen,
    polishDialogOpen,
    setPolishDialogOpen,
    approveMutation,
    openPrMutation,
    polishMutation,
    autonomyMutation,
  } = useTaskDetailMutations(taskId);

  if (!taskId) {
    return renderMissingTaskId();
  }

  if (taskQuery.isPending) {
    return <TaskDetailPageSkeleton />;
  }

  if (taskQuery.isError) {
    return renderTaskLoadError(taskQuery.error, () => void taskQuery.refetch());
  }

  const task = taskQuery.data;
  const autonomyMode = resolveAutonomyMode(task.status);

  return (
    <TaskDetailLoadedView
      task={task}
      taskId={taskId}
      taskQuerySuccess={taskQuery.isSuccess}
      saving={saving}
      modals={modals}
      scheduling={scheduling}
      dependencySummaries={dependencySummaries}
      autonomyMode={autonomyMode}
      autonomyConfirmOpen={autonomyConfirmOpen}
      setAutonomyConfirmOpen={setAutonomyConfirmOpen}
      autonomyMutation={autonomyMutation}
      approveConfirmOpen={approveConfirmOpen}
      setApproveConfirmOpen={setApproveConfirmOpen}
      approveMutation={approveMutation}
      openPrConfirmOpen={openPrConfirmOpen}
      setOpenPrConfirmOpen={setOpenPrConfirmOpen}
      openPrMutation={openPrMutation}
      polishDialogOpen={polishDialogOpen}
      setPolishDialogOpen={setPolishDialogOpen}
      polishMutation={polishMutation}
      modelConfigOpen={modelConfigOpen}
      setModelConfigOpen={setModelConfigOpen}
    />
  );
}
