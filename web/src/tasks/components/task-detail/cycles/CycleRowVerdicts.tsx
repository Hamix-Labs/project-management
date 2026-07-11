import { errorMessage } from "@/lib/errorMessage";
import {
  verdictPillClass,
  verifierKindLabel,
} from "@/tasks/task-events/parseVerificationSnapshot";
import type {
  CycleCommandRun,
  CycleCommit,
  CycleGitContext,
} from "@/types/cycle";
import { useTaskCycleVerdicts } from "../../../hooks/useTaskCycles";
import { CommitList } from "../commits/CommitList";
import { GitContextMeta } from "../commits/GitContextMeta";
import {
  commandRunAttemptSeqs,
  commandRunsForAttempt,
  groupVerdictsByAttempt,
} from "./cycleVerdictUtils";

type CycleRowVerdictsProps = {
  taskId: string;
  cycleId: string;
};

export function CycleRowVerdicts({ taskId, cycleId }: CycleRowVerdictsProps) {
  const verdictsQuery = useTaskCycleVerdicts(taskId, cycleId);
  if (verdictsQuery.isPending) {
    return (
      <p className="task-cycle-row-verdicts muted" aria-busy="true">
        Loading verdicts…
      </p>
    );
  }
  if (verdictsQuery.isError) {
    return (
      <p className="task-cycle-row-verdicts err" role="alert">
        {errorMessage(verdictsQuery.error, "Could not load verdicts.")}
      </p>
    );
  }
  const data = verdictsQuery.data;
  if (
    data.criteria_reports.length === 0 &&
    data.verify_reports.length === 0 &&
    data.command_runs.length === 0 &&
    data.commits.length === 0
  ) {
    return (
      <p
        className="task-cycle-row-verdicts muted"
        data-testid="task-cycle-verdicts-empty"
      >
        No verdicts captured for this cycle.
      </p>
    );
  }
  const groups = groupVerdictsByAttempt(
    data.criteria_reports,
    data.verify_reports,
  );
  return (
    <div className="task-cycle-row-verdicts" data-testid="task-cycle-verdicts">
      <CycleCommitsSummary
        taskId={taskId}
        gitContext={data.git_context}
        commits={data.commits}
      />
      <h4 className="task-cycle-row-verdicts-heading">Verdicts</h4>
      {groups.map((group) => (
        <section
          key={group.attemptSeq}
          className="task-cycle-verdicts-attempt"
          aria-label={`Attempt ${group.attemptSeq}`}
        >
          <p className="task-cycle-verdicts-attempt-eyebrow muted">
            Attempt #{group.attemptSeq}
          </p>
          <ul className="task-cycle-verdicts-list">
            {group.rows.map((row) => (
              <li
                key={row.criterionId}
                className="task-cycle-verdict-item"
                data-verified={String(row.verified)}
              >
                <header className="task-cycle-verdict-item-header">
                  <span className={`cell-pill ${verdictPillClass(row.verified)}`}>
                    {row.verified ? "Verified" : "Not verified"}
                  </span>
                  {row.verifierKind ? (
                    <span className="task-cycle-verdict-kind muted">
                      {verifierKindLabel(row.verifierKind)}
                    </span>
                  ) : null}
                  <span className="task-cycle-verdict-criterion">
                    {row.criterionId}
                  </span>
                </header>
                {row.reasoning ? (
                  <p className="task-cycle-verdict-reasoning">{row.reasoning}</p>
                ) : row.evidence ? (
                  <p className="task-cycle-verdict-evidence muted">
                    Agent-claimed evidence: {row.evidence}
                  </p>
                ) : null}
              </li>
            ))}
          </ul>
          <CycleCommandRunsSummary
            runs={commandRunsForAttempt(data.command_runs, group.attemptSeq)}
          />
        </section>
      ))}
      {groups.length === 0 && data.command_runs.length > 0
        ? commandRunAttemptSeqs(data.command_runs).map((attemptSeq) => (
            <section
              key={attemptSeq}
              className="task-cycle-verdicts-attempt"
              aria-label={`Attempt ${attemptSeq}`}
            >
              <p className="task-cycle-verdicts-attempt-eyebrow muted">
                Attempt #{attemptSeq}
              </p>
              <CycleCommandRunsSummary
                runs={commandRunsForAttempt(data.command_runs, attemptSeq)}
              />
            </section>
          ))
        : null}
    </div>
  );
}

function CycleCommitsSummary({
  taskId,
  gitContext,
  commits,
}: {
  taskId: string;
  gitContext?: CycleGitContext;
  commits: ReadonlyArray<CycleCommit>;
}) {
  if (commits.length === 0) {
    return null;
  }
  const ctx = gitContext ?? {
    repo: commits[0]?.repo ?? "",
    worktree: commits[0]?.worktree ?? "",
    branch: commits[commits.length - 1]?.branch ?? commits[0]?.branch ?? "",
  };
  return (
    <section
      className="task-cycle-commits"
      data-testid="task-cycle-commits"
      aria-label="Git commits"
    >
      <h4 className="task-cycle-row-verdicts-heading">Commits</h4>
      <GitContextMeta context={ctx} />
      <CommitList taskId={taskId} commits={commits} />
    </section>
  );
}

function CycleCommandRunsSummary({
  runs,
}: {
  runs: ReadonlyArray<CycleCommandRun>;
}) {
  if (runs.length === 0) {
    return null;
  }
  return (
    <ul className="task-cycle-command-runs-list">
      {runs.map((run) => (
        <li key={run.id} className="task-cycle-command-run muted">
          <span className="task-cycle-command-run-label">
            [{run.criterion_id}] command {run.command_seq}
          </span>
          <span className="task-cycle-command-run-meta">exit {run.exit_code}</span>
        </li>
      ))}
    </ul>
  );
}
