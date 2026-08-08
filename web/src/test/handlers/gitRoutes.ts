import type { JsonBodyType } from "msw";
import {
  globalGitBranchesResponse,
  globalGitRepositoriesResponse,
  globalGitWorktreesResponse,
  gitBranchFactory,
  gitRepositoryFactory,
  gitWorktreeFactory,
} from "../factories/git";
import { FACTORY_REPO_DEFAULT_PROJECT_ID, repoDefaultProjectFactory } from "../factories/project";

export type GitRouteScope = "project" | "global";

export type GitRouteMatch =
  | { kind: "repositories-list" }
  | { kind: "repositories-create" }
  | { kind: "repositories-get" }
  | { kind: "worktrees-list" }
  | { kind: "branches-list" }
  | { kind: "repo-projects" };

export function matchGitRoute(
  url: string,
  method: string,
  base: string,
  scope: GitRouteScope,
): GitRouteMatch | null {
  if (method === "POST" && url.endsWith(`${base}/repositories`) && scope === "global") {
    return { kind: "repositories-create" };
  }
  if (method !== "GET") return null;

  if (url.endsWith(`${base}/repositories`)) return { kind: "repositories-list" };
  if (url.includes(`${base}/repositories/`) && url.endsWith("/worktrees")) {
    return { kind: "worktrees-list" };
  }
  if (url.includes(`${base}/repositories/`) && url.endsWith("/branches")) {
    return { kind: "branches-list" };
  }
  if (scope === "global" && url.includes(`${base}/repositories/`) && url.endsWith("/projects")) {
    return { kind: "repo-projects" };
  }
  if (
    scope === "global" &&
    url.includes(`${base}/repositories/`) &&
    !url.slice(url.indexOf(`${base}/repositories/`) + `${base}/repositories/`.length).includes("/")
  ) {
    return { kind: "repositories-get" };
  }
  return null;
}

export function gitRouteJsonBody(match: GitRouteMatch, scope: GitRouteScope): JsonBodyType {
  switch (match.kind) {
    case "repositories-list":
      if (scope === "global") return globalGitRepositoriesResponse();
      return {
        repositories: [
          { ...gitRepositoryFactory(), project_id: FACTORY_REPO_DEFAULT_PROJECT_ID },
        ],
      };
    case "repositories-create": {
      const body = globalGitRepositoriesResponse() as { repositories: unknown[] };
      return body.repositories[0] as JsonBodyType;
    }
    case "repositories-get":
      return gitRepositoryFactory() as JsonBodyType;
    case "worktrees-list":
      return scope === "global"
        ? globalGitWorktreesResponse()
        : { worktrees: [gitWorktreeFactory()] };
    case "branches-list":
      return scope === "global"
        ? globalGitBranchesResponse()
        : { branches: [gitBranchFactory()] };
    case "repo-projects":
      return { projects: [repoDefaultProjectFactory()], limit: 100 };
  }
}

export function respondGitRoute(
  url: string,
  method: string,
  base: string,
  scope: GitRouteScope,
): Response | null {
  const match = matchGitRoute(url, method, base, scope);
  if (!match) return null;
  const body = gitRouteJsonBody(match, scope);
  const status = match.kind === "repositories-create" ? 201 : 200;
  return Response.json(body, { status });
}
