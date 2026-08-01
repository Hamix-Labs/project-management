import type {
  GitBranch,
  GitRepository,
  GitReconcileNeedsBranchBind,
  GitReconcileReport,
  GitReconcileResult,
  GitReconcileSkipped,
  GitWorktree,
  GitWorktreeCheckoutStatus,
  GitWorktreeDetail,
} from "@/types/git";
import { isRecord, parseBooleanField, parseNonEmptyString, parseOptionalNonEmptyId, parseString } from "./parseTaskApiCore";

function parseLinkedWorktreeCount(value: unknown, path: string): number {
  if (typeof value !== "number" || !Number.isFinite(value) || value < 0) {
    throw new Error(`Invalid API response: ${path} must be a non-negative number`);
  }
  return Math.trunc(value);
}

function parseGitRepositoryRow(value: unknown, path: string): GitRepository {
  if (!isRecord(value)) {
    throw new Error(`Invalid API response: ${path} must be object`);
  }
  return {
    id: parseNonEmptyString(value.id, `${path}.id`),
    path: parseString(value.path, `${path}.path`),
    git_common_dir: isRecord(value) && value.git_common_dir != null
      ? parseString(value.git_common_dir, `${path}.git_common_dir`)
      : "",
    host_path: parseString(value.host_path, `${path}.host_path`),
    default_branch: isRecord(value) && value.default_branch != null
      ? parseString(value.default_branch, `${path}.default_branch`)
      : "",
    main_branch_name: isRecord(value) && value.main_branch_name != null
      ? parseString(value.main_branch_name, `${path}.main_branch_name`)
      : "",
    linked_worktree_count:
      isRecord(value) && value.linked_worktree_count != null
        ? parseLinkedWorktreeCount(value.linked_worktree_count, `${path}.linked_worktree_count`)
        : 0,
    created_at: parseString(value.created_at, `${path}.created_at`),
    updated_at: parseString(value.updated_at, `${path}.updated_at`),
  };
}

function parseGitWorktreeRow(value: unknown, path: string): GitWorktree {
  if (!isRecord(value)) {
    throw new Error(`Invalid API response: ${path} must be object`);
  }
  const row: GitWorktree = {
    id: parseNonEmptyString(value.id, `${path}.id`),
    repository_id: parseNonEmptyString(value.repository_id, `${path}.repository_id`),
    path: parseString(value.path, `${path}.path`),
    host_path:
      value.host_path === undefined || value.host_path === null
        ? ""
        : parseString(value.host_path, `${path}.host_path`),
    name: parseString(value.name, `${path}.name`),
    is_main: parseBooleanField(value.is_main, `${path}.is_main`),
    created_at: parseString(value.created_at, `${path}.created_at`),
  };
  const branchID = parseOptionalNonEmptyId(value.branch_id, `${path}.branch_id`);
  if (branchID) {
    row.branch_id = branchID;
  }
  if (value.stale === true) {
    row.stale = true;
  }
  return row;
}

function parseGitBranchRow(value: unknown, path: string): GitBranch {
  if (!isRecord(value)) {
    throw new Error(`Invalid API response: ${path} must be object`);
  }
  return {
    id: parseNonEmptyString(value.id, `${path}.id`),
    repository_id: parseNonEmptyString(value.repository_id, `${path}.repository_id`),
    name: parseString(value.name, `${path}.name`),
    head_sha: parseString(value.head_sha, `${path}.head_sha`),
    created_at: parseString(value.created_at, `${path}.created_at`),
  };
}

export function parseGitRepositoryList(raw: unknown): GitRepository[] {
  if (!isRecord(raw)) {
    throw new Error("Invalid API response: body must be object");
  }
  const rows = raw.repositories;
  if (!Array.isArray(rows)) {
    throw new Error("Invalid API response: repositories must be array");
  }
  return rows.map((row, i) => parseGitRepositoryRow(row, `repositories[${i}]`));
}

export function parseGitRepository(raw: unknown): GitRepository {
  return parseGitRepositoryRow(raw, "repository");
}

export function parseGitWorktreeList(raw: unknown): GitWorktree[] {
  if (!isRecord(raw)) {
    throw new Error("Invalid API response: body must be object");
  }
  const rows = raw.worktrees;
  if (!Array.isArray(rows)) {
    throw new Error("Invalid API response: worktrees must be array");
  }
  return rows.map((row, i) => parseGitWorktreeRow(row, `worktrees[${i}]`));
}

export function parseGitWorktree(raw: unknown): GitWorktree {
  return parseGitWorktreeRow(raw, "worktree");
}

export function parseGitWorktreeDetail(raw: unknown): GitWorktreeDetail {
  const base = parseGitWorktreeRow(raw, "worktree");
  if (!isRecord(raw)) {
    throw new Error("Invalid API response: worktree must be object");
  }
  return {
    ...base,
    repository_path: parseString(raw.repository_path, "worktree.repository_path"),
    repository_host_path:
      raw.repository_host_path === undefined || raw.repository_host_path === null
        ? ""
        : parseString(raw.repository_host_path, "worktree.repository_host_path"),
    branch_name:
      raw.branch_name === undefined || raw.branch_name === null
        ? ""
        : parseString(raw.branch_name, "worktree.branch_name"),
  };
}

