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
import { mockPatch, mockPatchVerify } from "./useTaskDetailChecklist.testMocks";

describe("useTaskDetailChecklist verify commands", () => {
  beforeEach(() => {
    resetChecklistMocks();
  });

  describe("updateChecklistVerifyCommandsMutation", () => {
    it("verify-only submit patches commands, invalidates checklist + detail, closes modal", async () => {
      const { Wrapper, queryClient } = setupChecklistTest();
      const inv = vi.spyOn(queryClient, "invalidateQueries");
      const { result } = renderHook(() => useTaskDetailChecklist(TASK_A), {
        wrapper: Wrapper,
      });

      const ev = { preventDefault: vi.fn() } as unknown as FormEvent;
      const newCommands = [{ command: "npm test", expected_outcome: "pass" }];

      act(() => {
        result.current.openEditCriterionModal(ITEM_ID, "criterion", []);
        result.current.setEditChecklistVerifyCommands(newCommands);
      });

      await act(async () => {
        result.current.submitEditChecklistCriterion(ev);
      });

      await waitFor(() => {
        expect(mockPatchVerify).toHaveBeenCalledWith(TASK_A, ITEM_ID, newCommands);
      });
      expect(mockPatch).not.toHaveBeenCalled();
      await waitFor(() => {
        const keys = inv.mock.calls.map((call) => call[0]?.queryKey);
        expect(keys).toEqual(
          expect.arrayContaining([
            taskQueryKeys.checklist(TASK_A),
            taskQueryKeys.detail(TASK_A),
          ]),
        );
      });
      await waitFor(() => {
        expect(result.current.editCriterionModalOpen).toBe(false);
      });
    });

    it("commands-only no-op closes modal without API call", async () => {
      const { Wrapper } = setupChecklistTest();
      const { result } = renderHook(() => useTaskDetailChecklist(TASK_A), {
        wrapper: Wrapper,
      });

      const ev = { preventDefault: vi.fn() } as unknown as FormEvent;
      const commands = [{ command: "npm test" }];

      act(() => {
        result.current.openEditCriterionModal(ITEM_ID, "criterion", commands);
      });

      await act(async () => {
        result.current.submitEditChecklistCriterion(ev);
      });

      expect(mockPatchVerify).not.toHaveBeenCalled();
      expect(mockPatch).not.toHaveBeenCalled();
      await waitFor(() => {
        expect(result.current.editCriterionModalOpen).toBe(false);
      });
    });

    it("optimistically updates verify_commands on verify-only edit", async () => {
      let resolveFn: ((v: unknown) => void) | undefined;
      mockPatchVerify.mockImplementationOnce(
        () => new Promise((resolve) => { resolveFn = resolve; }),
      );
      const { Wrapper, queryClient } = setupChecklistTest();
      const newCommands = [{ command: "go test ./..." }];
      queryClient.setQueryData<TaskChecklistResponse>(taskQueryKeys.checklist(TASK_A), {
        items: [{ id: ITEM_ID, sort_order: 0, text: "criterion", done: false, verify_commands: [] }],
      });
      const { result } = renderHook(() => useTaskDetailChecklist(TASK_A), {
        wrapper: Wrapper,
      });

      act(() => {
        result.current.openEditCriterionModal(ITEM_ID, "criterion", []);
        result.current.setEditChecklistVerifyCommands(newCommands);
      });
      act(() => {
        const ev = { preventDefault: vi.fn() } as unknown as FormEvent;
        result.current.submitEditChecklistCriterion(ev);
      });
      await waitFor(() => {
        expect(result.current.updateChecklistVerifyCommandsMutation.isPending).toBe(true);
      });
      const cached = queryClient.getQueryData<TaskChecklistResponse>(taskQueryKeys.checklist(TASK_A));
      expect(cached?.items[0]?.verify_commands).toEqual(newCommands);
      act(() => {
        resolveFn?.({
          items: [{
            id: ITEM_ID,
            sort_order: 0,
            text: "criterion",
            done: false,
            verify_commands: newCommands,
          }],
        });
      });
      await waitFor(() => {
        expect(result.current.updateChecklistVerifyCommandsMutation.isPending).toBe(false);
      });
    });

    it("suppresses SSE echoes during verify mutation flight", async () => {
      let resolveFn: ((v: unknown) => void) | undefined;
      mockPatchVerify.mockImplementationOnce(
        () => new Promise((resolve) => { resolveFn = resolve; }),
      );
      const { Wrapper } = setupChecklistTest();
      const { result } = renderHook(() => useTaskDetailChecklist(TASK_A), {
        wrapper: Wrapper,
      });

      act(() => {
        result.current.openEditCriterionModal(ITEM_ID, "criterion", []);
        result.current.setEditChecklistVerifyCommands([{ command: "npm test" }]);
      });
      act(() => {
        const ev = { preventDefault: vi.fn() } as unknown as FormEvent;
        result.current.submitEditChecklistCriterion(ev);
      });
      await waitFor(() => {
        expect(result.current.updateChecklistVerifyCommandsMutation.isPending).toBe(true);
      });
      expect(shouldSuppressTaskMutationEcho(TASK_A)).toBe(true);
      expect(shouldSuppressTaskMutationEcho(TASK_B)).toBe(false);
      act(() => {
        resolveFn?.({ items: [] });
      });
      await waitFor(() => {
        expect(result.current.updateChecklistVerifyCommandsMutation.isPending).toBe(false);
      });
      expect(shouldSuppressTaskMutationEcho(TASK_A)).toBe(false);
    });

    it("rolls back optimistic verify_commands on server error", async () => {
      mockPatchVerify.mockRejectedValueOnce(new Error("server says no"));
      const { Wrapper, queryClient } = setupChecklistTest();
      const originalCommands = [{ command: "old cmd" }];
      queryClient.setQueryData<TaskChecklistResponse>(taskQueryKeys.checklist(TASK_A), {
        items: [{
          id: ITEM_ID,
          sort_order: 0,
          text: "criterion",
          done: false,
          verify_commands: originalCommands,
        }],
      });
      const { result } = renderHook(() => useTaskDetailChecklist(TASK_A), {
        wrapper: Wrapper,
      });

      act(() => {
        result.current.openEditCriterionModal(ITEM_ID, "criterion", originalCommands);
        result.current.setEditChecklistVerifyCommands([{ command: "new cmd" }]);
      });
      act(() => {
        const ev = { preventDefault: vi.fn() } as unknown as FormEvent;
        result.current.submitEditChecklistCriterion(ev);
      });
      await waitFor(() => {
        expect(result.current.updateChecklistVerifyCommandsMutation.isError).toBe(true);
      });
      const cached = queryClient.getQueryData<TaskChecklistResponse>(taskQueryKeys.checklist(TASK_A));
      expect(cached?.items[0]?.verify_commands).toEqual(originalCommands);
    });

    it("combined text + verify calls both mocks in order", async () => {
      const { Wrapper } = setupChecklistTest();
      const { result } = renderHook(() => useTaskDetailChecklist(TASK_A), {
        wrapper: Wrapper,
      });

      const ev = { preventDefault: vi.fn() } as unknown as FormEvent;
      const newCommands = [{ command: "npm test" }];

      act(() => {
        result.current.openEditCriterionModal(ITEM_ID, "old text", []);
        result.current.setEditChecklistText("new text");
        result.current.setEditChecklistVerifyCommands(newCommands);
      });

      await act(async () => {
        result.current.submitEditChecklistCriterion(ev);
      });

      await waitFor(() => {
        expect(mockPatch).toHaveBeenCalledWith(TASK_A, ITEM_ID, "new text");
        expect(mockPatchVerify).toHaveBeenCalledWith(TASK_A, ITEM_ID, newCommands);
      });
      expect(mockPatch.mock.invocationCallOrder[0]).toBeLessThan(
        mockPatchVerify.mock.invocationCallOrder[0],
      );
    });
  });
});
