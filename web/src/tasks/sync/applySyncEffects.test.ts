// @vitest-environment jsdom
import { QueryClient } from "@tanstack/react-query";
import { renderHook } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { makeTask } from "@/test/taskDefaults";
import { taskQueryKeys } from "../task-query";
import { applySyncEffects } from "./applySyncEffects";
import { useAgentRunProgress } from "../hooks/useAgentRunProgress";

function newTestQueryClient(): QueryClient {
  return new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
}

const validCycleDetail = {
  id: "cyc-1",
  task_id: "task-1",
  attempt_seq: 1,
  status: "running",
  started_at: "2026-04-18T10:00:00.000Z",
  triggered_by: "user",
  meta: { source: "manual" },
  phases: [
    {
      id: "ph-1",
      cycle_id: "cyc-1",
      phase: "execute",
      phase_seq: 1,
      status: "running",
      started_at: "2026-04-18T10:00:01.000Z",
      details: {},
    },
  ],
};

describe("applySyncEffects patch_task_detail", () => {
  it("invalidates list and stats immediately for terminal enriched task", () => {
    const queryClient = newTestQueryClient();
    const invalidateSpy = vi.spyOn(queryClient, "invalidateQueries");
    const setDataSpy = vi.spyOn(queryClient, "setQueryData");
    const task = makeTask({ id: "task-done", status: "done" });

    applySyncEffects(queryClient, [
      {
        kind: "patch_task_detail",
        taskId: "task-done",
        data: task,
      },
    ]);

    expect(setDataSpy).toHaveBeenCalledWith(
      taskQueryKeys.detail("task-done"),
      expect.objectContaining({ id: "task-done", status: "done" }),
    );
    expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: taskQueryKeys.listRoot() });
    expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: taskQueryKeys.stats() });
  });

  it("does not invalidate list for non-terminal enriched task", () => {
    const queryClient = newTestQueryClient();
    const invalidateSpy = vi.spyOn(queryClient, "invalidateQueries");
    const task = makeTask({ id: "task-running", status: "running" });

    applySyncEffects(queryClient, [
      {
        kind: "patch_task_detail",
        taskId: "task-running",
        data: task,
      },
    ]);

    expect(invalidateSpy).not.toHaveBeenCalled();
  });

  it("skips setQueryData when task payload fails to parse", () => {
    const queryClient = newTestQueryClient();
    const setDataSpy = vi.spyOn(queryClient, "setQueryData");
    const invalidateSpy = vi.spyOn(queryClient, "invalidateQueries");

    const marks = applySyncEffects(queryClient, [
      {
        kind: "patch_task_detail",
        taskId: "bad",
        data: { not: "a task" },
      },
    ]);

    expect(setDataSpy).not.toHaveBeenCalled();
    expect(invalidateSpy).not.toHaveBeenCalled();
    expect(marks.markTaskEnriched).toBeUndefined();
  });
});

describe("applySyncEffects other kinds", () => {
  it("invalidates the provided query key", () => {
    const queryClient = newTestQueryClient();
    const invalidateSpy = vi.spyOn(queryClient, "invalidateQueries");
    const key = taskQueryKeys.stats();

    applySyncEffects(queryClient, [{ kind: "invalidate", queryKey: key }]);

    expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: key });
  });

  it("busts query persist on rum_sse_resync", () => {
    sessionStorage.setItem("hamix:react-query", '{"client":true}');
    const queryClient = newTestQueryClient();

    applySyncEffects(queryClient, [{ kind: "rum_sse_resync" }]);

    expect(sessionStorage.getItem("hamix:react-query")).toBeNull();
  });

  it("writes parsed cycle detail into the cycle cache", () => {
    const queryClient = newTestQueryClient();
    const marks = applySyncEffects(queryClient, [
      {
        kind: "patch_cycle_detail",
        taskId: "task-1",
        cycleId: "cyc-1",
        data: validCycleDetail,
      },
    ]);

    expect(queryClient.getQueryData(taskQueryKeys.cycle("task-1", "cyc-1"))).toEqual(
      expect.objectContaining({
        id: "cyc-1",
        task_id: "task-1",
        phases: expect.any(Array),
      }),
    );
    expect(marks.markCycleEnriched).toEqual({
      taskId: "task-1",
      cycleId: "cyc-1",
    });
  });

  it("skips cycle cache write when cycle payload fails to parse", () => {
    const queryClient = newTestQueryClient();
    const setDataSpy = vi.spyOn(queryClient, "setQueryData");

    const marks = applySyncEffects(queryClient, [
      {
        kind: "patch_cycle_detail",
        taskId: "task-1",
        cycleId: "cyc-1",
        data: { id: "cyc-1" },
      },
    ]);

    expect(setDataSpy).not.toHaveBeenCalled();
    expect(marks.markCycleEnriched).toBeUndefined();
  });

  it("forwards agent run progress payloads", () => {
    const queryClient = newTestQueryClient();
    const payload = {
      taskId: "task-1",
      cycleId: "cyc-1",
      phaseSeq: 1,
      progress: { kind: "tool_call", message: "Reading" },
    };

    applySyncEffects(queryClient, [
      { kind: "push_agent_run_progress", payload },
    ]);

    const { result } = renderHook(() =>
      useAgentRunProgress("task-1", "cyc-1", 1),
    );
    expect(result.current).toEqual(
      expect.arrayContaining([
        expect.objectContaining({
          taskId: "task-1",
          cycleId: "cyc-1",
          phaseSeq: 1,
          progress: { kind: "tool_call", message: "Reading" },
        }),
      ]),
    );
  });
});
