import "./useTaskEventStream.testSetup";
import { act, renderHook } from "@testing-library/react";
import { QueryClient } from "@tanstack/react-query";
import { describe, expect, it, vi } from "vitest";
import { taskQueryKeys } from "../task-query";
import { useTaskEventStream } from "./useTaskEventStream";
import { createWrapper, getCurrentMockES } from "./useTaskEventStream.testSetup";

describe("useTaskEventStream invalidation", () => {
  it("task_event_changed invalidates events queries only without detail or list refetch", () => {
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
        data: '{"type":"task_event_changed","id":"task-events-only","event_seq":7}',
      });
    });
    // Immediate invalidation — no debounce window.
    const calls = inv.mock.calls.map(
      (c) => (c[0] as { queryKey: readonly unknown[] }).queryKey,
    );
    expect(calls).toContainEqual(taskQueryKeys.eventsRoot("task-events-only"));
    expect(calls).toContainEqual(
      taskQueryKeys.eventDetail("task-events-only", 7),
    );
    for (const key of calls) {
      expect(key).not.toEqual(["tasks", "list"]);
      expect(key).not.toEqual(["tasks", "detail"]);
    }
  });

  it("debounced SSE message triggers query invalidation after delay", () => {
    const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const inv = vi.spyOn(qc, "invalidateQueries");

    renderHook(() => useTaskEventStream(), {
      wrapper: createWrapper(qc),
    });

    const mockES = getCurrentMockES();
    expect(mockES).not.toBeNull();
    act(() => {
      mockES!.onopen?.();
    });
    act(() => {
      mockES!.onmessage?.({
        data: '{"type":"task_updated","id":"11111111-1111-4111-8111-111111111111"}',
      });
    });
    expect(inv).not.toHaveBeenCalled();

    // Halfway through the coalesce window: still nothing, the agent
    // worker emits frames ~1s apart so the trailing debounce must
    // outlast a single inter-frame gap to actually batch them.
    act(() => {
      vi.advanceTimersByTime(450);
    });
    expect(inv).not.toHaveBeenCalled();

    act(() => {
      vi.advanceTimersByTime(500);
    });
    expect(inv).toHaveBeenCalled();
  });

  it("invalidates cycles on task_cycle_changed without checklist or detailRoot", () => {
    // Cycle frames own the phase ledger only. Checklist completions and
    // task-row status publish task_updated (ADR-0022); coupling cycle
    // hints to checklist caused premature partial criteria refetches.
    const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const inv = vi.spyOn(qc, "invalidateQueries");

    renderHook(() => useTaskEventStream(), {
      wrapper: createWrapper(qc),
    });

    const mockES = getCurrentMockES();
    expect(mockES).not.toBeNull();
    act(() => {
      mockES!.onopen?.();
    });
    act(() => {
      mockES!.onmessage?.({
        data: '{"type":"task_cycle_changed","id":"task-1","cycle_id":"cyc-1"}',
      });
    });
    act(() => {
      vi.advanceTimersByTime(950);
    });

    const calls = inv.mock.calls.map((c) => c[0]);
    expect(calls).toContainEqual({ queryKey: ["tasks", "list"] });
    expect(calls).not.toContainEqual({
      queryKey: ["tasks", "detail"],
    });
    expect(calls).toContainEqual({
      queryKey: taskQueryKeys.commits("task-1"),
    });
    expect(calls).toContainEqual({
      queryKey: taskQueryKeys.cycles("task-1"),
    });
    expect(calls).not.toContainEqual({
      queryKey: taskQueryKeys.checklist("task-1"),
    });
  });

  it("invalidates the task-stats query on task/cycle frames", () => {
    // Regression: taskQueryKeys.stats() lives outside the taskQueryKeys.all tree.
    // SSE used to invalidate only listRoot + detail, so list rows updated
    // but aggregated stats stayed frozen until a manual mutation or hard
    // refresh.
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
        data: '{"type":"task_cycle_changed","id":"task-stats-1","cycle_id":"cyc-1"}',
      });
    });
    act(() => {
      vi.advanceTimersByTime(950);
    });

    const calls = inv.mock.calls.map((c) => c[0]);
    expect(calls).toContainEqual({ queryKey: taskQueryKeys.stats() });
    expect(calls).toContainEqual({
      queryKey: taskQueryKeys.cycleFailuresRoot(),
    });
  });

  it("refreshes the persisted cycle stream on a slower cadence than live progress", () => {
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
        data: '{"type":"agent_run_progress","id":"task-1","cycle_id":"cyc-1","phase_seq":2,"progress":{"kind":"tool_call","subtype":"started","tool":"ReadFile","message":"Started ReadFile"}}',
      });
    });
    act(() => {
      vi.advanceTimersByTime(3000);
    });
    expect(inv).not.toHaveBeenCalled();

    act(() => {
      vi.advanceTimersByTime(2000);
    });

    const calls = inv.mock.calls.map((c) => c[0]);
    expect(calls).toContainEqual({
      queryKey: taskQueryKeys.cycleStream("task-1", "cyc-1"),
    });
    for (const arg of calls) {
      const key = (arg as { queryKey: readonly unknown[] }).queryKey;
      expect(key).not.toEqual(["tasks"]);
      expect(key).not.toEqual(["tasks", "detail"]);
    }
  });

});
