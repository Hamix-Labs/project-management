import { http, HttpResponse, type JsonBodyType } from "msw";
import { FACTORY_REPO_DEFAULT_PROJECT_ID } from "../factories/project";
import {
  FACTORY_GIT_REPO_ID,
  gitRepositoryFactory,
} from "../factories/git";

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

/** GET /git/repositories — list used by project create/default labeling. */
export function gitRepositoriesList(
  repositories: JsonBodyType[] = [gitRepositoryFactory()],
) {
  return http.get("/git/repositories", () =>
    HttpResponse.json({ repositories }),
  );
}

/** POST /projects — create success; captures body via onPost. */
export function projectCreate(
  created: JsonBodyType,
  onPost?: (body: unknown) => void,
) {
  return http.post("/projects", async ({ request }) => {
    const body = await request.json();
    onPost?.(body);
    return HttpResponse.json(created, { status: 201 });
  });
}

export { FACTORY_GIT_REPO_ID };
