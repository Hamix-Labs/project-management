import { respondGitRoute } from "./gitRoutes";

/** Responds to global `/git/*` REST paths (ADR-0037). */
export function respondGlobalGitApi(url: string, method = "GET"): Response | null {
  return respondGitRoute(url, method, "/git", "global");
}
