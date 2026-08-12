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
      repositoryID?: string;
      worktreeID?: string;
      lockProjectAssignment?: boolean;
      lockGitAssignment?: boolean;
      target?: ComposeTarget;
      operation?: ComposeOperation;
      skipDraftPicker?: boolean;
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
      const skipDraftPicker =
        opts?.skipDraftPicker === true ||
        (opts?.lockGitAssignment === true && Boolean(opts?.worktreeID?.trim()));
      if (target === "task" && operation === "create" && !skipDraftPicker) {
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
      if (target === "template" || skipDraftPicker) {
        input.modal.resetNewTaskForm();
        input.modal.applyCreateModalPrefill();
        input.modal.openComposePhase({ target, operation });
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
        input.modal.showDraftPickerPhase();
        return;
      }
      input.modal.setCreateEntryDraftErrorHint(decision.entryDraftErrorHint);
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
    openComposeModal({ target: "template", operation: "create", skipDraftPicker: true });
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
