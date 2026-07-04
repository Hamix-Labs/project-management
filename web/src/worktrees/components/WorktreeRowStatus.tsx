import type { GitWorktreeCheckoutStatus } from "@/types/git";
import { useNow } from "@/shared/useNow";
import { formatWorktreeStatusRelativeTime } from "../formatWorktreeStatusRelativeTime";
import { worktreeGitCopy } from "../worktreeGitCopy";

type Props = {
  checkoutStatus?: GitWorktreeCheckoutStatus;
};

export function WorktreeRowStatus({ checkoutStatus }: Props) {
  const now = useNow({ intervalMs: 60_000 });

  if (!checkoutStatus) {
    return null;
  }

  if (!checkoutStatus.available) {
    return (
      <div className="worktree-row__status" title={worktreeGitCopy.statusUnavailableTitle}>
        <span className="worktree-row__status-label">{worktreeGitCopy.statusUnavailable}</span>
      </div>
    );
  }

  let tone: "clean" | "dirty" | "detached" = "clean";
  let label: string = worktreeGitCopy.statusClean;
  if (checkoutStatus.dirty) {
    tone = "dirty";
    label = worktreeGitCopy.statusDirty;
  } else if (checkoutStatus.detached) {
    tone = "detached";
    label = worktreeGitCopy.detachedHead;
  }

  const relative = checkoutStatus.head_commit_at
    ? formatWorktreeStatusRelativeTime(checkoutStatus.head_commit_at, new Date(now))
    : "";
  const lastCommitLabel = relative ? worktreeGitCopy.statusLastCommit(relative) : "";

  return (
    <div className="worktree-row__status">
      <span className={`worktree-row__status-dot worktree-row__status-dot--${tone}`} aria-hidden />
      <span className="worktree-row__status-label">{label}</span>
      {lastCommitLabel ? (
        <>
          <span className="worktree-row__status-sep" aria-hidden>
            ·
          </span>
          <span
            className="worktree-row__status-time"
            title={checkoutStatus.head_commit_at}
          >
            {lastCommitLabel}
          </span>
        </>
      ) : null}
    </div>
  );
}
