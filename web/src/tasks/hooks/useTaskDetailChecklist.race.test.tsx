import "./useTaskDetailChecklist.testMocks";
import { act, renderHook, waitFor } from "@testing-library/react";
import type { FormEvent } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { taskQueryKeys } from "../task-query";
import { useTaskDetailChecklist } from "../checklist/hooks/useTaskDetailChecklist";
import {
  ITEM_ID,
  resetChecklistMocks,
  setupChecklistTest,
  TASK_A,
} from "./useTaskDetailChecklist.testHelpers";
import { mockAdd, mockPatch } from "./useTaskDetailChecklist.testMocks";

describe("useTaskDetailChecklist race guards", () => {
  beforeEach(() => {
    resetChecklistMocks();
  });

  describe("addChecklistMutation race", () => {
    it("drops the form-clear + modal-close branch when the user dismissed and reopened mid-flight", async () => {
      const { Wrapper, queryClient } = setupChecklistTest();
      const inv = vi.spyOn(queryClient, "invalidateQueries");
      const { result } = renderHook(() => useTaskDetailChecklist(TASK_A), {
        wrapper: Wrapper,
      });

      let resolveA: ((value: unknown) => void) | undefined;
      mockAdd.mockImplementationOnce(
        () =>
          new Promise((resolve) => {
            resolveA = resolve;
          }),
      );

      const ev = { preventDefault: vi.fn() } as unknown as FormEvent;

      act(() => {
        result.current.openChecklistModal();
        result.current.setNewChecklistText("Criterion A");
      });
      await act(async () => {
        result.current.submitNewChecklistCriterion(ev);
        await Promise.resolve();
      });
      await waitFor(() => {
        expect(result.current.addChecklistMutation.isPending).toBe(true);
      });

      act(() => {
        result.current.closeChecklistModal();
        result.current.openChecklistModal();
        result.current.setNewChecklistText("Criterion B");
      });
      expect(result.current.checklistModalOpen).toBe(true);
      expect(result.current.newChecklistText).toBe("Criterion B");

      await act(async () => {
        resolveA?.({
          id: ITEM_ID,
          task_id: TASK_A,
          text: "Criterion A",
          done: false,
        });
        await Promise.resolve();
      });

      await waitFor(() => {
        const keys = inv.mock.calls.map((call) => call[0]?.queryKey);
        expect(keys).toEqual(
          expect.arrayContaining([
            taskQueryKeys.checklist(TASK_A),
            taskQueryKeys.detail(TASK_A),
          ]),
        );
      });
      expect(result.current.checklistModalOpen).toBe(true);
      expect(result.current.newChecklistText).toBe("Criterion B");
    });

    it("happy path: in-flight resolution closes the add modal and clears the text", async () => {
      const { Wrapper } = setupChecklistTest();
      const { result } = renderHook(() => useTaskDetailChecklist(TASK_A), {
        wrapper: Wrapper,
      });

      const ev = { preventDefault: vi.fn() } as unknown as FormEvent;

      act(() => {
        result.current.openChecklistModal();
        result.current.setNewChecklistText("Sole");
      });

      await act(async () => {
        result.current.submitNewChecklistCriterion(ev);
      });

      await waitFor(() => {
        expect(result.current.checklistModalOpen).toBe(false);
      });
      expect(result.current.newChecklistText).toBe("");
    });
  });

  describe("updateChecklistTextMutation race", () => {
    it("drops closeEditCriterionModal() when the user reopened the edit modal on a different item mid-flight", async () => {
      const otherItemId = "44444444-4444-4444-8444-444444444444";
      const { Wrapper, queryClient } = setupChecklistTest();
      const inv = vi.spyOn(queryClient, "invalidateQueries");
      const { result } = renderHook(() => useTaskDetailChecklist(TASK_A), {
        wrapper: Wrapper,
      });

      let resolveA: ((value: unknown) => void) | undefined;
      mockPatch.mockImplementationOnce(
        () =>
          new Promise((resolve) => {
            resolveA = resolve;
          }),
      );

      const ev = { preventDefault: vi.fn() } as unknown as FormEvent;

      act(() => {
        result.current.openEditCriterionModal(ITEM_ID, "old A");
        result.current.setEditChecklistText("new A");
      });
      await act(async () => {
        result.current.submitEditChecklistCriterion(ev);
        await Promise.resolve();
      });
      await waitFor(() => {
        expect(result.current.updateChecklistTextMutation.isPending).toBe(true);
      });

      act(() => {
        result.current.openEditCriterionModal(otherItemId, "old B");
      });
      expect(result.current.editingChecklistItemId).toBe(otherItemId);
      expect(result.current.editChecklistText).toBe("old B");

      await act(async () => {
        resolveA?.({
          id: ITEM_ID,
          task_id: TASK_A,
          text: "new A",
          done: false,
        });
        await Promise.resolve();
      });

      await waitFor(() => {
        const keys = inv.mock.calls.map((call) => call[0]?.queryKey);
        expect(keys).toEqual(
          expect.arrayContaining([
            taskQueryKeys.checklist(TASK_A),
            taskQueryKeys.detail(TASK_A),
          ]),
        );
      });
      expect(result.current.editCriterionModalOpen).toBe(true);
      expect(result.current.editingChecklistItemId).toBe(otherItemId);
      expect(result.current.editChecklistText).toBe("old B");
    });

    it("happy path: in-flight resolution closes the edit modal", async () => {
      const { Wrapper } = setupChecklistTest();
      const { result } = renderHook(() => useTaskDetailChecklist(TASK_A), {
        wrapper: Wrapper,
      });

      const ev = { preventDefault: vi.fn() } as unknown as FormEvent;

      act(() => {
        result.current.openEditCriterionModal(ITEM_ID, "old");
        result.current.setEditChecklistText("new");
      });
      await act(async () => {
        result.current.submitEditChecklistCriterion(ev);
      });

      await waitFor(() => {
        expect(result.current.editCriterionModalOpen).toBe(false);
      });
      expect(result.current.editingChecklistItemId).toBeNull();
    });
  });
});
