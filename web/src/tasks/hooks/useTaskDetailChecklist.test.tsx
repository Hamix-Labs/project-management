import "./useTaskDetailChecklist.testMocks";
import { act, renderHook, waitFor } from "@testing-library/react";
import type { FormEvent } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { shouldSuppressTaskMutationEcho } from "@/tasks/sync/mutationGuard";
import { taskQueryKeys } from "../task-query";
import { useTaskDetailChecklist } from "./useTaskDetailChecklist";
import type { TaskChecklistResponse } from "@/types";
import {
  ITEM_ID,
  resetChecklistMocks,
  setupChecklistTest,
  TASK_A,
  TASK_B,
} from "./useTaskDetailChecklist.testHelpers";
import { mockAdd, mockDelete, mockPatch } from "./useTaskDetailChecklist.testMocks";

describe("useTaskDetailChecklist", () => {
  beforeEach(() => {
    resetChecklistMocks();
  });

  it("clears checklist modals when taskId changes", () => {
    const { Wrapper } = setupChecklistTest();
    const { result, rerender } = renderHook(
      ({ taskId }: { taskId: string }) => useTaskDetailChecklist(taskId),
      {
        wrapper: Wrapper,
        initialProps: { taskId: TASK_A },
      },
    );

    act(() => {
      result.current.openChecklistModal();
      result.current.setNewChecklistText("x");
    });
    expect(result.current.checklistModalOpen).toBe(true);

    rerender({ taskId: TASK_B });
    expect(result.current.checklistModalOpen).toBe(false);
    expect(result.current.newChecklistText).toBe("");
    expect(result.current.editCriterionModalOpen).toBe(false);
  });

  it("openChecklistModal and closeChecklistModal", () => {
    const { Wrapper } = setupChecklistTest();
    const { result } = renderHook(() => useTaskDetailChecklist(TASK_A), {
      wrapper: Wrapper,
    });

    act(() => {
      result.current.openChecklistModal();
      result.current.setNewChecklistText("draft");
    });
    expect(result.current.checklistModalOpen).toBe(true);

    act(() => {
      result.current.closeChecklistModal();
    });
    expect(result.current.checklistModalOpen).toBe(false);
    expect(result.current.newChecklistText).toBe("");
  });

  it("openEditCriterionModal closes add modal and sets edit fields", () => {
    const { Wrapper } = setupChecklistTest();
    const { result } = renderHook(() => useTaskDetailChecklist(TASK_A), {
      wrapper: Wrapper,
    });

    act(() => {
      result.current.openChecklistModal();
      result.current.setNewChecklistText("n");
    });
    act(() => {
      result.current.openEditCriterionModal(ITEM_ID, "old text");
    });
    expect(result.current.checklistModalOpen).toBe(false);
    expect(result.current.newChecklistText).toBe("");
    expect(result.current.editCriterionModalOpen).toBe(true);
    expect(result.current.editingChecklistItemId).toBe(ITEM_ID);
    expect(result.current.editChecklistText).toBe("old text");
  });

  it("submitNewChecklistCriterion no-ops when text is blank", () => {
    const { Wrapper } = setupChecklistTest();
    const { result } = renderHook(() => useTaskDetailChecklist(TASK_A), {
      wrapper: Wrapper,
    });

    const ev = { preventDefault: vi.fn() } as unknown as FormEvent;
    act(() => {
      result.current.setNewChecklistText("   ");
      result.current.submitNewChecklistCriterion(ev);
    });
    expect(ev.preventDefault).toHaveBeenCalled();
    expect(mockAdd).not.toHaveBeenCalled();
  });

  it("submitNewChecklistCriterion adds item, invalidates, closes add modal", async () => {
    const { Wrapper, queryClient } = setupChecklistTest();
    const inv = vi.spyOn(queryClient, "invalidateQueries");
    const { result } = renderHook(() => useTaskDetailChecklist(TASK_A), {
      wrapper: Wrapper,
    });

    const ev = { preventDefault: vi.fn() } as unknown as FormEvent;

    act(() => {
      result.current.openChecklistModal();
      result.current.setNewChecklistText("  New  ");
    });

    await act(async () => {
      result.current.submitNewChecklistCriterion(ev);
    });

    await waitFor(() => {
      expect(mockAdd).toHaveBeenCalledWith(TASK_A, "New", { verify_commands: [] });
    });
    expect(inv).toHaveBeenCalled();
    await waitFor(() => {
      expect(result.current.checklistModalOpen).toBe(false);
    });
  });

  it("submitEditChecklistCriterion patches and closes edit modal", async () => {
    const { Wrapper, queryClient } = setupChecklistTest();
    const inv = vi.spyOn(queryClient, "invalidateQueries");
    const { result } = renderHook(() => useTaskDetailChecklist(TASK_A), {
      wrapper: Wrapper,
    });

    const ev = { preventDefault: vi.fn() } as unknown as FormEvent;

    act(() => {
      result.current.openEditCriterionModal(ITEM_ID, "a");
      result.current.setEditChecklistText("  b  ");
    });

    await act(async () => {
      result.current.submitEditChecklistCriterion(ev);
    });

    await waitFor(() => {
      expect(mockPatch).toHaveBeenCalledWith(TASK_A, ITEM_ID, "b");
    });
    expect(inv).toHaveBeenCalled();
    await waitFor(() => {
      expect(result.current.editCriterionModalOpen).toBe(false);
    });
  });

  it("optimistically appends a synthetic checklist item on submit", async () => {
    let resolveFn: (() => void) | undefined;
    mockAdd.mockImplementationOnce(
      () => new Promise<void>((resolve) => { resolveFn = resolve; }),
    );
    const { Wrapper, queryClient } = setupChecklistTest();
    queryClient.setQueryData<TaskChecklistResponse>(taskQueryKeys.checklist(TASK_A), {
      items: [{ id: "i1", sort_order: 0, text: "existing", done: false }],
    });
    const { result } = renderHook(() => useTaskDetailChecklist(TASK_A), {
      wrapper: Wrapper,
    });
    act(() => {
      result.current.openChecklistModal();
      result.current.setNewChecklistText("new criterion");
    });
    act(() => {
      const ev = { preventDefault: vi.fn() } as unknown as FormEvent;
      result.current.submitNewChecklistCriterion(ev);
    });
    await waitFor(() => {
      expect(result.current.addChecklistMutation.isPending).toBe(true);
    });
    const cached = queryClient.getQueryData<TaskChecklistResponse>(taskQueryKeys.checklist(TASK_A));
    expect(cached?.items).toHaveLength(2);
    expect(cached?.items[1]?.text).toBe("new criterion");
    expect(cached?.items[1]?.id.startsWith("optimistic-")).toBe(true);
    act(() => {
      resolveFn?.();
    });
    await waitFor(() => {
      expect(result.current.addChecklistMutation.isPending).toBe(false);
    });
  });

  it("optimistically updates checklist item text on edit", async () => {
    let resolveFn: ((v: unknown) => void) | undefined;
    mockPatch.mockImplementationOnce(
      () => new Promise((resolve) => { resolveFn = resolve; }),
    );
    const { Wrapper, queryClient } = setupChecklistTest();
    queryClient.setQueryData<TaskChecklistResponse>(taskQueryKeys.checklist(TASK_A), {
      items: [{ id: ITEM_ID, sort_order: 0, text: "old", done: false }],
    });
    const { result } = renderHook(() => useTaskDetailChecklist(TASK_A), {
      wrapper: Wrapper,
    });
    act(() => {
      result.current.openEditCriterionModal(ITEM_ID, "old");
      result.current.setEditChecklistText("new");
    });
    act(() => {
      const ev = { preventDefault: vi.fn() } as unknown as FormEvent;
      result.current.submitEditChecklistCriterion(ev);
    });
    await waitFor(() => {
      expect(result.current.updateChecklistTextMutation.isPending).toBe(true);
    });
    const cached = queryClient.getQueryData<TaskChecklistResponse>(taskQueryKeys.checklist(TASK_A));
    expect(cached?.items[0]?.text).toBe("new");
    act(() => {
      resolveFn?.({ items: [{ id: ITEM_ID, sort_order: 0, text: "new", done: false }] });
    });
    await waitFor(() => {
      expect(result.current.updateChecklistTextMutation.isPending).toBe(false);
    });
  });

  it("optimistically removes checklist item AND invalidates detail on delete success", async () => {
    let resolveFn: (() => void) | undefined;
    mockDelete.mockImplementationOnce(
      () => new Promise<void>((resolve) => { resolveFn = resolve; }),
    );
    const { Wrapper, queryClient } = setupChecklistTest();
    queryClient.setQueryData<TaskChecklistResponse>(taskQueryKeys.checklist(TASK_A), {
      items: [
        { id: ITEM_ID, sort_order: 0, text: "doomed", done: false },
        { id: "keep", sort_order: 1, text: "keep", done: false },
      ],
    });
    const inv = vi.spyOn(queryClient, "invalidateQueries");
    const { result } = renderHook(() => useTaskDetailChecklist(TASK_A), {
      wrapper: Wrapper,
    });
    act(() => {
      result.current.deleteChecklistMutation.mutate(ITEM_ID);
    });
    await waitFor(() => {
      expect(result.current.deleteChecklistMutation.isPending).toBe(true);
    });
    const cached = queryClient.getQueryData<TaskChecklistResponse>(taskQueryKeys.checklist(TASK_A));
    expect(cached?.items.map((i) => i.id)).toEqual(["keep"]);
    act(() => {
      resolveFn?.();
    });
    await waitFor(() => {
      expect(result.current.deleteChecklistMutation.isPending).toBe(false);
    });
    expect(inv).toHaveBeenCalledWith({ queryKey: taskQueryKeys.checklist(TASK_A) });
    expect(inv).toHaveBeenCalledWith({ queryKey: taskQueryKeys.detail(TASK_A) });
  });

  it("rolls back the optimistic add on server error", async () => {
    mockAdd.mockRejectedValueOnce(new Error("server says no"));
    const { Wrapper, queryClient } = setupChecklistTest();
    queryClient.setQueryData<TaskChecklistResponse>(taskQueryKeys.checklist(TASK_A), {
      items: [{ id: "i1", sort_order: 0, text: "existing", done: false }],
    });
    const { result } = renderHook(() => useTaskDetailChecklist(TASK_A), {
      wrapper: Wrapper,
    });
    act(() => {
      result.current.openChecklistModal();
      result.current.setNewChecklistText("doomed");
    });
    act(() => {
      const ev = { preventDefault: vi.fn() } as unknown as FormEvent;
      result.current.submitNewChecklistCriterion(ev);
    });
    await waitFor(() => {
      expect(result.current.addChecklistMutation.isError).toBe(true);
    });
    const cached = queryClient.getQueryData<TaskChecklistResponse>(taskQueryKeys.checklist(TASK_A));
    expect(cached?.items.map((i) => i.id)).toEqual(["i1"]);
  });

  it("bumps the optimistic-version counter so SSE echoes are suppressed in flight", async () => {
    let resolveFn: (() => void) | undefined;
    mockAdd.mockImplementationOnce(
      () =>
        new Promise<void>((resolve) => {
          resolveFn = resolve;
        }),
    );
    const { Wrapper } = setupChecklistTest();
    const { result } = renderHook(() => useTaskDetailChecklist(TASK_A), {
      wrapper: Wrapper,
    });
    act(() => {
      result.current.openChecklistModal();
      result.current.setNewChecklistText("new criterion");
    });
    act(() => {
      const ev = { preventDefault: vi.fn() } as unknown as FormEvent;
      result.current.submitNewChecklistCriterion(ev);
    });
    await waitFor(() => {
      expect(result.current.addChecklistMutation.isPending).toBe(true);
    });
    expect(shouldSuppressTaskMutationEcho(TASK_A)).toBe(true);
    expect(shouldSuppressTaskMutationEcho(TASK_B)).toBe(false);
    act(() => {
      resolveFn?.();
    });
    await waitFor(() => {
      expect(result.current.addChecklistMutation.isPending).toBe(false);
    });
    expect(shouldSuppressTaskMutationEcho(TASK_A)).toBe(false);
  });
});
