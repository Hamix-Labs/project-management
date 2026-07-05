import { DEFAULT_PROJECT_ID } from "@/types";
import type { JsonBodyType } from "msw";
import {
  FACTORY_GIT_BRANCH_ID,
  FACTORY_GIT_REPO_ID,
  FACTORY_GIT_WORKTREE_ID,
  gitBranchFactory,
  gitRepositoryFactory,
  gitWorktreeFactory,
} from "../factories/git";
import { respondGitRoute } from "./gitRoutes";

export const GIT_TEST_REPO_ID = FACTORY_GIT_REPO_ID;
export const GIT_TEST_WORKTREE_ID = FACTORY_GIT_WORKTREE_ID;
export const GIT_TEST_BRANCH_ID = FACTORY_GIT_BRANCH_ID;

export function gitRepositoriesResponse(): JsonBodyType {
  return {
    repositories: [
      { ...gitRepositoryFactory(), project_id: DEFAULT_PROJECT_ID },
    ],
  };
}

export function gitWorktreesResponse(): JsonBodyType {
  return { worktrees: [gitWorktreeFactory()] };
}

export function gitBranchesResponse(): JsonBodyType {
  return { branches: [gitBranchFactory()] };
}

/** Responds to project-scoped git REST paths used by the Worktrees UI and task create modal. */
export function respondGitApi(url: string, method = "GET"): Response | null {
  const base = `/projects/${DEFAULT_PROJECT_ID}/git`;
  return respondGitRoute(url, method, base, "project");
}
