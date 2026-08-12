import { fetchRepoFilesPage } from "@/api/repoFiles";

export type RepoFileIndexStatus = "idle" | "warming" | "ready" | "error";

export type RepoFileIndexSnapshot = {
  worktreeId: string;
  status: RepoFileIndexStatus;
  paths: readonly string[];
  version: number;
  errorMessage?: string;
};

type Listener = () => void;

type Entry = {
  status: RepoFileIndexStatus;
  paths: string[];
  version: number;
  errorMessage?: string;
  abort: AbortController | null;
  listeners: Set<Listener>;
};

const entries = new Map<string, Entry>();

function getOrCreate(worktreeId: string): Entry {
  let ent = entries.get(worktreeId);
  if (!ent) {
    ent = {
      status: "idle",
      paths: [],
      version: 0,
      abort: null,
      listeners: new Set(),
    };
    entries.set(worktreeId, ent);
  }
  return ent;
}

function notify(ent: Entry) {
  for (const listener of ent.listeners) {
    listener();
  }
}

export function getRepoFileIndexSnapshot(
  worktreeId: string,
): RepoFileIndexSnapshot {
  const id = worktreeId.trim();
  if (id === "") {
    return { worktreeId: "", status: "idle", paths: [], version: 0 };
  }
  const ent = getOrCreate(id);
  return {
    worktreeId: id,
    status: ent.status,
    paths: ent.paths,
    version: ent.version,
    errorMessage: ent.errorMessage,
  };
}

export function subscribeRepoFileIndex(
  worktreeId: string,
  listener: Listener,
): () => void {
  const id = worktreeId.trim();
  if (id === "") {
    return () => {};
  }
  const ent = getOrCreate(id);
  ent.listeners.add(listener);
  return () => {
    ent.listeners.delete(listener);
  };
}

/** Start or resume warming the in-memory file index for a worktree. */
export function warmRepoFileIndex(worktreeId: string): void {
  const id = worktreeId.trim();
  if (id === "") return;
  const ent = getOrCreate(id);
  if (ent.status === "warming" || ent.status === "ready") return;

  ent.abort?.abort();
  const ac = new AbortController();
  ent.abort = ac;
  ent.paths = [];
  ent.status = "warming";
  ent.errorMessage = undefined;
  ent.version += 1;
  notify(ent);

  void (async () => {
    let after: string | undefined;
    try {
      for (;;) {
        const page = await fetchRepoFilesPage({
          worktreeId: id,
          after,
          limit: 500,
          signal: ac.signal,
        });
        if (ac.signal.aborted) return;
        if (page === null) {
          ent.status = "error";
          ent.errorMessage = "File index unavailable for this worktree.";
          ent.version += 1;
          notify(ent);
          return;
        }
        if (page.paths.length > 0) {
          ent.paths = ent.paths.concat(page.paths);
          ent.version += 1;
          notify(ent);
        }
        if (!page.has_more || !page.next_after) {
          ent.status = "ready";
          ent.version += 1;
          notify(ent);
          return;
        }
        after = page.next_after;
      }
    } catch (err) {
      if (ac.signal.aborted) return;
      ent.status = "error";
      ent.errorMessage =
        err instanceof Error ? err.message : "File index failed to load.";
      ent.version += 1;
      notify(ent);
    }
  })();
}

/** Cancel warm and drop cached paths (e.g. repository switched). */
export function clearRepoFileIndex(worktreeId?: string): void {
  if (worktreeId === undefined) {
    for (const [id, ent] of entries) {
      ent.abort?.abort();
      entries.delete(id);
    }
    return;
  }
  const id = worktreeId.trim();
  const ent = entries.get(id);
  if (!ent) return;
  ent.abort?.abort();
  entries.delete(id);
}

/** Test helper: seed index state without network. */
export function seedRepoFileIndexForTest(
  worktreeId: string,
  paths: string[],
  status: RepoFileIndexStatus = "ready",
): void {
  clearRepoFileIndex(worktreeId);
  const ent = getOrCreate(worktreeId.trim());
  ent.paths = paths.slice();
  ent.status = status;
  ent.version += 1;
  notify(ent);
}
