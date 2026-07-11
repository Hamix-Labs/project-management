import type { Editor } from "@tiptap/react";
import type { RepoWorkspaceProbe } from "@/api";
import type { ProjectContextItem } from "@/types";

export type PendingFileInsert = {
  insertAt: number;
  path: string;
};

export type PendingProjectChoice = {
  item: ProjectContextItem;
  insertAt: number | null;
};

export type RepoHintFlags = {
  showSelectWorktreeHint: boolean;
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

export function insertProjectContextChipAt(
  editor: Editor,
  item: ProjectContextItem,
  insertAt: number | null,
) {
  const chip = {
    type: "projectContextMention",
    attrs: { id: item.id, title: item.title || "" },
  } as const;
  const text = { type: "text", text: " " } as const;
  if (insertAt == null) {
    // Came from the "Choose context" chooser, not the `#` plugin —
    // append the chip to the current cursor instead of jumping to a
    // stale document position from a different surface.
    editor.chain().focus().insertContent([chip, text]).run();
    return;
  }
  editor.chain().focus().insertContentAt(insertAt, [chip, text]).run();
}

export function computeRepoHintFlags(
  workspaceProbe: RepoWorkspaceProbe | "pending",
  fileSearchUnavailable: boolean,
  showFileSearchSpinner: boolean,
  worktreeId?: string,
): RepoHintFlags {
  const worktreeScoped = worktreeId !== undefined;
  const worktreeSelected = worktreeId?.trim() !== "";
  const probeDone = workspaceProbe !== "pending";
  const showSelectWorktreeHint =
    worktreeScoped && probeDone && !worktreeSelected;
  const showRepoMisconfigHint =
    worktreeScoped &&
    worktreeSelected &&
    probeDone &&
    (workspaceProbe.state === "unavailable" ||
      workspaceProbe.state === "broken" ||
      (workspaceProbe.state === "available" && fileSearchUnavailable));
  const showRepoUnknownHint =
    worktreeScoped &&
    worktreeSelected &&
    probeDone &&
    workspaceProbe.state === "unknown";

  return {
    showSelectWorktreeHint,
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
