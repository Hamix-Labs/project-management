import { useCallback, useRef, useState } from "react";
import type { GitRepository } from "@/types";
import { useOptionalToast } from "@/shared/toast";
import type { GitDeleteTarget } from "../gitDeleteErrors";
import { formatReconcileSuccess } from "../gitReconcileErrors";
import { useGlobalGitMutations } from "./useGlobalGitMutations";

type ActiveWorktreeModal = "register-worktree" | "create-worktree" | null;
type ReconcileFlowOutcome = "ok" | "needs_bootstrap" | "error";
type ReconcileIntent = "manual" | "silent";

type Options = {
  repository: GitRepository | null | undefined;
  onRepositoryDeleted?: () => void;
};

export function useRepositoryGitActions({ repository, onRepositoryDeleted }: Options) {
  const mutations = useGlobalGitMutations();
  const toast = useOptionalToast();

  const [activeWorktreeModal, setActiveWorktreeModal] = useState<ActiveWorktreeModal>(null);
  const [deleteTarget, setDeleteTarget] = useState<GitDeleteTarget | null>(null);
  const [deleteError, setDeleteError] = useState<unknown>(null);
  const [relocateRepository, setRelocateRepository] = useState<GitRepository | null>(null);
  const [reconcileErrors, setReconcileErrors] = useState<Record<string, unknown>>({});
  const [autoReconcileBlocked, setAutoReconcileBlocked] = useState<Record<string, true>>({});
  const [reconcileIntent, setReconcileIntent] = useState<ReconcileIntent | null>(null);
  const reconcileInFlightRef = useRef<{
    repoId: string;
    promise: Promise<ReconcileFlowOutcome>;
  } | null>(null);

  const closeDelete = () => {
    setDeleteTarget(null);
    setDeleteError(null);
  };

  const runDelete = async (options?: { force?: boolean }) => {
    if (!deleteTarget) return;
    setDeleteError(null);
    try {
      if (deleteTarget.kind === "repository") {
        await mutations.deleteRepository.mutateAsync(deleteTarget.id);
        closeDelete();
        onRepositoryDeleted?.();
      } else if (deleteTarget.mode === "remove_from_disk") {
        await mutations.removeWorktreeFromDisk.mutateAsync({
          worktreeId: deleteTarget.id,
          repositoryId: deleteTarget.repositoryId,
          force: options?.force,
        });
        closeDelete();
      } else {
        await mutations.unregisterWorktree.mutateAsync({
          worktreeId: deleteTarget.id,
          repositoryId: deleteTarget.repositoryId,
        });
        closeDelete();
      }
    } catch (err) {
      setDeleteError(err);
    }
  };

  const deletePending =
    mutations.deleteRepository.isPending ||
    mutations.unregisterWorktree.isPending ||
    mutations.removeWorktreeFromDisk.isPending;

  const reconcilingRepositoryId =
    mutations.reconcile.isPending || mutations.relocateRepository.isPending
      ? mutations.reconcile.variables?.repositoryId ??
        mutations.relocateRepository.variables?.repositoryId
      : undefined;

  const handleReconcile = useCallback(
    async (
      repo: GitRepository,
      options?: { silent?: boolean },
    ): Promise<ReconcileFlowOutcome> => {
      const intent: ReconcileIntent = options?.silent ? "silent" : "manual";
      setReconcileIntent(intent);
      setReconcileErrors((prev) => {
        const next = { ...prev };
        delete next[repo.id];
        return next;
      });
      try {
        const result = await mutations.reconcile.mutateAsync({
          repositoryId: repo.id,
          input: { repair: true },
        });
        if (result.status === "needs_bootstrap_path") {
          setAutoReconcileBlocked((prev) => ({ ...prev, [repo.id]: true }));
          setRelocateRepository(repo);
          return "needs_bootstrap";
        }
        if (!options?.silent) {
          toast?.success(formatReconcileSuccess(result));
        }
        return "ok";
      } catch (err) {
        setReconcileErrors((prev) => ({ ...prev, [repo.id]: err }));
        return "error";
      } finally {
        setReconcileIntent((current) => (current === intent ? null : current));
      }
    },
    [mutations.reconcile, toast],
  );

  const ensureInventoryFresh = useCallback(
    async (repo: GitRepository): Promise<ReconcileFlowOutcome> => {
      const inFlight = reconcileInFlightRef.current;
      if (inFlight?.repoId === repo.id) {
        return inFlight.promise;
      }
      const promise = handleReconcile(repo, { silent: true });
      reconcileInFlightRef.current = { repoId: repo.id, promise };
      try {
        return await promise;
      } finally {
        if (reconcileInFlightRef.current?.repoId === repo.id) {
          reconcileInFlightRef.current = null;
        }
      }
    },
    [handleReconcile],
  );

  const openWorktreeModal = useCallback(
    (modal: Exclude<ActiveWorktreeModal, null>) => {
      if (!repository) return;
      setActiveWorktreeModal(modal);
      void ensureInventoryFresh(repository).then((outcome) => {
        if (outcome === "needs_bootstrap") setActiveWorktreeModal(null);
      });
    },
    [repository, ensureInventoryFresh],
  );

  const closeRelocateModal = () => {
    setRelocateRepository(null);
    mutations.relocateRepository.reset();
  };

  const reconcilePending = repository != null && reconcilingRepositoryId === repository.id;
  const inventoryRefreshPending = reconcilePending && reconcileIntent === "silent";
  const manualReconcilePending = reconcilePending && reconcileIntent === "manual";
  const reconcileError =
    repository != null ? reconcileErrors[repository.id] : undefined;
  const reconcileBlocked =
    repository != null && autoReconcileBlocked[repository.id] === true;

  const openDeleteRepository = () => {
    if (!repository) return;
    setDeleteTarget({
      kind: "repository",
      id: repository.id,
      label: repository.path,
      repositoryId: repository.id,
    });
  };

  const openDeleteWorktree = (worktreeId: string, label: string) => {
    if (!repository) return;
    setDeleteTarget({
      kind: "worktree",
      mode: "unregister",
      id: worktreeId,
      label,
      repositoryId: repository.id,
    });
  };

  const openRemoveWorktreeFromDisk = (worktreeId: string, label: string) => {
    if (!repository) return;
    setDeleteTarget({
      kind: "worktree",
      mode: "remove_from_disk",
      id: worktreeId,
      label,
      repositoryId: repository.id,
    });
  };

  return {
    mutations,
    activeWorktreeModal,
    setActiveWorktreeModal,
    deleteTarget,
    deleteError,
    deletePending,
    relocateRepository,
    reconcilePending,
    inventoryRefreshPending,
    manualReconcilePending,
    reconcileError,
    reconcileBlocked,
    closeDelete,
    runDelete,
    closeRelocateModal,
    handleReconcile,
    ensureInventoryFresh,
    openWorktreeModal,
    openDeleteRepository,
    openDeleteWorktree,
    openRemoveWorktreeFromDisk,
    setAutoReconcileBlocked,
  };
}
