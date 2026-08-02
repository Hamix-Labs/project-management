import { apiErrorFromResponse, fetchWithTimeout } from "./shared";

/** Match pkgs/tasks/handler/repo_handlers.go and docs/api.md (abuse guards). */
export const maxRepoPathQueryBytes = 4096;
export const maxRepoSearchQueryBytes = 512;
export const maxRepoLineQueryParamBytes = 32;
export const maxRepoShaQueryBytes = 64;

const repoShaPattern = /^[0-9a-fA-F]{7,40}$/;

export const repoQueryKeys = {
  all: ["repo"] as const,
  diff: (worktreeId: string, sha: string) =>
    [...repoQueryKeys.all, "diff", worktreeId, sha] as const,
  file: (worktreeId: string, path: string) =>
    [...repoQueryKeys.all, "file", worktreeId, path] as const,
};

/**
 * Hard ceiling on every `/repo/*` and `/health/ready` fetch in this module
 * (probe, search, validate-range, file). Long enough to absorb a cold
 * filesystem walk on first hit, short enough that a hung backend cannot
 * pin a UI request forever.
 */
const repoFetchTimeoutMs = 45_000;

function assertRepoRelPath(path: string): string {
  const t = path.trim();
  if (t.length === 0) {
    throw new Error("path is required");
  }
  if (t.length > maxRepoPathQueryBytes) {
    throw new Error("path is too long");
  }
  return t;
}

function assertRepoSearchQueryLength(q: string): void {
  if (q.length > maxRepoSearchQueryBytes) {
    throw new Error("search query is too long");
  }
}

function assertRepoLineQueryParam(name: string, n: number): string {
  if (!Number.isFinite(n) || !Number.isInteger(n) || n < 1) {
    throw new Error(`${name} must be a positive integer`);
  }
  const s = String(n);
  if (s.length > maxRepoLineQueryParamBytes) {
    throw new Error(`${name} is too large`);
  }
  return s;
}

/** Result of probing whether taskapi has a usable workspace repo (see GET /health/ready). */
export type RepoWorkspaceProbe =
  | { state: "available" }
  | { state: "unavailable" }
  | { state: "broken" }
  | { state: "unknown" };

/**
 * Lightweight check: can `/repo/search` resolve the given task worktree?
 * Uses an empty `q` so the handler opens the worktree without walking paths.
 */
export async function probeWorktreeRepo(
  worktreeId: string,
  options?: { signal?: AbortSignal },
): Promise<RepoWorkspaceProbe> {
  const id = worktreeId.trim();
  if (id === "") {
    return { state: "unavailable" };
  }
  try {
    const paths = await searchRepoFiles("", {
      worktreeId: id,
      signal: options?.signal,
    });
    if (paths === null) {
      return { state: "unavailable" };
    }
    return { state: "available" };
  } catch {
    return { state: "unknown" };
  }
}

/**
 * Lightweight check: does the running taskapi have a workspace repo configured (via Settings) and on disk?
 * Prefer this over GET /repo/search?q= on mount (avoids walking the tree).
 */
export async function probeRepoWorkspace(
  options?: { signal?: AbortSignal },
): Promise<RepoWorkspaceProbe> {
  try {
    const res = await fetchWithTimeout(
      "/health/ready",
      {
        headers: { Accept: "application/json" },
        signal: options?.signal,
      },
      { timeoutMs: repoFetchTimeoutMs },
    );
    let raw: unknown;
    try {
      raw = await res.json();
    } catch {
      return { state: "unknown" };
    }
    if (raw === null || typeof raw !== "object") {
      return { state: "unknown" };
    }
    const body = raw as {
      status?: string;
      checks?: Record<string, string>;
    };
    const checks = body.checks ?? {};
    const st = body.status ?? "";

    if (!res.ok) {
      if (
        st === "degraded" &&
        checks.database === "ok" &&
        checks.workspace_repo === "fail"
      ) {
        return { state: "broken" };
      }
      return { state: "unknown" };
    }

    if (st === "ok" && checks.database === "ok") {
      if (checks.workspace_repo === "ok") return { state: "available" };
      if (checks.workspace_repo === undefined) return { state: "unavailable" };
      if (checks.workspace_repo === "fail") return { state: "broken" };
    }

    return { state: "unknown" };
  } catch {
    return { state: "unknown" };
  }
}