export function parseGitBranchList(raw: unknown): GitBranch[] {
  if (!isRecord(raw)) {
    throw new Error("Invalid API response: body must be object");
  }
  const rows = raw.branches;
  if (!Array.isArray(rows)) {
    throw new Error("Invalid API response: branches must be array");
  }
  return rows.map((row, i) => parseGitBranchRow(row, `branches[${i}]`));
}

export function parseGitBranch(raw: unknown): GitBranch {
  return parseGitBranchRow(raw, "branch");
}

function parseOptionalInt(value: unknown, path: string): number | undefined {
  if (value == null) return undefined;
  if (typeof value !== "number" || !Number.isFinite(value) || value < 0) {
    throw new Error(`Invalid API response: ${path} must be a non-negative number`);
  }
  return Math.trunc(value);
}

function parseGitWorktreeCheckoutStatusRow(
  value: unknown,
  path: string,
): GitWorktreeCheckoutStatus {
  if (!isRecord(value)) {
    throw new Error(`Invalid API response: ${path} must be object`);
  }
  const row: GitWorktreeCheckoutStatus = {
    worktree_id: parseNonEmptyString(value.worktree_id, `${path}.worktree_id`),
    available: Boolean(value.available),
  };
  if (value.reason != null) {
    row.reason = parseString(value.reason, `${path}.reason`);
  }
  if (!row.available) {
    return row;
  }
  if (value.dirty != null) {
    row.dirty = Boolean(value.dirty);
  }
  if (value.detached != null) {
    row.detached = Boolean(value.detached);
  }
  if (value.head_commit_at != null) {
    row.head_commit_at = parseString(value.head_commit_at, `${path}.head_commit_at`);
  }
  if (value.has_upstream != null) {
    row.has_upstream = Boolean(value.has_upstream);
  }
  if (row.has_upstream) {
    if (value.upstream != null) {
      row.upstream = parseString(value.upstream, `${path}.upstream`);
    }
    row.ahead = parseOptionalInt(value.ahead, `${path}.ahead`);
    row.behind = parseOptionalInt(value.behind, `${path}.behind`);
  }
  return row;
}

export function parseGitWorktreeCheckoutStatusList(raw: unknown): GitWorktreeCheckoutStatus[] {
  if (!isRecord(raw)) {
    throw new Error("Invalid API response: body must be object");
  }
  const rows = raw.worktrees;
  if (!Array.isArray(rows)) {
    throw new Error("Invalid API response: worktrees must be array");
  }
  return rows.map((row, i) => parseGitWorktreeCheckoutStatusRow(row, `worktrees[${i}]`));
}

function parseGitReconcileSkipped(value: unknown, path: string): GitReconcileSkipped {
  if (!isRecord(value)) {
    throw new Error(`Invalid API response: ${path} must be object`);
  }
  return {
    worktree_id: parseNonEmptyString(value.worktree_id, `${path}.worktree_id`),
    reason: parseString(value.reason, `${path}.reason`),
  };
}

function parseGitReconcileNeedsBranchBind(
  value: unknown,
  path: string,
): GitReconcileNeedsBranchBind {
  if (!isRecord(value)) {
    throw new Error(`Invalid API response: ${path} must be object`);
  }
  return {
    path: parseString(value.path, `${path}.path`),
    branch: parseString(value.branch, `${path}.branch`),
  };
}

function parseGitReconcileReport(value: unknown): GitReconcileReport {
  if (!isRecord(value)) {
    throw new Error("Invalid API response: report must be object");
  }
  const skippedRaw = value.worktrees_skipped;
  const needsBindRaw = value.needs_branch_bind;
  const skipped: GitReconcileSkipped[] = Array.isArray(skippedRaw)
    ? skippedRaw.map((row, i) => parseGitReconcileSkipped(row, `report.worktrees_skipped[${i}]`))
    : [];
  const needsBranchBind: GitReconcileNeedsBranchBind[] = Array.isArray(needsBindRaw)
    ? needsBindRaw.map((row, i) =>
        parseGitReconcileNeedsBranchBind(row, `report.needs_branch_bind[${i}]`),
      )
    : [];
  return {
    repo_path_updated: value.repo_path_updated === true,
    worktrees_path_updated:
      typeof value.worktrees_path_updated === "number" && Number.isFinite(value.worktrees_path_updated)
        ? value.worktrees_path_updated
        : 0,
    worktrees_added:
      typeof value.worktrees_added === "number" && Number.isFinite(value.worktrees_added)
        ? value.worktrees_added
        : 0,
    worktrees_removed:
      typeof value.worktrees_removed === "number" && Number.isFinite(value.worktrees_removed)
        ? value.worktrees_removed
        : 0,
    branches_head_updated:
      typeof value.branches_head_updated === "number" && Number.isFinite(value.branches_head_updated)
        ? value.branches_head_updated
        : 0,
    resolution_source:
      typeof value.resolution_source === "string" ? value.resolution_source : undefined,
    discovered_path:
      typeof value.discovered_path === "string" ? value.discovered_path : undefined,
    worktrees_skipped: skipped,
    needs_branch_bind: needsBranchBind,
  };
}

export function parseGitReconcileResult(raw: unknown): GitReconcileResult {
  if (!isRecord(raw)) {
    throw new Error("Invalid API response: body must be object");
  }
  return {
    status: parseString(raw.status, "status"),
    report: parseGitReconcileReport(raw.report),
  };
}
