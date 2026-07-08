import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
  createGitBranch,
  createGitRepository,
  createGitWorktree,
  deleteGitBranch,
  deleteGitRepository,
  unregisterGitWorktree,
  reconcileGitRepository,
} from "@/api/git";
import { invalidateGitCache } from "./invalidateGitCache";

export function useGitMutations(projectId: string) {
  const queryClient = useQueryClient();

  const invalidateRepo = (repositoryId: string) => {
    invalidateGitCache(queryClient, {
      scope: "legacyRepository",
      projectId,
      repositoryId,
    });
  };

  const registerRepository = useMutation({
    mutationFn: (input: { path: string; host_path?: string; default_branch?: string }) =>
      createGitRepository(projectId, input),
    onSuccess: () => {
      invalidateGitCache(queryClient, { scope: "legacyRepositories", projectId });
    },
  });

  const removeRepository = useMutation({
    mutationFn: (repositoryId: string) => deleteGitRepository(projectId, repositoryId),
    onSuccess: () => {
      invalidateGitCache(queryClient, { scope: "legacyRepositories", projectId });
    },
  });

  const addWorktree = useMutation({
    mutationFn: (input: {
      repositoryId: string;
      path: string;
      name?: string;
      branch: string;
      create_branch?: boolean;
      start_point?: string;
    }) =>
      createGitWorktree(projectId, input.repositoryId, {
        path: input.path,
        name: input.name,
        branch: input.branch,
        create_branch: input.create_branch,
        start_point: input.start_point,
      }),
    onSuccess: (_data, vars) => invalidateRepo(vars.repositoryId),
  });

  const removeWorktree = useMutation({
    mutationFn: (input: { worktreeId: string; repositoryId: string }) =>
      unregisterGitWorktree(projectId, input.worktreeId),
    onSuccess: (_data, vars) => invalidateRepo(vars.repositoryId),
  });

  const addBranch = useMutation({
    mutationFn: (input: { repositoryId: string; name: string; start_point?: string }) =>
      createGitBranch(projectId, input.repositoryId, {
        name: input.name,
        start_point: input.start_point,
      }),
    onSuccess: (_data, vars) => invalidateRepo(vars.repositoryId),
  });

  const removeBranch = useMutation({
    mutationFn: (input: { branchId: string; repositoryId: string; force?: boolean }) =>
      deleteGitBranch(projectId, input.branchId, { force: input.force }),
    onSuccess: (_data, vars) => invalidateRepo(vars.repositoryId),
  });

  const reconcile = useMutation({
    mutationFn: (repositoryId: string) => reconcileGitRepository(projectId, repositoryId),
    onSuccess: (_data, repositoryId) => invalidateRepo(repositoryId),
  });

  return {
    registerRepository,
    removeRepository,
    addWorktree,
    removeWorktree,
    addBranch,
    removeBranch,
    reconcile,
  };
}
