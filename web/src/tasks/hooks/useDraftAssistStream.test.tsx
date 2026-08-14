import { act, renderHook } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  DRAFT_ASSIST_HEARTBEAT_LOSS_MS,
  useDraftAssistStream,
} from "./useDraftAssistStream";
import {
  draftAssistCancelFrames,
  draftAssistReplayFrames,
  draftAssistSessionFrame,
} from "@/test/handlers/draftAssist";

type NamedListener = (ev: { data?: string }) => void;

class MockEventSource {
  static all: MockEventSource[] = [];
  static latest(): MockEventSource | null {
    return MockEventSource.all[MockEventSource.all.length - 1] ?? null;
  }
  onopen: (() => void) | null = null;
  onmessage: ((ev: { data?: string }) => void) | null = null;
  onerror: (() => void) | null = null;
  readonly listeners = new Map<string, Set<NamedListener>>();
  readyState = 0;
  closed = false;
  close = vi.fn(() => {
    this.closed = true;
  });
  constructor(public url: string) {
    MockEventSource.all.push(this);
  }
  addEventListener(kind: string, cb: NamedListener) {
    let set = this.listeners.get(kind);
    if (!set) {
      set = new Set();
      this.listeners.set(kind, set);
    }
    set.add(cb);
  }
  dispatch(kind: string, payload: unknown) {
    const data = typeof payload === "string" ? payload : JSON.stringify(payload);
    const set = this.listeners.get(kind);
    if (set) for (const cb of set) cb({ data });
  }
}

function latestSource(): MockEventSource {
  const es = MockEventSource.latest();
  if (!es) throw new Error("no EventSource was opened");
  return es;
}

beforeEach(() => {
  MockEventSource.all = [];
  // Fake the watchdog interval and `performance.now()` so gap detection
  // is deterministic; leave setTimeout/microtasks real so React effects
  // and Testing Library helpers keep running normally.
  vi.useFakeTimers({
    toFake: ["setInterval", "clearInterval", "performance"],
  });
  vi.stubGlobal("EventSource", MockEventSource);
});

afterEach(() => {
  vi.unstubAllGlobals();
  vi.useRealTimers();
});

describe("useDraftAssistStream", () => {
  it("stays idle when sessionId is null", () => {
    const { result } = renderHook(() => useDraftAssistStream(null));
    expect(result.current.status).toBe("idle");
    expect(MockEventSource.all).toHaveLength(0);
  });

  it("opens EventSource once sessionId is known and marks live on first frame", () => {
    const { result } = renderHook(() => useDraftAssistStream("sess-1"));
    expect(MockEventSource.all).toHaveLength(1);
    expect(latestSource().url).toBe("/draft-assist/sessions/sess-1/events");

    act(() => {
      latestSource().onopen?.();
    });
    expect(result.current.status).toBe("live");

    act(() => {
      latestSource().dispatch("session", draftAssistSessionFrame("sess-1", 1));
    });
    expect(result.current.events).toHaveLength(1);
    expect(result.current.events[0].kind).toBe("session");
    expect(result.current.lastEventId).toBe(1);
  });

  it("deduplicates events replayed by Last-Event-ID", () => {
    const onEvent = vi.fn();
    const { result } = renderHook(() =>
      useDraftAssistStream("sess-replay", { onEvent }),
    );
    const { initial, replay } = draftAssistReplayFrames("sess-replay");

    act(() => {
      for (const frame of initial) {
        latestSource().dispatch(frame.kind, frame);
      }
    });
    expect(result.current.events).toHaveLength(initial.length);

    // Simulate a browser-managed reconnect: same MockEventSource re-dispatches
    // frames including the last-seen id, then continues past it.
    act(() => {
      for (const frame of replay) {
        latestSource().dispatch(frame.kind, frame);
      }
    });
    // initial: 3 frames (ids 1,2,3); replay: 3,4,5 → dedup keeps 1..5.
    expect(result.current.events.map((e) => e.id)).toEqual([1, 2, 3, 4, 5]);
    expect(onEvent).toHaveBeenCalledTimes(5);
    expect(result.current.lastEventId).toBe(5);
  });

  it("emits the cancel two-frame sequence in order", () => {
    const { result } = renderHook(() => useDraftAssistStream("sess-cancel"));
    act(() => {
      for (const frame of draftAssistCancelFrames) {
        latestSource().dispatch(frame.kind, frame);
      }
    });
    expect(result.current.events).toHaveLength(2);
    expect(result.current.events[0].kind).toBe("status");
    if (result.current.events[0].kind === "status") {
      expect(result.current.events[0].data.status).toBe("cancelling");
    }
    expect(result.current.events[1].kind).toBe("done");
    if (result.current.events[1].kind === "done") {
      expect(result.current.events[1].data.status).toBe("cancelled");
    }
  });

  it("marks disconnected on transport error and reopens after the heartbeat gap", () => {
    const { result } = renderHook(() => useDraftAssistStream("sess-recon"));
    act(() => {
      latestSource().onopen?.();
    });
    act(() => {
      latestSource().dispatch(
        "session",
        draftAssistSessionFrame("sess-recon", 1),
      );
    });
    expect(result.current.status).toBe("live");

    act(() => {
      latestSource().onerror?.();
    });
    expect(result.current.status).toBe("disconnected");

    // Watchdog polls at 1s; advance past the loss threshold to force a reopen.
    // performance.now() is real, so no fresh frames means gap > threshold.
    act(() => {
      vi.advanceTimersByTime(DRAFT_ASSIST_HEARTBEAT_LOSS_MS + 1_500);
    });
    expect(MockEventSource.all.length).toBeGreaterThanOrEqual(2);
    expect(result.current.reconnectCount).toBeGreaterThanOrEqual(1);
    expect(latestSource().url).toBe(
      "/draft-assist/sessions/sess-recon/events",
    );
  });

  it("surfaces parse errors on schema_version mismatch and does not append the frame", () => {
    const { result } = renderHook(() => useDraftAssistStream("sess-bad"));
    act(() => {
      latestSource().dispatch("session", {
        id: 1,
        kind: "session",
        at: "2026-08-14T00:00:00Z",
        data: {
          session_id: "sess-bad",
          snapshot: {},
          schema_version: 999,
        },
      });
    });
    expect(result.current.error?.message ?? "").toMatch(/schema mismatch/);
    expect(result.current.events).toHaveLength(0);
  });

  it("closes the source on unmount", () => {
    const { unmount } = renderHook(() => useDraftAssistStream("sess-unmount"));
    const es = latestSource();
    unmount();
    expect(es.close).toHaveBeenCalled();
  });

  it("does not open a source while enabled is false", () => {
    renderHook(() =>
      useDraftAssistStream("sess-off", { enabled: false }),
    );
    expect(MockEventSource.all).toHaveLength(0);
  });
});
