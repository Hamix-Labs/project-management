import { useQueryClient } from "@tanstack/react-query";
import type { MutableRefObject } from "react";
import { useTaskCreateDraftMutations } from "./useTaskCreateDraftMutations";
import { useTaskCreateMutationResets } from "./useTaskCreateMutationResets";
import { useTaskCreateTaskMutations } from "./useTaskCreateTaskMutations";
import { useTaskCreateTemplateMutations } from "./useTaskCreateTemplateMutations";

export function useTaskCreateMutations(input: {
  queryClient: ReturnType<typeof useQueryClient>;
  newDraftIDRef: MutableRefObject<string>;
  newDraftID: string;
  closeCreateModal: () => void;
  setNewDraftID: (id: string) => void;
  setDraftAutosaveBaseline: (baseline: string) => void;
  setDraftAutosaveBaselineID: (id: string) => void;
  setLastDraftSavedAt: (timestamp: number | null) => void;
  createModalOpen: boolean;
  editingTemplateId: string | null;
}) {
  const { createMutation, instantiateTemplatesMutation } = useTaskCreateTaskMutations({
    queryClient: input.queryClient,
    newDraftIDRef: input.newDraftIDRef,
    closeCreateModal: input.closeCreateModal,
  });

  const { saveDraftMutation, deleteDraftMutation, resumeDraftMutation } =
    useTaskCreateDraftMutations({
      queryClient: input.queryClient,
      newDraftIDRef: input.newDraftIDRef,
      newDraftID: input.newDraftID,
      setNewDraftID: input.setNewDraftID,
      setDraftAutosaveBaseline: input.setDraftAutosaveBaseline,
      setDraftAutosaveBaselineID: input.setDraftAutosaveBaselineID,
      setLastDraftSavedAt: input.setLastDraftSavedAt,
      createModalOpen: input.createModalOpen,
    });

  const {
    saveTemplateMutation,
    patchTemplateMutation,
    loadTemplateMutation,
    deleteTemplateMutation,
  } = useTaskCreateTemplateMutations({
    queryClient: input.queryClient,
    closeCreateModal: input.closeCreateModal,
    editingTemplateId: input.editingTemplateId,
  });

  useTaskCreateMutationResets({
    createModalOpen: input.createModalOpen,
    createMutation,
    saveDraftMutation,
    saveTemplateMutation,
    patchTemplateMutation,
  });

  return {
    createMutation,
    saveDraftMutation,
    deleteDraftMutation,
    resumeDraftMutation,
    saveTemplateMutation,
    patchTemplateMutation,
    loadTemplateMutation,
    deleteTemplateMutation,
    instantiateTemplatesMutation,
  };
}
