import { http, HttpResponse } from "msw";
import {
  FACTORY_GIT_WORKTREE_ID,
  gitWorktreeDetailFactory,
} from "../factories/git";
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
    http.get(new RegExp(`${base}/repositories/.+/branches`), () =>
      HttpResponse.json(gitRouteJsonBody({ kind: "branches-list" }, "global")),
    ),
    http.get(new RegExp(`${base}/repositories/.+/projects`), () =>
      HttpResponse.json(gitRouteJsonBody({ kind: "repo-projects" }, "global")),
    ),
    http.get(`${base}/worktrees/:worktreeId`, ({ params }) => {
      const id = String(params.worktreeId ?? "");
      if (id !== FACTORY_GIT_WORKTREE_ID) {
        return HttpResponse.json(
          { error: "worktree not found", code: "worktree_not_found" },
          { status: 404 },
        );
      }
      return HttpResponse.json(gitWorktreeDetailFactory({ id }));
    }),
    http.post(`${base}/repositories`, () =>
      HttpResponse.json(gitRouteJsonBody({ kind: "repositories-create" }, "global"), {
        status: 201,
      }),
    ),
  ];
}
