import "./useTasksApp.testSetup";
import { act, renderHook, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
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

describe("useTasksApp saveDraftMutation race", () => {
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
    mockedSaveDraft.mockReset();
    mockedGetDraft.mockReset();
  });

  it("does not stamp the autosave baseline / 'Draft saved' label onto the now-current draft when a save for a previous draft resolves late", async () => {
    // Hold the autosave for draft A so we can switch to draft B before it
    // resolves. Capture the id we sent so the resolution can echo it back -
    // that's what the server does today.
    let resolveSaveA: ((v: { id: string; name: string }) => void) | undefined;
    let savedAId: string | undefined;
    mockedSaveDraft.mockImplementationOnce((input) => {
      savedAId = input.id;
      return new Promise<{ id: string; name: string }>((resolve) => {
        resolveSaveA = resolve;
      });
    });

    // Draft B comes back from the picker with all the fields the resume
    // path stamps onto state.
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

    // Touch a field so the autosave signature differs from the baseline -
    // otherwise saveDraftNow short-circuits before calling the API.
    act(() => {
      result.current.setNewTitle("Draft A title");
    });

    act(() => {
      result.current.saveDraftNow();
    });
    await waitFor(() => {
      expect(result.current.draftSavePending).toBe(true);
    });
    expect(savedAId).toBeDefined();
    expect(savedAId).not.toBe("draft-B-id");

    // Mid-flight: user picks draft B from the picker. resumeDraftByID
    // synchronously updates newDraftIDRef to "draft-B-id" via the
    // setNewDraftID wrapper, then stamps the form + autosave baseline with
    // draft B's data.
    await act(async () => {
      await result.current.resumeDraftByID("draft-B-id");
    });
    expect(result.current.newTitle).toBe("Draft B title");
    expect(result.current.newPrompt).toBe("Draft B prompt");
    expect(result.current.newPriority).toBe("high");

    // Now resolve draft A's save. The server echoes the id we sent, so
    // saved.id !== newDraftIDRef.current and the guard must fire.
    await act(async () => {
      resolveSaveA?.({ id: savedAId!, name: "Untitled draft" });
    });
    await waitFor(() => {
      expect(result.current.draftSavePending).toBe(false);
    });

    // The form must STILL be showing draft B - the stale resolution must
    // not have stomped any of the fields newDraftID / newTitle / etc.
    expect(result.current.newTitle).toBe("Draft B title");
    expect(result.current.newPrompt).toBe("Draft B prompt");
    expect(result.current.newPriority).toBe("high");

    // The "Draft saved" label is the user-visible proof the baseline was
    // updated. With the bug, lastDraftSavedAt was set to Date.now() and
    // the label flips to "Draft saved" - falsely claiming draft B was
    // just saved when in reality the save was for draft A. With the
    // guard, lastDraftSavedAt stays null and the label stays null.
    expect(result.current.draftSaveLabel).toBeNull();
  });

  it("updates the autosave baseline + 'Draft saved' label on the happy path (no draft switch)", async () => {
    let savedId: string | undefined;
    mockedSaveDraft.mockImplementationOnce(async (input) => {
      savedId = input.id;
      return { id: input.id!, name: input.name };
    });

    const { Wrapper } = makeWrapper();
    const { result } = renderHook(() => useTasksApp({ sseLive: false }), { wrapper: Wrapper });

    await waitFor(() => {
      expect(result.current.draftListLoading).toBe(false);
    });

    await openCreateModalReady(result);
    act(() => {
      result.current.setNewTitle("Draft A title");
    });

    act(() => {
      result.current.saveDraftNow();
    });

    await waitFor(() => {
      expect(result.current.draftSaveLabel).toBe("Draft saved");
    });
    expect(savedId).toBeDefined();

    // Proof the baseline was actually updated to the just-saved state: the
    // debounced autosave has nothing left to write. (The explicit saveDraftNow
    // is no longer a witness for this - under I8 it writes unconditionally.)
    mockedSaveDraft.mockClear();
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

  it("baseline tracks the snapshot that was sent, not live form state, so edits made while a save is in flight still autosave on the next dispatch", async () => {
    // First save is held so we can edit the form mid-flight. The second
    // save resolves immediately so we can assert it actually fired.
    let resolveFirst: (() => void) | undefined;
    let firstInputTitle: string | undefined;
    let secondInputTitle: string | undefined;
    mockedSaveDraft.mockImplementationOnce(async (input) => {
      firstInputTitle = input.payload.title;
      await new Promise<void>((resolve) => {
        resolveFirst = resolve;
      });
      return { id: input.id!, name: input.name };
    });
    mockedSaveDraft.mockImplementationOnce(async (input) => {
      secondInputTitle = input.payload.title;
      return { id: input.id!, name: input.name };
    });

    const { Wrapper } = makeWrapper();
    const { result } = renderHook(() => useTasksApp({ sseLive: false }), { wrapper: Wrapper });

    await waitFor(() => {
      expect(result.current.draftListLoading).toBe(false);
    });

    await openCreateModalReady(result);

    // First edit: kicks the autosave off with snapshot S1.
    act(() => {
      result.current.setNewTitle("Title v1");
    });
    act(() => {
      result.current.saveDraftNow();
    });
    await waitFor(() => {
      expect(result.current.draftSavePending).toBe(true);
    });
    expect(firstInputTitle).toBe("Title v1");

    // Mid-flight: user keeps typing. Live form signature is now S2 (the
    // "Title v2" string). With the bug, onSuccess will rebuild the
    // baseline from live form state at resolve time -> baseline = S2.
    // currentSig = S2 too, so the next saveDraftNow gate matches and
    // autosave silently skips, even though the server still has v1.
    // With the fix, onSuccess uses variables.signature = S1, so the
    // next saveDraftNow gate sees S2 != S1 and fires.
    act(() => {
      result.current.setNewTitle("Title v2");
    });

    // Resolve the first save now (after the mid-flight edit landed in
    // state). draftSavePending flips back to false.
    await act(async () => {
      resolveFirst?.();
    });
    await waitFor(() => {
      expect(result.current.draftSavePending).toBe(false);
    });

    // The user-visible damage check: the debounced autosave MUST send
    // "Title v2" to the server. Without the fix, the baseline matched the
    // current signature and the gated autosave effect never scheduled a
    // write, so the v2 edit would sit unsaved until the next state change.
    // This has to go through the debounce rather than saveDraftNow, because
    // the explicit path is unconditional under I8 and would pass either way.
    vi.useFakeTimers();
    try {
      await act(async () => {
        await vi.advanceTimersByTimeAsync(2000);
      });
    } finally {
      vi.useRealTimers();
    }
    await waitFor(() => {
      expect(mockedSaveDraft).toHaveBeenCalledTimes(2);
    });
    expect(secondInputTitle).toBe("Title v2");
  });
});
