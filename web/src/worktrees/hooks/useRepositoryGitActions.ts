import { useCallback, useState } from "react";
import type { GitRepository } from "@/types";
import { useOptionalToast } from "@/shared/toast";
import type { GitDeleteTarget } from "../gitDeleteErrors";
import {
  formatReconcileSuccess,
  gitReconcileErrorMessage,
} from "../gitReconcileErrors";
import { useGlobalGitMutations } from "./useGlobalGitMutations";

type ReconcileFlowOutcome = "ok" | "needs_bootstrap" | "error";
type ReconcileIntent = "manual" | "silent";

type Options = {
  onRepositoryDeleted?: () => void;
};

export function useRepositoryGitActions(options?: Options) {
  const onRepositoryDeleted = options?.onRepositoryDeleted;
  const mutations = useGlobalGitMutations();
  const toast = useOptionalToast();

  const [deleteTarget, setDeleteTarget] = useState<GitDeleteTarget | null>(null);
  const [deleteError, setDeleteError] = useState<unknown>(null);
  const [relocateRepository, setRelocateRepository] = useState<GitRepository | null>(null);
  const [reconcileIntent, setReconcileIntent] = useState<ReconcileIntent | null>(null);

  const closeDelete = () => {
    setDeleteTarget(null);
    setDeleteError(null);
  };

  const runDelete = async () => {
    if (!deleteTarget) return;
    setDeleteError(null);
    try {
      await mutations.deleteRepository.mutateAsync(deleteTarget.id);
      closeDelete();
      onRepositoryDeleted?.();
    } catch (err) {
      setDeleteError(err);
    }
  };

  const deletePending = mutations.deleteRepository.isPending;

  const reconcilingRepositoryId =
    mutations.reconcile.isPending || mutations.relocateRepository.isPending
      ? (mutations.reconcile.variables?.repositoryId ??
        mutations.relocateRepository.variables?.repositoryId)
      : undefined;

  const handleReconcile = useCallback(
    async (
      repo: GitRepository,
      options?: { silent?: boolean },
    ): Promise<ReconcileFlowOutcome> => {
      const intent: ReconcileIntent = options?.silent ? "silent" : "manual";
      setReconcileIntent(intent);
      try {
        const result = await mutations.reconcile.mutateAsync({
          repositoryId: repo.id,
          input: { repair: true },
        });
        if (result.status === "needs_bootstrap_path") {
          setRelocateRepository(repo);
          return "needs_bootstrap";
        }
        if (!options?.silent) {
          toast?.success(formatReconcileSuccess(result));
        }
        return "ok";
      } catch (err) {
        if (!options?.silent) {
          toast?.error(gitReconcileErrorMessage(err));
        }
        return "error";
      } finally {
        setReconcileIntent((current) => (current === intent ? null : current));
      }
    },
    [mutations.reconcile, toast],
  );

  const closeRelocateModal = () => {
    setRelocateRepository(null);
    mutations.relocateRepository.reset();
  };

  const isManualReconciling = (repositoryId: string) =>
    reconcilingRepositoryId === repositoryId && reconcileIntent === "manual";

  const openDeleteRepository = (repo: GitRepository) => {
    setDeleteTarget({
      kind: "repository",
      id: repo.id,
      label: repo.path,
      repositoryId: repo.id,
    });
  };

  return {
    mutations,
    deleteTarget,
    deleteError,
    deletePending,
    relocateRepository,
    reconcilingRepositoryId,
    isManualReconciling,
    closeDelete,
    runDelete,
    closeRelocateModal,
    handleReconcile,
    openDeleteRepository,
  };
}
