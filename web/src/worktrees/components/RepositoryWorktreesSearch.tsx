import { worktreeGitCopy } from "../worktreeGitCopy";
import { WorktreesCloseIcon, WorktreesSearchIcon } from "./WorktreesIcons";

type Props = {
  value: string;
  onChange: (value: string) => void;
};

export function RepositoryWorktreesSearch({ value, onChange }: Props) {
  const trimmed = value.trim();

  return (
    <div
      className="repository-detail-card__search"
      role="search"
      aria-label="Search worktrees"
    >
      <label htmlFor="repository-worktrees-search" className="visually-hidden">
        Search worktrees
      </label>
      <div className="repository-detail-card__search-field">
        <WorktreesSearchIcon className="repository-detail-card__search-icon" aria-hidden />
        <input
          id="repository-worktrees-search"
          type="search"
          className="repository-detail-card__search-input"
          placeholder={worktreeGitCopy.searchWorktreesPlaceholder}
          autoComplete="off"
          value={value}
          onChange={(e) => onChange(e.target.value)}
        />
        {trimmed ? (
          <button
            type="button"
            className="repository-detail-card__search-clear"
            aria-label="Clear search"
            onClick={() => onChange("")}
          >
            <WorktreesCloseIcon aria-hidden />
          </button>
        ) : null}
      </div>
    </div>
  );
}
