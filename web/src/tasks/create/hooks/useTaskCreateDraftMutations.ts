import { useMutation } from "@tanstack/react-query";
import type { MutableRefObject } from "react";
import {
  deleteTaskDraft as apiDeleteDraft,
  getTaskDraft as apiGetDraft,
  saveTaskDraft as apiSaveDraft,
} from "@/api";
import type { PriorityChoice } from "@/types";
import { invalidateTaskCacheAsync } from "@/tasks/mutations";

export function useTaskCreateDraftMutations(input: {
  queryClient: import("@tanstack/react-query").QueryClient;
  newDraftIDRef: MutableRefObject<string>;
  newDraftID: string;
  setNewDraftID: (id: string) => void;
  setDraftAutosaveBaseline: (baseline: string) => void;
  setDraftAutosaveBaselineID: (id: string) => void;
  setLastDraftSavedAt: (timestamp: number | null) => void;
}) {
  const saveDraftMutation = useMutation({
    mutationFn: (mutationInput: {
      id: string;
      name: string;
      payload: {
        title: string;
        initial_prompt: string;
        priority: PriorityChoice;
        runner: string;
        cursor_model: string;
        project_id: string;
        project_context_item_ids: string[];
        checklist_items: import("@/types").TaskDraftChecklistItem[];
      };
      signature: string;
    }) => apiSaveDraft(mutationInput),
    onSuccess: async (saved, variables) => {
      // I2 - baseline stamp only when save response matches active draft ref.
      if (input.newDraftIDRef.current !== saved.id) {
        await invalidateTaskCacheAsync(input.queryClient, { scope: "drafts" });
        return;
      }
      if (saved.id !== input.newDraftID) {
        input.setNewDraftID(saved.id);
      }
      input.setDraftAutosaveBaseline(variables.signature);
      input.setDraftAutosaveBaselineID(saved.id);
      input.setLastDraftSavedAt(Date.now());
      await invalidateTaskCacheAsync(input.queryClient, { scope: "drafts" });
    },
  });

  const deleteDraftMutation = useMutation({
    mutationFn: (id: string) => apiDeleteDraft(id),
    onSuccess: async () => {
      await invalidateTaskCacheAsync(input.queryClient, { scope: "drafts" });
    },
  });

  const resumeDraftMutation = useMutation({
    mutationFn: (id: string) => apiGetDraft(id),
  });

  return { saveDraftMutation, deleteDraftMutation, resumeDraftMutation };
}
