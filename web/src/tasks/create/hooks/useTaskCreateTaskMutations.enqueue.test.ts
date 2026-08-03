// @vitest-environment jsdom
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, renderHook } from "@testing-library/react";
import { createElement, type ReactNode } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { makeTask } from "@/test/taskDefaults";
import { useTaskCreateTaskMutations } from "./useTaskCreateTaskMutations";

vi.mock("@/api", () => ({
  createTask: vi.fn(),
  instantiateTaskTemplates: vi.fn(),
}));

import { createTask } from "@/api";

const mockedCreate = vi.mocked(createTask);

describe("useTaskCreateTaskMutations enqueue", () => {
  afterEach(() => {
    vi.clearAllMocks();
  });

  it("forwards worktree_id when enqueueing onto an existing worktree", async () => {
    mockedCreate.mockResolvedValue(makeTask({ id: "task-enq", worktree_id: "wt-shared" }));
    const queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
    });
    function Wrapper({ children }: { children: ReactNode }) {
      return createElement(QueryClientProvider, { client: queryClient }, children);
    }
    const { result } = renderHook(
      () =>
        useTaskCreateTaskMutations({
          queryClient,
          newDraftIDRef: { current: "draft-1" },
          closeCreateModal: vi.fn(),
        }),
      { wrapper: Wrapper },
    );

    await act(async () => {
      await result.current.createMutation.mutateAsync({
        title: "Enqueued",
        initial_prompt: "p",
        status: "ready",
        priority: "medium",
        draft_id: "draft-1",
        runner: "cursor",
        cursor_model: "",
        project_id: "proj-1",
        pickup_not_before: null,
        tags: [],
        depends_on: [],
        repository_id: "repo-1",
        worktree_id: "wt-shared",
        checklistItems: [{ text: "c" }],
      });
    });

    expect(mockedCreate).toHaveBeenCalledWith(
      expect.objectContaining({
        worktree_id: "wt-shared",
        repository_id: "repo-1",
        project_id: "proj-1",
      }),
    );
  });
});
