import { Link } from "react-router-dom";
import type { GitRepository } from "@/types/git";
import { Button } from "@/components/ui";
import { repositoryDisplayName, repositoryPathsEquivalent } from "../repositoryDisplay";
import { worktreeGitCopy } from "../worktreeGitCopy";
import {
  WorktreesChevronLeftIcon,
  WorktreesFolderIcon,
  WorktreesRefreshIcon,
  WorktreesTrashIcon,
} from "./WorktreesIcons";

type Props = {
  repository: GitRepository;
  headingId: string;
  reconcilePending: boolean;
  onReconcile: () => void;
  onDeleteRepository: () => void;
};

export function RepositoryDetailHeader({
  repository,
  headingId,
  reconcilePending,
  onReconcile,
  onDeleteRepository,
}: Props) {
  const displayName = repositoryDisplayName(repository.path);

  return (
    <header className="repository-detail-card__header">
      <div className="repository-detail-card__header-top">
        <nav aria-label="Repository navigation">
          <Link to="/worktrees" className="repository-detail-card__back">
            <WorktreesChevronLeftIcon className="repository-detail-card__back-icon" aria-hidden />
            All repositories
          </Link>
        </nav>

        <div className="repository-detail-card__actions">
          <Button
            type="button"
            variant="secondary"
            disabled={reconcilePending}
            aria-busy={reconcilePending || undefined}
            className="repository-detail-card__action-btn"
            onClick={onReconcile}
          >
            <WorktreesRefreshIcon className="repository-detail-card__action-icon" aria-hidden />
            {reconcilePending ? worktreeGitCopy.reconciling : worktreeGitCopy.reconcile}
          </Button>
          <Button
            type="button"
            variant="destructive"
            disabled={reconcilePending}
            className="repository-detail-card__action-btn repository-detail-card__action-btn--destructive"
            onClick={onDeleteRepository}
          >
            <WorktreesTrashIcon className="repository-detail-card__action-icon" aria-hidden />
            {worktreeGitCopy.deleteRepository}
          </Button>
        </div>
      </div>

      <div className="repository-detail-card__identity">
        <WorktreesFolderIcon className="repository-detail-card__folder-icon" aria-hidden />
        <div className="repository-detail-card__titles">
          <h1 id={headingId} className="repository-detail-card__title">
            {displayName}
          </h1>
          <p className="repository-detail-card__path" title={repository.path}>
            {repository.path}
          </p>
          {repository.host_path &&
          !repositoryPathsEquivalent(repository.host_path, repository.path) ? (
            <p className="repository-detail-card__host-path" title={repository.host_path}>
              {worktreeGitCopy.hostPathLabel}: {repository.host_path}
            </p>
          ) : null}
        </div>
      </div>
    </header>
  );
}
