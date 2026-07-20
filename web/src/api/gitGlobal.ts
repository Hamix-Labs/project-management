import type {
  GitBranch,
  GitReconcileResult,
  GitRepository,
  GitWorktree,
} from "@/types/git";
import type { ProjectListResponse } from "@/types/project";
import { parseProjectListResponse } from "./projects";
import {
  parseGitBranchList,
  parseGitRepository,
  parseGitRepositoryList,
  parseGitWorktreeList,
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
