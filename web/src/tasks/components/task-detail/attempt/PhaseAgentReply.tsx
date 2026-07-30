import { useState } from "react";
import type { PhaseStatus } from "@/types";
import type { PhaseAgentReply as PhaseAgentReplyData } from "./latestAgentReplyByPhase";

const DISPLAY_CAP = 800;

type PhaseAgentReplyProps = {
  reply: PhaseAgentReplyData;
  phaseLabel: string;
  phaseStatus: PhaseStatus;
  phaseSeq: number;
  onViewInActivity?: (phaseSeq: number) => void;
};

export function PhaseAgentReply({
  reply,
  phaseLabel,
  phaseStatus,
  phaseSeq,
  onViewInActivity,
}: PhaseAgentReplyProps) {
  const [expanded, setExpanded] = useState(false);
  const needsTruncate = reply.text.length > DISPLAY_CAP;
  const displayText =
    needsTruncate && !expanded
      ? `${reply.text.slice(0, DISPLAY_CAP - 1).trimEnd()}…`
      : reply.text;

  return (
    <div
      className="task-attempt-phase-reply"
      data-status={phaseStatus}
      aria-label={`Agent reply for ${phaseLabel}`}
    >
      <div className="task-attempt-phase-reply-header">
        <span
          className="task-attempt-stream-kind task-attempt-stream-kind--reply"
          title="Message from the Cursor agent"
        >
          Agent reply
        </span>
        {reply.at ? (
          <time
            className="task-attempt-phase-reply-time"
            dateTime={reply.at}
          >
            {new Date(reply.at).toLocaleTimeString(undefined, {
              hour: "numeric",
              minute: "2-digit",
            })}
          </time>
        ) : null}
      </div>
      <p className="task-attempt-phase-reply-body">{displayText}</p>
      <div className="task-attempt-phase-reply-actions">
        {needsTruncate ? (
          <button
            type="button"
            className="task-attempt-phase-reply-toggle"
            onClick={() => setExpanded((v) => !v)}
          >
            {expanded ? "Show less" : "Show more"}
          </button>
        ) : null}
        {onViewInActivity ? (
          <button
            type="button"
            className="task-attempt-phase-reply-activity"
            onClick={() => onViewInActivity(phaseSeq)}
          >
            View in Activity
          </button>
        ) : null}
      </div>
    </div>
  );
}
