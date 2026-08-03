import { act, renderHook, waitFor } from "@testing-library/react";
import { HttpResponse } from "msw";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { useTaskCloseFlow } from "./useTaskCloseFlow";
import { taskQueryKeys } from "../task-query";
import {
  __resetMutationGuardForTests,
  shouldSuppressTaskMutationEcho,
} from "@/tasks/sync/mutationGuard";
import type { TaskListResponse } from "@/types";
import { makeMutationTestWrapper } from "@/test/reactQuery";
import { makeTask } from "@/test/taskDefaults";
import {
  taskClose,
  taskCloseError,
  taskClosePending,
  taskReopen,
  taskReopenError,
} from "@/test/handlers/tasks";
import { ensureMswListening } from "@/test/mswLifecycle";
import { server } from "@/test/server";

ensureMswListening();

describe("useTaskCloseFlow", () => {
  beforeEach(() => {
    __resetMutationGuardForTests();
  });
  afterEach(() => {
    __resetMutationGuardForTests();
    vi.restoreAllMocks();
  });

  it("starts with no target, no pending, no success, no error", () => {
    const { Wrapper } = makeMutationTestWrapper();
    const { result } = renderHook(() => useTaskCloseFlow(), {
      wrapper: Wrapper,
    });
    expect(result.current.closeTarget).toBeNull();
    expect(result.current.closePending).toBe(false);
    expect(result.current.closeSuccess).toBe(false);
    expect(result.current.closeError).toBeNull();
    expect(result.current.closeVariables).toBeUndefined();
  });

  it("requestClose captures id, title, and number", () => {
    const { Wrapper } = makeMutationTestWrapper();
    const { result } = renderHook(() => useTaskCloseFlow(), {
      wrapper: Wrapper,
    });
    act(() => {
      result.current.requestClose({ id: "t1", title: "Hello", number: 12 });
    });
    expect(result.current.closeTarget).toEqual({
      id: "t1",
      title: "Hello",
      number: 12,
    });
  });

  it("cancelClose clears the target without calling the API", () => {
    let closed = 0;
    server.use(taskClose("t1", () => {
      closed += 1;
    }));
    const { Wrapper } = makeMutationTestWrapper();
    const { result } = renderHook(() => useTaskCloseFlow(), {
      wrapper: Wrapper,
    });
    act(() => {
      result.current.requestClose({ id: "t1", title: "X" });
    });
    act(() => {
      result.current.cancelClose();
    });
    expect(result.current.closeTarget).toBeNull();
    expect(closed).toBe(0);
  });

  it("confirmClose is a no-op when no target is set", () => {
    let closed = 0;
    server.use(taskClose("t1", () => {
      closed += 1;
    }));
    const { Wrapper } = makeMutationTestWrapper();
    const { result } = renderHook(() => useTaskCloseFlow(), {
      wrapper: Wrapper,
    });
    act(() => {
      result.current.confirmClose();
    });
    expect(closed).toBe(0);
  });

  it("confirmClose calls the API, invalidates list+stats, clears target, fires onClosed", async () => {
    let closed = 0;
    server.use(taskClose("t1", () => {
      closed += 1;
    }));
    const { Wrapper, invalidateSpy } = makeMutationTestWrapper();
    const onClosed = vi.fn();
    const { result } = renderHook(() => useTaskCloseFlow({ onClosed }), {
      wrapper: Wrapper,
    });

    act(() => {
      result.current.requestClose({ id: "t1", title: "X" });
    });
    act(() => {
      result.current.confirmClose();
    });

    await waitFor(() => {
      expect(result.current.closeSuccess).toBe(true);
    });

    expect(closed).toBe(1);
    expect(result.current.closeTarget).toBeNull();
    expect(result.current.closeVariables).toEqual({ id: "t1" });
    expect(onClosed).toHaveBeenCalledWith("t1");
    expect(invalidateSpy).toHaveBeenCalledWith({
      queryKey: ["tasks", "list"],
    });
    expect(invalidateSpy).toHaveBeenCalledWith({
      queryKey: taskQueryKeys.stats(),
    });
  });

  it("surfaces API errors via closeError without clearing the target", async () => {
    server.use(taskCloseError("t1", 403, "nope"));
    const { Wrapper } = makeMutationTestWrapper();
    const onClosed = vi.fn();
    const { result } = renderHook(() => useTaskCloseFlow({ onClosed }), {
      wrapper: Wrapper,
    });

    act(() => {
      result.current.requestClose({ id: "t1", title: "X" });
    });
    act(() => {
      result.current.confirmClose();
    });

    await waitFor(() => {
      expect(result.current.closeError).toBe("nope");
    });
    expect(result.current.closeSuccess).toBe(false);
    expect(result.current.closeTarget).toEqual({
      id: "t1",
      title: "X",
      number: null,
    });
    expect(onClosed).not.toHaveBeenCalled();
  });

  it("resetCloseError clears a settled error without firing a new request", async () => {
    server.use(taskCloseError("t1", 500, "boom"));
    const { Wrapper } = makeMutationTestWrapper();
    const { result } = renderHook(() => useTaskCloseFlow(), {
      wrapper: Wrapper,
    });
    act(() => {
      result.current.requestClose({ id: "t1", title: "X" });
    });
    act(() => {
      result.current.confirmClose();
    });
    await waitFor(() => {
      expect(result.current.closeError).toBe("boom");
    });
    act(() => {
      result.current.resetCloseError();
    });
    await waitFor(() => {
      expect(result.current.closeError).toBeNull();
    });
  });

  it("optimistically flips status to closed in cached list data", async () => {
    const [handler, deferred] = taskClosePending("t1");
    server.use(handler);
    const { Wrapper, queryClient } = makeMutationTestWrapper();
    const list: TaskListResponse = {
      tasks: [
        makeTask({ id: "t1", status: "running" }),
        makeTask({ id: "t2" }),
      ],
      limit: 50,
      offset: 0,
      has_more: false,
    };
    queryClient.setQueryData<TaskListResponse>(
      taskQueryKeys.list({ limit: 20, offset: 0 }),
      list,
    );

    const { result } = renderHook(() => useTaskCloseFlow(), {
      wrapper: Wrapper,
    });
    act(() => {
      result.current.requestClose({ id: "t1", title: "Some task" });
    });
    act(() => {
      result.current.confirmClose();
    });
    await waitFor(() => {
      expect(result.current.closePending).toBe(true);
    });
    const cached = queryClient.getQueryData<TaskListResponse>(
      taskQueryKeys.list({ limit: 20, offset: 0 }),
    );
    const t1 = cached?.tasks.find((t) => t.id === "t1");
    expect(t1?.status).toBe("closed");
    act(() => {
      deferred.resolve(HttpResponse.json(makeTask({ id: "t1", status: "closed" })));
    });
    await waitFor(() => {
      expect(result.current.closePending).toBe(false);
    });
  });

  it("restores the list cache on server error", async () => {
    server.use(taskCloseError("t1", 403, "perm denied"));
    const { Wrapper, queryClient } = makeMutationTestWrapper();
    const list: TaskListResponse = {
      tasks: [
        makeTask({ id: "t1", status: "running" }),
        makeTask({ id: "t2" }),
      ],
      limit: 50,
      offset: 0,
      has_more: false,
    };
    queryClient.setQueryData<TaskListResponse>(
      taskQueryKeys.list({ limit: 20, offset: 0 }),
      list,
    );

    const { result } = renderHook(() => useTaskCloseFlow(), {
      wrapper: Wrapper,
    });
    act(() => {
      result.current.requestClose({ id: "t1", title: "Some task" });
    });
    act(() => {
      result.current.confirmClose();
    });
    await waitFor(() => {
      expect(result.current.closeError).toBe("perm denied");
    });
    const restored = queryClient.getQueryData<TaskListResponse>(
      taskQueryKeys.list({ limit: 20, offset: 0 }),
    );
    expect(restored?.tasks.find((t) => t.id === "t1")?.status).toBe("running");
  });

  it("bumps the optimistic-version counter so SSE echoes are suppressed in flight", async () => {
    const [handler, deferred] = taskClosePending("t1");
    server.use(handler);
    const { Wrapper } = makeMutationTestWrapper();
    const { result } = renderHook(() => useTaskCloseFlow(), {
      wrapper: Wrapper,
    });
    act(() => {
      result.current.requestClose({ id: "t1", title: "Some task" });
    });
    act(() => {
      result.current.confirmClose();
    });
    await waitFor(() => {
      expect(result.current.closePending).toBe(true);
    });
    expect(shouldSuppressTaskMutationEcho("t1")).toBe(true);
    act(() => {
      deferred.resolve(
        HttpResponse.json(makeTask({ id: "t1", status: "closed" })),
      );
    });
    await waitFor(() => {
      expect(result.current.closePending).toBe(false);
    });
    expect(shouldSuppressTaskMutationEcho("t1")).toBe(false);
  });

  it("reopen flips status back to ready optimistically then invalidates", async () => {
    server.use(taskReopen("t1"));
    const { Wrapper, queryClient, invalidateSpy } = makeMutationTestWrapper();
    queryClient.setQueryData(
      taskQueryKeys.detail("t1"),
      makeTask({ id: "t1", status: "closed" }),
    );
    const { result } = renderHook(() => useTaskCloseFlow(), {
      wrapper: Wrapper,
    });
    act(() => {
      result.current.reopen("t1");
    });
    await waitFor(() => {
      expect(result.current.reopenPending).toBe(false);
    });
    const detail = queryClient.getQueryData<{ status: string }>(
      taskQueryKeys.detail("t1"),
    );
    expect(detail?.status).toBe("ready");
    expect(invalidateSpy).toHaveBeenCalledWith({
      queryKey: ["tasks", "list"],
    });
  });

  it("reopen surfaces errors and restores prior status", async () => {
    server.use(taskReopenError("t1", 409, "already open"));
    const { Wrapper, queryClient } = makeMutationTestWrapper();
    queryClient.setQueryData(
      taskQueryKeys.detail("t1"),
      makeTask({ id: "t1", status: "closed" }),
    );
    const { result } = renderHook(() => useTaskCloseFlow(), {
      wrapper: Wrapper,
    });
    act(() => {
      result.current.reopen("t1");
    });
    await waitFor(() => {
      expect(result.current.reopenError).toBe("already open");
    });
    const detail = queryClient.getQueryData<{ status: string }>(
      taskQueryKeys.detail("t1"),
    );
    expect(detail?.status).toBe("closed");
  });
});
