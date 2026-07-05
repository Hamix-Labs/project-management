export const worktreeGitCopy = {
  sectionTitle: "Worktrees",
  addWorktree: "Add worktree",
  registerRepository: "Register repository",
  registerWorktree: "Register worktree",
  createWorktree: "Create worktree",
  reconcile: "Reconcile",
  reconciling: "Reconciling…",
  reconcilingStatus: "Syncing registered worktrees with git…",
  inventoryRefreshStatus: "Refreshing worktree list…",
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
  listColumnWorktreeCount: "Worktrees",
  searchRepositoriesPlaceholder: "Search by name or path…",
  repositoriesPageSubtitle:
    "Register and manage the Git repositories powering your worktrees.",
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
    "The primary checkout from git clone or git init. Unregistering removes Hamix tracking only — the checkout stays on disk.",
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
  emptyWorktreesTitle: "No worktrees yet",
  emptyWorktreesDescription:
    "Register an existing linked directory or create a new one with git worktree add.",
  registerModalTitle: "Register worktree",
  registerModalLead:
    "Link an existing git worktree directory and choose the branch Hamix should track.",
  registerModalPathLabel: "Worktree path",
  registerModalDisplayNameLabel: "Display name",
  liveInventoryReconcileLead:
    "Hamix can't read linked worktrees because the registered checkout path isn't available on disk. Reconcile refreshes paths from git so you can continue.",
  liveInventoryReconcileAction: "Reconcile repository",
  registerModalDisplayNamePlaceholder: "Optional",
  registerModalSubmit: "Register worktree",
  registerModalSubmitting: "Registering…",
  createModalTitle: "Create worktree",
  createModalLead:
    "Git will create a new checkout folder on disk and link it to this repository. Pick where the folder should live and which branch it tracks.",
  createModalStartFromLabel: "Branch starting point",
  createModalStartFromMain: "Main repository checkout",
  createModalStartFromReference: "Reference worktree",
  createModalReferenceLabel: "Reference worktree",
  createModalReferenceDetached:
    "The selected worktree has a detached HEAD. Pick a worktree checked out on a branch.",
  createModalLocationLabel: "Parent directory",
  createModalLocationHint: "Where the worktree folder will be created on disk.",
  createModalChooseParentFolder: "Choose parent folder",
  createModalChangeParentFolder: "Change parent folder",
  createModalParentSelectedPrefix: "Selected parent directory",
  createModalFolderNameLabel: "Checkout folder name",
  createModalFolderNameHint:
    "Directory Git creates on disk.",
  createModalFolderNamePlaceholder: "e.g. Hamix-wt-feature",
  createModalFullPathPrefix: "Full path:",
  createModalPickerTitle: "Choose parent folder",
  createModalPickerLead:
    "Pick where the new checkout folder should be created. Browse from Home or Documents if you want a location outside this repository.",
  createModalPickerSelectionLabel: "Parent folder",
  createModalPickerConfirmLabel: "Use as parent folder",
  createModalDisplayNameLabel: "Display name",
  createModalDisplayNameHint:
    "Label in the worktree list. Defaults to the checkout folder name if left blank.",
  createModalDisplayNamePlaceholder: "Same as folder name",
  createModalSubmit: "Create worktree",
  createModalSubmitting: "Creating…",
  cancel: "Cancel",
  relocateModalTitle: "Relocate repository",
  relocateModalLead:
    "Hamix cannot find this repository at its registered path. Browse from the parent folder or Documents to find the renamed checkout — Hamix verifies it is the same repo before updating paths.",
  relocateModalStoredPathLabel: "Registered path",
  relocateModalPathLabel: "New checkout path",
  relocateModalChoosePath: "Choose folder",
  relocateModalSelectedPrefix: "Selected:",
  relocateModalNoPath: "No folder selected yet.",
  relocateModalSubmit: "Relocate and reconcile",
  relocateModalSubmitting: "Relocating…",
  reconcileErrorTitle: "Reconcile failed",
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

export function liveWorktreeOptionLabel(path: string, isMain: boolean): string {
  return isMain ? `${path} (main worktree)` : path;
}
