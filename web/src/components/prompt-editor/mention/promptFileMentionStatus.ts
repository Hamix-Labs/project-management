/**
 * Outcome of the most recent `@` file-mention lookup.
 *
 * Every distinguishable failure gets its own variant. A single sticky
 * "unavailable" boolean cannot tell "the repository is not configured" from
 * "the worktree was deleted" from "we have not resolved a worktree yet", so it
 * reported search failures for prompts that never issued a request.
 */
export type MentionSearchStatus =
  /** Nothing has been searched yet. */
  | { kind: "idle" }
  /** No worktree and no repository to fall back to. */
  | { kind: "unbound" }
  /** Repository worktree lookup is in flight. */
  | { kind: "resolving" }
  /** Repository is known but exposes no main worktree row. */
  | { kind: "no-main-worktree" }
  /** Repository worktree lookup itself failed. */
  | { kind: "unresolved" }
  | { kind: "searching" }
  | { kind: "ready"; matched: number }
  /** 409 / 503 — repo root is not configured for this worktree. */
  | { kind: "no-repo" }
  /** 404 — the worktree row is gone. */
  | { kind: "worktree-missing" }
  /** 500 — the worktree path is missing or not a directory. */
  | { kind: "worktree-broken" }
  /** Query exceeded the byte ceiling the API accepts. */
  | { kind: "query-rejected" }
  /** The request hit the client-side fetch deadline. */
  | { kind: "timed-out" }
  | { kind: "failed"; status?: number };

export type MentionHintTone = "pending" | "info" | "error";

/**
 * Renderable description of a status. Kept as data rather than JSX so the
 * copy can be unit tested without a renderer.
 */
export type MentionSearchHint = {
  tone: MentionHintTone;
  message: string;
  action?: { label: string; href: string };
  /** Sentence fragment rendered after `action`. */
  trailing?: string;
};

const repositoriesPage = { label: "Repositories page", href: "/repositories" };

/** Hint to show under the editor, or null when the status needs no words. */
export function describeMentionSearchStatus(
  status: MentionSearchStatus,
): MentionSearchHint | null {
  switch (status.kind) {
    case "idle":
    case "ready":
      return null;
    case "unbound":
      return {
        tone: "info",
        message:
          "Bind a repository to this prompt to reference files with @.",
      };
    case "resolving":
      return { tone: "pending", message: "Finding the repository worktree…" };
    case "searching":
      return { tone: "pending", message: "Searching files…" };
    case "no-main-worktree":
      return {
        tone: "error",
        message: "This repository has no main worktree registered. Add one on the",
        action: repositoriesPage,
        trailing: "to reference files with @.",
      };
    case "unresolved":
      return {
        tone: "error",
        message:
          "Could not look up the repository worktree. Check your connection and reopen the menu.",
      };
    case "no-repo":
      return {
        tone: "error",
        message: "File search is not available for the selected worktree. Check the repository on the",
        action: repositoriesPage,
        trailing: "to enable @ file mentions.",
      };
    case "worktree-missing":
      return {
        tone: "error",
        message: "The selected worktree no longer exists. Rebind this prompt from the",
        action: repositoriesPage,
        trailing: "to reference files again.",
      };
    case "worktree-broken":
      return {
        tone: "error",
        message:
          "The selected worktree path is missing or not a directory. Update the repository on the",
        action: repositoriesPage,
        trailing: "to restore @ file mentions.",
      };
    case "query-rejected":
      return {
        tone: "info",
        message: "That search text is too long — shorten it to keep searching files.",
      };
    case "timed-out":
      return {
        tone: "error",
        message:
          "File search timed out. The repository may be very large — try a narrower search.",
      };
    case "failed":
      return {
        tone: "error",
        message: "File search failed for the selected worktree. Check the repository on the",
        action: repositoriesPage,
        trailing: "or inspect the server logs.",
      };
  }
}

/**
 * Maps a rejected `searchRepoFiles` call to a status.
 *
 * `status` is the `ApiError.status` when the rejection carried one. The repo
 * handler answers 404 for an unknown worktree id and 500 when the recorded
 * path will not open, which is the only signal separating "deleted" from
 * "broken" — its `reason` field is dropped by the shared error parser.
 */
export function mentionStatusFromHttpStatus(
  status: number | undefined,
): MentionSearchStatus {
  if (status === 404) return { kind: "worktree-missing" };
  if (status === 500) return { kind: "worktree-broken" };
  return { kind: "failed", status };
}
