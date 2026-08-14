import {
  DRAFT_ASSIST_EVENT_KINDS,
  type DraftAssistEvent,
} from "@/types/draftAssist";
import { openDraftAssistEventsSource } from "@/api/draftAssist";
import { parseDraftAssistEvent } from "@/api/parseTaskApiDraftAssist";

/**
 * Server heartbeat cadence (3s while active) plus jitter allowance.
 * 8s is ~3× the server interval — catches a wedged sidecar without
 * flapping on a single dropped frame.
 */
export const DRAFT_ASSIST_HEARTBEAT_LOSS_MS = 8_000;

/** Watchdog poll cadence. Well under the loss threshold. */
export const DRAFT_ASSIST_HEARTBEAT_POLL_MS = 1_000;

export type DraftAssistStreamStatus =
  | "idle"
  | "connecting"
  | "live"
  | "disconnected";

export type DraftAssistStreamCallbacks = {
  onStatus: (status: DraftAssistStreamStatus) => void;
  onEvent: (event: DraftAssistEvent) => void;
  onError: (err: Error) => void;
  onReconnect: () => void;
};

/**
 * Opens an EventSource, wires named-event listeners, dedupes replayed
 * frames by monotonic `id`, and runs the heartbeat-loss watchdog. Kept
 * pure and side-effectful so `useDraftAssistStream` can stay small and
 * so tests can hit both layers independently.
 */
export function connectDraftAssistStream(
  sessionId: string,
  callbacks: DraftAssistStreamCallbacks,
): () => void {
  let cancelled = false;
  let source: EventSource | null = null;
  let lastMessageAt = performance.now();
  let watchdog: ReturnType<typeof setInterval> | null = null;
  const seenIds = new Set<number>();

  const clearWatchdog = () => {
    if (watchdog !== null) {
      clearInterval(watchdog);
      watchdog = null;
    }
  };

  const handleFrame = (raw: string) => {
    if (cancelled) return;
    lastMessageAt = performance.now();
    let parsed: DraftAssistEvent;
    try {
      parsed = parseDraftAssistEvent(JSON.parse(raw));
    } catch (e: unknown) {
      callbacks.onError(e instanceof Error ? e : new Error(String(e)));
      return;
    }
    if (seenIds.has(parsed.id)) {
      return;
    }
    seenIds.add(parsed.id);
    callbacks.onStatus("live");
    callbacks.onEvent(parsed);
  };

  const open = (isReconnect: boolean) => {
    if (cancelled) return;
    clearWatchdog();
    try {
      source = openDraftAssistEventsSource(sessionId);
    } catch (e: unknown) {
      callbacks.onError(e instanceof Error ? e : new Error(String(e)));
      callbacks.onStatus("disconnected");
      return;
    }
    callbacks.onStatus("connecting");
    if (isReconnect) callbacks.onReconnect();
    lastMessageAt = performance.now();

    source.onopen = () => {
      if (cancelled) return;
      lastMessageAt = performance.now();
      callbacks.onStatus("live");
    };
    source.onerror = () => {
      if (cancelled) return;
      // Browser auto-reconnect will fire onopen again with the native
      // Last-Event-ID header; the watchdog handles a stalled retry.
      callbacks.onStatus("disconnected");
    };
    for (const kind of DRAFT_ASSIST_EVENT_KINDS) {
      source.addEventListener(kind, (ev: MessageEvent) => {
        handleFrame(String(ev.data ?? ""));
      });
    }
    watchdog = setInterval(() => {
      if (cancelled) return;
      const gap = performance.now() - lastMessageAt;
      if (gap > DRAFT_ASSIST_HEARTBEAT_LOSS_MS) {
        callbacks.onStatus("disconnected");
        source?.close();
        source = null;
        open(true);
      }
    }, DRAFT_ASSIST_HEARTBEAT_POLL_MS);
  };

  open(false);

  return () => {
    cancelled = true;
    clearWatchdog();
    source?.close();
    source = null;
  };
}
