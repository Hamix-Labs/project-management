import type { QueryClient } from "@tanstack/react-query";
import { useCallback } from "react";
import { errorMessage } from "@/lib/errorMessage";
import { ensureRepositoriesRegistered } from "@/lib/ensureRepositoriesRegistered";
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
      repositoryID?: string;
      worktreeID?: string;
      lockProjectAssignment?: boolean;
      lockGitAssignment?: boolean;
      target?: ComposeTarget;
      operation?: ComposeOperation;
    }) => {
      const requestId = input.modal.beginEntryRequest();
      input.modal.setCreateEntryDraftErrorHint(null);
      const projectID = opts?.projectID?.trim();
      input.modal.createModalPrefillRef.current = projectID
        ? {
            projectID,
            repositoryID: opts?.repositoryID?.trim() || undefined,
            worktreeID: opts?.worktreeID?.trim() || undefined,
            lockProjectAssignment: opts?.lockProjectAssignment === true,
            lockGitAssignment: opts?.lockGitAssignment === true,
          }
        : null;
      const target = opts?.target ?? "task";
      const operation = opts?.operation ?? "create";
      const skipRepoCheck =
        target !== "task" ||
        operation !== "create" ||
        (opts?.lockGitAssignment === true && Boolean(opts?.worktreeID?.trim()));
      if (!skipRepoCheck) {
        const hasRepos = await ensureRepositoriesRegistered(input.queryClient);
        if (!input.modal.isEntryRequestCurrent(requestId)) {
          return;
        }
        if (!hasRepos) {
          input.modal.showRepositorySetupPhase();
          return;
        }
      }
      if (!input.modal.isEntryRequestCurrent(requestId)) {
        return;
      }
      if (target === "task" && operation === "create" && input.draftsQuery.isError) {
        input.modal.setCreateEntryDraftErrorHint(
          errorMessage(input.draftsQuery.error),
        );
      }
      input.modal.resetNewTaskForm();
      input.modal.applyCreateModalPrefill();
      input.modal.openComposePhase({ target, operation });
    },
    [input],
  );

  const openCreateModal = useCallback(
    (prefill?: {
      projectID: string;
      repositoryID?: string;
      worktreeID?: string;
      lockProjectAssignment?: boolean;
      lockGitAssignment?: boolean;
    }) => {
      return openComposeModal({
        projectID: prefill?.projectID,
        repositoryID: prefill?.repositoryID,
        worktreeID: prefill?.worktreeID,
        lockProjectAssignment: prefill?.lockProjectAssignment,
        lockGitAssignment: prefill?.lockGitAssignment,
        target: "task",
        operation: "create",
      });
    },
    [openComposeModal],
  );

  const openTemplateCreateModal = useCallback(() => {
    openComposeModal({ target: "template", operation: "create" });
  }, [openComposeModal]);

  const startFreshDraft = useCallback(async () => {
    const requestId = input.modal.beginEntryRequest();
    const hasRepos = await ensureRepositoriesRegistered(input.queryClient);
    if (!input.modal.isEntryRequestCurrent(requestId)) {
      return;
    }
    if (!hasRepos) {
      input.modal.showRepositorySetupPhase();
      return;
    }
    input.modal.resetNewTaskForm();
    input.modal.applyCreateModalPrefill();
    input.modal.openComposePhase({ target: "task", operation: "create" });
  }, [input]);

  return {
    openComposeModal,
    openCreateModal,
    openTemplateCreateModal,
    startFreshDraft,
  };
}
