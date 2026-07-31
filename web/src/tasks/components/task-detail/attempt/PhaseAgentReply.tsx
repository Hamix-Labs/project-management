import { useId, useState } from "react";
import { Modal } from "@/shared/Modal";
import type { PhaseStatus } from "@/types";
import type { PhaseAgentReply as PhaseAgentReplyData } from "./latestAgentReplyByPhase";

const DISPLAY_CAP = 800;

type PhaseAgentReplyProps = {
  reply: PhaseAgentReplyData;
  phaseLabel: string;
  phaseStatus: PhaseStatus;
};

export function PhaseAgentReply({
  reply,
  phaseLabel,
  phaseStatus,
}: PhaseAgentReplyProps) {
  const [expanded, setExpanded] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const titleId = useId();
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
        <button
          type="button"
          className="task-attempt-phase-reply-open"
          onClick={() => setModalOpen(true)}
        >
          View reply
        </button>
      </div>
      {modalOpen ? (
        <Modal
          onClose={() => setModalOpen(false)}
          labelledBy={titleId}
          size="wide"
        >
          <section className="panel modal-sheet task-attempt-phase-reply-modal">
            <header className="task-attempt-phase-reply-modal-head">
              <p className="task-attempt-phase-reply-modal-eyebrow">
                Agent reply
              </p>
              <h2 id={titleId} className="task-attempt-phase-reply-modal-title">
                {phaseLabel}
              </h2>
              {reply.at ? (
                <time
                  className="task-attempt-phase-reply-modal-time"
                  dateTime={reply.at}
                >
                  {new Date(reply.at).toLocaleString(undefined, {
                    dateStyle: "medium",
                    timeStyle: "short",
                  })}
                </time>
              ) : null}
            </header>
            <pre className="task-attempt-phase-reply-modal-body">
              {reply.text}
            </pre>
            <div className="row stack-row-actions task-attempt-phase-reply-modal-footer">
              <button
                type="button"
                className="secondary"
                onClick={() => setModalOpen(false)}
              >
                Close
              </button>
            </div>
          </section>
        </Modal>
      ) : null}
    </div>
  );
}
