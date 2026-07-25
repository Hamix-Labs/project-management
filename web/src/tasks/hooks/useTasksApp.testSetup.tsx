import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, waitFor } from "@testing-library/react";
import { expect, vi } from "vitest";
import type { ReactNode } from "react";
import { useTasksApp } from "./useTasksApp";

vi.mock("../../api", () => ({
  listTasks: vi.fn(),
  getTaskStats: vi.fn(),
  listTaskDrafts: vi.fn(),
  getTaskDraft: vi.fn(),
  saveTaskDraft: vi.fn(),
  deleteTaskDraft: vi.fn(),
  createTask: vi.fn(),
  patchTask: vi.fn(),
  closeTask: vi.fn(),
  reopenTask: vi.fn(),
  addChecklistItem: vi.fn(),
}));

vi.mock("@/lib/ensureRepositoriesRegistered", () => ({
  ensureRepositoriesRegistered: vi.fn().mockResolvedValue(true),
}));

import {
  listTasks,
  getTaskStats,
  listTaskDrafts,
  saveTaskDraft,
  getTaskDraft,
  createTask,
  closeTask,
  reopenTask,
} from "../../api";
import { ensureRepositoriesRegistered } from "@/lib/ensureRepositoriesRegistered";

const mockedListTasks = vi.mocked(listTasks);
const mockedGetStats = vi.mocked(getTaskStats);
const mockedListDrafts = vi.mocked(listTaskDrafts);
const mockedSaveDraft = vi.mocked(saveTaskDraft);
const mockedGetDraft = vi.mocked(getTaskDraft);
const mockedCreateTask = vi.mocked(createTask);
const mockedCloseTask = vi.mocked(closeTask);
const mockedReopenTask = vi.mocked(reopenTask);
const mockedEnsureRepos = vi.mocked(ensureRepositoriesRegistered);

async function openCreateModalReady(
  result: { current: ReturnType<typeof useTasksApp> },
) {
  await act(async () => {
    result.current.openCreateModal();
  });
  await waitFor(() => {
    expect(result.current.createModalOpen).toBe(true);
  });
}

function makeWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  });
  function Wrapper({ children }: { children: ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>
        {children}
      </QueryClientProvider>
    );
  }
  return { Wrapper, queryClient };
}


export {
  mockedListTasks,
  mockedGetStats,
  mockedListDrafts,
  mockedSaveDraft,
  mockedGetDraft,
  mockedCreateTask,
  mockedCloseTask,
  mockedReopenTask,
  mockedEnsureRepos,
  openCreateModalReady,
  makeWrapper,
};
