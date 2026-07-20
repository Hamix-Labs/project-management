export const worktreeGitCopy = {
  registerRepository: "Register repository",
  deleteRepository: "Delete",
  hostPathLabel: "Host path",
  listColumnName: "Name",
  listColumnActions: "Actions",
  searchRepositoriesPlaceholder: "Search by name or path…",
  repositoriesPageSubtitle:
    "Register repositories; Hamix allocates worktrees when you create tasks.",
  repositoriesSearchCount: (filtered: number, total: number) => {
    const noun = total === 1 ? "repository" : "repositories";
    return `${filtered} of ${total} ${noun}`;
  },
  repositoriesSearchEmptyTitle: "No repositories found",
  repositoriesSearchEmptyDescription: (query: string) =>
    `No results for "${query}". Try a different name or path.`,
  clearSearch: "Clear search",
  copyPath: "Copy path",
  pathCopied: "Path copied",
  cancel: "Cancel",
} as const;
