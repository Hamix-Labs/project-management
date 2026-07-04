import { Link } from "react-router-dom";
import type { GitRepository } from "@/types/git";
import { Button } from "@/components/ui";
import { repositoryDisplayName, repositoryPathsEquivalent } from "../repositoryDisplay";
import { worktreeGitCopy } from "../worktreeGitCopy";
import {
  WorktreesChevronLeftIcon,
  WorktreesFolderIcon,
  WorktreesPlusIcon,
  WorktreesRefreshIcon,
  WorktreesTrashIcon,
} from "./WorktreesIcons";
import { WorktreesMenu, type WorktreesMenuItem } from "./WorktreesMenu";

type Props = {
  repository: GitRepository;
  headingId: string;
  reconcilePending: boolean;
  onReconcile: () => void;
  onDeleteRepository: () => void;
  addWorktreeMenuItems: WorktreesMenuItem[];
  onAddMenuOpenChange: (open: boolean) => void;
};

export function RepositoryDetailHeader({
  repository,
  headingId,
  reconcilePending,
  onReconcile,
  onDeleteRepository,
  addWorktreeMenuItems,
  onAddMenuOpenChange,
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
          <WorktreesMenu
            triggerLabel={worktreeGitCopy.addWorktree}
            className="ui-btn ui-btn--primary repository-detail-card__add-btn"
            menuClassName="repository-detail-card__add-menu"
            icon={<WorktreesPlusIcon className="repository-detail-card__action-icon" aria-hidden />}
            chevron
            split
            onOpenChange={onAddMenuOpenChange}
            items={addWorktreeMenuItems}
          />
        </div>
      </div>

      <div className="repository-detail-card__identity">
        <h1 id={headingId} className="repository-detail-card__title">
          {displayName}
        </h1>
        <p className="repository-detail-card__path" title={repository.path}>
          <WorktreesFolderIcon className="repository-detail-card__path-icon" aria-hidden />
          <span className="repository-detail-card__path-text">{repository.path}</span>
        </p>
        {repository.host_path.trim() !== "" &&
        !repositoryPathsEquivalent(repository.path, repository.host_path) ? (
          <p className="repository-detail-card__host-path">
            <span className="worktrees-repo-row__meta-label">{worktreeGitCopy.hostPathLabel}</span>
            <code>{repository.host_path}</code>
          </p>
        ) : null}
      </div>
    </header>
  );
}
