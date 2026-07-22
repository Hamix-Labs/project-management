import { useQuery } from "@tanstack/react-query";
import { useMemo } from "react";
import { Link, useParams } from "react-router-dom";
import { getTask } from "@/api";
import { maxRepoShaQueryBytes } from "@/api/repo";
import { taskQueryKeys } from "@/lib/taskQueryKeys";
import { useDocumentTitle } from "@/shared/useDocumentTitle";
import { useNow } from "@/shared/useNow";
import { CommitDiffPanel } from "../components/task-detail/commits/CommitDiffPanel";
import {
  commitShaParamPattern,
  shortSha,
} from "../components/task-detail/commits/commitDisplay";
import { useCommitDiff } from "../hooks/useCommitDiff";
import { useTaskCommits } from "../hooks/useTaskCommits";
import { QUERY_POLICY } from "../queryPolicy";
import { TaskCommitDiffPageHeader } from "./TaskCommitDiffPageHeader";

export function TaskCommitDiffPage() {
  const now = useNow();
  const { taskId = "", sha: shaParam = "" } = useParams<{
    taskId: string;
    sha: string;
  }>();
  const sha = decodeURIComponent(shaParam).trim();
  const shaValid =
    sha.length > 0 &&
    sha.length <= maxRepoShaQueryBytes &&
    commitShaParamPattern.test(sha);

  const taskQuery = useQuery({
    queryKey: taskQueryKeys.detail(taskId),
    queryFn: ({ signal }) => getTask(taskId, { signal }),
    enabled: Boolean(taskId) && shaValid,
    staleTime: QUERY_POLICY.detailStaleTimeMs,
  });
  const worktreeId = (taskQuery.data?.worktree_id ?? "").trim();

  const commitsQuery = useTaskCommits(taskId, {
    enabled: Boolean(taskId) && shaValid,
  });
  const diffQuery = useCommitDiff(sha, {
    worktreeId,
    enabled: Boolean(taskId) && shaValid && worktreeId !== "",
  });
  const commit = useMemo(
    () => commitsQuery.data?.commits.find((c) => c.sha === sha),
    [commitsQuery.data?.commits, sha],
  );

  const pageTitle = shaValid
    ? commit?.message
      ? `${shortSha(sha)}: ${commit.message}`
      : `Commit ${shortSha(sha)}`
    : "Invalid commit";
  useDocumentTitle(pageTitle);

  if (!taskId) {
    return (
      <p className="muted" role="status">
        Missing task id.
      </p>
    );
  }

  if (!shaValid) {
    return (
      <section className="panel task-detail-panel task-detail-content--enter">
        <div className="err" role="alert">
          <p>Invalid commit SHA in the URL.</p>
          <div className="task-detail-error-actions">
            <Link
              to={`/tasks/${encodeURIComponent(taskId)}`}
              className="pd__back project-context-back-link"
            >
              <span aria-hidden="true">&#8249;</span>
              Back to task
            </Link>
          </div>
        </div>
      </section>
    );
  }

  return (
    <section
      className="panel task-detail-panel task-commit-diff-page task-detail-content--enter"
      data-testid="task-commit-diff-page"
    >
      <TaskCommitDiffPageHeader
        taskId={taskId}
        sha={sha}
        now={now}
        commit={commit}
        commitsQuery={commitsQuery}
        diffQuery={diffQuery}
      />
      <CommitDiffPanel
        sha={sha}
        worktreeId={worktreeId}
        viewClassName="task-commit-diff-view task-commit-diff-view--page"
      />
    </section>
  );
}
