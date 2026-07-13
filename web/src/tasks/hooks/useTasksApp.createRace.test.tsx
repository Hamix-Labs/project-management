import "./useTasksApp.testSetup";
import { act, renderHook, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { FormEvent } from "react";
import { useTasksApp } from "./useTasksApp";
import { stubEventSource } from "../../test/browserMocks";
import { TASK_TEST_DEFAULTS } from "@/test/taskDefaults";
import {
  makeWrapper,
  mockedCreateTask,
  mockedEnsureRepos,
  mockedGetDraft,
  mockedGetStats,
  mockedListDrafts,
  mockedListTasks,
  openCreateModalReady,
} from "./useTasksApp.testSetup";

describe("useTasksApp createMutation race", () => {
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
  });

  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
    mockedCreateTask.mockReset();
    mockedGetDraft.mockReset();
  });

  it("does not closeCreateModal when a stale create resolves after the user switched drafts (defensive guard)", async () => {
    // Hold the create mutation so we can interleave a draft switch
    // before it resolves. Today this race is unreachable in production
    // because `Modal busy={pending}` blocks ESC/backdrop close while
    // the create is in flight, but the *hook itself* doesn't refuse a
    // programmatic resume - the moment the modal lock is loosened (or
    // an out-of-modal "submit and continue editing" path lands), the
    // unconditional `closeCreateModal()` would slam shut a draft the
    // user has since switched to. The guard documents and pins that
    // contract: the modal close is gated on the just-resolved create
    // matching the currently-active draft id.
    let resolveCreate: (() => void) | undefined;
    let createdDraftId: string | undefined;
    mockedCreateTask.mockImplementationOnce(async (input) => {
      createdDraftId = input.draft_id;
      await new Promise<void>((r) => {
        resolveCreate = r;
      });
      return {
        id: "task-1",
        title: input.title,
        initial_prompt: input.initial_prompt ?? "",
        status: "ready",
        priority: input.priority,
        runner: input.runner ?? TASK_TEST_DEFAULTS.runner,
        cursor_model: input.cursor_model ?? TASK_TEST_DEFAULTS.cursor_model,
      };
    });

    mockedGetDraft.mockResolvedValueOnce({
      id: "draft-B-id",
      name: "Draft B",
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
      payload: {
        title: "Draft B title",
        initial_prompt: "Draft B prompt",
        priority: "high",
        checklist_items: [],
      },
    });

    const { Wrapper } = makeWrapper();
    const { result } = renderHook(() => useTasksApp({ sseLive: false }), { wrapper: Wrapper });

    await waitFor(() => {
      expect(result.current.draftListLoading).toBe(false);
    });

    await openCreateModalReady(result);
    act(() => {
      result.current.setNewTitle("Draft A");
    });
    act(() => {
      result.current.setNewPriority("medium");
      result.current.appendNewChecklistCriterion("Criterion");
    });

    act(() => {
      result.current.submitCreate({
        preventDefault: () => {},
      } as unknown as FormEvent);
    });
    await waitFor(() => {
      expect(result.current.createPending).toBe(true);
    });
    expect(createdDraftId).toBeDefined();
    expect(createdDraftId).not.toBe("draft-B-id");

    // Programmatically resume draft B mid-flight. Bypasses the UI lock
    // (which is the defensive scenario this guard exists for).
    await act(async () => {
      await result.current.resumeDraftByID("draft-B-id");
    });
    expect(result.current.newTitle).toBe("Draft B title");
    expect(result.current.createModalOpen).toBe(true);

    await act(async () => {
      resolveCreate?.();
    });
    await waitFor(() => {
      expect(result.current.createPending).toBe(false);
    });

    // The modal must STILL be open showing draft B - the stale create
    // resolution must not have closed it. Without the guard,
    // `closeCreateModal()` runs unconditionally and the modal shuts.
    expect(result.current.createModalOpen).toBe(true);
    expect(result.current.newTitle).toBe("Draft B title");
  });

  it("closeCreateModal still fires on the happy path (no draft switch)", async () => {
    mockedCreateTask.mockImplementationOnce(async (input) => ({
      id: "task-2",
      title: input.title,
      initial_prompt: input.initial_prompt ?? "",
      status: "ready",
      priority: input.priority,
      runner: input.runner ?? TASK_TEST_DEFAULTS.runner,
      cursor_model: input.cursor_model ?? TASK_TEST_DEFAULTS.cursor_model,
    }));

    const { Wrapper } = makeWrapper();
    const { result } = renderHook(() => useTasksApp({ sseLive: false }), { wrapper: Wrapper });

    await waitFor(() => {
      expect(result.current.draftListLoading).toBe(false);
    });

    await openCreateModalReady(result);
    act(() => {
      result.current.setNewTitle("Draft A");
    });
    act(() => {
      result.current.setNewPriority("medium");
      result.current.appendNewChecklistCriterion("Criterion");
    });

    act(() => {
      result.current.submitCreate({
        preventDefault: () => {},
      } as unknown as FormEvent);
    });

    await waitFor(() => {
      expect(result.current.createModalOpen).toBe(false);
    });
  });
});
