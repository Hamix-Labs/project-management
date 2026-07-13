import type { QueryClient } from "@tanstack/react-query";
import { useCallback } from "react";
import { errorMessage } from "@/lib/errorMessage";
import { ensureRepositoriesRegistered } from "@/lib/ensureRepositoriesRegistered";
import { decideCreateEntry } from "../decideCreateEntry";
import type { ComposeOperation, ComposeTarget, TaskDraftsQuery } from "../types";
import type { useTaskCreateModalState } from "./useTaskCreateModalState";

export function useTaskCreateComposeEntryActions(input: {
  modal: ReturnType<typeof useTaskCreateModalState>;
  draftsQuery: TaskDraftsQuery;
  queryClient: QueryClient;
}) {
  const openComposeModal = useCallback(
    async (opts?: {
      projectID?: string;
      lockProjectAssignment?: boolean;
      target?: ComposeTarget;
      operation?: ComposeOperation;
      skipDraftPicker?: boolean;
    }) => {
      input.modal.setCreateEntryDraftErrorHint(null);
      const projectID = opts?.projectID?.trim();
      input.modal.createModalPrefillRef.current = projectID
        ? {
            projectID,
            lockProjectAssignment: opts?.lockProjectAssignment === true,
          }
        : null;
      const target = opts?.target ?? "task";
      const operation = opts?.operation ?? "create";
      if (target === "task" && operation === "create" && !opts?.skipDraftPicker) {
        const hasRepos = await ensureRepositoriesRegistered(input.queryClient);
        if (!hasRepos) {
          input.modal.setRepositorySetupPromptOpen(true);
          return;
        }
      }
      if (target === "template" || opts?.skipDraftPicker) {
        input.modal.resetNewTaskForm();
        input.modal.setComposeTarget(target);
        input.modal.setComposeOperation(operation);
        input.modal.applyCreateModalPrefill();
        input.modal.setCreateModalOpen(true);
        return;
      }
      const decision = decideCreateEntry({
        isPending: input.draftsQuery.isPending,
        isError: input.draftsQuery.isError,
        errorMessage: input.draftsQuery.isError
          ? errorMessage(input.draftsQuery.error)
          : null,
        draftCount: input.draftsQuery.data?.length ?? 0,
      });
      if (decision.kind === "showPicker") {
        input.modal.setDraftPickerOpen(true);
        return;
      }
      input.modal.setCreateEntryDraftErrorHint(decision.entryDraftErrorHint);
      input.modal.resetNewTaskForm();
      input.modal.applyCreateModalPrefill();
      input.modal.setCreateModalOpen(true);
    },
    [input],
  );

  const openCreateModal = useCallback(
    (prefill?: { projectID: string; lockProjectAssignment?: boolean }) => {
      void openComposeModal({
        projectID: prefill?.projectID,
        lockProjectAssignment: prefill?.lockProjectAssignment,
        target: "task",
        operation: "create",
      });
    },
    [openComposeModal],
  );

  const openTemplateCreateModal = useCallback(() => {
    openComposeModal({ target: "template", operation: "create", skipDraftPicker: true });
  }, [openComposeModal]);

  const startFreshDraft = useCallback(async () => {
    const hasRepos = await ensureRepositoriesRegistered(input.queryClient);
    if (!hasRepos) {
      input.modal.setDraftPickerOpen(false);
      input.modal.setRepositorySetupPromptOpen(true);
      return;
    }
    input.modal.resetNewTaskForm();
    input.modal.applyCreateModalPrefill();
    input.modal.setDraftPickerOpen(false);
    input.modal.setCreateModalOpen(true);
  }, [input]);

  return {
    openComposeModal,
    openCreateModal,
    openTemplateCreateModal,
    startFreshDraft,
  };
}
