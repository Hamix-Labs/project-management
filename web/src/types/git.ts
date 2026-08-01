export type GitRepository = {
  id: string;
  path: string;
  git_common_dir: string;
  host_path: string;
  default_branch: string;
  main_branch_name: string;
  linked_worktree_count: number;
  created_at: string;
  updated_at: string;
};

export type GitWorktree = {
  id: string;
  repository_id: string;
  path: string;
  host_path: string;
  name: string;
  is_main: boolean;
  branch_id?: string;
  created_at: string;
  stale?: boolean;
};

/** GET /git/worktrees/{id} — worktree plus resolved repo paths and branch name. */
export type GitWorktreeDetail = GitWorktree & {
  repository_path: string;
  repository_host_path: string;
  branch_name: string;
};

export type GitBranch = {
  id: string;
  repository_id: string;
  name: string;
  head_sha: string;
  created_at: string;
};

/** Checkout status from `GET /git/repositories/{repoId}/worktrees/checkout-status`. */
export type GitWorktreeCheckoutStatus = {
  worktree_id: string;
  available: boolean;
  reason?: string;
  dirty?: boolean;
  detached?: boolean;
  head_commit_at?: string;
  has_upstream?: boolean;
  ahead?: number;
  behind?: number;
  upstream?: string;
};

export type GitReconcileSkipped = {
  worktree_id: string;
  reason: string;
};

export type GitReconcileNeedsBranchBind = {
  path: string;
  branch: string;
};

export type GitReconcileReport = {
  repo_path_updated: boolean;
  worktrees_path_updated: number;
  worktrees_added: number;
  worktrees_removed: number;
  branches_head_updated: number;
  resolution_source?: string;
  discovered_path?: string;
  worktrees_skipped: GitReconcileSkipped[];
  needs_branch_bind: GitReconcileNeedsBranchBind[];
};

export type GitReconcileResult = {
  status: string;
  report: GitReconcileReport;
};

export type GitReconcileInput = {
  bootstrap_path?: string;
  repair?: boolean;
  dry_run?: boolean;
};
