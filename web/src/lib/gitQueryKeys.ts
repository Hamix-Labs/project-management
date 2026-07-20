export const gitQueryKeys = {
  all: ["git"] as const,
  /** Global git tree (ADR-0037). */
  globalRepositories: () => [...gitQueryKeys.all, "global", "repositories"] as const,
  globalRepository: (repositoryId: string) =>
    [...gitQueryKeys.all, "global", "repository", repositoryId] as const,
  globalWorktrees: (repositoryId: string) =>
    [...gitQueryKeys.all, "global", "worktrees", repositoryId] as const,
  globalBranches: (repositoryId: string) =>
    [...gitQueryKeys.all, "global", "branches", repositoryId] as const,
  projectsByRepo: (repositoryId: string) =>
    [...gitQueryKeys.all, "global", "projects", repositoryId] as const,
  taskBinding: (worktreeId: string, repositoryIdHint?: string) =>
    [
      ...gitQueryKeys.all,
      "task-binding",
      worktreeId,
      repositoryIdHint ?? "all",
    ] as const,
};
