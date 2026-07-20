export const worktreeGitCopy = {
  registerRepository: "Register repository",
  reconcile: "Sync",
  reconciling: "Syncing…",
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
  relocateModalTitle: "Relocate repository",
  relocateModalLead:
    "Hamix cannot find this repository at its registered path. Browse from the parent folder or Documents to find the renamed checkout — Hamix verifies it is the same repo before updating paths.",
  relocateModalStoredPathLabel: "Registered path",
  relocateModalPathLabel: "New checkout path",
  relocateModalChoosePath: "Choose folder",
  relocateModalSelectedPrefix: "Selected:",
  relocateModalNoPath: "No folder selected yet.",
  relocateModalSubmit: "Relocate and sync",
  relocateModalSubmitting: "Relocating…",
  reconcileErrorTitle: "Sync failed",
} as const;
