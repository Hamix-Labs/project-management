import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, renderHook, waitFor } from "@testing-library/react";
import { createElement, type ReactNode } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { taskQueryKeys } from "../../task-query";
import { useTaskCreateMutations } from "./useTaskCreateMutations";
import {
  __resetMutationGuardForTests,
  shouldSuppressTaskMutationEcho,
} from "@/tasks/sync/mutationGuard";
import { makeTask } from "@/test/taskDefaults";

vi.mock("@/api", () => ({
  createTask: vi.fn(),
  deleteTaskDraft: vi.fn(),
  deleteTaskTemplate: vi.fn(),
  getTaskDraft: vi.fn(),
  getTaskTemplate: vi.fn(),
  instantiateTaskTemplates: vi.fn(),
  patchTaskTemplate: vi.fn(),
  saveTaskDraft: vi.fn(),
  saveTaskTemplate: vi.fn(),
}));

import { createTask, instantiateTaskTemplates } from "@/api";

const mockedCreate = vi.mocked(createTask);
const mockedInstantiate = vi.mocked(instantiateTaskTemplates);

function makeMutationInput(queryClient: QueryClient) {
  return {
    queryClient,
    newDraftIDRef: { current: "draft-1" },
    newDraftID: "draft-1",
    closeCreateModal: vi.fn(),
    setNewDraftID: vi.fn(),
    setDraftAutosaveBaseline: vi.fn(),
    setDraftAutosaveBaselineID: vi.fn(),
    setLastDraftSavedAt: vi.fn(),
    createModalOpen: false,
    editingTemplateId: null,
  };
}

describe("useTaskCreateMutations", () => {
  afterEach(() => {
    __resetMutationGuardForTests();
    vi.clearAllMocks();
  });

  it("create arms mutation guard while seeding cache and invalidating list", async () => {
    const created = makeTask({ id: "task-new" });
    mockedCreate.mockResolvedValue(created);

    const queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
        mutations: { retry: false },
      },
    });
    const invalidateSpy = vi.spyOn(queryClient, "invalidateQueries");

    function Wrapper({ children }: { children: ReactNode }) {
      return createElement(QueryClientProvider, { client: queryClient }, children);
    }

    const { result } = renderHook(
      () => useTaskCreateMutations(makeMutationInput(queryClient)),
      { wrapper: Wrapper },
    );

    await act(async () => {
      await result.current.createMutation.mutateAsync({
        title: "T",
        initial_prompt: "p",
        status: "ready",
        priority: "medium",
        draft_id: "draft-1",
        runner: "cursor",
        cursor_model: "",
        project_id: "",
        pickup_not_before: null,
        tags: [],
        depends_on: [],
        repository_id: "repo-1",
        worktree_id: "",
        checklistItems: [],
      });
    });

    expect(queryClient.getQueryData(taskQueryKeys.detail("task-new"))).toEqual(
      created,
    );
    expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: taskQueryKeys.listRoot() });
    expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: taskQueryKeys.stats() });
    expect(shouldSuppressTaskMutationEcho("task-new")).toBe(false);
  });

  it("instantiate awaits list invalidation before the mutation settles", async () => {
    mockedInstantiate.mockResolvedValue({
      tasks: [makeTask()],
      errors: [],
    });

    const queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
        mutations: { retry: false },
      },
    });

    let resolveInvalidate: () => void = () => {};
    const slowInvalidate = new Promise<void>((resolve) => {
      resolveInvalidate = resolve;
    });
    const invalidateSpy = vi
      .spyOn(queryClient, "invalidateQueries")
      .mockImplementation(() => slowInvalidate);

    function Wrapper({ children }: { children: ReactNode }) {
      return createElement(QueryClientProvider, { client: queryClient }, children);
    }

    const { result } = renderHook(
      () => useTaskCreateMutations(makeMutationInput(queryClient)),
      { wrapper: Wrapper },
    );

    let mutationSettled = false;
    const pending = result.current.instantiateTemplatesMutation
      .mutateAsync([{ template_id: "tmpl-1", count: 1 }])
      .then(() => {
        mutationSettled = true;
      });

    await waitFor(() => {
      expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: taskQueryKeys.listRoot() });
    });
    expect(mutationSettled).toBe(false);

    await act(async () => {
      resolveInvalidate();
      await pending;
    });

    expect(mutationSettled).toBe(true);
    expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: taskQueryKeys.stats() });
  });
});
