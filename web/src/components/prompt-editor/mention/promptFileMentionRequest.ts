import type { QueryClient } from "@tanstack/react-query";
import { ApiError, listRepoFiles, repoQueryKeys, type RepoFileList } from "@/api";
import type {
  FileWorktreeGap,
  FileWorktreeResolution,
} from "../usePromptEditorFileWorktree";
import type { PromptFileMentionItem } from "./PromptEditorMentionMenu";
import { rankMentionPaths } from "./promptFileMentionRank";
import {
  mentionStatusFromHttpStatus,
  type MentionSearchStatus,
} from "./promptFileMentionStatus";

/**
 * How long a cached listing is trusted before the next `@` refetches it. Long
 * enough that a burst of mentions costs one request, short enough that files an
 * agent just created show up while the operator is still writing the prompt.
 */
export const mentionFileListStaleTimeMs = 60_000;

/** A status only ever describes the binding that produced it. */
export type BoundMentionStatus = {
  worktreeId: string | undefined;
  value: MentionSearchStatus;
};

export type MentionSearchOutcome = BoundMentionStatus & {
  items: PromptFileMentionItem[];
};

function statusForGap(gap: FileWorktreeGap | undefined): MentionSearchStatus {
  switch (gap) {
    case "no-main-worktree":
      return { kind: "no-main-worktree" };
    case "lookup-failed":
      return { kind: "unresolved" };
    default:
      return { kind: "unbound" };
  }
}

function isAbortError(error: unknown): boolean {
  return (error as { name?: string } | null)?.name === "AbortError";
}

/**
 * Runs one `@` lookup against the cached file list.
 *
 * Never throws: the caller hands the promise to BlockNote, which attaches no
 * rejection handler and would leave the menu loading forever.
 */
export async function runPromptFileMentionSearch({
  query,
  worktree,
  queryClient,
  onSelectPath,
  onProgress,
}: {
  query: string;
  worktree: FileWorktreeResolution;
  queryClient: QueryClient;
  onSelectPath: (path: string) => void;
  onProgress: (progress: BoundMentionStatus) => void;
}): Promise<MentionSearchOutcome> {
  const empty = (
    worktreeId: string | undefined,
    value: MentionSearchStatus,
  ): MentionSearchOutcome => ({ worktreeId, value, items: [] });

  try {
    let target = worktree.worktreeId?.trim();
    if (!target && worktree.resolving) {
      onProgress({ worktreeId: undefined, value: { kind: "resolving" } });
      target = (await worktree.whenResolved())?.trim();
    }
    if (!target) {
      return empty(undefined, statusForGap(worktree.gap));
    }

    const worktreeId = target;
    const cached = queryClient.getQueryData<RepoFileList | null>(
      repoQueryKeys.files(worktreeId),
    );
    // Only announce a wait when there is nothing cached to rank yet; a warm
    // list answers within the frame and a spinner would only flicker.
    if (cached === undefined) {
      onProgress({ worktreeId, value: { kind: "searching" } });
    }

    const listing = await queryClient.fetchQuery({
      queryKey: repoQueryKeys.files(worktreeId),
      queryFn: ({ signal }) => listRepoFiles(worktreeId, { signal }),
      staleTime: mentionFileListStaleTimeMs,
    });

    if (listing == null) return empty(worktreeId, { kind: "no-repo" });
    if (listing.paths.length === 0) {
      return empty(worktreeId, { kind: "empty-repo" });
    }

    const ranked = rankMentionPaths(listing.paths, query);
    const items = ranked.map((path) => ({
      title: path,
      query,
      onItemClick: () => onSelectPath(path),
    }));
    return {
      worktreeId,
      value: {
        kind: "ready",
        matched: items.length,
        truncated: listing.truncated,
      },
      items,
    };
  } catch (error) {
    const target = worktree.worktreeId?.trim();
    if (isAbortError(error)) {
      return empty(target, { kind: "timed-out" });
    }
    return empty(
      target,
      error instanceof ApiError
        ? mentionStatusFromHttpStatus(error.status)
        : { kind: "failed" },
    );
  }
}
