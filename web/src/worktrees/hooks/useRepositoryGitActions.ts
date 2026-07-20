import { useState } from "react";
import type { GitRepository } from "@/types";
import type { GitDeleteTarget } from "../gitDeleteErrors";
import { useGlobalGitMutations } from "./useGlobalGitMutations";

type Options = {
  onRepositoryDeleted?: () => void;
};

export function useRepositoryGitActions(options?: Options) {
  const onRepositoryDeleted = options?.onRepositoryDeleted;
  const mutations = useGlobalGitMutations();

  const [deleteTarget, setDeleteTarget] = useState<GitDeleteTarget | null>(null);
  const [deleteError, setDeleteError] = useState<unknown>(null);

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
    deletePending: mutations.deleteRepository.isPending,
    closeDelete,
    runDelete,
    openDeleteRepository,
  };
}
