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
    () => ({
      ...buildDraftSavePayload(input.formFields),
      signature: currentDraftAutosaveSignature,
    }),
    [currentDraftAutosaveSignature, input.formFields],
  );

  // I1 — no draft writes while editing an existing task or composing a template.
  // The baseline-id check keeps a save bound to the draft it was computed from.
  const draftSaveAllowed = useCallback(() => {
    if (
      input.editingTaskId ||
      input.composeTarget !== "task" ||
      !input.createModalOpen ||
      !input.formFields.newDraftID
    ) {
      return false;
    }
    return input.draftAutosaveBaselineID === input.formFields.newDraftID;
  }, [input]);

  const cancelPendingAutosave = useCallback(() => {
    if (input.autosaveTimerRef.current) {
      clearTimeout(input.autosaveTimerRef.current);
      input.autosaveTimerRef.current = null;
    }
  }, [input.autosaveTimerRef]);

  // I8 — explicit saves are not gated on the dirty bit. The operator asked.
  const saveDraftNow = useCallback(() => {
    if (!draftSaveAllowed()) return;
    cancelPendingAutosave();
    input.saveDraftMutation.mutate(buildDraftSaveInput());
  }, [buildDraftSaveInput, cancelPendingAutosave, draftSaveAllowed, input.saveDraftMutation]);

  /** Awaitable explicit save, for callers that navigate away on completion. */
  const saveDraftNowAsync = useCallback(async () => {
    if (!draftSaveAllowed()) return;
    cancelPendingAutosave();
    await input.saveDraftMutation.mutateAsync(buildDraftSaveInput());
  }, [buildDraftSaveInput, cancelPendingAutosave, draftSaveAllowed, input.saveDraftMutation]);

  useEffect(() => {
    // I8 — implicit saves stay gated, so an untouched modal writes nothing.
    if (!draftSaveAllowed()) return;
    if (currentDraftAutosaveSignature === input.draftAutosaveBaseline) return;
    // buildDraftSaveInput carries the signature from this render, so the write
    // is stamped with the state that scheduled it, not whatever lands later.
    const scheduled = buildDraftSaveInput();
    input.autosaveTimerRef.current = setTimeout(() => {
      input.saveDraftMutation.mutate(scheduled);
      input.autosaveTimerRef.current = null;
    }, DRAFT_AUTOSAVE_DEBOUNCE_MS);
    return cancelPendingAutosave;
  }, [
    buildDraftSaveInput,
    cancelPendingAutosave,
    currentDraftAutosaveSignature,
    draftSaveAllowed,
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
    saveDraftNowAsync,
    draftSaveLabel,
    draftSaveError: input.createModalOpen && saveFailedVisibly,
  };
}
