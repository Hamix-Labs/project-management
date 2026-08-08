import "./useTasksApp.testSetup";
import { act, renderHook, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { TaskDraftDetail } from "@/types";
import { useTasksApp } from "./useTasksApp";
import { stubEventSource } from "../../test/browserMocks";
import {
  makeWrapper,
  mockedEnsureRepos,
  mockedGetDraft,
  mockedGetStats,
  mockedListDrafts,
  mockedListTasks,
  mockedSaveDraft,
  openCreateModalReady,
} from "./useTasksApp.testSetup";

function draftWithRepository(repositoryID: string): TaskDraftDetail {
  return {
    id: "draft-1",
    name: "Draft 1",
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
    payload: {
      title: "Draft title",
      initial_prompt: "Draft prompt",
      priority: "medium",
      checklist_items: [],
      repository_id: repositoryID,
    },
  } as TaskDraftDetail;
}

describe("draft repository persistence", () => {
  beforeEach(() => {
    stubEventSource();
    mockedListTasks.mockResolvedValue({
      tasks: [],
      limit: 200,
      offset: 0,
      has_more: false,
    });
    mockedGetStats.mockResolvedValue(null as unknown as never);
    mockedListDrafts.mockResolvedValue([]);
    mockedEnsureRepos.mockResolvedValue(true);
    mockedSaveDraft.mockImplementation(async (input) => ({
      id: input.id!,
      name: input.name,
    }));
  });

  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
    mockedSaveDraft.mockReset();
    mockedGetDraft.mockReset();
  });

  it("restores the repository a draft was saved with", async () => {
    mockedGetDraft.mockResolvedValueOnce(draftWithRepository("repo-1"));

    const { Wrapper } = makeWrapper();
    const { result } = renderHook(() => useTasksApp({ sseLive: false }), {
      wrapper: Wrapper,
    });
    await waitFor(() => expect(result.current.draftListLoading).toBe(false));

    await act(async () => {
      await result.current.resumeDraftByID("draft-1");
    });

    expect(result.current.newRepositoryID).toBe("repo-1");
  });

  // The reported bug: the repository was the only edit, the dirty gate could not
  // see it, so nothing was written and reopening the draft showed no repository.
  it("autosaves when the repository is the only thing that changed", async () => {
    mockedGetDraft.mockResolvedValueOnce(draftWithRepository("repo-1"));

    const { Wrapper } = makeWrapper();
    const { result } = renderHook(() => useTasksApp({ sseLive: false }), {
      wrapper: Wrapper,
    });
    await waitFor(() => expect(result.current.draftListLoading).toBe(false));

    await act(async () => {
      await result.current.resumeDraftByID("draft-1");
    });
    expect(mockedSaveDraft).not.toHaveBeenCalled();

    act(() => {
      result.current.setNewRepositoryID("repo-2");
    });
    act(() => {
      result.current.saveDraftNow();
    });

    await waitFor(() => expect(mockedSaveDraft).toHaveBeenCalledTimes(1));
    expect(mockedSaveDraft.mock.calls[0][0].payload.repository_id).toBe("repo-2");
  });

  it("keeps an untouched resumed draft clean", async () => {
    mockedGetDraft.mockResolvedValueOnce(draftWithRepository("repo-1"));

    const { Wrapper } = makeWrapper();
    const { result } = renderHook(() => useTasksApp({ sseLive: false }), {
      wrapper: Wrapper,
    });
    await waitFor(() => expect(result.current.draftListLoading).toBe(false));

    await act(async () => {
      await result.current.resumeDraftByID("draft-1");
    });

    // The baseline is derived from the same mapper that filled the form, so a
    // resumed draft must not look dirty and autosave itself back.
    vi.useFakeTimers();
    try {
      await act(async () => {
        await vi.advanceTimersByTimeAsync(2000);
      });
      expect(mockedSaveDraft).not.toHaveBeenCalled();
    } finally {
      vi.useRealTimers();
    }
  });

  it("sends the repository on a fresh draft's first autosave", async () => {
    const { Wrapper } = makeWrapper();
    const { result } = renderHook(() => useTasksApp({ sseLive: false }), {
      wrapper: Wrapper,
    });
    await waitFor(() => expect(result.current.draftListLoading).toBe(false));
    await openCreateModalReady(result);

    act(() => {
      result.current.setNewRepositoryID("repo-9");
    });
    act(() => {
      result.current.saveDraftNow();
    });

    await waitFor(() => expect(mockedSaveDraft).toHaveBeenCalledTimes(1));
    expect(mockedSaveDraft.mock.calls[0][0].payload.repository_id).toBe("repo-9");
  });
});
