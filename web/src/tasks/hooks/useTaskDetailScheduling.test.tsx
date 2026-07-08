import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, renderHook, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { taskQueryKeys } from "../task-query";
import { useTaskDetailScheduling } from "./useTaskDetailScheduling";

const {
  mockAddTaskDependency,
  mockRemoveTaskDependency,
  mockPatchTaskGate,
  mockPatchTask,
} = vi.hoisted(() => ({
  mockAddTaskDependency: vi.fn(),
  mockRemoveTaskDependency: vi.fn(),
  mockPatchTaskGate: vi.fn(),
  mockPatchTask: vi.fn(),
}));

vi.mock("@/api", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/api")>();
  return {
    ...actual,
    addTaskDependency: mockAddTaskDependency,
    removeTaskDependency: mockRemoveTaskDependency,
    patchTaskGate: mockPatchTaskGate,
    patchTask: mockPatchTask,
  };
});

const TASK_ID = "11111111-1111-4111-8111-111111111111";
const UPSTREAM_ID = "22222222-2222-4222-8222-222222222222";

function createWrapper(qc: QueryClient) {
  return function Wrapper({ children }: { children: ReactNode }) {
    return <QueryClientProvider client={qc}>{children}</QueryClientProvider>;
  };
}

function newQueryClient() {
  return new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
}

describe("useTaskDetailScheduling", () => {
  beforeEach(() => {
    mockAddTaskDependency.mockReset();
    mockRemoveTaskDependency.mockReset();
    mockPatchTaskGate.mockReset();
    mockPatchTask.mockReset();
    mockAddTaskDependency.mockResolvedValue(undefined);
    mockRemoveTaskDependency.mockResolvedValue(undefined);
    mockPatchTaskGate.mockResolvedValue(undefined);
    mockPatchTask.mockResolvedValue(undefined);
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("adds a dependency and clears the input on success", async () => {
    const qc = newQueryClient();
    const invalidateSpy = vi.spyOn(qc, "invalidateQueries");
    const { result } = renderHook(() => useTaskDetailScheduling(TASK_ID), {
      wrapper: createWrapper(qc),
    });

    act(() => {
      result.current.setDepAddValue(UPSTREAM_ID);
    });

    act(() => {
      result.current.addDepMutation.mutate(undefined);
    });

    await waitFor(() => {
      expect(mockAddTaskDependency).toHaveBeenCalledWith(TASK_ID, UPSTREAM_ID);
    });
    await waitFor(() => {
      expect(result.current.depAddValue).toBe("");
    });
    expect(invalidateSpy).toHaveBeenCalledWith({
      queryKey: taskQueryKeys.detail(TASK_ID),
    });
  });

  it("removes a dependency and invalidates task queries", async () => {
    const qc = newQueryClient();
    const invalidateSpy = vi.spyOn(qc, "invalidateQueries");
    const { result } = renderHook(() => useTaskDetailScheduling(TASK_ID), {
      wrapper: createWrapper(qc),
    });

    act(() => {
      result.current.removeDepMutation.mutate(UPSTREAM_ID);
    });

    await waitFor(() => {
      expect(mockRemoveTaskDependency).toHaveBeenCalledWith(TASK_ID, UPSTREAM_ID);
    });
    expect(invalidateSpy).toHaveBeenCalledWith({
      queryKey: taskQueryKeys.detail(TASK_ID),
    });
  });

  it("releases the gate via patchTaskGate", async () => {
    const qc = newQueryClient();
    const { result } = renderHook(() => useTaskDetailScheduling(TASK_ID), {
      wrapper: createWrapper(qc),
    });

    act(() => {
      result.current.gateMutation.mutate("release");
    });

    await waitFor(() => {
      expect(mockPatchTaskGate).toHaveBeenCalledWith(TASK_ID, "release");
    });
  });

  it("patches tags and milestone", async () => {
    const qc = newQueryClient();
    const { result } = renderHook(() => useTaskDetailScheduling(TASK_ID), {
      wrapper: createWrapper(qc),
    });

    act(() => {
      result.current.tagsMutation.mutate(["backend", "api"]);
    });
    act(() => {
      result.current.milestoneMutation.mutate("M1");
    });

    await waitFor(() => {
      expect(mockPatchTask).toHaveBeenCalledWith(TASK_ID, {
        tags: ["backend", "api"],
      });
    });
    await waitFor(() => {
      expect(mockPatchTask).toHaveBeenCalledWith(TASK_ID, { milestone: "M1" });
    });
  });
});
