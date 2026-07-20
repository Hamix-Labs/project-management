import { CopyableId } from "@/shared/CopyableId";
import type { CycleTerminalOverviewModel } from "../../task-events/parsePhaseEventOverview";

export function CycleTerminalOverviewBody({
  model,
}: {
  model: CycleTerminalOverviewModel;
}) {
  const tone = model.terminal === "failed" ? "failed" : "success";
  return (
    <div
      className="task-event-cycle-overview task-event-phase-overview"
      data-terminal={model.terminal}
    >
      <div className="task-event-phase-overview-header">
        <span
          className="task-event-status-pill"
          data-tone={tone}
          data-status={model.status.toLowerCase()}
        >
          {model.status}
        </span>
      </div>
      <dl className="task-event-phase-meta">
        <div>
          <dt>Cycle</dt>
          <dd>
            <CopyableId value={model.cycleId} />
          </dd>
        </div>
        <div>
          <dt>Attempt</dt>
          <dd>#{model.attemptSeq}</dd>
        </div>
      </dl>
      {model.terminal === "failed" ? (
        <div className="task-event-cycle-failure-block">
          {model.failureSummary ? (
            <div
              className="task-event-phase-alert"
              role="alert"
              data-severity="error"
            >
              <p className="task-event-phase-alert-msg">{model.failureSummary}</p>
              {model.reason ? (
                <p className="task-event-cycle-reason-code">
                  <span className="muted">Reason code</span>{" "}
                  <code>{model.reason}</code>
                </p>
              ) : null}
            </div>
          ) : model.reason ? (
            <p className="task-event-cycle-reason-only">
              <span className="muted">Reason code</span>{" "}
              <code>{model.reason}</code>
            </p>
          ) : null}
        </div>
      ) : null}
    </div>
  );
}
