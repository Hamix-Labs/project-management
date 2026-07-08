import type { GitRepository } from "@/types/git";
import { Button } from "@/components/ui";
import { MutationErrorBanner } from "@/shared/MutationErrorBanner";
import { useGlobalBranches } from "../hooks/useGlobalBranches";
import { useGlobalWorktrees } from "../hooks/useGlobalWorktrees";
import { useWorktreeCheckoutStatus } from "../hooks/useWorktreeCheckoutStatus";
import { worktreeMatchesSearchQuery } from "../repositoryDisplay";
import { worktreeGitCopy } from "../worktreeGitCopy";
import { isDetailPageWorktree, sortDetailPageWorktrees } from "../worktreeRegistration";
import { gitReconcileErrorMessage } from "../gitReconcileErrors";
import { WorktreeListFooter } from "./WorktreeListFooter";
import { WorktreeRow } from "./WorktreeRow";
import { WorktreeRowSkeleton } from "./WorktreeRowSkeleton";
import { WorktreeReconcileStatus } from "./WorktreeReconcileStatus";
import { WorktreesSearchIcon } from "./WorktreesIcons";

type Props = {
  repository: GitRepository;
  searchQuery?: string;
  onClearSearch?: () => void;
  onUnregisterWorktree: (worktreeId: string, label: string) => void;
  onDeleteWorktreeFromDisk: (worktreeId: string, label: string) => void;
  reconcilePending?: boolean;
  reconcileError?: unknown;
};

export function RepositoryWorktreesSection({
  repository,
  searchQuery = "",
  onClearSearch,
  onUnregisterWorktree,
  onDeleteWorktreeFromDisk,
  reconcilePending = false,
  reconcileError,
}: Props) {
  const worktreesQuery = useGlobalWorktrees(repository.id);
  const branchesQuery = useGlobalBranches(repository.id);
  const worktrees = sortDetailPageWorktrees(
    (worktreesQuery.data ?? []).filter(isDetailPageWorktree),
  );
  const checkoutStatusQuery = useWorktreeCheckoutStatus(repository.id, {
    enabled: !worktreesQuery.isLoading && worktrees.length > 0,
  });
  const statusByWorktreeId = new Map(
    (checkoutStatusQuery.data ?? []).map((row) => [row.worktree_id, row]),
  );
  const branches = branchesQuery.data ?? [];
  const branchById = new Map(branches.map((branch) => [branch.id, branch]));
  const filteredWorktrees = worktrees.filter((worktree) => {
    const branch = worktree.branch_id ? branchById.get(worktree.branch_id) : undefined;
    return worktreeMatchesSearchQuery(worktree, branch?.name, searchQuery);
  });
  const loading = worktreesQuery.isLoading || branchesQuery.isLoading;
  const worktreesError =
    worktreesQuery.isError && !worktreesQuery.isLoading
      ? worktreesQuery.error instanceof Error
        ? worktreesQuery.error.message
        : "Could not load worktrees."
      : null;
  const reconcileErrorMessage =
    reconcileError != null ? gitReconcileErrorMessage(reconcileError) : null;
  const uniqueBranchIds = new Set(
    worktrees.map((worktree) => worktree.branch_id).filter((id): id is string => Boolean(id)),
  );

  return (
    <div className="worktree-list">
      <div className="worktree-list__head worktree-list__row-layout" role="row">
        <span className="worktree-list__head-spacer" aria-hidden />
        <div className="worktree-row__content">
          <span className="worktree-list__head-label" role="columnheader">
            {worktreeGitCopy.listColumnName}
          </span>
        </div>
        <div className="worktree-row__branch">
          <span className="worktree-list__head-label" role="columnheader">
            {worktreeGitCopy.listColumnBranch}
          </span>
        </div>
        <span className="worktree-list__head-menu-spacer" aria-hidden />
      </div>

      <ul className="worktree-list__rows" aria-label="Worktrees">
        {reconcilePending ? (
          <li className="worktree-list__status">
            <WorktreeReconcileStatus className="worktree-list__banner" />
          </li>
        ) : null}

        {reconcileErrorMessage ? (
          <li className="worktree-list__status">
            <MutationErrorBanner
              error={reconcileErrorMessage}
              className="worktree-list__banner"
            />
          </li>
        ) : null}

        {worktreesError ? (
          <li className="worktree-list__status">
            <MutationErrorBanner error={worktreesError} className="worktree-list__banner" />
          </li>
        ) : null}

        {loading ? (
          <>
            <WorktreeRowSkeleton />
            <WorktreeRowSkeleton />
            <WorktreeRowSkeleton />
          </>
        ) : null}

        {!loading && !worktreesError
          ? filteredWorktrees.map((worktree) => (
              <WorktreeRow
                key={worktree.id}
                worktree={worktree}
                branches={branches}
                checkoutStatus={statusByWorktreeId.get(worktree.id)}
                onUnregister={() =>
                  onUnregisterWorktree(worktree.id, worktree.name.trim() || worktree.path)
                }
                onDeleteFromDisk={() =>
                  onDeleteWorktreeFromDisk(worktree.id, worktree.name.trim() || worktree.path)
                }
              />
            ))
          : null}

        {!loading && !worktreesError && worktrees.length === 0 ? (
          <li className="worktree-list__empty">
            <div className="worktree-list__empty-inner">
              <div className="worktree-list__empty-icon-wrap" aria-hidden>
                <WorktreesSearchIcon />
              </div>
              <p className="worktree-list__empty-title">{worktreeGitCopy.emptyWorktreesTitle}</p>
              <p className="worktree-list__empty-description">
                {worktreeGitCopy.emptyWorktreesDescription}
              </p>
            </div>
          </li>
        ) : null}

        {!loading && !worktreesError && worktrees.length > 0 && filteredWorktrees.length === 0 ? (
          <li className="worktree-list__empty">
            <div className="worktree-list__empty-inner">
              <div className="worktree-list__empty-icon-wrap" aria-hidden>
                <WorktreesSearchIcon />
              </div>
              <p className="worktree-list__empty-title">{worktreeGitCopy.noMatchingWorktreesTitle}</p>
              <p className="worktree-list__empty-description">
                Nothing matches &ldquo;{searchQuery}&rdquo;. Try a different name or branch.
              </p>
              {onClearSearch ? (
                <Button type="button" variant="secondary" onClick={onClearSearch}>
                  {worktreeGitCopy.clearSearch}
                </Button>
              ) : null}
            </div>
          </li>
        ) : null}
      </ul>

      {!loading && !worktreesError && worktrees.length > 0 ? (
        <WorktreeListFooter
          filteredCount={filteredWorktrees.length}
          totalCount={worktrees.length}
          branchCount={uniqueBranchIds.size}
        />
      ) : null}
    </div>
  );
}
