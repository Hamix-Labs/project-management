import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { Link, useParams } from "react-router-dom";
import { getTaskEvent } from "@/api";
import { useDocumentTitle } from "@/shared/useDocumentTitle";
import { errorMessage } from "@/lib/errorMessage";
import { useRolloutFlags } from "@/settings";
import { buildPatchTaskEventUserResponseMutationOptions } from "@/tasks/mutations/patchTaskEventUserResponseMutation";
import {
  awaitingUserReply,
  eventTypeLabel,
  eventTypeNeedsUserInput,
} from "../task-events";
import { TaskEventDataPanel } from "../components/task-event-detail/TaskEventDataPanel";
import { TaskEventDetailLoadedHeader } from "../components/task-event-detail/TaskEventDetailLoadedHeader";
import { TaskEventDetailReplyForm } from "../components/task-event-detail/TaskEventDetailReplyForm";
import { TaskEventDetailThread } from "../components/task-event-detail/TaskEventDetailThread";
import { TaskEventDetailSkeleton } from "../components/skeletons";
import { taskQueryKeys } from "../task-query";

export function TaskEventDetailPage() {
  const qc = useQueryClient();
  const { optimisticMutationsEnabled } = useRolloutFlags();
  const { taskId = "", eventSeq: eventSeqParam = "" } = useParams<{
    taskId: string;
    eventSeq: string;
  }>();
  const seqLooksLikePositiveInt = /^[1-9]\d*$/.test(eventSeqParam);
  const eventSeq = seqLooksLikePositiveInt
    ? Number.parseInt(eventSeqParam, 10)
    : Number.NaN;
  const seqValid = Number.isSafeInteger(eventSeq) && eventSeq >= 1;

  const [draft, setDraft] = useState("");

  const q = useQuery({
    queryKey: taskQueryKeys.eventDetail(taskId, eventSeq),
    queryFn: ({ signal }) => getTaskEvent(taskId, eventSeq, { signal }),
    enabled: Boolean(taskId) && seqValid,
  });

  const saveMutation = useMutation(
    buildPatchTaskEventUserResponseMutationOptions({
      taskId,
      eventSeq,
      queryClient: qc,
      optimisticMutationsEnabled,
      onDraftCleared: () => setDraft(""),
    }),
  );

  const eventDocPageTitle = (() => {
    if (!taskId) return undefined;
    if (!seqValid) return "Invalid event";
    if (q.isSuccess && q.data) {
      return `Event #${q.data.seq}: ${eventTypeLabel(q.data.type)}`;
    }
    return undefined;
  })();
  useDocumentTitle(eventDocPageTitle);

  if (!taskId) {
    return (
      <p className="muted" role="status">
        Missing task id.
      </p>
    );
  }

  if (!seqValid) {
    return (
      <section className="panel task-detail-panel task-detail-content--enter">
        <div className="err" role="alert">
          <p>Invalid event sequence in the URL.</p>
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

  if (q.isPending) {
    return <TaskEventDetailSkeleton />;
  }

  if (q.isError) {
    return (
      <section className="panel task-detail-panel task-detail-content--enter">
        <div className="err" role="alert">
          <p>{errorMessage(q.error, "Could not load event.")}</p>
          <div className="task-detail-error-actions">
            <button
              type="button"
              className="secondary"
              onClick={() => void q.refetch()}
            >
              Try again
            </button>
            <Link
              to={`/tasks/${encodeURIComponent(taskId)}`}
              className="pd__back project-context-back-link"
            >
              <span aria-hidden="true">&#8249;</span>
              Back to task
            </Link>
            <Link to="/" className="pd__back project-context-back-link">
              <span aria-hidden="true">&#8249;</span>
              All tasks
            </Link>
          </div>
        </div>
      </section>
    );
  }

  const ev = q.data;
  const needsInput = eventTypeNeedsUserInput(ev.type);
  const awaitingUser = needsInput && awaitingUserReply(ev);

  return (
    <section className="panel task-detail-panel task-event-detail-panel task-detail-content--enter">
      <TaskEventDetailLoadedHeader
        taskId={taskId}
        event={ev}
        needsInput={needsInput}
        awaitingUser={awaitingUser}
      />
      {needsInput ? (
        <TaskEventDetailReplyForm
          draft={draft}
          onDraftChange={setDraft}
          saveMutation={saveMutation}
        />
      ) : null}
      <TaskEventDataPanel event={ev} />
      {ev.response_thread ? (
        <TaskEventDetailThread thread={ev.response_thread} />
      ) : null}
    </section>
  );
}
