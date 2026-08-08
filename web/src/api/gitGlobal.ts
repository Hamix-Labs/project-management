import type {
  GitBranch,
  GitRepository,
  GitWorktree,
  GitWorktreeDetail,
} from "@/types/git";
import type { ProjectListResponse } from "@/types/project";
import { parseProjectListResponse } from "./projects";
import {
  parseGitBranchList,
  parseGitRepository,
  parseGitRepositoryList,
  parseGitWorktreeDetail,
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

export async function getGlobalGitWorktree(
  worktreeId: string,
  options?: { signal?: AbortSignal },
): Promise<GitWorktreeDetail> {
  const wtId = assertTaskPathId(worktreeId, "worktree id");
  const raw = await gitFetchJson(
    `${gitRoot}/worktrees/${encodeURIComponent(wtId)}`,
    gitJsonGetInit(options?.signal),
  );
  return parseGitWorktreeDetail(raw);
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
