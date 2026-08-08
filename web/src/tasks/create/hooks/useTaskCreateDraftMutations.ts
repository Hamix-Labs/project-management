import { useMutation } from "@tanstack/react-query";
import { useCallback, useEffect, useRef, type MutableRefObject } from "react";
import {
  deleteTaskDraft as apiDeleteDraft,
  getTaskDraft as apiGetDraft,
  saveTaskDraft as apiSaveDraft,
} from "@/api";
import { isAbortError } from "@/lib/isAbortError";
import type { DraftSavePayload } from "@/types";
import { invalidateTaskCacheAsync } from "@/tasks/mutations";

export function useTaskCreateDraftMutations(input: {
  queryClient: import("@tanstack/react-query").QueryClient;
  newDraftIDRef: MutableRefObject<string>;
  newDraftID: string;
  setNewDraftID: (id: string) => void;
  setDraftAutosaveBaseline: (baseline: string) => void;
  setDraftAutosaveBaselineID: (id: string) => void;
  setLastDraftSavedAt: (timestamp: number | null) => void;
  createModalOpen: boolean;
}) {
  const saveAbortRef = useRef<AbortController | null>(null);

  const cancelInFlightSave = useCallback(() => {
    saveAbortRef.current?.abort();
    saveAbortRef.current = null;
  }, []);

  const saveDraftMutation = useMutation({
    mutationFn: async (mutationInput: DraftSavePayload & { signature: string }) => {
      // Supersede any prior in-flight save (draft change or rapid autosave).
      cancelInFlightSave();
      const ac = new AbortController();
      saveAbortRef.current = ac;
      try {
        return await apiSaveDraft({
          id: mutationInput.id,
          name: mutationInput.name,
          payload: mutationInput.payload,
          signal: ac.signal,
        });
      } finally {
        if (saveAbortRef.current === ac) {
          saveAbortRef.current = null;
        }
      }
    },
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

  // Abort in-flight save when the modal closes.
  useEffect(() => {
    if (!input.createModalOpen) {
      cancelInFlightSave();
    }
  }, [input.createModalOpen, cancelInFlightSave]);

  // Abort when the active draft id changes (resume / fresh) and on unmount.
  useEffect(() => {
    return () => {
      cancelInFlightSave();
    };
  }, [input.newDraftID, cancelInFlightSave]);

  // Aborted saves must not linger as mutation errors (UI treats them as no-ops).
  useEffect(() => {
    if (
      saveDraftMutation.isError &&
      isAbortError(saveDraftMutation.error)
    ) {
      saveDraftMutation.reset();
    }
  }, [saveDraftMutation]);

  return {
    saveDraftMutation,
    deleteDraftMutation,
    resumeDraftMutation,
    cancelInFlightSave,
  };
}
