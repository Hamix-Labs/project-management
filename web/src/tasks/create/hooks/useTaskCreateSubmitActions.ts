import { useCallback, type FormEvent } from "react";
import { buildCreateTaskMutationInput } from "../buildCreateMutationInput";
import { validateCreateFormChecklist } from "../validateCreateForm";
import { validateTagsCsv } from "../taskTagValidation";
import type { useTaskCreateFormState } from "./useTaskCreateFormState";
import type { useTaskCreateModalState } from "./useTaskCreateModalState";
import type { useTaskCreateMutations } from "./useTaskCreateMutations";

function validateCreateOrTemplateForm(form: ReturnType<typeof useTaskCreateFormState>): string | null {
  const checklistError = validateCreateFormChecklist(
    form.newTitle,
    form.newPriority,
    form.newChecklistItems,
  );
  if (checklistError) return checklistError;
  return validateTagsCsv(form.newTagsCsv);
}

export function useTaskCreateSubmitActions(input: {
  form: ReturnType<typeof useTaskCreateFormState>;
  modal: ReturnType<typeof useTaskCreateModalState>;
  mutations: ReturnType<typeof useTaskCreateMutations>;
}) {
  const submitCreate = useCallback(async (event: FormEvent) => {
    event.preventDefault();
    if (!input.form.newTitle.trim() || !input.form.newPriority) return;
    const validationError = validateCreateOrTemplateForm(input.form);
    if (validationError) {
      input.form.setCreateFormError(validationError);
      return;
    }
    input.form.setCreateFormError(null);
    input.mutations.createMutation.mutate(buildCreateTaskMutationInput(input.form.formFields));
  }, [input]);

  const submitTemplate = useCallback(async (event: FormEvent) => {
    event.preventDefault();
    if (!input.form.newTitle.trim() || !input.form.newPriority) return;
    const validationError = validateCreateOrTemplateForm(input.form);
    if (validationError) {
      input.form.setCreateFormError(validationError);
      return;
    }
    input.form.setCreateFormError(null);
    const name = input.form.newTitle.trim();
    const fields = input.form.formFields;
    if (input.modal.composeOperation === "edit" && input.modal.editingTemplateId) {
      input.mutations.patchTemplateMutation.mutate({
        id: input.modal.editingTemplateId,
        name,
        fields,
      });
      return;
    }
    input.mutations.saveTemplateMutation.mutate({
      name,
      fields,
    });
  }, [input]);

  return {
    submitCreate,
    submitTemplate,
  };
}
