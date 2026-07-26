import type { QueryClient } from "@tanstack/react-query";
import { useCallback } from "react";
import type { AppSettings } from "@/api/settings";
import { settingsQueryKeys } from "@/lib/settingsQueryKeys";
import { hydrateFormFromComposePayload } from "../composePayload";
import type { useTaskCreateFormState } from "./useTaskCreateFormState";
import type { useTaskCreateModalState } from "./useTaskCreateModalState";
import type { useTaskCreateMutations } from "./useTaskCreateMutations";

export function useTaskCreateTemplateEntryActions(input: {
  form: ReturnType<typeof useTaskCreateFormState>;
  modal: ReturnType<typeof useTaskCreateModalState>;
  mutations: ReturnType<typeof useTaskCreateMutations>;
  queryClient: QueryClient;
}) {
  const editTemplateByID = useCallback(
    async (id: string) => {
      input.modal.createModalPrefillRef.current = null;
      input.modal.setCreateModalAssignmentLocked(false);
      const template = await input.mutations.loadTemplateMutation.mutateAsync(id);
      const settings = input.queryClient.getQueryData<AppSettings>(settingsQueryKeys.app());
      const hydrated = hydrateFormFromComposePayload(template.payload, settings);
      input.modal.resetNewTaskForm();
      input.form.setNewTitle(hydrated.title);
      input.form.setNewPrompt(hydrated.prompt);
      input.form.setNewPriority(hydrated.priority);
      input.form.setNewTaskRunner(hydrated.runner);
      input.form.setNewTaskCursorModel(hydrated.cursorModel);
      input.form.setNewProjectID(hydrated.projectID);
      input.form.setNewRepositoryID(hydrated.repositoryID);
      input.form.setNewWorktreeID(hydrated.worktreeID);
      input.form.setNewSchedule(hydrated.schedule);
      input.form.setNewAutonomyEnabled(hydrated.autonomyEnabled);
      input.form.setNewTagsCsv(hydrated.tagsCsv);
      input.form.setNewMilestone(hydrated.milestone);
      input.form.setNewDependsOn(hydrated.dependsOn);
      input.form.setNewChecklistItems(hydrated.checklistItems);
      input.form.setNewFunctionInputs(hydrated.functionInputs);
      input.modal.openComposePhase({
        target: "template",
        operation: "edit",
        editingTemplateId: template.id,
      });
    },
    [input],
  );

  const deleteTemplateByID = useCallback(
    async (id: string) => {
      await input.mutations.deleteTemplateMutation.mutateAsync(id);
    },
    [input.mutations.deleteTemplateMutation],
  );

  const instantiateTemplates = useCallback(
    async (items: import("@/api").TaskTemplateInstantiateItem[]) =>
      input.mutations.instantiateTemplatesMutation.mutateAsync(items),
    [input.mutations.instantiateTemplatesMutation],
  );

  return {
    editTemplateByID,
    deleteTemplateByID,
    instantiateTemplates,
  };
}
