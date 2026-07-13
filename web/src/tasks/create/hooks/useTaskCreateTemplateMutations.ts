import { useMutation } from "@tanstack/react-query";
import {
  deleteTaskTemplate as apiDeleteTemplate,
  getTaskTemplate as apiGetTemplate,
  patchTaskTemplate as apiPatchTemplate,
  saveTaskTemplate as apiSaveTemplate,
} from "@/api";
import { invalidateTaskCacheAsync } from "@/tasks/mutations";
import { buildComposePayloadFromForm } from "../composePayload";
import type { TaskCreateFormFields } from "../types";

export function useTaskCreateTemplateMutations(input: {
  queryClient: import("@tanstack/react-query").QueryClient;
  closeCreateModal: () => void;
  editingTemplateId: string | null;
}) {
  const saveTemplateMutation = useMutation({
    mutationFn: (mutationInput: {
      id?: string;
      name: string;
      fields: TaskCreateFormFields;
    }) =>
      apiSaveTemplate({
        ...(mutationInput.id ? { id: mutationInput.id } : {}),
        name: mutationInput.name,
        payload: buildComposePayloadFromForm(mutationInput.fields),
      }),
    onSuccess: async () => {
      input.closeCreateModal();
      await invalidateTaskCacheAsync(input.queryClient, { scope: "templates" });
    },
  });

  const patchTemplateMutation = useMutation({
    mutationFn: (mutationInput: { id: string; fields: TaskCreateFormFields; name: string }) =>
      apiPatchTemplate(mutationInput.id, {
        name: mutationInput.name,
        payload: buildComposePayloadFromForm(mutationInput.fields),
      }),
    onSuccess: async (_result, variables) => {
      if (input.editingTemplateId === variables.id) {
        input.closeCreateModal();
      }
      await invalidateTaskCacheAsync(input.queryClient, { scope: "templates" });
    },
  });

  const loadTemplateMutation = useMutation({
    mutationFn: (id: string) => apiGetTemplate(id),
  });

  const deleteTemplateMutation = useMutation({
    mutationFn: (id: string) => apiDeleteTemplate(id),
    onSuccess: async () => {
      await invalidateTaskCacheAsync(input.queryClient, { scope: "templates" });
    },
  });

  return {
    saveTemplateMutation,
    patchTemplateMutation,
    loadTemplateMutation,
    deleteTemplateMutation,
  };
}
