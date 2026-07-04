import type { GitWorktreeCheckoutStatus } from "@/types/git";
import { worktreeGitCopy } from "../worktreeGitCopy";

type Props = {
  checkoutStatus?: GitWorktreeCheckoutStatus;
};

export function BranchSyncIndicator({ checkoutStatus }: Props) {
  if (!checkoutStatus?.available || checkoutStatus.detached || !checkoutStatus.has_upstream) {
    return null;
  }

  const ahead = checkoutStatus.ahead ?? 0;
  const behind = checkoutStatus.behind ?? 0;

  if (ahead === 0 && behind === 0) {
    return (
      <span className="worktree-row__branch-sync">{worktreeGitCopy.syncUpToDate}</span>
    );
  }

  const parts: string[] = [];
  if (ahead > 0) {
    parts.push(`↑ ${ahead}`);
  }
  if (behind > 0) {
    parts.push(`↓ ${behind}`);
  }

  return <span className="worktree-row__branch-sync">{parts.join(" ")}</span>;
}
