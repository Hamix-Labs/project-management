import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useCallback } from "react";
import type { FormEvent } from "react";
import type { ChecklistVerifyCommandInput } from "@/types";
import { useOptionalToast } from "@/shared/toast";
import { useRolloutFlags } from "@/settings";
import type { TaskChecklistResponse } from "@/types";
import {
  buildAddChecklistMutationOptions,
  buildDeleteChecklistMutationOptions,
  buildUpdateChecklistTextMutationOptions,
  buildUpdateChecklistVerifyCommandsMutationOptions,
  createSubmitNewChecklistCriterionHandler,
  submitEditChecklistCriterionForm,
} from "../checklistMutations";
import type {
  ChecklistMutationDeps,
  ChecklistOptimisticContext,
} from "../checklistOptimistic";
import { useChecklistModalControls } from "../useChecklistModalControls";

export function useTaskDetailChecklist(taskId: string) {
  const queryClient = useQueryClient();
  const toast = useOptionalToast();
  const { optimisticMutationsEnabled } = useRolloutFlags();
  const modal = useChecklistModalControls(taskId);

  const mutationDeps: ChecklistMutationDeps = {
    taskId,
    queryClient,
    toast,
    optimisticMutationsEnabled,
  };

  const addChecklistMutation = useMutation<
    void,
    unknown,
    { text: string; verify_commands: ChecklistVerifyCommandInput[]; submissionToken: number },
    ChecklistOptimisticContext
  >(
    buildAddChecklistMutationOptions({
      ...mutationDeps,
      addSubmissionTokenRef: modal.addSubmissionTokenRef,
      setNewChecklistText: modal.setNewChecklistText,
      setNewChecklistVerifyCommands: modal.setNewChecklistVerifyCommands,
      setChecklistModalOpen: modal.setChecklistModalOpen,
    }),
  );

  const submitNewChecklistCriterion = useCallback(
    createSubmitNewChecklistCriterionHandler(modal, addChecklistMutation),
    [modal, addChecklistMutation],
  );

  const updateChecklistTextMutation = useMutation<
    TaskChecklistResponse,
    unknown,
    { itemId: string; text: string },
    ChecklistOptimisticContext
  >(buildUpdateChecklistTextMutationOptions(mutationDeps));

  const updateChecklistVerifyCommandsMutation = useMutation<
    TaskChecklistResponse,
    unknown,
    { itemId: string; verify_commands: ChecklistVerifyCommandInput[] },
    ChecklistOptimisticContext
  >(buildUpdateChecklistVerifyCommandsMutationOptions(mutationDeps));

  const submitEditChecklistCriterion = useCallback(
    (e: FormEvent) =>
      submitEditChecklistCriterionForm(e, {
        modal,
        updateChecklistTextMutation,
        updateChecklistVerifyCommandsMutation,
        closeEditCriterionModal: modal.closeEditCriterionModal,
      }),
    [modal, updateChecklistTextMutation, updateChecklistVerifyCommandsMutation],
  );

  const deleteChecklistMutation = useMutation<
    void,
    unknown,
    string,
    ChecklistOptimisticContext
  >(buildDeleteChecklistMutationOptions(mutationDeps));

  return {
    checklistModalOpen: modal.checklistModalOpen,
    newChecklistText: modal.newChecklistText,
    setNewChecklistText: modal.setNewChecklistText,
    newChecklistVerifyCommands: modal.newChecklistVerifyCommands,
    setNewChecklistVerifyCommands: modal.setNewChecklistVerifyCommands,
    editCriterionModalOpen: modal.editCriterionModalOpen,
    editingChecklistItemId: modal.editingChecklistItemId,
    editChecklistText: modal.editChecklistText,
    setEditChecklistText: modal.setEditChecklistText,
    editChecklistVerifyCommands: modal.editChecklistVerifyCommands,
    setEditChecklistVerifyCommands: modal.setEditChecklistVerifyCommands,
    closeChecklistModal: modal.closeChecklistModal,
    closeEditCriterionModal: modal.closeEditCriterionModal,
    openChecklistModal: modal.openChecklistModal,
    openEditCriterionModal: modal.openEditCriterionModal,
    addChecklistMutation,
    submitNewChecklistCriterion,
    updateChecklistTextMutation,
    updateChecklistVerifyCommandsMutation,
    submitEditChecklistCriterion,
    deleteChecklistMutation,
  };
}
