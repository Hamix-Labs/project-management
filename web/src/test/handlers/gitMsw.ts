import { http, HttpResponse } from "msw";
import { DEFAULT_PROJECT_ID } from "@/types";
import { gitRouteJsonBody } from "./gitRoutes";

/** MSW handlers for project-scoped git REST paths. */
export function gitApiHandlers() {
  const base = `/projects/${DEFAULT_PROJECT_ID}/git`;
  return [
    http.get(`${base}/repositories`, () =>
      HttpResponse.json(gitRouteJsonBody({ kind: "repositories-list" }, "project")),
    ),
    http.get(new RegExp(`${base}/repositories/.+/worktrees`), () =>
      HttpResponse.json(gitRouteJsonBody({ kind: "worktrees-list" }, "project")),
    ),
    http.get(new RegExp(`${base}/repositories/.+/branches`), () =>
      HttpResponse.json(gitRouteJsonBody({ kind: "branches-list" }, "project")),
    ),
  ];
}

/** MSW handlers for global `/git/*` REST paths. */
export function globalGitApiHandlers() {
  const base = "/git";
  return [
    http.get(`${base}/repositories`, () =>
      HttpResponse.json(gitRouteJsonBody({ kind: "repositories-list" }, "global")),
    ),
    http.get(new RegExp(`${base}/repositories/.+/worktrees`), () =>
      HttpResponse.json(gitRouteJsonBody({ kind: "worktrees-list" }, "global")),
    ),
    http.get(new RegExp(`${base}/repositories/.+/branches/live`), () =>
      HttpResponse.json(gitRouteJsonBody({ kind: "branches-live" }, "global")),
    ),
    http.get(new RegExp(`${base}/repositories/.+/branches`), () =>
      HttpResponse.json(gitRouteJsonBody({ kind: "branches-list" }, "global")),
    ),
    http.get(new RegExp(`${base}/repositories/.+/projects`), () =>
      HttpResponse.json(gitRouteJsonBody({ kind: "repo-projects" }, "global")),
    ),
    http.post(`${base}/repositories`, () =>
      HttpResponse.json(gitRouteJsonBody({ kind: "repositories-create" }, "global"), {
        status: 201,
      }),
    ),
  ];
}
