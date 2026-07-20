import { http, HttpResponse, type JsonBodyType } from "msw";
import type {
  GitBranch,
  GitReconcileResult,
  GitRepository,
  GitWorktree,
  GitWorktreeCheckoutStatus,
} from "@/types/git";
import {
  FACTORY_GIT_REPO_ID,
  gitBranchFactory,
  gitRepositoryFactory,
  gitWorktreeFactory,
} from "../factories/git";

/** GET /git/repositories — explicit list (empty, error, or fixtures). */
export function gitRepositoriesListOk(
  repositories: GitRepository[] = [gitRepositoryFactory()],
) {
  return http.get("/git/repositories", () =>
    HttpResponse.json({ repositories }),
  );
}

export function gitRepositoriesListEmpty() {
  return gitRepositoriesListOk([]);
}

export function gitRepositoriesListError(
  status = 404,
  error = "Not Found",
) {
  return http.get("/git/repositories", () =>
    HttpResponse.json({ error }, { status }),
  );
}

export function gitRepositoryGet(
  repoId: string,
  repository: Partial<GitRepository> = {},
) {
  return http.get(`/git/repositories/${repoId}`, () =>
    HttpResponse.json({
      ...gitRepositoryFactory({ id: repoId }),
      ...repository,
    }),
  );
}

export function gitRepositoryWorktreesList(
  repoId: string,
  worktrees: GitWorktree[],
) {
  return http.get(`/git/repositories/${repoId}/worktrees`, () =>
    HttpResponse.json({ worktrees }),
  );
}

export function gitRepositoryBranchesList(
  repoId: string,
  branches: GitBranch[],
) {
  return http.get(`/git/repositories/${repoId}/branches`, () =>
    HttpResponse.json({ branches }),
  );
}

export function gitRepositoryCheckoutStatus(
  repoId: string,
  worktrees: GitWorktreeCheckoutStatus[],
) {
  return http.get(
    `/git/repositories/${repoId}/worktrees/checkout-status`,
    () => HttpResponse.json({ worktrees }),
  );
}

export function gitRepositorySyncOk(
  repoId: string,
  onSync?: () => void,
  result: GitReconcileResult = {
    status: "ok",
    report: {
      repo_path_updated: false,
      worktrees_path_updated: 0,
      worktrees_added: 0,
      worktrees_removed: 0,
      branches_head_updated: 0,
      worktrees_skipped: [],
      needs_branch_bind: [],
    },
  },
) {
  return http.post(`/git/repositories/${repoId}/sync`, () => {
    onSync?.();
    return HttpResponse.json(result);
  });
}

/** DELETE /git/worktrees/:id — 409 has_running_task by default. */
export function gitWorktreeDeleteConflict(
  error = "task still running",
  code = "has_running_task",
) {
  return http.delete(/\/git\/worktrees\/.+/, () =>
    HttpResponse.json({ error, code }, { status: 409 }),
  );
}

/** Default detail page handlers used by RepositoryWorktreesPage tests. */
export function repositoryDetailHandlers(options?: {
  repoId?: string;
  repository?: Partial<GitRepository>;
  worktrees?: GitWorktree[];
  branches?: GitBranch[];
  checkoutStatus?: GitWorktreeCheckoutStatus[];
  onSync?: () => void;
  deleteBody?: JsonBodyType;
  deleteStatus?: number;
}) {
  const repoId = options?.repoId ?? FACTORY_GIT_REPO_ID;
  const wtMain = "00000000-0000-4000-8000-000000000020";
  const wtB = "00000000-0000-4000-8000-000000000030";
  const branchId = "00000000-0000-4000-8000-000000000040";
  const mainBranchId = "00000000-0000-4000-8000-000000000041";

  const defaultWorktrees: GitWorktree[] = [
    gitWorktreeFactory({
      id: wtMain,
      repository_id: repoId,
      path: "/repo/main",
      name: "main",
      is_main: true,
      branch_id: mainBranchId,
    }),
    gitWorktreeFactory({
      id: wtB,
      repository_id: repoId,
      path: "/repo/feature",
      name: "feature",
      is_main: false,
      branch_id: branchId,
    }),
  ];

  const defaultBranches: GitBranch[] = [
    gitBranchFactory({
      id: mainBranchId,
      repository_id: repoId,
      name: "main",
      head_sha: "def456",
    }),
    gitBranchFactory({
      id: branchId,
      repository_id: repoId,
      name: "feature",
      head_sha: "abc123",
    }),
  ];

  const defaultCheckout: GitWorktreeCheckoutStatus[] = [
    {
      worktree_id: wtMain,
      available: true,
      dirty: false,
      detached: false,
      head_commit_at: "2026-06-22T12:00:00Z",
      has_upstream: true,
      ahead: 0,
      behind: 0,
      upstream: "origin/main",
    },
    {
      worktree_id: wtB,
      available: true,
      dirty: false,
      detached: false,
      head_commit_at: "2026-06-22T12:00:00Z",
      has_upstream: true,
      ahead: 0,
      behind: 0,
      upstream: "origin/feature",
    },
  ];

  const deleteStatus = options?.deleteStatus ?? 409;
  const deleteBody = options?.deleteBody ?? {
    error: "task still running",
    code: "has_running_task",
  };

  return [
    gitRepositoryGet(repoId, options?.repository),
    gitRepositoryCheckoutStatus(repoId, options?.checkoutStatus ?? defaultCheckout),
    gitRepositoryWorktreesList(repoId, options?.worktrees ?? defaultWorktrees),
    gitRepositoryBranchesList(repoId, options?.branches ?? defaultBranches),
    gitRepositorySyncOk(repoId, options?.onSync),
    http.delete(/\/git\/worktrees\/.+/, () =>
      HttpResponse.json(deleteBody, { status: deleteStatus }),
    ),
  ];
}

export { FACTORY_GIT_REPO_ID };
