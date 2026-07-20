import { act, renderHook, waitFor } from "@testing-library/react";
import { HttpResponse } from "msw";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { useTaskDeleteFlow } from "./useTaskDeleteFlow";
import { taskQueryKeys } from "../task-query";
import {
  __resetMutationGuardForTests,
  shouldSuppressTaskMutationEcho,
} from "@/tasks/sync/mutationGuard";
import type { TaskListResponse } from "@/types";
import { makeMutationTestWrapper } from "@/test/reactQuery";
import { makeTask } from "@/test/taskDefaults";
import {
  taskDelete,
  taskDeleteError,
  taskDeletePending,
} from "@/test/handlers/tasks";
import { server } from "@/test/server";

describe("useTaskDeleteFlow", () => {
  beforeEach(() => {
    __resetMutationGuardForTests();
  });
  afterEach(() => {
    __resetMutationGuardForTests();
    vi.restoreAllMocks();
  });

  it("starts with no target, no pending, no success, no error", () => {
    const { Wrapper } = makeMutationTestWrapper();
    const { result } = renderHook(() => useTaskDeleteFlow(), {
      wrapper: Wrapper,
    });
    expect(result.current.deleteTarget).toBeNull();
    expect(result.current.deletePending).toBe(false);
    expect(result.current.deleteSuccess).toBe(false);
    expect(result.current.deleteError).toBeNull();
    expect(result.current.deleteVariables).toBeUndefined();
  });

  it("requestDelete captures id and title", () => {
    const { Wrapper } = makeMutationTestWrapper();
    const { result } = renderHook(() => useTaskDeleteFlow(), {
      wrapper: Wrapper,
    });
    act(() => {
      result.current.requestDelete({ id: "t1", title: "Hello" });
    });
    expect(result.current.deleteTarget).toEqual({
      id: "t1",
      title: "Hello",
    });
  });

  it("cancelDelete clears the target without calling the API", () => {
    let deleted = 0;
    server.use(taskDelete("t1", () => {
      deleted += 1;
    }));
    const { Wrapper } = makeMutationTestWrapper();
    const { result } = renderHook(() => useTaskDeleteFlow(), {
      wrapper: Wrapper,
    });
    act(() => {
      result.current.requestDelete({ id: "t1", title: "X" });
    });
    act(() => {
      result.current.cancelDelete();
    });
    expect(result.current.deleteTarget).toBeNull();
    expect(deleted).toBe(0);
  });

  it("confirmDelete is a no-op when no target is set", () => {
    let deleted = 0;
    server.use(taskDelete("t1", () => {
      deleted += 1;
    }));
    const { Wrapper } = makeMutationTestWrapper();
    const { result } = renderHook(() => useTaskDeleteFlow(), {
      wrapper: Wrapper,
    });
    act(() => {
      result.current.confirmDelete();
    });
    expect(deleted).toBe(0);
  });

  it("confirmDelete calls the API, invalidates list+stats, clears target, fires onDeleted", async () => {
    let deleted = 0;
    server.use(taskDelete("t1", () => {
      deleted += 1;
    }));
    const { Wrapper, invalidateSpy } = makeMutationTestWrapper();
    const onDeleted = vi.fn();
    const { result } = renderHook(() => useTaskDeleteFlow({ onDeleted }), {
      wrapper: Wrapper,
    });

    act(() => {
      result.current.requestDelete({ id: "t1", title: "X" });
    });
    act(() => {
      result.current.confirmDelete();
    });

    await waitFor(() => {
      expect(result.current.deleteSuccess).toBe(true);
    });

    expect(deleted).toBe(1);
    expect(result.current.deleteTarget).toBeNull();
    expect(result.current.deleteVariables).toEqual({ id: "t1" });
    expect(onDeleted).toHaveBeenCalledWith("t1");
    expect(invalidateSpy).toHaveBeenCalledWith({
      queryKey: ["tasks", "list"],
    });
    expect(invalidateSpy).toHaveBeenCalledWith({
      queryKey: taskQueryKeys.stats(),
    });
  });

  it("surfaces API errors via deleteError without clearing the target", async () => {
    server.use(taskDeleteError("t1", 403, "nope"));
    const { Wrapper } = makeMutationTestWrapper();
    const onDeleted = vi.fn();
    const { result } = renderHook(() => useTaskDeleteFlow({ onDeleted }), {
      wrapper: Wrapper,
    });

    act(() => {
      result.current.requestDelete({ id: "t1", title: "X" });
    });
    act(() => {
      result.current.confirmDelete();
    });

    await waitFor(() => {
      expect(result.current.deleteError).toBe("nope");
    });
    expect(result.current.deleteSuccess).toBe(false);
    expect(result.current.deleteTarget).toEqual({
      id: "t1",
      title: "X",
    });
    expect(onDeleted).not.toHaveBeenCalled();
  });

  it("resetError clears a settled error without firing a new request (session #34)", async () => {
    server.use(taskDeleteError("t1", 500, "boom"));
    const { Wrapper } = makeMutationTestWrapper();
    const { result } = renderHook(() => useTaskDeleteFlow(), {
      wrapper: Wrapper,
    });
    act(() => {
      result.current.requestDelete({ id: "t1", title: "X" });
    });
    act(() => {
      result.current.confirmDelete();
    });
    await waitFor(() => {
      expect(result.current.deleteError).toBe("boom");
    });
    act(() => {
      result.current.resetError();
    });
    await waitFor(() => {
      expect(result.current.deleteError).toBeNull();
    });
  });

  it("resetError is a no-op while idle (no extra reset churn)", () => {
    let deleted = 0;
    server.use(taskDelete("t1", () => {
      deleted += 1;
    }));
    const { Wrapper } = makeMutationTestWrapper();
    const { result } = renderHook(() => useTaskDeleteFlow(), {
      wrapper: Wrapper,
    });
    expect(result.current.deleteError).toBeNull();
    act(() => {
      result.current.resetError();
    });
    expect(result.current.deleteError).toBeNull();
    expect(deleted).toBe(0);
  });

  it("omits parent_id from delete variables", async () => {
    server.use(taskDelete("root"));
    const { Wrapper } = makeMutationTestWrapper();
    const { result } = renderHook(() => useTaskDeleteFlow(), {
      wrapper: Wrapper,
    });

    act(() => {
      result.current.requestDelete({ id: "root", title: "X" });
    });
    act(() => {
      result.current.confirmDelete();
    });

    await waitFor(() => {
      expect(result.current.deleteSuccess).toBe(true);
    });
    expect(result.current.deleteVariables).toEqual({ id: "root" });
    expect(result.current.deleteVariables).not.toHaveProperty("parent_id");
  });

  it("does not clobber a freshly-opened confirm dialog when a previous delete settles", async () => {
    const [handlerA, deferredA] = taskDeletePending("A");
    server.use(handlerA, taskDelete("B"));

    const { Wrapper } = makeMutationTestWrapper();
    const onDeleted = vi.fn();
    const { result } = renderHook(() => useTaskDeleteFlow({ onDeleted }), {
      wrapper: Wrapper,
    });

    act(() => {
      result.current.requestDelete({ id: "A", title: "A" });
    });
    act(() => {
      result.current.confirmDelete();
    });

    await waitFor(() => {
      expect(result.current.deletePending).toBe(true);
    });

    act(() => {
      result.current.requestDelete({ id: "B", title: "B" });
    });
    expect(result.current.deleteTarget).toEqual({
      id: "B",
      title: "B",
    });

    act(() => {
      deferredA.resolve(new HttpResponse(null, { status: 204 }));
    });

    await waitFor(() => {
      expect(onDeleted).toHaveBeenCalledWith("A");
    });

    expect(result.current.deleteTarget).toEqual({
      id: "B",
      title: "B",
    });
  });

  it("optimistically removes the row from cached list data before the server resolves", async () => {
    const [handler, deferred] = taskDeletePending("t1");
    server.use(handler);
    const { Wrapper, queryClient } = makeMutationTestWrapper();
    const list: TaskListResponse = {
      tasks: [makeTask({ id: "t1" }), makeTask({ id: "t2" })],
      limit: 50,
      offset: 0,
      has_more: false,
    };
    queryClient.setQueryData<TaskListResponse>(taskQueryKeys.list({ limit: 20, offset: 0 }), list);

    const { result } = renderHook(() => useTaskDeleteFlow(), {
      wrapper: Wrapper,
    });
    act(() => {
      result.current.requestDelete({ id: "t1", title: "Some task" });
    });
    act(() => {
      result.current.confirmDelete();
    });
    await waitFor(() => {
      expect(result.current.deletePending).toBe(true);
    });
    const cached = queryClient.getQueryData<TaskListResponse>(taskQueryKeys.list({ limit: 20, offset: 0 }));
    expect(cached?.tasks.map((t) => t.id)).toEqual(["t2"]);
    act(() => {
      deferred.resolve(new HttpResponse(null, { status: 204 }));
    });
    await waitFor(() => {
      expect(result.current.deletePending).toBe(false);
    });
  });

  it("restores the list cache on server error", async () => {
    server.use(taskDeleteError("t1", 403, "perm denied"));
    const { Wrapper, queryClient } = makeMutationTestWrapper();
    const list: TaskListResponse = {
      tasks: [makeTask({ id: "t1" }), makeTask({ id: "t2" })],
      limit: 50,
      offset: 0,
      has_more: false,
    };
    queryClient.setQueryData<TaskListResponse>(taskQueryKeys.list({ limit: 20, offset: 0 }), list);

    const { result } = renderHook(() => useTaskDeleteFlow(), {
      wrapper: Wrapper,
    });
    act(() => {
      result.current.requestDelete({ id: "t1", title: "Some task" });
    });
    act(() => {
      result.current.confirmDelete();
    });
    await waitFor(() => {
      expect(result.current.deleteError).toBe("perm denied");
    });
    const restored = queryClient.getQueryData<TaskListResponse>(taskQueryKeys.list({ limit: 20, offset: 0 }));
    expect(restored?.tasks.map((t) => t.id)).toEqual(["t1", "t2"]);
  });

  it("bumps the optimistic-version counter so SSE echoes are suppressed in flight", async () => {
    const [handler, deferred] = taskDeletePending("t1");
    server.use(handler);
    const { Wrapper } = makeMutationTestWrapper();
    const { result } = renderHook(() => useTaskDeleteFlow(), {
      wrapper: Wrapper,
    });
    act(() => {
      result.current.requestDelete({ id: "t1", title: "Some task" });
    });
    act(() => {
      result.current.confirmDelete();
    });
    await waitFor(() => {
      expect(result.current.deletePending).toBe(true);
    });
    expect(shouldSuppressTaskMutationEcho("t1")).toBe(true);
    act(() => {
      deferred.resolve(new HttpResponse(null, { status: 204 }));
    });
    await waitFor(() => {
      expect(result.current.deletePending).toBe(false);
    });
    expect(shouldSuppressTaskMutationEcho("t1")).toBe(false);
  });
});
