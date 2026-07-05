import { http, HttpResponse } from "msw";
import { FACTORY_REPO_DEFAULT_PROJECT_ID } from "../factories/project";

export function projectsListEmpty(limit = 100) {
  return http.get("/projects", () =>
    HttpResponse.json({ projects: [], limit }),
  );
}

export function projectContextEmpty(
  projectId = FACTORY_REPO_DEFAULT_PROJECT_ID,
  limit = 100,
) {
  return http.get(`/projects/${projectId}/context`, () =>
    HttpResponse.json({ items: [], edges: [], limit }),
  );
}
