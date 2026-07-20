import type { GitRepository } from "@/types/git";
import { worktreeGitCopy } from "../worktreeGitCopy";
import { RepositoryListRow } from "./RepositoryListRow";

type Props = {
  repositories: GitRepository[];
  reconcilingRepositoryId?: string;
  onReconcile: (repository: GitRepository) => void;
  onDelete: (repository: GitRepository) => void;
};

export function RepositoriesListTable({
  repositories,
  reconcilingRepositoryId,
  onReconcile,
  onDelete,
}: Props) {
  return (
    <div className="repositories-list">
      <div className="repositories-list-head" role="row">
        <span className="repositories-list-head__label" role="columnheader">
          {worktreeGitCopy.listColumnName}
        </span>
        <span
          className="repositories-list-head__label repositories-list-head__label--actions"
          role="columnheader"
        >
          {worktreeGitCopy.listColumnActions}
        </span>
      </div>
      <ul className="repositories-list-rows" aria-label="Repositories">
        {repositories.map((repository) => (
          <RepositoryListRow
            key={repository.id}
            repository={repository}
            reconcilePending={reconcilingRepositoryId === repository.id}
            onReconcile={onReconcile}
            onDelete={onDelete}
          />
        ))}
      </ul>
    </div>
  );
}
