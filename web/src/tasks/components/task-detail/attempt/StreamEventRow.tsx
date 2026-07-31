import type { TaskCycleStreamEvent } from "@/types";
import {
  agentProgressKindDescriptor,
  resolveAgentProgressMessage,
} from "@/tasks/cycleDisplay/agentProgressDisplay";
import { PhaseSeqBadge } from "./AttemptPhaseSeqBadge";

type StreamEventRowProps = {
  ev: TaskCycleStreamEvent;
  showPhaseBadge: boolean;
};

export function StreamEventRow({ ev, showPhaseBadge }: StreamEventRowProps) {
  const preview = resolveAgentProgressMessage(
    ev.subtype,
    ev.message,
    ev.tool,
    "Agent reported progress.",
  );
  const kind = agentProgressKindDescriptor(ev.kind, ev.subtype, ev.tool);
  return (
    <li className="task-attempt-stream-row">
      <details className="task-attempt-stream-details">
        <summary className="task-attempt-stream-summary">
          <time className="task-attempt-stream-time" dateTime={ev.at}>
            {new Date(ev.at).toLocaleTimeString(undefined, {
              hour: "numeric",
              minute: "2-digit",
            })}
          </time>
          <span className="task-attempt-stream-label">
            <span
              className={`task-attempt-stream-kind task-attempt-stream-kind--${kind.tone}`}
              title={kind.title}
            >
              {kind.label}
            </span>
            <span className="task-attempt-stream-message" title={preview}>
              {preview}
            </span>
          </span>
          {showPhaseBadge ? <PhaseSeqBadge seq={ev.phase_seq} /> : null}
        </summary>
        <div className="task-attempt-stream-detail-panel">
          <dl className="task-attempt-stream-detail-list">
            {ev.tool ? (
              <div>
                <dt>Tool</dt>
                <dd>{ev.tool}</dd>
              </div>
            ) : null}
            <div>
              <dt>Phase</dt>
              <dd>#{ev.phase_seq}</dd>
            </div>
          </dl>
          <div className="task-attempt-stream-detail-block">
            <h4>Raw payload</h4>
            <pre>{JSON.stringify(ev.payload, null, 2)}</pre>
          </div>
        </div>
      </details>
    </li>
  );
}
