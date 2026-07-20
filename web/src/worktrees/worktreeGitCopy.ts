export const worktreeGitCopy = {
  sectionTitle: "Worktrees",
  registerRepository: "Register repository",
  reconcile: "Sync",
  reconciling: "Syncing…",
  reconcilingStatus: "Fetching origin and refreshing worktree metadata…",
  staleWorktreeHint: "Idle for over 24 hours — safe to remove from disk.",
  removeStaleWorktree: "Remove from disk",
  inspectPath: "On disk",
  deleteRepository: "Delete",
  unregisterWorktree: "Unregister worktree",
  deleteWorktree: "Delete worktree",
  unregisterWorktreeConfirmTitle: "Unregister worktree?",
  unregisterWorktreeConfirmDescription:
    "will be removed from Hamix only. The checkout stays on disk.",
  unregisterWorktreeConfirmFootnote:
    "Use Delete worktree to remove the checkout from git.",
  deleteWorktreeConfirmTitle: "Delete worktree?",
  deleteWorktreeConfirmDescription:
    "will be removed from Hamix and deleted from disk.",
  deleteWorktreeConfirmFootnote:
    "Use Unregister to keep the checkout on disk.",
  deleteWorktreeForceLabel: "Force remove (discard uncommitted changes)",
  repositoryActions: "Repository actions",
  worktreeActions: (name: string) => `Worktree actions for ${name}`,
  hostPathLabel: "Host path",
  listColumnName: "Name",
  listColumnBranch: "Branch",
  listColumnActions: "Actions",
  listColumnWorktreeCount: "Worktrees",
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
  searchWorktreesPlaceholder: "Search by name or branch…",
  cellNotApplicable: "—",
  mainWorktreeShortLabel: "main",
  mainWorktreeLabel: "main worktree",
  mainWorktreeHint:
    "The primary checkout from git clone or git init. Tasks never bind to this worktree.",
  statusUnavailable: "—",
  statusUnavailableTitle: "Worktree checkout status is not available yet",
  primaryWorktreeBadge: "Primary",
  statusClean: "Clean",
  statusDirty: "Uncommitted changes",
  statusLastCommit: (relative: string) => `Last commit ${relative}`,
  syncUpToDate: "Up to date",
  locationLabel: "Location",
  detachedHead: "Detached HEAD",
  noMatchingWorktreesTitle: "No worktrees found",
  emptyWorktreesTitle: "No managed worktrees yet",
  emptyWorktreesDescription:
    "Create a task on this repository and Hamix will allocate a worktree automatically.",
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

export function worktreeCountLabel(count: number): string {
  return count === 1 ? "1 worktree" : `${count} worktrees`;
}

/** Numeric count for repository list cells; the column header already says "Worktrees". */
export function repositoryListWorktreeCountDisplay(count: number): string {
  return String(count);
}

export function worktreeAriaLabel(displayName: string): string {
  return `Worktree: ${displayName}`;
}

export function unregisterWorktreeAriaLabel(displayName: string): string {
  return `Unregister worktree "${displayName}"`;
}
