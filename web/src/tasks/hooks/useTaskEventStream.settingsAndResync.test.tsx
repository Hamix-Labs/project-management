import "./useTaskEventStream.testSetup";
import { act, renderHook } from "@testing-library/react";
import { QueryClient } from "@tanstack/react-query";
import { describe, expect, it, vi } from "vitest";
import { taskQueryKeys } from "../task-query";
import { useTaskEventStream } from "./useTaskEventStream";
import { createWrapper, getCurrentMockES } from "./useTaskEventStream.testSetup";

describe("useTaskEventStream settings and resync", () => {
  it("settings_changed invalidates only the settings cache and never the task tree", () => {
    // Documented in docs/api.md: settings/cancel frames must "invalidate
    // only the settings cache slot ... without disturbing task caches; they
    // bypass the debounce batch". Regression: previously the trailing
    // debounce was armed for *every* frame, so settings/cancel frames
    // (which add nothing to pendingRef) fell through to the broad-fallback
    // branch in flushStreamInvalidation and refetched every active task
    // query SSE_INVALIDATE_WINDOW_MS later. The fix gates the timer on
    // whether there is anything pending to flush.
    const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const inv = vi.spyOn(qc, "invalidateQueries");

    renderHook(() => useTaskEventStream(), {
      wrapper: createWrapper(qc),
    });
    const mockES = getCurrentMockES();
    act(() => {
      mockES!.onopen?.();
    });
    act(() => {
      mockES!.onmessage?.({ data: '{"type":"settings_changed"}' });
    });
    // Drain well past the debounce window + maxWait safety valve so any
    // timer-triggered fallback would have fired.
    act(() => {
      vi.advanceTimersByTime(3000);
    });

    const calls = inv.mock.calls.map(
      (c) => (c[0] as { queryKey: readonly unknown[] }).queryKey,
    );
    expect(calls).toContainEqual(["settings", "app"]);
    for (const key of calls) {
      expect(key).not.toEqual(["tasks"]);
      expect((key as readonly unknown[])[0]).not.toBe("tasks");
    }
  });

  it("agent_run_cancelled invalidates only the settings cache and never the task tree", () => {
    // Same contract as settings_changed: id-less notification, must not
    // disturb task caches. See docs/api.md.
    const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const inv = vi.spyOn(qc, "invalidateQueries");

    renderHook(() => useTaskEventStream(), {
      wrapper: createWrapper(qc),
    });
    const mockES = getCurrentMockES();
    act(() => {
      mockES!.onopen?.();
    });
    act(() => {
      mockES!.onmessage?.({ data: '{"type":"agent_run_cancelled"}' });
    });
    act(() => {
      vi.advanceTimersByTime(3000);
    });

    const calls = inv.mock.calls.map(
      (c) => (c[0] as { queryKey: readonly unknown[] }).queryKey,
    );
    expect(calls).toContainEqual(["settings", "app"]);
    for (const key of calls) {
      expect(key).not.toEqual(["tasks"]);
      expect((key as readonly unknown[])[0]).not.toBe("tasks");
    }
  });

  it("a settings frame followed by a task frame still flushes the task invalidation", () => {
    // Guard against an over-eager fix that disables the debounce timer
    // entirely on settings frames: a *subsequent* task/cycle frame in the
    // same batch must still flush the broader task invalidation after the
    // coalesce window. Without this pin, gating the timer on "nothing
    // pending" could regress real task-driven invalidation if the order
    // of frames matters.
    const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const inv = vi.spyOn(qc, "invalidateQueries");

    renderHook(() => useTaskEventStream(), {
      wrapper: createWrapper(qc),
    });
    const mockES = getCurrentMockES();
    act(() => {
      mockES!.onopen?.();
    });
    act(() => {
      mockES!.onmessage?.({ data: '{"type":"settings_changed"}' });
      mockES!.onmessage?.({
        data: '{"type":"task_updated","id":"task-mix"}',
      });
    });
    act(() => {
      vi.advanceTimersByTime(950);
    });

    const calls = inv.mock.calls.map(
      (c) => (c[0] as { queryKey: readonly unknown[] }).queryKey,
    );
    expect(calls).toContainEqual(["settings", "app"]);
    expect(calls).toContainEqual(["tasks", "list"]);
    expect(calls).toContainEqual(["tasks", "detail"]);
  });

  it("resync directive forces an immediate broad invalidation across tasks/stats/settings", () => {
    // Phase 2 of the realtime smoothness plan: the hub emits a
    // `{"type":"resync"}` directive whenever it cannot bridge a
    // reconnect gap (Last-Event-ID outside the ring buffer) or it
    // had to evict a slow consumer. The client must drop every
    // cached query and refetch from REST — no debounce, no
    // coalescing, because we just demonstrated we can't trust the
    // delta stream. Pinning broad invalidation across all three
    // cache slots (tasks tree, task-stats, settings)
    // guards against a future "optimisation" that would handle
    // resync as a regular settings frame and silently leave the
    // task tree stale.
    const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const inv = vi.spyOn(qc, "invalidateQueries");

    renderHook(() => useTaskEventStream(), {
      wrapper: createWrapper(qc),
    });
    const mockES = getCurrentMockES();
    act(() => {
      mockES!.onopen?.();
    });
    // Queue a regular task frame first — the resync MUST cancel
    // the pending debounce and refetch broadly instead of letting
    // the per-task invalidation fire on its own delayed timer.
    act(() => {
      mockES!.onmessage?.({
        data: '{"type":"task_updated","id":"task-pending"}',
      });
      mockES!.onmessage?.({ data: '{"type":"resync"}' });
    });
    // Resync handling is synchronous; no need to advance timers.
    const calls = inv.mock.calls.map(
      (c) => (c[0] as { queryKey: readonly unknown[] }).queryKey,
    );
    expect(calls).toContainEqual(["tasks"]);
    expect(calls).toContainEqual(taskQueryKeys.stats());
    expect(calls).toContainEqual(taskQueryKeys.cycleFailuresRoot());
    expect(calls).toContainEqual(["settings", "app"]);

    // Even after letting any pending debounce drain, the broad
    // refetch should NOT fire a second time — the pending frame
    // was cleared by the resync handler.
    const beforeDrainCount = inv.mock.calls.length;
    act(() => {
      vi.advanceTimersByTime(3000);
    });
    expect(inv.mock.calls.length).toBe(beforeDrainCount);
  });

  it("does not invalidate queries after unmount before debounce elapses", () => {
    const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const inv = vi.spyOn(qc, "invalidateQueries");

    const { unmount } = renderHook(() => useTaskEventStream(), {
      wrapper: createWrapper(qc),
    });

    const mockES = getCurrentMockES();
    expect(mockES).not.toBeNull();
    act(() => {
      mockES!.onopen?.();
    });
    act(() => {
      mockES!.onmessage?.({ data: "{}" });
    });
    act(() => {
      vi.advanceTimersByTime(100);
    });
    unmount();
    act(() => {
      vi.advanceTimersByTime(400);
    });
    expect(inv).not.toHaveBeenCalled();
  });
});
