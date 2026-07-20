import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
  createGlobalGitRepository,
  deleteGlobalGitRepository,
  relocateGlobalGitRepository,
  syncGlobalGitRepository,
} from "@/api/gitGlobal";
import type { GitReconcileInput } from "@/types/git";
import {
  invalidateGitCache,
} from "./invalidateGitCache";

export function useGlobalGitMutations() {
  const qc = useQueryClient();

  const invalidateRepo = (repositoryId: string) => {
    invalidateGitCache(qc, { scope: "repository", repositoryId });
  };

  const createRepository = useMutation({
    mutationFn: createGlobalGitRepository,
    onSuccess: () => {
      invalidateGitCache(qc, { scope: "repositories" });
    },
  });

  const deleteRepository = useMutation({
    mutationFn: deleteGlobalGitRepository,
    onSuccess: () => {
      invalidateGitCache(qc, { scope: "repositories" });
    },
  });

  const reconcile = useMutation({
    mutationFn: (vars: { repositoryId: string; input?: GitReconcileInput }) =>
      syncGlobalGitRepository(vars.repositoryId),
    onSuccess: (_data, vars) => invalidateRepo(vars.repositoryId),
  });

  const relocateRepository = useMutation({
    mutationFn: (vars: { repositoryId: string; input: { path: string } }) =>
      relocateGlobalGitRepository(vars.repositoryId, vars.input),
    onSuccess: (_data, vars) => invalidateRepo(vars.repositoryId),
  });

  return {
    createRepository,
    deleteRepository,
    reconcile,
    relocateRepository,
  };
}
