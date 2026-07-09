import { http, HttpResponse } from "msw";
import { gitRouteJsonBody } from "./gitRoutes";

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
