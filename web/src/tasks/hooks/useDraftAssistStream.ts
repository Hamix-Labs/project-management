import { useEffect, useRef, useState } from "react";
import type { DraftAssistEvent } from "@/types/draftAssist";
import {
  DRAFT_ASSIST_HEARTBEAT_LOSS_MS,
  connectDraftAssistStream,
  type DraftAssistStreamStatus,
} from "./draftAssistStreamConnection";

export {
  DRAFT_ASSIST_HEARTBEAT_LOSS_MS,
  type DraftAssistStreamStatus,
} from "./draftAssistStreamConnection";

export type UseDraftAssistStreamOptions = {
  /** When false the hook does not open a stream even if `sessionId` is set. */
  enabled?: boolean;
  /**
   * Called with each parsed event in wire order (deduplicated by
   * `event.id`). Callers should not mutate the argument.
   */
  onEvent?: (event: DraftAssistEvent) => void;
};

export type UseDraftAssistStreamReturn = {
  events: DraftAssistEvent[];
  status: DraftAssistStreamStatus;
  /** Last SSE `id:` seen. */
  lastEventId: number;
  /** Number of watchdog-driven or error-driven reopens observed. */
  reconnectCount: number;
  /** Last parse/transport error, cleared on the next successful frame. */
  error: Error | null;
};

/**
 * Opens `GET /draft-assist/sessions/{id}/events` once `sessionId` is
 * known and surfaces parsed frames. Transport, watchdog, and
 * deduplication live in {@link connectDraftAssistStream}; this hook
 * just wires that state into React.
 *
 * The browser handles Last-Event-ID replay natively on transport
 * errors. When the {@link DRAFT_ASSIST_HEARTBEAT_LOSS_MS} watchdog
 * fires the source is force-reopened so the server replays any
 * missed events from its ring buffer.
 *
 * Unmount closes the EventSource and clears watchdog timers.
 */
export function useDraftAssistStream(
  sessionId: string | null | undefined,
  options?: UseDraftAssistStreamOptions,
): UseDraftAssistStreamReturn {
  const enabled = options?.enabled ?? true;
  const onEvent = options?.onEvent;

  const [events, setEvents] = useState<DraftAssistEvent[]>([]);
  const [status, setStatus] = useState<DraftAssistStreamStatus>("idle");
  const [lastEventId, setLastEventId] = useState<number>(0);
  const [reconnectCount, setReconnectCount] = useState<number>(0);
  const [error, setError] = useState<Error | null>(null);

  const onEventRef = useRef<typeof onEvent>(onEvent);
  useEffect(() => {
    onEventRef.current = onEvent;
  }, [onEvent]);

  useEffect(() => {
    if (!enabled) {
      setStatus("idle");
      return undefined;
    }
    const id = sessionId?.trim();
    if (!id) {
      setStatus("idle");
      return undefined;
    }
    return connectDraftAssistStream(id, {
      onStatus: setStatus,
      onEvent: (ev) => {
        setLastEventId((prev) => (ev.id > prev ? ev.id : prev));
        setError(null);
        setEvents((prev) => [...prev, ev]);
        onEventRef.current?.(ev);
      },
      onError: setError,
      onReconnect: () => setReconnectCount((n) => n + 1),
    });
  }, [enabled, sessionId]);

  useEffect(() => {
    setEvents([]);
    setLastEventId(0);
    setReconnectCount(0);
    setError(null);
  }, [sessionId, enabled]);

  return { events, status, lastEventId, reconnectCount, error };
}
