import type { UseMutationResult } from "@tanstack/react-query";
import { FieldRequirementBadge } from "@/shared/FieldLabel";
import { errorMessage } from "@/lib/errorMessage";

type ReplyMutation = UseMutationResult<unknown, unknown, string, unknown>;

export function TaskEventDetailReplyForm(props: {
  draft: string;
  onDraftChange: (value: string) => void;
  saveMutation: ReplyMutation;
}) {
  const { draft, onDraftChange, saveMutation } = props;
  return (
    <div className="task-event-detail-response-block">
      <div className="field-heading-with-req task-event-response-heading-row">
        <h3
          className="task-detail-subheading"
          id="task-event-response-heading"
        >
          Add a message
        </h3>
        <FieldRequirementBadge requirement="required" />
      </div>
      <p className="muted task-event-detail-thread-hint">
        Each send appends to this conversation and appears on the task timeline.
      </p>
      {saveMutation.isError ? (
        <div className="err" role="alert">
          <p>
            {errorMessage(saveMutation.error, "Could not send message.")}
          </p>
        </div>
      ) : null}
      <textarea
        id="task-event-user-response"
        className="task-event-detail-response-field"
        rows={5}
        value={draft}
        onChange={(e) => onDraftChange(e.target.value)}
        disabled={saveMutation.isPending}
        aria-labelledby="task-event-response-heading"
        aria-required="true"
        placeholder="Type a message and send. It is stored on this event and shown on the task timeline."
      />
      <div className="task-event-detail-response-actions">
        <button
          type="button"
          onClick={() => {
            const t = draft.trim();
            if (t) saveMutation.mutate(t);
          }}
          disabled={saveMutation.isPending || !draft.trim()}
        >
          {saveMutation.isPending ? "Sending…" : "Send"}
        </button>
      </div>
    </div>
  );
}
