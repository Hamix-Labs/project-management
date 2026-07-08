import { QueryClient } from "@tanstack/react-query";
import { describe, expect, it, vi } from "vitest";
import { makeTask } from "@/test/taskDefaults";
import { taskQueryKeys } from "../task-query";
import { applySyncEffects } from "./applySyncEffects";

function newTestQueryClient(): QueryClient {
  return new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
}

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
});
