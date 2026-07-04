import { WorktreesBranchIcon } from "./WorktreesIcons";

type Props = {
  filteredCount: number;
  totalCount: number;
  branchCount: number;
};

export function WorktreeListFooter({ filteredCount, totalCount, branchCount }: Props) {
  const worktreeWord = totalCount === 1 ? "worktree" : "worktrees";
  const branchWord = branchCount === 1 ? "branch" : "branches";

  return (
    <footer className="worktree-list-footer">
      <span className="worktree-list-footer__count">
        {filteredCount} of {totalCount} {worktreeWord}
      </span>
      <span className="worktree-list-footer__branches">
        <WorktreesBranchIcon className="worktree-list-footer__branch-icon" aria-hidden />
        {branchCount} {branchWord} checked out
      </span>
    </footer>
  );
}
