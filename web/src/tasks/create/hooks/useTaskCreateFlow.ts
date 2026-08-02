import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useLayoutEffect, useMemo } from "react";
import { listTaskDrafts as apiListDrafts } from "@/api";
import { TASK_DRAFTS } from "@/constants/tasks";
import { taskQueryKeys } from "../../task-query";
import { computeDraftAutosaveSignature } from "../draftPayload";
import { deriveCreateFlowError, mapCreateFlowViewModel } from "../mapCreateFlowViewModel";
import { normalizeDraftPromptForDirty } from "../../task-drafts";
import { useTaskCreateChecklistActions } from "./useTaskCreateChecklistActions";
import { useTaskCreateDraftAutosave } from "./useTaskCreateDraftAutosave";
import { useTaskCreateEntryActions } from "./useTaskCreateEntryActions";
import { useTaskCreateFormState } from "./useTaskCreateFormState";
import { useTaskCreateModalState } from "./useTaskCreateModalState";
import { useTaskCreateMutations } from "./useTaskCreateMutations";
import { useTaskCreateSubmitActions } from "./useTaskCreateSubmitActions";
import { createOpenComposePromptEditor } from "../../prompt-editor/useOpenPromptEditor";

/**
 * Create-task modal, draft autosave, draft picker, and related mutations.
 * Composed by `useTasksApp`.
 */
export function useTaskCreateFlow() {
  const queryClient = useQueryClient();
  const form = useTaskCreateFormState(queryClient);
  const modal = useTaskCreateModalState(
    form.resetFormFields,
    form.populateFromTask,
    form.setNewChecklistItems,
    form.setNewProjectID,
    form.setNewRepositoryID,
    form.setNewWorktreeID,
  );
  const draftsQuery = useQuery({
    queryKey: taskQueryKeys.drafts(),
    queryFn: ({ signal }) =>
      apiListDrafts(TASK_DRAFTS.createModalDraftListLimit, { signal }),
  });
  const mutations = useTaskCreateMutations({
    queryClient,
    newDraftIDRef: form.newDraftIDRef,
    newDraftID: form.newDraftID,
    closeCreateModal: modal.closeCreateModal,
    setNewDraftID: form.setNewDraftID,
    setDraftAutosaveBaseline: form.setDraftAutosaveBaseline,
    setDraftAutosaveBaselineID: form.setDraftAutosaveBaselineID,
    setLastDraftSavedAt: form.setLastDraftSavedAt,
    // Keep draft saves alive while Prompt Editor is open (compose is suspended).
    createModalOpen: modal.createModalOpen || modal.promptEditorSuspended,
    editingTemplateId: modal.editingTemplateId,
  });

  // Repo/project/worktree cascade pre-selects system defaults on open. That is not
  // operator input — fold it into the autosave baseline so an untouched modal stays clean.
  // useLayoutEffect runs before draft autosave's useEffect on the same commit.
  useLayoutEffect(() => {
    if (!modal.createModalOpen || modal.editingTaskId || modal.composeTarget !== "task") {
      return;
    }
    if (form.draftAutosaveBaselineID !== form.newDraftID) return;
    if (!form.newProjectID) return;
    const hasUserContent =
      form.newTitle.trim() !== "" ||
      form.newPriority !== "" ||
      form.newChecklistItems.length > 0 ||
      normalizeDraftPromptForDirty(form.newPrompt) !== "";
    if (hasUserContent) return;
    const sig = computeDraftAutosaveSignature(form.formFields);
    if (sig !== form.draftAutosaveBaseline) {
      form.setDraftAutosaveBaseline(sig);
    }
  }, [
    form.draftAutosaveBaseline,
    form.draftAutosaveBaselineID,
    form.formFields,
    form.newChecklistItems.length,
    form.newDraftID,
    form.newPriority,
    form.newProjectID,
    form.newPrompt,
    form.newTitle,
    form.setDraftAutosaveBaseline,
    modal.composeTarget,
    modal.createModalOpen,
    modal.editingTaskId,
  ]);

  const autosave = useTaskCreateDraftAutosave({
    formFields: form.formFields,
    draftAutosaveBaseline: form.draftAutosaveBaseline,
    draftAutosaveBaselineID: form.draftAutosaveBaselineID,
    editingTaskId: modal.editingTaskId,
    composeTarget: modal.composeTarget,
    createModalOpen: modal.createModalOpen,
    autosaveTimerRef: form.autosaveTimerRef,
    saveDraftMutation: mutations.saveDraftMutation,
    lastDraftSavedAt: form.lastDraftSavedAt,
  });
  const entryActions = useTaskCreateEntryActions({
    form,
    modal,
    mutations,
    draftsQuery,
    queryClient,
  });
  const submitActions = useTaskCreateSubmitActions({ form, modal, mutations });
  const checklistActions = useTaskCreateChecklistActions({ form });
  const openPromptEditor = useMemo(
    () => createOpenComposePromptEditor({ form, modal, mutations }),
    [form, modal, mutations],
  );

  const actions = { ...entryActions, ...submitActions, ...checklistActions };
  const createFlowError = useMemo(
    () => deriveCreateFlowError(mutations),
    [mutations],
  );

  return mapCreateFlowViewModel({
    createFlowError,
    form,
    modal,
    mutations,
    autosave,
    actions,
    draftsQuery,
    openPromptEditor,
  });
}
