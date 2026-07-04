export function WorktreeRowSkeleton() {
  return (
    <li className="worktree-row worktree-row--skeleton" aria-hidden>
      <div className="worktree-row__main">
        <span className="worktree-row__skeleton worktree-row__skeleton--chevron" />
        <div className="worktree-row__content">
          <span className="worktree-row__skeleton worktree-row__skeleton--title" />
          <span className="worktree-row__skeleton worktree-row__skeleton--status" />
        </div>
        <span className="worktree-row__skeleton worktree-row__skeleton--branch" />
        <span className="worktree-row__skeleton worktree-row__skeleton--menu" />
      </div>
    </li>
  );
}
