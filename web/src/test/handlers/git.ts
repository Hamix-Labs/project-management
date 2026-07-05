import type { JsonBodyType } from "msw";
import {
  FACTORY_GIT_BRANCH_ID,
  FACTORY_GIT_REPO_ID,
  FACTORY_GIT_WORKTREE_ID,
  gitBranchFactory,
  gitRepositoryFactory,
  gitWorktreeFactory,
} from "../factories/git";
import { FACTORY_REPO_DEFAULT_PROJECT_ID, repoDefaultProjectFactory } from "../factories/project";
import { respondGitRoute } from "./gitRoutes";

export const GIT_TEST_REPO_ID = FACTORY_GIT_REPO_ID;
export const GIT_TEST_WORKTREE_ID = FACTORY_GIT_WORKTREE_ID;
export const GIT_TEST_BRANCH_ID = FACTORY_GIT_BRANCH_ID;
export const GIT_TEST_DEFAULT_PROJECT_ID = FACTORY_REPO_DEFAULT_PROJECT_ID;

export function gitRepositoriesResponse(): JsonBodyType {
  return {
    repositories: [gitRepositoryFactory()],
  };
}

export function gitWorktreesResponse(): JsonBodyType {
  return { worktrees: [gitWorktreeFactory()] };
}

export function gitBranchesResponse(): JsonBodyType {
  return { branches: [gitBranchFactory()] };
}

export function repoProjectsResponse(): JsonBodyType {
  return { projects: [repoDefaultProjectFactory()], limit: 100 };
}

/** Responds to project-scoped git REST paths used by legacy tests. */
export function respondGitApi(url: string, method = "GET"): Response | null {
  const base = `/projects/${FACTORY_REPO_DEFAULT_PROJECT_ID}/git`;
  return respondGitRoute(url, method, base, "project");
}
