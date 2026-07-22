import { useMemo } from "react";
import { errorMessage } from "@/lib/errorMessage";
import {
  EmptyState,
  EmptyStateCommitsGlyph,
} from "@/shared/EmptyState";
import { useTaskCommits } from "@/tasks/hooks/useTaskCommits";
import { TaskDetailCollapsibleSection } from "../layout/TaskDetailCollapsibleSection";
import { TaskDetailGitBranchGlyph } from "../layout/TaskDetailActionGlyphs";
import { CommitList } from "./CommitList";

type Props = {
  taskId: string;
  enabled?: boolean;
};

export function TaskCommitsPanel({ taskId, enabled = true }: Props) {
  const commitsQuery = useTaskCommits(taskId, { enabled });
  const commits = useMemo(
    () => commitsQuery.data?.commits ?? [],
    [commitsQuery.data?.commits],
  );

  const branch = useMemo(() => {
    if (commits.length === 0) return "";
    const first = commits[0];
    const last = commits[commits.length - 1];
    return (last.branch || first.branch || "").trim();
  }, [commits]);

  const count =
    !commitsQuery.isPending && !commitsQuery.isError ? commits.length : null;

  return (
    <TaskDetailCollapsibleSection
      className="task-commits-panel"
      title="Commits"
      headingId="task-commits-heading"
      count={count}
      defaultOpen
      data-testid="task-commits-panel"
    >
      {commitsQuery.isPending ? (
        <CommitsLoading />
      ) : commitsQuery.isError ? (
        <div className="err" role="alert">
          <p>
            {errorMessage(
              commitsQuery.error,
              "Could not load commits.",
            )}
          </p>
          <div className="task-detail-error-actions">
            <button
              type="button"
              className="secondary"
              onClick={() => {
                void commitsQuery.refetch();
              }}
            >
              Try again
            </button>
          </div>
        </div>
      ) : commits.length === 0 ? (
        <div className="task-commits-empty-well" data-testid="task-commits-empty-well">
          <EmptyState
            icon={<EmptyStateCommitsGlyph />}
            title="No commits indexed yet"
            description="Recorded when an agent run commits to git."
            className="task-commits-empty-state"
          />
        </div>
      ) : (
        <>
          {branch !== "" ? (
            <div
              className="task-commits-branch-line"
              data-testid="task-commits-branch-line"
            >
              <TaskDetailGitBranchGlyph className="task-detail-action-glyph" />
              <span className="task-commits-branch-name">{branch}</span>
            </div>
          ) : null}
          <CommitList taskId={taskId} commits={commits} showAttempt />
        </>
      )}
    </TaskDetailCollapsibleSection>
  );
}

function CommitsLoading() {
  return (
    <div className="task-commits-empty-well" aria-busy="true">
      <ul
        className="task-commits-list task-commits-list--loading"
        aria-label="Loading commits"
      >
        <li className="task-commit-row task-commit-row--skeleton" />
        <li className="task-commit-row task-commit-row--skeleton" />
      </ul>
    </div>
  );
}
