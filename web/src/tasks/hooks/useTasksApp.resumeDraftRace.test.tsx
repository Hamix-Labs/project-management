import "./useTasksApp.testSetup";
import { act, renderHook, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { useTasksApp } from "./useTasksApp";
import { stubEventSource } from "../../test/browserMocks";
import type { TaskDraftDetail } from "@/types";
import {
  makeWrapper,
  mockedEnsureRepos,
  mockedGetDraft,
  mockedGetStats,
  mockedListDrafts,
  mockedListTasks,
} from "./useTasksApp.testSetup";

describe("useTasksApp resumeDraftMutation race", () => {
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
    mockedGetDraft.mockReset();
  });

  function makeDraftDetail(id: string, title: string): TaskDraftDetail {
    return {
      id,
      name: `Draft ${id}`,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
      payload: {
        title,
        initial_prompt: `${title} prompt`,
        priority: "medium",
        checklist_items: [],
      },
    } as TaskDraftDetail;
  }

  it("ignores out-of-order resume resolutions (last-clicked draft wins)", async () => {
    // Two concurrent resume requests. The user clicks B (slow GET),
    // then quickly clicks C before B resolves. We resolve C first
    // (so the form populates with C), then resolve B. Without the
    // `requestedResumeRef` guard, B's late resolution stamps the form
    // with B's payload *after* C already populated it - silently
    // reverting the user to the draft they navigated away from.
    let resolveB: ((v: TaskDraftDetail) => void) | undefined;
    let resolveC: ((v: TaskDraftDetail) => void) | undefined;
    mockedGetDraft.mockImplementation((id: string) => {
      if (id === "draft-B") {
        return new Promise((resolve) => {
          resolveB = resolve;
        });
      }
      if (id === "draft-C") {
        return new Promise((resolve) => {
          resolveC = resolve;
        });
      }
      throw new Error(`unexpected getDraft id ${id}`);
    });

    const { Wrapper } = makeWrapper();
    const { result } = renderHook(() => useTasksApp({ sseLive: false }), { wrapper: Wrapper });
    await waitFor(() => {
      expect(result.current.draftListLoading).toBe(false);
    });

    let bPromise: Promise<void> | undefined;
    let cPromise: Promise<void> | undefined;
    await act(async () => {
      bPromise = result.current.resumeDraftByID("draft-B");
      cPromise = result.current.resumeDraftByID("draft-C");
      // Yield so both `mutateAsync` calls have run their synchronous
      // prologue (registering with the mutation observer + invoking
      // the mocked `mockedGetDraft`) before we resolve either side.
      await Promise.resolve();
    });
    expect(resolveB).toBeDefined();
    expect(resolveC).toBeDefined();

    await act(async () => {
      resolveC?.(makeDraftDetail("draft-C", "Title C"));
      await cPromise;
    });
    expect(result.current.newTitle).toBe("Title C");
    expect(result.current.createModalOpen).toBe(true);

    await act(async () => {
      resolveB?.(makeDraftDetail("draft-B", "Title B"));
      await bPromise;
    });
    // The late B resolution must NOT have clobbered the form. The
    // user's most recent intent (C) is preserved.
    expect(result.current.newTitle).toBe("Title C");
  });

  it("drops a late resume resolution when the user starts a fresh draft mid-flight", async () => {
    // User clicks resume B, then changes their mind and hits "Start
    // fresh" before B resolves. `startFreshDraft` runs
    // `resetNewTaskForm`, which clears `requestedResumeRef`. When B
    // finally resolves, the post-await guard sees the ref no longer
    // matches "draft-B" and skips the form-stamp branch, preserving
    // the fresh draft the user is now editing.
    let resolveB: ((v: TaskDraftDetail) => void) | undefined;
    mockedGetDraft.mockImplementationOnce(
      () =>
        new Promise((resolve) => {
          resolveB = resolve;
        }),
    );

    const { Wrapper } = makeWrapper();
    const { result } = renderHook(() => useTasksApp({ sseLive: false }), { wrapper: Wrapper });
    await waitFor(() => {
      expect(result.current.draftListLoading).toBe(false);
    });

    let bPromise: Promise<void> | undefined;
    act(() => {
      bPromise = result.current.resumeDraftByID("draft-B");
    });

    await act(async () => {
      await result.current.startFreshDraft();
    });
    const freshTitleBeforeResolve = result.current.newTitle;
    expect(freshTitleBeforeResolve).toBe("");
    expect(result.current.createModalOpen).toBe(true);

    await act(async () => {
      resolveB?.(makeDraftDetail("draft-B", "Title B"));
      await bPromise;
    });

    expect(result.current.newTitle).toBe("");
    expect(result.current.createModalOpen).toBe(true);
  });

  it("happy path: a normal resume populates the form", async () => {
    mockedGetDraft.mockResolvedValueOnce(makeDraftDetail("draft-A", "Title A"));

    const { Wrapper } = makeWrapper();
    const { result } = renderHook(() => useTasksApp({ sseLive: false }), { wrapper: Wrapper });
    await waitFor(() => {
      expect(result.current.draftListLoading).toBe(false);
    });

    await act(async () => {
      await result.current.resumeDraftByID("draft-A");
    });

    expect(result.current.newTitle).toBe("Title A");
    expect(result.current.createModalOpen).toBe(true);
  });
});
