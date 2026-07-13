import type { QueryClient } from "@tanstack/react-query";
import type { TaskDraftsQuery } from "../types";
import type { useTaskCreateFormState } from "./useTaskCreateFormState";
import { useTaskCreateComposeEntryActions } from "./useTaskCreateComposeEntryActions";
import { useTaskCreateDraftEntryActions } from "./useTaskCreateDraftEntryActions";
import type { useTaskCreateModalState } from "./useTaskCreateModalState";
import type { useTaskCreateMutations } from "./useTaskCreateMutations";
import { useTaskCreateTemplateEntryActions } from "./useTaskCreateTemplateEntryActions";

export function useTaskCreateEntryActions(input: {
  form: ReturnType<typeof useTaskCreateFormState>;
  modal: ReturnType<typeof useTaskCreateModalState>;
  mutations: ReturnType<typeof useTaskCreateMutations>;
  draftsQuery: TaskDraftsQuery;
  queryClient: QueryClient;
}) {
  const compose = useTaskCreateComposeEntryActions({
    modal: input.modal,
    draftsQuery: input.draftsQuery,
    queryClient: input.queryClient,
  });

  const draft = useTaskCreateDraftEntryActions({
    form: input.form,
    modal: input.modal,
    mutations: input.mutations,
    draftsQuery: input.draftsQuery,
    queryClient: input.queryClient,
  });

  const template = useTaskCreateTemplateEntryActions({
    form: input.form,
    modal: input.modal,
    mutations: input.mutations,
    queryClient: input.queryClient,
  });

  return {
    openCreateModal: compose.openCreateModal,
    openComposeModal: compose.openComposeModal,
    openTemplateCreateModal: compose.openTemplateCreateModal,
    startFreshDraft: compose.startFreshDraft,
    resumeDraftByID: draft.resumeDraftByID,
    deleteDraftByID: draft.deleteDraftByID,
    editTemplateByID: template.editTemplateByID,
    deleteTemplateByID: template.deleteTemplateByID,
    instantiateTemplates: template.instantiateTemplates,
    retryDraftList: draft.retryDraftList,
    retryCreateEntryDraftLoad: draft.retryCreateEntryDraftLoad,
  };
}
