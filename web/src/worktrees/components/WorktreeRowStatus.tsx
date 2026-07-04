import { worktreeGitCopy } from "../worktreeGitCopy";

type Props = {
  hasBranch: boolean;
};

export function WorktreeRowStatus({ hasBranch }: Props) {
  const label = hasBranch ? worktreeGitCopy.statusClean : worktreeGitCopy.detachedHead;
  const tone = hasBranch ? "clean" : "detached";

  return (
    <div className="worktree-row__status">
      <span className={`worktree-row__status-dot worktree-row__status-dot--${tone}`} aria-hidden />
      <span className="worktree-row__status-label">{label}</span>
    </div>
  );
}
