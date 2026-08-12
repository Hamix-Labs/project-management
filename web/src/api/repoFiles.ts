import { apiErrorFromResponse, fetchWithTimeout } from "./shared";

/** Match web/src/api/repo.ts maxRepoSearchQueryBytes / docs/api.md. */
const maxRepoSearchQueryBytes = 512;
const repoFetchTimeoutMs = 45_000;
const defaultRepoFilesPageLimit = 500;

export type RepoFilesPage = {
  paths: string[];
  next_after?: string;
  has_more: boolean;
  source: string;
  truncated?: boolean;
};

function assertRepoSearchQueryLength(q: string): void {
  if (q.length > maxRepoSearchQueryBytes) {
    throw new Error("search query is too long");
  }
}

/** One cursor page of gitignore-aware paths for `@` index warm / fallback search. */
export async function fetchRepoFilesPage(options: {
  worktreeId: string;
  q?: string;
  after?: string;
  limit?: number;
  signal?: AbortSignal;
}): Promise<RepoFilesPage | null> {
  const worktreeId = options.worktreeId.trim();
  if (worktreeId === "") {
    throw new Error("worktree_id is required");
  }
  const q = options.q ?? "";
  assertRepoSearchQueryLength(q);
  const params = new URLSearchParams({
    worktree_id: worktreeId,
    limit: String(options.limit ?? defaultRepoFilesPageLimit),
  });
  if (q !== "") {
    params.set("q", q);
  }
  if (options.after?.trim()) {
    params.set("after", options.after.trim());
  }
  const res = await fetchWithTimeout(
    `/repo/files?${params}`,
    {
      headers: { Accept: "application/json" },
      signal: options.signal,
    },
    { timeoutMs: repoFetchTimeoutMs },
  );
  if (res.status === 503 || res.status === 409) {
    return null;
  }
  if (!res.ok) throw await apiErrorFromResponse(res);
  const raw: unknown = await res.json();
  if (raw === null || typeof raw !== "object") {
    throw new Error("unexpected files response");
  }
  const o = raw as Record<string, unknown>;
  if (!Array.isArray(o.paths) || typeof o.has_more !== "boolean") {
    throw new Error("unexpected files response shape");
  }
  const paths = o.paths.filter((p): p is string => typeof p === "string");
  const page: RepoFilesPage = {
    paths,
    has_more: o.has_more,
    source: typeof o.source === "string" ? o.source : "git",
  };
  if (typeof o.next_after === "string" && o.next_after !== "") {
    page.next_after = o.next_after;
  }
  if (typeof o.truncated === "boolean") {
    page.truncated = o.truncated;
  }
  return page;
}