/** File paths under the configured workspace repo matching q, or null if repo is not configured (409/503). */
export async function searchRepoFiles(
  q: string,
  options?: { signal?: AbortSignal; worktreeId?: string },
): Promise<string[] | null> {
  assertRepoSearchQueryLength(q);
  const params = new URLSearchParams({ q });
  const worktreeId = options?.worktreeId?.trim();
  if (worktreeId) {
    params.set("worktree_id", worktreeId);
  }
  const res = await fetchWithTimeout(
    `/repo/search?${params}`,
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
  if (
    raw !== null &&
    typeof raw === "object" &&
    "paths" in raw &&
    Array.isArray((raw as { paths: unknown }).paths)
  ) {
    return (raw as { paths: string[] }).paths.filter(
      (p): p is string => typeof p === "string",
    );
  }
  throw new Error("unexpected search response");
}

export type RepoValidateRangeResult = {
  ok: boolean;
  line_count?: number;
  warning?: string;
};

/** Returns null if repo is not configured (503). */
export async function validateRepoRange(
  path: string,
  start: number,
  end: number,
  options?: { signal?: AbortSignal; worktreeId?: string },
): Promise<RepoValidateRangeResult | null> {
  const p = assertRepoRelPath(path);
  const params = new URLSearchParams({
    path: p,
    start: assertRepoLineQueryParam("start", start),
    end: assertRepoLineQueryParam("end", end),
  });
  const worktreeId = options?.worktreeId?.trim();
  if (worktreeId) {
    params.set("worktree_id", worktreeId);
  }
  const res = await fetchWithTimeout(
    `/repo/validate-range?${params}`,
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
  if (raw !== null && typeof raw === "object" && "ok" in raw) {
    const o = raw as {
      ok: boolean;
      line_count?: number;
      warning?: string;
    };
    return {
      ok: Boolean(o.ok),
      line_count: typeof o.line_count === "number" ? o.line_count : undefined,
      warning: typeof o.warning === "string" ? o.warning : undefined,
    };
  }
  throw new Error("unexpected validate-range response");
}

export type RepoFileResult = {
  path: string;
  content: string;
  binary: boolean;
  truncated: boolean;
  size_bytes: number;
  line_count: number;
  warning?: string;
};

/** Full file text for @ line-range UI, or null if repo is not configured (503). */
export async function fetchRepoFile(
  path: string,
  options?: { signal?: AbortSignal; worktreeId?: string },
): Promise<RepoFileResult | null> {
  const p = assertRepoRelPath(path);
  const params = new URLSearchParams({ path: p });
  const worktreeId = options?.worktreeId?.trim();
  if (worktreeId) {
    params.set("worktree_id", worktreeId);
  }
  const res = await fetchWithTimeout(
    `/repo/file?${params}`,
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
    throw new Error("unexpected file response");
  }
  const o = raw as Record<string, unknown>;
  const pathVal = o.path;
  const contentVal = o.content;
  const binaryVal = o.binary;
  const truncatedVal = o.truncated;
  const sizeVal = o.size_bytes;
  const linesVal = o.line_count;
  if (
    typeof pathVal !== "string" ||
    typeof contentVal !== "string" ||
    typeof binaryVal !== "boolean" ||
    typeof truncatedVal !== "boolean" ||
    typeof sizeVal !== "number" ||
    typeof linesVal !== "number"
  ) {
    throw new Error("unexpected file response shape");
  }
  const out: RepoFileResult = {
    path: pathVal,
    content: contentVal,
    binary: binaryVal,
    truncated: truncatedVal,
    size_bytes: sizeVal,
    line_count: linesVal,
  };
  if (typeof o.warning === "string") {
    out.warning = o.warning;
  }
  return out;
}

export type RepoDiffResult = {
  sha: string;
  patch: string;
  truncated: boolean;
  size_bytes: number;
  author?: string;
  author_email?: string;
  parent_sha?: string;
  files_changed?: number;
  insertions?: number;
  deletions?: number;
};

function assertRepoSha(sha: string): string {
  const t = sha.trim();
  if (t.length === 0) {
    throw new Error("sha is required");
  }
  if (t.length > maxRepoShaQueryBytes) {
    throw new Error("sha is too long");
  }
  if (!repoShaPattern.test(t)) {
    throw new Error("invalid sha");
  }
  return t;
}

export function parseRepoDiffResponse(raw: unknown): RepoDiffResult {
  if (raw === null || typeof raw !== "object") {
    throw new Error("unexpected diff response");
  }
  const o = raw as Record<string, unknown>;
  const shaVal = o.sha;
  const patchVal = o.patch;
  const truncatedVal = o.truncated;
  const sizeVal = o.size_bytes;
  if (
    typeof shaVal !== "string" ||
    typeof patchVal !== "string" ||
    typeof truncatedVal !== "boolean" ||
    typeof sizeVal !== "number"
  ) {
    throw new Error("unexpected diff response shape");
  }
  const out: RepoDiffResult = {
    sha: shaVal,
    patch: patchVal,
    truncated: truncatedVal,
    size_bytes: sizeVal,
  };
  if (typeof o.author === "string" && o.author !== "") {
    out.author = o.author;
  }
  if (typeof o.author_email === "string" && o.author_email !== "") {
    out.author_email = o.author_email;
  }
  if (typeof o.parent_sha === "string" && o.parent_sha !== "") {
    out.parent_sha = o.parent_sha;
  }
  if (typeof o.files_changed === "number") {
    out.files_changed = o.files_changed;
  }
  if (typeof o.insertions === "number") {
    out.insertions = o.insertions;
  }
  if (typeof o.deletions === "number") {
    out.deletions = o.deletions;
  }
  return out;
}

/** Unified diff for one commit, or null if worktree repo is unavailable (409/503). */
export async function fetchRepoCommitDiff(
  sha: string,
  options: { worktreeId: string; signal?: AbortSignal },
): Promise<RepoDiffResult | null> {
  const s = assertRepoSha(sha);
  const worktreeId = options.worktreeId.trim();
  if (worktreeId === "") {
    throw new Error("worktree_id is required");
  }
  const params = new URLSearchParams({
    sha: s,
    worktree_id: worktreeId,
  });
  const res = await fetchWithTimeout(
    `/repo/diff?${params}`,
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
  return parseRepoDiffResponse(raw);
}
