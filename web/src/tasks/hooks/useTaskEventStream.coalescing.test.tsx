import "./useTaskEventStream.testSetup";
import { act, renderHook } from "@testing-library/react";
import { QueryClient } from "@tanstack/react-query";
import { describe, expect, it, vi } from "vitest";
import { taskQueryKeys } from "../task-query";
import { useTaskEventStream } from "./useTaskEventStream";
import { createWrapper, getCurrentMockES } from "./useTaskEventStream.testSetup";

describe("useTaskEventStream coalescing", () => {
  it("coalesces many progress frames into one persisted stream refresh", () => {
    const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const inv = vi.spyOn(qc, "invalidateQueries");

    renderHook(() => useTaskEventStream(), {
      wrapper: createWrapper(qc),
    });
    const mockES = getCurrentMockES();
    act(() => {
      mockES!.onopen?.();
    });

    for (let i = 0; i < 4; i++) {
      act(() => {
        mockES!.onmessage?.({
          data: '{"type":"agent_run_progress","id":"task-1","cycle_id":"cyc-1","phase_seq":2,"progress":{"kind":"tool_call","subtype":"started","tool":"ReadFile","message":"Started ReadFile"}}',
        });
      });
      act(() => {
        vi.advanceTimersByTime(1000);
      });
    }
    act(() => {
      vi.advanceTimersByTime(5000);
    });

    const streamCalls = inv.mock.calls
      .map((c) => (c[0] as { queryKey: readonly unknown[] }).queryKey)
      .filter(
        (key) =>
          JSON.stringify(key) ===
          JSON.stringify(taskQueryKeys.cycleStream("task-1", "cyc-1")),
      );
    expect(streamCalls).toHaveLength(1);
  });

  it("ignores unrecognised frames without scheduling invalidation", () => {
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
      mockES!.onmessage?.({ data: "{}" });
    });
    act(() => {
      vi.advanceTimersByTime(950);
    });
    expect(inv).not.toHaveBeenCalled();
  });

  it("still invalidates cycles and checklist when task_updated shares a flush with cycle hints", () => {
    // Task pending owns checklist; cycle pending owns cycles. Both must
    // survive enrichment skipping detailRoot after verify success.
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
      mockES!.onmessage?.({
        data: '{"type":"task_updated","id":"task-1"}',
      });
      mockES!.onmessage?.({
        data: '{"type":"task_cycle_changed","id":"task-1","cycle_id":"cyc-1"}',
      });
    });
    act(() => {
      vi.advanceTimersByTime(950);
    });
    const calls = inv.mock.calls.map((c) => c[0]);
    expect(calls).toContainEqual({
      queryKey: ["tasks", "detail"],
    });
    expect(calls).toContainEqual({
      queryKey: taskQueryKeys.cycles("task-1"),
    });
    expect(calls).toContainEqual({
      queryKey: taskQueryKeys.checklist("task-1"),
    });
  });

  it("coalesces a burst of agent worker cycle frames into a single flush", () => {
    // Regression: in production the agent worker emits ~6 task_cycle_changed
    // frames per task run, ~1s apart. A short trailing debounce never
    // batched them and each frame fired its own refetch storm. With the
    // new ~900ms window plus maxWait, frames arriving every ~700ms should
    // collapse into ONE flush (one cycles invalidation per task) instead
    // of six.
    const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const inv = vi.spyOn(qc, "invalidateQueries");

    renderHook(() => useTaskEventStream(), {
      wrapper: createWrapper(qc),
    });
    const mockES = getCurrentMockES();
    act(() => {
      mockES!.onopen?.();
    });

    // Emit 4 cycle frames 700ms apart — each one resets the debounce so
    // nothing flushes mid-burst.
    for (let i = 0; i < 4; i++) {
      act(() => {
        mockES!.onmessage?.({
          data: '{"type":"task_cycle_changed","id":"task-burst","cycle_id":"cyc-1"}',
        });
      });
      act(() => {
        vi.advanceTimersByTime(700);
      });
    }
    // Drain the trailing debounce.
    act(() => {
      vi.advanceTimersByTime(950);
    });

    const cycleCalls = inv.mock.calls
      .map((c) => (c[0] as { queryKey: readonly unknown[] }).queryKey)
      .filter(
        (k) =>
          JSON.stringify(k) === JSON.stringify(taskQueryKeys.cycles("task-burst")),
      );
    expect(cycleCalls).toHaveLength(1);
    const detailCalls = inv.mock.calls
      .map((c) => (c[0] as { queryKey: readonly unknown[] }).queryKey)
      .filter((k) => JSON.stringify(k) === JSON.stringify(["tasks", "detail"]));
    expect(detailCalls).toHaveLength(0);
  });

  it("forces a flush at maxWait so a continuous SSE stream cannot starve the UI", () => {
    // The trailing debounce alone could be reset forever by frames
    // arriving inside the coalesce window (e.g. multiple concurrent
    // tasks). The maxWait safety valve must force a flush so the open
    // task page still receives cycle updates under sustained load.
    const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const inv = vi.spyOn(qc, "invalidateQueries");

    renderHook(() => useTaskEventStream(), {
      wrapper: createWrapper(qc),
    });
    const mockES = getCurrentMockES();
    act(() => {
      mockES!.onopen?.();
    });

    // 12 frames every 300ms = 3.6s of continuous activity, all *inside*
    // the 900ms coalesce window — a naive debounce would never flush.
    for (let i = 0; i < 12; i++) {
      act(() => {
        mockES!.onmessage?.({
          data: '{"type":"task_cycle_changed","id":"task-stream","cycle_id":"cyc-1"}',
        });
      });
      act(() => {
        vi.advanceTimersByTime(300);
      });
    }

    const cycleCalls = inv.mock.calls
      .map((c) => (c[0] as { queryKey: readonly unknown[] }).queryKey)
      .filter(
        (k) =>
          JSON.stringify(k) ===
          JSON.stringify(taskQueryKeys.cycles("task-stream")),
      );
    expect(cycleCalls.length).toBeGreaterThanOrEqual(1);
  });

});
