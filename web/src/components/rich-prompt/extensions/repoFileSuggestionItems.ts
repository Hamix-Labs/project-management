import { fetchRepoFilesPage } from "@/api/repoFiles";
import { rankRepoFilePaths } from "@/lib/rankRepoFilePaths";
import { getRepoFileIndexSnapshot, warmRepoFileIndex } from "@/lib/repoFileIndex";
import type { RepoSuggestionItem } from "./repoFileSuggestionList";

export function itemsFromIndex(
  worktreeId: string,
  query: string,
): { items: RepoSuggestionItem[]; indexing: boolean; emptyBecauseCold: boolean } {
  const snap = getRepoFileIndexSnapshot(worktreeId);
  const indexing = snap.status === "warming" || snap.status === "idle";
  const ranked = rankRepoFilePaths([...snap.paths], query);
  return {
    items: ranked.map((path) => ({ path })),
    indexing,
    emptyBecauseCold: indexing && snap.paths.length === 0,
  };
}

export async function resolveSuggestionItems(
  worktreeId: string,
  query: string,
): Promise<{
  items: RepoSuggestionItem[];
  unavailable: boolean;
  available: boolean;
}> {
  warmRepoFileIndex(worktreeId);
  const snap = getRepoFileIndexSnapshot(worktreeId);
  if (snap.status === "error") {
    return { items: [], unavailable: true, available: false };
  }
  const fromIndex = itemsFromIndex(worktreeId, query);
  if (
    fromIndex.items.length > 0 ||
    query.trim() === "" ||
    snap.status === "ready"
  ) {
    return {
      items: fromIndex.items,
      unavailable: false,
      available: fromIndex.items.length > 0 || snap.status === "ready",
    };
  }

  const page = await fetchRepoFilesPage({
    worktreeId,
    q: query,
    limit: 200,
  });
  if (page === null) {
    return { items: [], unavailable: true, available: false };
  }
  return {
    items: rankRepoFilePaths(page.paths, query).map((path) => ({ path })),
    unavailable: false,
    available: true,
  };
}
