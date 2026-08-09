import { repoFetchTimeoutMs } from "./repo";
import { apiErrorFromResponse, fetchWithTimeout } from "./shared";

/** How GET /repo/files produced a listing; only "git" applies gitignore rules. */
export type RepoFileListSource = "git" | "walk";

export type RepoFileList = {
  paths: string[];
  truncated: boolean;
  source: RepoFileListSource;
};

/**
 * Every referenceable file under a worktree, or null when the repo is not
 * configured (409/503).
 *
 * Unfiltered by design: callers cache this once per worktree and rank locally,
 * so typing in the `@` menu costs no request at all.
 */
export async function listRepoFiles(
  worktreeId: string,
  options?: { signal?: AbortSignal },
): Promise<RepoFileList | null> {
  const id = worktreeId.trim();
  if (id === "") {
    throw new Error("worktree_id is required");
  }
  const res = await fetchWithTimeout(
    `/repo/files?${new URLSearchParams({ worktree_id: id })}`,
    {
      headers: { Accept: "application/json" },
      signal: options?.signal,
    },
    { timeoutMs: repoFetchTimeoutMs },
  );
  if (res.status === 503 || res.status === 409) {
    return null;
  }
  if (!res.ok) throw await apiErrorFromResponse(res);
  const raw: unknown = await res.json();
  if (raw === null || typeof raw !== "object") {
    throw new Error("unexpected file list response");
  }
  const body = raw as { paths?: unknown; truncated?: unknown; source?: unknown };
  if (!Array.isArray(body.paths)) {
    throw new Error("unexpected file list response");
  }
  return {
    paths: body.paths.filter((p): p is string => typeof p === "string"),
    truncated: body.truncated === true,
    source: body.source === "walk" ? "walk" : "git",
  };
}
