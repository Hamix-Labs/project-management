import type { GitBranch, GitReconcileResult, GitRepository, GitWorktree } from "@/types/git";
import {
  parseGitBranch,
  parseGitBranchList,
  parseGitReconcileResult,
  parseGitRepository,
  parseGitRepositoryList,
  parseGitWorktree,
  parseGitWorktreeList,
} from "./parseGitApi";
import { assertTaskPathId } from "./taskRequestBounds";
import {
  gitApiRoot,
  gitDeleteInit,
  gitFetchJson,
  gitFetchVoid,
  gitJsonGetInit,
  gitJsonPostInit,
} from "./gitClient";

export async function listGitRepositories(
  projectId: string,
  options?: { signal?: AbortSignal },
): Promise<GitRepository[]> {
  const raw = await gitFetchJson(
    `${gitApiRoot(projectId)}/repositories`,
    gitJsonGetInit(options?.signal),
  );
  return parseGitRepositoryList(raw);
}

export async function createGitRepository(
  projectId: string,
  input: { path: string; host_path?: string; default_branch?: string },
): Promise<GitRepository> {
  const raw = await gitFetchJson(
    `${gitApiRoot(projectId)}/repositories`,
    gitJsonPostInit(input),
  );
  return parseGitRepository(raw);
}

export async function deleteGitRepository(
  projectId: string,
  repositoryId: string,
): Promise<void> {
  const repoId = assertTaskPathId(repositoryId, "repository id");
  await gitFetchVoid(
    `${gitApiRoot(projectId)}/repositories/${encodeURIComponent(repoId)}`,
    gitDeleteInit(),
  );
}

export async function listGitWorktrees(
  projectId: string,
  repositoryId: string,
  options?: { signal?: AbortSignal },
): Promise<GitWorktree[]> {
  const repoId = assertTaskPathId(repositoryId, "repository id");
  const raw = await gitFetchJson(
    `${gitApiRoot(projectId)}/repositories/${encodeURIComponent(repoId)}/worktrees`,
    gitJsonGetInit(options?.signal),
  );
  return parseGitWorktreeList(raw);
}

export async function createGitWorktree(
  projectId: string,
  repositoryId: string,
  input: {
    path: string;
    name?: string;
    branch: string;
    create_branch?: boolean;
    start_point?: string;
  },
): Promise<GitWorktree> {
  const repoId = assertTaskPathId(repositoryId, "repository id");
  const raw = await gitFetchJson(
    `${gitApiRoot(projectId)}/repositories/${encodeURIComponent(repoId)}/worktrees`,
    gitJsonPostInit(input),
  );
  return parseGitWorktree(raw);
}

export async function unregisterGitWorktree(
  projectId: string,
  worktreeId: string,
): Promise<void> {
  const wtId = assertTaskPathId(worktreeId, "worktree id");
  await gitFetchVoid(
    `${gitApiRoot(projectId)}/worktrees/${encodeURIComponent(wtId)}`,
    gitDeleteInit(),
  );
}

export async function listGitBranches(
  projectId: string,
  repositoryId: string,
  options?: { signal?: AbortSignal },
): Promise<GitBranch[]> {
  const repoId = assertTaskPathId(repositoryId, "repository id");
  const raw = await gitFetchJson(
    `${gitApiRoot(projectId)}/repositories/${encodeURIComponent(repoId)}/branches`,
    gitJsonGetInit(options?.signal),
  );
  return parseGitBranchList(raw);
}

export async function createGitBranch(
  projectId: string,
  repositoryId: string,
  input: { name: string; start_point?: string },
): Promise<GitBranch> {
  const repoId = assertTaskPathId(repositoryId, "repository id");
  const raw = await gitFetchJson(
    `${gitApiRoot(projectId)}/repositories/${encodeURIComponent(repoId)}/branches`,
    gitJsonPostInit(input),
  );
  return parseGitBranch(raw);
}

export async function deleteGitBranch(
  projectId: string,
  branchId: string,
  options?: { force?: boolean },
): Promise<void> {
  const bid = assertTaskPathId(branchId, "branch id");
  const params = new URLSearchParams();
  if (options?.force) params.set("force", "true");
  const qs = params.toString();
  await gitFetchVoid(
    `${gitApiRoot(projectId)}/branches/${encodeURIComponent(bid)}${qs ? `?${qs}` : ""}`,
    gitDeleteInit(),
  );
}

export async function reconcileGitRepository(
  projectId: string,
  repositoryId: string,
): Promise<GitReconcileResult> {
  const repoId = assertTaskPathId(repositoryId, "repository id");
  const raw = await gitFetchJson(
    `${gitApiRoot(projectId)}/repositories/${encodeURIComponent(repoId)}/reconcile`,
    gitJsonPostInit({}),
  );
  return parseGitReconcileResult(raw);
}
