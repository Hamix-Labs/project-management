import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  __resetMutationGuardForTests,
  shouldSuppressTaskMutationEcho,
} from "@/tasks/sync/mutationGuard";
import { runGuardedTaskMutation } from "./runGuardedMutation";

vi.mock("@/observability", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/observability")>();
  return {
    ...actual,
    rumMutationStarted: vi.fn(),
    rumMutationOptimisticApplied: vi.fn(),
  };
});

import {
  rumMutationOptimisticApplied,
  rumMutationStarted,
} from "@/observability";

describe("runGuardedTaskMutation", () => {
  beforeEach(() => {
    __resetMutationGuardForTests();
    vi.mocked(rumMutationStarted).mockClear();
    vi.mocked(rumMutationOptimisticApplied).mockClear();
  });

  afterEach(() => {
    __resetMutationGuardForTests();
  });

  it("applies optimistic work, runs the mutation, then ends the guard", async () => {
    const applyOptimistic = vi.fn();
    const run = vi.fn().mockResolvedValue("ok");

    const result = await runGuardedTaskMutation({
      taskId: "t1",
      optimisticEnabled: true,
      rumKind: "task_patch",
      applyOptimistic,
      run,
    });

    expect(applyOptimistic).toHaveBeenCalledTimes(1);
    expect(run).toHaveBeenCalledTimes(1);
    expect(result.value).toBe("ok");
    expect(result.guard.guarded).toBe(true);
    expect(rumMutationStarted).toHaveBeenCalledWith("task_patch");
    expect(rumMutationOptimisticApplied).toHaveBeenCalledWith(
      "task_patch",
      expect.any(Number),
    );
    expect(shouldSuppressTaskMutationEcho("t1")).toBe(false);
  });

  it("still ends the guard when run throws", async () => {
    const applyOptimistic = vi.fn();
    const run = vi.fn().mockRejectedValue(new Error("boom"));

    await expect(
      runGuardedTaskMutation({
        taskId: "t1",
        optimisticEnabled: true,
        rumKind: "task_close",
        applyOptimistic,
        run,
      }),
    ).rejects.toThrow("boom");

    expect(applyOptimistic).toHaveBeenCalledTimes(1);
    expect(shouldSuppressTaskMutationEcho("t1")).toBe(false);
  });

  it("skips optimistic apply and guard when optimisticEnabled is false", async () => {
    const applyOptimistic = vi.fn();
    const run = vi.fn().mockResolvedValue(42);

    const result = await runGuardedTaskMutation({
      taskId: "t1",
      optimisticEnabled: false,
      rumKind: "task_patch",
      applyOptimistic,
      run,
    });

    expect(applyOptimistic).not.toHaveBeenCalled();
    expect(rumMutationOptimisticApplied).not.toHaveBeenCalled();
    expect(result.guard.guarded).toBe(false);
    expect(result.value).toBe(42);
    expect(shouldSuppressTaskMutationEcho("t1")).toBe(false);
  });
});
