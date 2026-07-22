import { Link } from "react-router-dom";
import type { UseQueryResult } from "@tanstack/react-query";
import type { RepoDiffResult } from "@/api/repo";
import { errorMessage } from "@/lib/errorMessage";
import { CopyableId } from "@/shared/CopyableId";
import { formatRelativeTime } from "@/shared/time/relativeTime";
import type { TaskCommitsResponse } from "@/types";
import { shortSha } from "../components/task-detail/commits/commitDisplay";

type Props = {
  taskId: string;
  sha: string;
  now: number;
  commit: TaskCommitsResponse["commits"][number] | undefined;
  commitsQuery: UseQueryResult<TaskCommitsResponse, Error>;
  diffQuery: UseQueryResult<RepoDiffResult | null, Error>;
};

export function TaskCommitDiffPageHeader({
  taskId,
  sha,
  now,
  commit,
  commitsQuery,
  diffQuery,
}: Props) {
  const backTo = `/tasks/${encodeURIComponent(taskId)}`;
  const gitAuthor = diffQuery.data?.author;

  return (
    <header className="task-commit-diff-page-head">
      <Link to={backTo} className="pd__back project-context-back-link">
        <span aria-hidden="true">&#8249;</span>
        Back to task
      </Link>

      <div className="task-commit-diff-page-hero">
        {commit?.message ? (
          <h1 className="task-commit-diff-page-message">{commit.message}</h1>
        ) : (
          <h1 className="task-commit-diff-page-message muted">
            {shortSha(sha)}
          </h1>
        )}
      </div>

      <p className="task-commit-diff-page-meta muted">
        <CopyableId
          value={sha}
          displayValue={shortSha(sha)}
          copyLabel="Copy SHA"
          className="task-commit-diff-page-sha"
        />
        {commit ? (
          <>
            <span className="task-commit-meta-sep" aria-hidden="true">
              ·
            </span>
            <span>{formatRelativeTime(commit.committed_at, new Date(now))}</span>
            {commit.branch ? (
              <>
                <span className="task-commit-meta-sep" aria-hidden="true">
                  ·
                </span>
                <span>{commit.branch}</span>
              </>
            ) : null}
          </>
        ) : commitsQuery.isError ? (
          <>
            <span className="task-commit-meta-sep" aria-hidden="true">
              ·
            </span>
            <span role="status">
              {errorMessage(commitsQuery.error, "Could not load commit metadata.")}
            </span>
          </>
        ) : null}
        {gitAuthor ? (
          <>
            <span className="task-commit-meta-sep" aria-hidden="true">
              ·
            </span>
            <span
              title={
                diffQuery.data?.author_email
                  ? diffQuery.data.author_email
                  : undefined
              }
            >
              {gitAuthor}
            </span>
          </>
        ) : null}
      </p>
    </header>
  );
}
