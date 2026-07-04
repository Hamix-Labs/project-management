import type { GitRepository } from "@/types/git";
import { MutationErrorBanner } from "@/shared/MutationErrorBanner";
import { EmptyState } from "@/shared/EmptyState";
import { useGlobalBranches } from "../hooks/useGlobalBranches";
import { useGlobalWorktrees } from "../hooks/useGlobalWorktrees";
import { worktreeMatchesSearchQuery } from "../repositoryDisplay";
import { worktreeGitCopy } from "../worktreeGitCopy";
import { isLinkedWorktreeForDisplay } from "../worktreeRegistration";
import { gitReconcileErrorMessage } from "../gitReconcileErrors";
import { WorktreeRow } from "./WorktreeRow";
import { WorktreeReconcileStatus } from "./WorktreeReconcileStatus";

type Props = {
  repository: GitRepository;
  searchQuery?: string;
  onUnregisterWorktree: (worktreeId: string, label: string) => void;
  onDeleteWorktreeFromDisk: (worktreeId: string, label: string) => void;
  reconcilePending?: boolean;
  reconcileError?: unknown;
};

export function RepositoryWorktreesSection({
  repository,
  searchQuery = "",
  onUnregisterWorktree,
  onDeleteWorktreeFromDisk,
  reconcilePending = false,
  reconcileError,
}: Props) {
  const worktreesQuery = useGlobalWorktrees(repository.id);
  const branchesQuery = useGlobalBranches(repository.id);
  const worktrees = (worktreesQuery.data ?? []).filter(isLinkedWorktreeForDisplay);
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

  return (
    <div className="worktrees-list">
      <div className="worktrees-list-head" role="row">
        <span className="worktrees-list-head__label" role="columnheader">
          {worktreeGitCopy.listColumnName}
        </span>
        <span
          className="worktrees-list-head__label worktrees-list-head__label--branch"
          role="columnheader"
        >
          {worktreeGitCopy.listColumnBranch}
        </span>
        <span className="worktrees-list-head__spacer" aria-hidden />
      </div>
      <ul className="draft-row-list worktrees-list-rows" aria-label="Worktrees">
        {reconcilePending ? (
          <li className="worktrees-list-row worktrees-list-row--status">
            <WorktreeReconcileStatus className="worktrees-inventory-reconcile" />
          </li>
        ) : null}

        {reconcileErrorMessage ? (
          <li className="worktrees-list-row worktrees-list-row--status">
            <MutationErrorBanner
              error={reconcileErrorMessage}
              className="worktrees-inventory-error"
            />
          </li>
        ) : null}

        {worktreesError ? (
          <li className="worktrees-list-row worktrees-list-row--status">
            <MutationErrorBanner error={worktreesError} className="worktrees-inventory-error" />
          </li>
        ) : null}

        {loading ? (
          <li className="worktrees-list-row worktrees-list-row--status">
            <p className="worktrees-inventory-loading" aria-busy="true">
              Loading worktrees…
            </p>
          </li>
        ) : null}

        {!loading && !worktreesError
          ? filteredWorktrees.map((worktree) => (
              <WorktreeRow
                key={worktree.id}
                worktree={worktree}
                branches={branches}
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
          <li className="worktrees-list-row worktrees-list-row--empty">
            <p className="worktrees-inventory-empty">{worktreeGitCopy.emptyWorktreesTitle}</p>
          </li>
        ) : null}

        {!loading && !worktreesError && worktrees.length > 0 && filteredWorktrees.length === 0 ? (
          <li className="worktrees-list-row worktrees-list-row--empty">
            <EmptyState
              title="No matching worktrees"
              description="Try a different search term."
              hideIcon
              className="empty-state--task-list-fresh empty-state--in-table"
            />
          </li>
        ) : null}
      </ul>
    </div>
  );
}
