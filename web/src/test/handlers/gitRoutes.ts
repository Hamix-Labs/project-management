import type { JsonBodyType } from "msw";
import {
  globalGitBranchesResponse,
  globalGitLiveBranchesResponse,
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
  | { kind: "worktrees-list" }
  | { kind: "worktrees-live" }
  | { kind: "worktrees-probe"; path: string }
  | { kind: "branches-list" }
  | { kind: "branches-live" }
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
  if (url.includes(`${base}/repositories/`) && url.endsWith("/worktrees/probe")) {
    const probePath = new URL(url, "http://local").searchParams.get("path") ?? "";
    return { kind: "worktrees-probe", path: probePath };
  }
  if (url.includes(`${base}/repositories/`) && url.endsWith("/worktrees/live")) {
    return { kind: "worktrees-live" };
  }
  if (url.includes(`${base}/repositories/`) && url.endsWith("/worktrees")) {
    return { kind: "worktrees-list" };
  }
  if (url.includes(`${base}/repositories/`) && url.endsWith("/branches/live")) {
    return { kind: "branches-live" };
  }
  if (url.includes(`${base}/repositories/`) && url.endsWith("/branches")) {
    return { kind: "branches-list" };
  }
  if (scope === "global" && url.includes(`${base}/repositories/`) && url.endsWith("/projects")) {
    return { kind: "repo-projects" };
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
    case "worktrees-list":
      return scope === "global"
        ? globalGitWorktreesResponse()
        : { worktrees: [gitWorktreeFactory()] };
    case "worktrees-live":
      return {
        worktrees: [
          {
            path: "/repo/main",
            branch: "main",
            is_main: true,
            detached: false,
            registered: false,
          },
        ],
      };
    case "worktrees-probe": {
      const linked = match.path.includes("/repo/");
      return {
        path: match.path || "/repo/wt-feature",
        linked,
        is_main: false,
        branch: linked ? "feature" : "",
        registered: false,
      };
    }
    case "branches-list":
      return scope === "global"
        ? globalGitBranchesResponse()
        : { branches: [gitBranchFactory()] };
    case "branches-live":
      return globalGitLiveBranchesResponse();
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
