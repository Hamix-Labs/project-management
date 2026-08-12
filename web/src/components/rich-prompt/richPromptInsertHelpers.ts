import type { Editor } from "@tiptap/react";
import type { RepoWorkspaceProbe } from "@/api";

export type PendingFileInsert = {
  insertAt: number;
  path: string;
};

export type RepoHintFlags = {
  showSelectWorktreeHint: boolean;
  showSelectRepositoryHint: boolean;
  showRepoMisconfigHint: boolean;
  workspaceBroken: boolean;
  fileSearchFailedWhileAvailable: boolean;
  showRepoUnknownHint: boolean;
  showFileSearchSpinner: boolean;
};

export function insertRepoFileMentionAt(
  editor: Editor,
  insertAt: number,
  path: string,
  lineStart?: number,
  lineEnd?: number,
) {
  const attrs =
    lineStart != null && lineEnd != null
      ? { path, lineStart, lineEnd }
      : { path };
  editor
    .chain()
    .focus()
    .insertContentAt(insertAt, [
      { type: "repoFileMention", attrs },
      { type: "text", text: " " },
    ])
    .run();
}

export function computeRepoHintFlags(
  workspaceProbe: RepoWorkspaceProbe | "pending",
  fileSearchUnavailable: boolean,
  showFileSearchSpinner: boolean,
  options: {
    /** Resolved mention worktree id (may be "" while resolving / unset). */
    mentionWorktreeId?: string;
    /** True when the editor participates in git-scoped @ mentions. */
    gitScoped: boolean;
    preferRepositoryHint?: boolean;
    repositoryId?: string;
  },
): RepoHintFlags {
  const mentionSelected = options.mentionWorktreeId?.trim() !== "";
  const probeDone = workspaceProbe !== "pending";
  const preferRepositoryHint = options.preferRepositoryHint === true;
  const repositorySelected = options.repositoryId?.trim() !== "";

  const showSelectRepositoryHint =
    options.gitScoped &&
    preferRepositoryHint &&
    probeDone &&
    !repositorySelected &&
    !mentionSelected;

  const showSelectWorktreeHint =
    options.gitScoped &&
    !preferRepositoryHint &&
    probeDone &&
    !mentionSelected;

  const showRepoMisconfigHint =
    options.gitScoped &&
    mentionSelected &&
    probeDone &&
    (workspaceProbe.state === "unavailable" ||
      workspaceProbe.state === "broken" ||
      (workspaceProbe.state === "available" && fileSearchUnavailable));
  const showRepoUnknownHint =
    options.gitScoped &&
    mentionSelected &&
    probeDone &&
    workspaceProbe.state === "unknown";

  return {
    showSelectWorktreeHint,
    showSelectRepositoryHint,
    showRepoMisconfigHint,
    workspaceBroken:
      workspaceProbe !== "pending" && workspaceProbe.state === "broken",
    fileSearchFailedWhileAvailable:
      workspaceProbe !== "pending" &&
      workspaceProbe.state === "available" &&
      fileSearchUnavailable,
    showRepoUnknownHint,
    showFileSearchSpinner,
  };
}
