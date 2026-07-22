import { errorMessage } from "@/lib/errorMessage";
import type { CycleCommit, CycleGitContext } from "@/types/cycle";
import { useTaskCycleVerdicts } from "../../../hooks/useTaskCycles";
import { CommitList } from "../commits/CommitList";
import { GitContextMeta } from "../commits/GitContextMeta";

type Props = {
  taskId: string;
  cycleId: string;
};

/** Commits indexed for this attempt (cycle), from GET .../verdicts. */
export function AttemptCommitsSection({ taskId, cycleId }: Props) {
  const verdictsQuery = useTaskCycleVerdicts(taskId, cycleId);
  if (verdictsQuery.isPending) {
    return (
      <section
        className="task-attempt-commits"
        data-testid="task-attempt-commits"
        aria-busy="true"
      >
        <h2 className="task-attempt-commits-heading">Commits</h2>
        <p className="muted">Loading commits…</p>
      </section>
    );
  }
  if (verdictsQuery.isError) {
    return (
      <section
        className="task-attempt-commits"
        data-testid="task-attempt-commits"
      >
        <h2 className="task-attempt-commits-heading">Commits</h2>
        <p className="err" role="alert">
          {errorMessage(verdictsQuery.error, "Could not load commits.")}
        </p>
      </section>
    );
  }
  const { commits, git_context: gitContext } = verdictsQuery.data;
  return (
    <AttemptCommitsBody
      taskId={taskId}
      commits={commits}
      gitContext={gitContext}
    />
  );
}

function AttemptCommitsBody({
  taskId,
  commits,
  gitContext,
}: {
  taskId: string;
  commits: ReadonlyArray<CycleCommit>;
  gitContext?: CycleGitContext;
}) {
  if (commits.length === 0) {
    return (
      <section
        className="task-attempt-commits"
        data-testid="task-attempt-commits"
        aria-label="Git commits"
      >
        <h2 className="task-attempt-commits-heading">Commits</h2>
        <p className="muted" data-testid="task-attempt-commits-empty">
          No commits indexed for this attempt.
        </p>
      </section>
    );
  }
  const ctx = gitContext ?? {
    repo: commits[0]?.repo ?? "",
    worktree: commits[0]?.worktree ?? "",
    branch: commits[commits.length - 1]?.branch ?? commits[0]?.branch ?? "",
  };
  return (
    <section
      className="task-attempt-commits"
      data-testid="task-attempt-commits"
      aria-label="Git commits"
    >
      <h2 className="task-attempt-commits-heading">Commits</h2>
      <GitContextMeta context={ctx} />
      <CommitList taskId={taskId} commits={commits} />
    </section>
  );
}
