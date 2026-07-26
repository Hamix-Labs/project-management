import { apiErrorFromResponse, fetchWithTimeout } from "./shared";
import { maxRepoSearchQueryBytes } from "./repo";

/** Same ceiling as other `/repo/*` fetches in `repo.ts`. */
const repoFetchTimeoutMs = 45_000;

function assertRepoSearchQueryLength(q: string): void {
  if (q.length > maxRepoSearchQueryBytes) {
    throw new Error("search query is too long");
  }
}

export type RepoSearchEntryKind = "file" | "dir";

export type RepoSearchEntry = {
  path: string;
  kind: RepoSearchEntryKind;
};

export type RepoSearchKinds = {
  file?: boolean;
  dir?: boolean;
};

/** Path entries (files and/or dirs) under the worktree matching q, or null if repo unavailable. */
export async function searchRepoEntries(
  q: string,
  options: {
    signal?: AbortSignal;
    worktreeId: string;
    kinds?: RepoSearchKinds;
  },
): Promise<RepoSearchEntry[] | null> {
  assertRepoSearchQueryLength(q);
  const worktreeId = options.worktreeId.trim();
  if (worktreeId === "") {
    throw new Error("worktree_id is required");
  }
  const params = new URLSearchParams({ q, worktree_id: worktreeId });
  const kindParts: string[] = [];
  if (options.kinds) {
    if (options.kinds.file) kindParts.push("file");
    if (options.kinds.dir) kindParts.push("dir");
    if (kindParts.length === 0) {
      kindParts.push("file");
    }
  } else {
    kindParts.push("file", "dir");
  }
  params.set("kinds", kindParts.join(","));
  const res = await fetchWithTimeout(
    `/repo/search?${params}`,
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
    throw new Error("unexpected search response");
  }
  const body = raw as { entries?: unknown; paths?: unknown };
  if (Array.isArray(body.entries)) {
    return body.entries
      .map((row): RepoSearchEntry | null => {
        if (row === null || typeof row !== "object") return null;
        const o = row as { path?: unknown; kind?: unknown };
        if (typeof o.path !== "string") return null;
        if (o.kind !== "file" && o.kind !== "dir") return null;
        return { path: o.path, kind: o.kind };
      })
      .filter((e): e is RepoSearchEntry => e !== null);
  }
  if (Array.isArray(body.paths)) {
    return body.paths
      .filter((p): p is string => typeof p === "string")
      .map((path) => ({ path, kind: "file" as const }));
  }
  throw new Error("unexpected search response");
}

export type RepoSymbolKind = "function" | "method" | "class";

export type RepoSymbolHit = {
  path: string;
  name: string;
  line: number;
  kind: RepoSymbolKind;
};

/** Best-effort symbol hits under the worktree matching q, or null if repo unavailable. */
export async function searchRepoSymbols(
  q: string,
  options: { signal?: AbortSignal; worktreeId: string },
): Promise<RepoSymbolHit[] | null> {
  assertRepoSearchQueryLength(q);
  const worktreeId = options.worktreeId.trim();
  if (worktreeId === "") {
    throw new Error("worktree_id is required");
  }
  const params = new URLSearchParams({ q, worktree_id: worktreeId });
  const res = await fetchWithTimeout(
    `/repo/symbols?${params}`,
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
  if (
    raw === null ||
    typeof raw !== "object" ||
    !("symbols" in raw) ||
    !Array.isArray((raw as { symbols: unknown }).symbols)
  ) {
    throw new Error("unexpected symbols response");
  }
  const out: RepoSymbolHit[] = [];
  for (const row of (raw as { symbols: unknown[] }).symbols) {
    if (row === null || typeof row !== "object") continue;
    const o = row as {
      path?: unknown;
      name?: unknown;
      line?: unknown;
      kind?: unknown;
    };
    if (typeof o.path !== "string" || typeof o.name !== "string") continue;
    if (typeof o.line !== "number" || !Number.isFinite(o.line) || o.line < 1) {
      continue;
    }
    if (o.kind !== "function" && o.kind !== "method" && o.kind !== "class") {
      continue;
    }
    out.push({ path: o.path, name: o.name, line: o.line, kind: o.kind });
  }
  return out;
}
