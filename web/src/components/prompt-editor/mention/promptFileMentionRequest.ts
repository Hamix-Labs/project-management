import { ApiError, maxRepoSearchQueryBytes, searchRepoFiles } from "@/api";
import type {
  FileWorktreeGap,
  FileWorktreeResolution,
} from "../usePromptEditorFileWorktree";
import type { PromptFileMentionItem } from "./PromptEditorMentionMenu";
import {
  mentionStatusFromHttpStatus,
  type MentionSearchStatus,
} from "./promptFileMentionStatus";

/** BlockNote renders the whole list; a local index lifts this later. */
const maxMentionItems = 20;

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

function toMentionItems(
  paths: string[],
  query: string,
  onSelectPath: (path: string) => void,
): PromptFileMentionItem[] {
  return paths.slice(0, maxMentionItems).map((rawPath) => {
    const path = rawPath.replace(/\\/g, "/");
    return {
      title: path,
      query,
      onItemClick: () => onSelectPath(path),
    };
  });
}

/**
 * Runs one `@` lookup. Never throws: the caller hands the promise to BlockNote,
 * which attaches no rejection handler and would leave the menu loading forever.
 */
export async function runPromptFileMentionSearch({
  query,
  worktree,
  controller,
  onSelectPath,
  onProgress,
}: {
  query: string;
  worktree: FileWorktreeResolution;
  controller: AbortController;
  onSelectPath: (path: string) => void;
  onProgress: (progress: BoundMentionStatus) => void;
}): Promise<MentionSearchOutcome> {
  const empty = (
    worktreeId: string | undefined,
    value: MentionSearchStatus,
  ): MentionSearchOutcome => ({ worktreeId, value, items: [] });

  try {
    if (query.length > maxRepoSearchQueryBytes) {
      return empty(worktree.worktreeId, { kind: "query-rejected" });
    }

    let target = worktree.worktreeId?.trim();
    if (!target && worktree.resolving) {
      onProgress({ worktreeId: undefined, value: { kind: "resolving" } });
      target = (await worktree.whenResolved())?.trim();
    }
    if (!target) {
      return empty(undefined, statusForGap(worktree.gap));
    }

    onProgress({ worktreeId: target, value: { kind: "searching" } });
    const paths = await searchRepoFiles(query, {
      worktreeId: target,
      signal: controller.signal,
    });
    if (controller.signal.aborted) return empty(target, { kind: "idle" });
    if (paths == null) return empty(target, { kind: "no-repo" });

    const items = toMentionItems(paths, query, onSelectPath);
    return { worktreeId: target, value: { kind: "ready", matched: items.length }, items };
  } catch (error) {
    const target = worktree.worktreeId?.trim();
    if (isAbortError(error)) {
      // Our own abort means a newer keystroke took over; any other abort came
      // from the fetch deadline, which the user should hear about.
      return empty(
        target,
        controller.signal.aborted ? { kind: "idle" } : { kind: "timed-out" },
      );
    }
    return empty(
      target,
      error instanceof ApiError
        ? mentionStatusFromHttpStatus(error.status)
        : { kind: "failed" },
    );
  }
}
