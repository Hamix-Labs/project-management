import type {
  GitBranch,
  GitLiveBranch,
  GitLiveWorktree,
  GitReconcileInput,
  GitReconcileResult,
  GitRepository,
  GitWorktree,
  GitWorktreeBranchBind,
  GitWorktreeCheckoutStatus,
  GitWorktreeProbe,
} from "@/types/git";
import type { ProjectListResponse } from "@/types/project";
import { parseProjectListResponse } from "./projects";
import {
  parseGitBranchList,
  parseGitLiveBranchList,
  parseGitLiveWorktreeList,
  parseGitRepository,
  parseGitRepositoryList,
  parseGitWorktree,
  parseGitWorktreeCheckoutStatusList,
  parseGitWorktreeList,
  parseGitWorktreeProbe,
  parseGitReconcileResult,
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

const gitRoot = gitApiRoot();

export async function listGlobalGitRepositories(
  options?: { signal?: AbortSignal },
): Promise<GitRepository[]> {
  const raw = await gitFetchJson(`${gitRoot}/repositories`, gitJsonGetInit(options?.signal));
  return parseGitRepositoryList(raw);
}

export async function createGlobalGitRepository(input: {
  path: string;
  host_path?: string;
  default_branch?: string;
}): Promise<GitRepository> {
  const raw = await gitFetchJson(`${gitRoot}/repositories`, gitJsonPostInit(input));
  return parseGitRepository(raw);
}

export async function getGlobalGitRepository(
  repositoryId: string,
  options?: { signal?: AbortSignal },
): Promise<GitRepository> {
  const repoId = assertTaskPathId(repositoryId, "repository id");
  const raw = await gitFetchJson(
    `${gitRoot}/repositories/${encodeURIComponent(repoId)}`,
    gitJsonGetInit(options?.signal),
  );
  return parseGitRepository(raw);
}

export async function deleteGlobalGitRepository(repositoryId: string): Promise<void> {
  const repoId = assertTaskPathId(repositoryId, "repository id");
  await gitFetchVoid(`${gitRoot}/repositories/${encodeURIComponent(repoId)}`, gitDeleteInit());
}

export async function listGlobalGitWorktrees(
  repositoryId: string,
  options?: { signal?: AbortSignal },
): Promise<GitWorktree[]> {
  const repoId = assertTaskPathId(repositoryId, "repository id");
  const raw = await gitFetchJson(
    `${gitRoot}/repositories/${encodeURIComponent(repoId)}/worktrees`,
    gitJsonGetInit(options?.signal),
  );
  return parseGitWorktreeList(raw);
}

export async function listGlobalGitWorktreeCheckoutStatus(
  repositoryId: string,
  options?: { signal?: AbortSignal },
): Promise<GitWorktreeCheckoutStatus[]> {
  const repoId = assertTaskPathId(repositoryId, "repository id");
  const raw = await gitFetchJson(
    `${gitRoot}/repositories/${encodeURIComponent(repoId)}/worktrees/checkout-status`,
    gitJsonGetInit(options?.signal),
  );
  return parseGitWorktreeCheckoutStatusList(raw);
}

export async function createGlobalGitWorktree(
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
    `${gitRoot}/repositories/${encodeURIComponent(repoId)}/worktrees`,
    gitJsonPostInit(input),
  );
  return parseGitWorktree(raw);
}

export async function registerGlobalGitWorktree(
  repositoryId: string,
  input: {
    path: string;
    name?: string;
    branch?: GitWorktreeBranchBind;
  },
): Promise<GitWorktree> {
  const repoId = assertTaskPathId(repositoryId, "repository id");
  const raw = await gitFetchJson(
    `${gitRoot}/repositories/${encodeURIComponent(repoId)}/worktrees/register`,
    gitJsonPostInit(input),
  );
  return parseGitWorktree(raw);
}

export async function unregisterGlobalGitWorktree(worktreeId: string): Promise<void> {
  const wtId = assertTaskPathId(worktreeId, "worktree id");
  await gitFetchVoid(`${gitRoot}/worktrees/${encodeURIComponent(wtId)}`, gitDeleteInit());
}

