import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
  createGlobalGitRepository,
  createGlobalGitWorktree,
  deleteGlobalGitRepository,
  deleteGlobalGitWorktreeFromDisk,
  unregisterGlobalGitWorktree,
  registerGlobalGitWorktree,
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

  const createWorktree = useMutation({
    mutationFn: (vars: {
      repositoryId: string;
      input: Parameters<typeof createGlobalGitWorktree>[1];
    }) => createGlobalGitWorktree(vars.repositoryId, vars.input),
    onSuccess: (_data, vars) => invalidateRepo(vars.repositoryId),
  });

  const registerWorktree = useMutation({
    mutationFn: (vars: {
      repositoryId: string;
      input: Parameters<typeof registerGlobalGitWorktree>[1];
    }) => registerGlobalGitWorktree(vars.repositoryId, vars.input),
    onSuccess: (_data, vars) => invalidateRepo(vars.repositoryId),
  });

  const unregisterWorktree = useMutation({
    mutationFn: (vars: { worktreeId: string; repositoryId: string }) =>
      unregisterGlobalGitWorktree(vars.worktreeId),
    onSuccess: (_data, vars) => invalidateRepo(vars.repositoryId),
  });

  const removeWorktreeFromDisk = useMutation({
    mutationFn: (vars: {
      worktreeId: string;
      repositoryId: string;
      force?: boolean;
    }) => deleteGlobalGitWorktreeFromDisk(vars.worktreeId, { force: vars.force }),
    onSuccess: (_data, vars) => invalidateRepo(vars.repositoryId),
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
    createWorktree,
    registerWorktree,
    unregisterWorktree,
    removeWorktreeFromDisk,
    reconcile,
    relocateRepository,
  };
}
