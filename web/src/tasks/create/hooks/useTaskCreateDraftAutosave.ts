import { TASK_TIMINGS } from "@/constants/tasks";
import { useCallback, useEffect, useMemo, type MutableRefObject } from "react";
import { isAbortError } from "@/lib/isAbortError";
import { buildDraftSavePayload, computeDraftAutosaveSignature } from "../draftPayload";
import type { TaskCreateFormFields } from "../types";
import type { useTaskCreateMutations } from "./useTaskCreateMutations";

const DRAFT_AUTOSAVE_DEBOUNCE_MS = TASK_TIMINGS.draftAutosaveDebounceMs;

export function useTaskCreateDraftAutosave(input: {
  formFields: TaskCreateFormFields;
  draftAutosaveBaseline: string;
  draftAutosaveBaselineID: string;
  editingTaskId: string | null;
  composeTarget: "task" | "template";
  createModalOpen: boolean;
  autosaveTimerRef: MutableRefObject<ReturnType<typeof setTimeout> | null>;
  saveDraftMutation: ReturnType<typeof useTaskCreateMutations>["saveDraftMutation"];
  lastDraftSavedAt: number | null;
}) {
  const currentDraftAutosaveSignature = useMemo(
    () => computeDraftAutosaveSignature(input.formFields),
    [input.formFields],
  );

  const buildDraftSaveInput = useCallback(
    () => buildDraftSavePayload(input.formFields),
    [input.formFields],
  );

  const saveDraftNow = useCallback(() => {
    // I1 — no autosave while editing an existing task or composing a template.
    if (
      input.editingTaskId ||
      input.composeTarget !== "task" ||
      !input.createModalOpen ||
      !input.formFields.newDraftID
    ) {
      return;
    }
    if (input.draftAutosaveBaselineID !== input.formFields.newDraftID) return;
    if (currentDraftAutosaveSignature === input.draftAutosaveBaseline) return;
    if (input.autosaveTimerRef.current) {
      clearTimeout(input.autosaveTimerRef.current);
      input.autosaveTimerRef.current = null;
    }
    input.saveDraftMutation.mutate({
      ...buildDraftSaveInput(),
      signature: currentDraftAutosaveSignature,
    });
  }, [
    buildDraftSaveInput,
    currentDraftAutosaveSignature,
    input,
  ]);

  useEffect(() => {
    // I1 — no autosave while editing an existing task or composing a template.
    if (
      input.editingTaskId ||
      input.composeTarget !== "task" ||
      !input.createModalOpen ||
      !input.formFields.newDraftID
    ) {
      return;
    }
    if (input.draftAutosaveBaselineID !== input.formFields.newDraftID) return;
    if (currentDraftAutosaveSignature === input.draftAutosaveBaseline) return;
    const signatureAtSchedule = currentDraftAutosaveSignature;
    input.autosaveTimerRef.current = setTimeout(() => {
      input.saveDraftMutation.mutate({
        ...buildDraftSaveInput(),
        signature: signatureAtSchedule,
      });
      input.autosaveTimerRef.current = null;
    }, DRAFT_AUTOSAVE_DEBOUNCE_MS);
    return () => {
      if (input.autosaveTimerRef.current) {
        clearTimeout(input.autosaveTimerRef.current);
        input.autosaveTimerRef.current = null;
      }
    };
  }, [
    buildDraftSaveInput,
    currentDraftAutosaveSignature,
    input,
  ]);

  const saveFailedVisibly =
    input.saveDraftMutation.isError &&
    !isAbortError(input.saveDraftMutation.error);

  const draftSaveLabel = useMemo(() => {
    if (input.editingTaskId || input.composeTarget !== "task" || !input.createModalOpen) {
      return null;
    }
    if (input.saveDraftMutation.isPending) return "Saving draft…";
    if (saveFailedVisibly) {
      return "Draft autosave failed. You can still create the task.";
    }
    if (input.lastDraftSavedAt == null) return null;
    return "Draft saved";
  }, [
    input.createModalOpen,
    input.composeTarget,
    input.editingTaskId,
    input.lastDraftSavedAt,
    input.saveDraftMutation.isPending,
    saveFailedVisibly,
  ]);

  return {
    saveDraftNow,
    draftSaveLabel,
    draftSaveError: input.createModalOpen && saveFailedVisibly,
  };
}