export async function deleteGlobalGitWorktreeFromDisk(
  worktreeId: string,
  options?: { force?: boolean },
): Promise<void> {
  const wtId = assertTaskPathId(worktreeId, "worktree id");
  const params = new URLSearchParams({ remove_from_disk: "true" });
  if (options?.force) {
    params.set("force", "true");
  }
  await gitFetchVoid(
    `${gitRoot}/worktrees/${encodeURIComponent(wtId)}?${params.toString()}`,
    gitDeleteInit(),
  );
}

export async function listGlobalGitBranches(
  repositoryId: string,
  options?: { signal?: AbortSignal },
): Promise<GitBranch[]> {
  const repoId = assertTaskPathId(repositoryId, "repository id");
  const raw = await gitFetchJson(
    `${gitRoot}/repositories/${encodeURIComponent(repoId)}/branches`,
    gitJsonGetInit(options?.signal),
  );
  return parseGitBranchList(raw);
}

export async function listGlobalGitLiveWorktrees(
  repositoryId: string,
  options?: { signal?: AbortSignal },
): Promise<GitLiveWorktree[]> {
  const repoId = assertTaskPathId(repositoryId, "repository id");
  const raw = await gitFetchJson(
    `${gitRoot}/repositories/${encodeURIComponent(repoId)}/worktrees/live`,
    gitJsonGetInit(options?.signal),
  );
  return parseGitLiveWorktreeList(raw);
}

export async function probeGlobalGitWorktree(
  repositoryId: string,
  path: string,
  options?: { signal?: AbortSignal },
): Promise<GitWorktreeProbe> {
  const repoId = assertTaskPathId(repositoryId, "repository id");
  const params = new URLSearchParams({ path });
  const raw = await gitFetchJson(
    `${gitRoot}/repositories/${encodeURIComponent(repoId)}/worktrees/probe?${params}`,
    gitJsonGetInit(options?.signal),
  );
  return parseGitWorktreeProbe(raw);
}

export async function listGlobalGitLiveBranches(
  repositoryId: string,
  options?: { signal?: AbortSignal },
): Promise<GitLiveBranch[]> {
  const repoId = assertTaskPathId(repositoryId, "repository id");
  const raw = await gitFetchJson(
    `${gitRoot}/repositories/${encodeURIComponent(repoId)}/branches/live`,
    gitJsonGetInit(options?.signal),
  );
  return parseGitLiveBranchList(raw);
}

export async function reconcileGlobalGitRepository(
  repositoryId: string,
  input?: GitReconcileInput,
): Promise<GitReconcileResult> {
  const repoId = assertTaskPathId(repositoryId, "repository id");
  const body: GitReconcileInput = input ?? {};
  const raw = await gitFetchJson(
    `${gitRoot}/repositories/${encodeURIComponent(repoId)}/reconcile`,
    gitJsonPostInit(body),
  );
  return parseGitReconcileResult(raw);
}

export async function syncGlobalGitRepository(
  repositoryId: string,
): Promise<GitReconcileResult> {
  const repoId = assertTaskPathId(repositoryId, "repository id");
  const raw = await gitFetchJson(
    `${gitRoot}/repositories/${encodeURIComponent(repoId)}/sync`,
    gitJsonPostInit({}),
  );
  return parseGitReconcileResult(raw);
}

export async function relocateGlobalGitRepository(
  repositoryId: string,
  input: { path: string },
): Promise<GitReconcileResult> {
  const repoId = assertTaskPathId(repositoryId, "repository id");
  const raw = await gitFetchJson(
    `${gitRoot}/repositories/${encodeURIComponent(repoId)}/relocate`,
    gitJsonPostInit(input),
  );
  return parseGitReconcileResult(raw);
}

export async function listProjectsByRepository(
  repositoryId: string,
  options?: { signal?: AbortSignal },
): Promise<ProjectListResponse> {
  const repoId = assertTaskPathId(repositoryId, "repository id");
  const raw = await gitFetchJson(
    `${gitRoot}/repositories/${encodeURIComponent(repoId)}/projects`,
    gitJsonGetInit(options?.signal),
  );
  return parseProjectListResponse(raw);
}
