import { useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { useDocumentTitle } from "@/shared/useDocumentTitle";
import { useOptionalToast } from "@/shared/toast";
import { EmptyState } from "@/shared/EmptyState";
import { useDelayedTrue } from "@/lib/useDelayedTrue";
import { TASK_TIMINGS } from "@/constants/tasks";
import { TaskDraftsListSkeleton } from "@/components/skeletons/TaskDraftsListSkeleton";
import {
  repositoryDisplayName,
  repositoryPathsEquivalent,
} from "./repositoryDisplay";
import { useGlobalRepository } from "./hooks/useGlobalRepository";
import { useRepositoryGitActions } from "./hooks/useRepositoryGitActions";
import { RepositoryWorktreesSection } from "./components/RepositoryWorktreesSection";
import { DeleteConfirmDialog } from "./components/DeleteConfirmDialog";
import { RegisterWorktreeModal } from "./modals/RegisterWorktreeModal";
import { CreateWorktreeModal } from "./modals/CreateWorktreeModal";
import { RelocateRepositoryModal } from "./modals/RelocateRepositoryModal";
import { formatReconcileSuccess } from "./gitReconcileErrors";
import { worktreeGitCopy } from "./worktreeGitCopy";
import { WorktreesFolderIcon, WorktreesPlusIcon } from "./components/WorktreesIcons";
import { WorktreesMenu } from "./components/WorktreesMenu";

function useDebouncedTrimmedValue(value: string, delayMs: number): string {
  const [debounced, setDebounced] = useState(value.trim());

  useEffect(() => {
    const timer = window.setTimeout(() => setDebounced(value.trim()), delayMs);
    return () => window.clearTimeout(timer);
  }, [value, delayMs]);

  return debounced;
}

export function RepositoryWorktreesPage() {
  const { repositoryId = "" } = useParams();
  const navigate = useNavigate();
  const toast = useOptionalToast();
  const repositoryQuery = useGlobalRepository(repositoryId);
  const repository = repositoryQuery.data;

  const actions = useRepositoryGitActions({
    repository,
    onRepositoryDeleted: () => navigate("/worktrees"),
  });

  const displayName = repository
    ? repositoryDisplayName(repository.path)
    : "Repository";
  useDocumentTitle(repository ? `${displayName} worktrees` : "Repository worktrees");

  const showSkeleton = useDelayedTrue(
    repositoryQuery.isLoading && !repositoryQuery.data,
    TASK_TIMINGS.draftResumeMinLoadingMs,
  );
  const [searchInput, setSearchInput] = useState("");
  const debouncedQ = useDebouncedTrimmedValue(searchInput, 300);

  if (!repositoryId) {
    return (
      <section className="panel task-list-section-panel task-detail-content--enter repository-detail">
        <EmptyState
          title="Missing repository id"
          description="Choose a repository from the list."
          hideIcon
          className="empty-state--task-list-fresh"
        />
      </section>
    );
  }

  return (
    <section
      className="panel task-list-section-panel task-detail-content--enter worktrees-page repository-detail"
      aria-labelledby="repository-detail-heading"
    >
      <div className="task-list-toolbar repository-detail__toolbar">
        <header className="repository-detail__header pd__header">
          <nav aria-label="Repository navigation">
            <Link to="/worktrees" className="pd__back project-context-back-link">
              <span aria-hidden="true">&#8249;</span>
              All repositories
            </Link>
          </nav>
          {repository ? (
            <div className="repository-detail__header-actions task-list-section-actions">
              <button
                type="button"
                className="secondary repository-detail__reconcile-btn"
                disabled={actions.manualReconcilePending}
                aria-busy={actions.manualReconcilePending || undefined}
                onClick={() => void actions.handleReconcile(repository)}
              >
                {actions.manualReconcilePending
                  ? worktreeGitCopy.reconciling
                  : worktreeGitCopy.reconcile}
              </button>
              <button
                type="button"
                className="danger repository-detail__delete-btn"
                disabled={actions.manualReconcilePending}
                onClick={actions.openDeleteRepository}
              >
                {worktreeGitCopy.deleteRepository}
              </button>
              <WorktreesMenu
                triggerLabel={worktreeGitCopy.addWorktree}
                className="task-home-new-task-btn worktrees-register-btn worktrees-menu-trigger"
                icon={<WorktreesPlusIcon className="worktrees-register-btn__icon" aria-hidden />}
                chevron
                onOpenChange={(open) => {
                  if (open && repository) void actions.ensureInventoryFresh(repository);
                }}
                items={[
                  {
                    id: "register-worktree",
                    label: worktreeGitCopy.registerWorktree,
                    onSelect: () => void actions.openWorktreeModal("register-worktree"),
                  },
                  {
                    id: "create-worktree",
                    label: worktreeGitCopy.createWorktree,
                    onSelect: () => void actions.openWorktreeModal("create-worktree"),
                  },
                ]}
              />
            </div>
          ) : null}
        </header>

        {repository ? (
          <>
            <div className="repository-detail__identity">
              <h1 id="repository-detail-heading" className="repository-detail__title">
                {displayName}
              </h1>
              <p className="repository-detail__path" title={repository.path}>
                <WorktreesFolderIcon className="repository-detail__path-icon" aria-hidden />
                <span className="repository-detail__path-text">{repository.path}</span>
              </p>
              {repository.host_path.trim() !== "" &&
              !repositoryPathsEquivalent(repository.path, repository.host_path) ? (
                <p className="repository-detail__host-path">
                  <span className="worktrees-repo-row__meta-label">
                    {worktreeGitCopy.hostPathLabel}
                  </span>
                  <code>{repository.host_path}</code>
                </p>
              ) : null}
            </div>

            <div
              className="task-templates-search field grow task-list-search-field repository-detail__search"
              role="search"
              aria-label="Search worktrees"
            >
              <label htmlFor="repository-worktrees-search" className="visually-hidden">
                Search worktrees
              </label>
              <input
                id="repository-worktrees-search"
                type="search"
                placeholder={worktreeGitCopy.searchWorktreesPlaceholder}
                autoComplete="off"
                value={searchInput}
                onChange={(e) => setSearchInput(e.target.value)}
              />
            </div>
          </>
        ) : null}
      </div>

      {repositoryQuery.isError && !repositoryQuery.isLoading ? (
        <div className="err" role="alert">
          <p>
            {repositoryQuery.error instanceof Error
              ? repositoryQuery.error.message
              : "Could not load repository."}
          </p>
          <div className="task-detail-error-actions">
            <button
              type="button"
              className="secondary"
              onClick={() => {
                void repositoryQuery.refetch();
              }}
            >
              Try again
            </button>
          </div>
        </div>
      ) : null}

      <div className="stack">
        {showSkeleton ? <TaskDraftsListSkeleton /> : null}
        {!showSkeleton && repository ? (
          <div className="stack task-list-content task-list-content--enter">
            <RepositoryWorktreesSection
              repository={repository}
              searchQuery={debouncedQ}
              reconcilePending={actions.manualReconcilePending}
              reconcileError={actions.reconcileError}
              onUnregisterWorktree={actions.openDeleteWorktree}
              onDeleteWorktreeFromDisk={actions.openRemoveWorktreeFromDisk}
            />
          </div>
        ) : null}
        {!repositoryQuery.isLoading && !repositoryQuery.isError && !repository ? (
          <EmptyState
            title="Repository not found"
            description="It may have been removed. Return to the repository list."
            hideIcon
            className="empty-state--task-list-fresh"
          />
        ) : null}
      </div>

      <RegisterWorktreeModal
        open={actions.activeWorktreeModal === "register-worktree"}
        pending={actions.mutations.registerWorktree.isPending}
        error={actions.mutations.registerWorktree.error}
        repositoryId={repository?.id ?? ""}
        storedPath={repository?.path ?? ""}
        reconcilePending={actions.reconcilePending}
        inventoryRefreshPending={actions.inventoryRefreshPending}
        reconcileError={actions.reconcileError}
        reconcileBlocked={actions.reconcileBlocked}
        onReconcile={() => {
          if (repository != null) void actions.handleReconcile(repository);
        }}
        onClose={() => {
          actions.setActiveWorktreeModal(null);
          actions.mutations.registerWorktree.reset();
        }}
        onSubmit={(input) => {
          if (!repository || actions.activeWorktreeModal !== "register-worktree") return;
          void actions.mutations.registerWorktree
            .mutateAsync({ repositoryId: repository.id, input })
            .then(() => actions.setActiveWorktreeModal(null));
        }}
      />

      <CreateWorktreeModal
        open={actions.activeWorktreeModal === "create-worktree"}
        pending={actions.mutations.createWorktree.isPending}
        error={actions.mutations.createWorktree.error}
        repositoryId={repository?.id ?? ""}
        storedPath={repository?.path ?? ""}
        reconcilePending={actions.reconcilePending}
        inventoryRefreshPending={actions.inventoryRefreshPending}
        reconcileError={actions.reconcileError}
        reconcileBlocked={actions.reconcileBlocked}
        onReconcile={() => {
          if (repository != null) void actions.handleReconcile(repository);
        }}
        onClose={() => {
          actions.setActiveWorktreeModal(null);
          actions.mutations.createWorktree.reset();
        }}
        onSubmit={(input) => {
          if (!repository || actions.activeWorktreeModal !== "create-worktree") return;
          void actions.mutations.createWorktree
            .mutateAsync({ repositoryId: repository.id, input })
            .then(() => actions.setActiveWorktreeModal(null));
        }}
      />

      <DeleteConfirmDialog
        target={actions.deleteTarget}
        pending={actions.deletePending}
        error={actions.deleteError}
        onClose={actions.closeDelete}
        onConfirm={(options) => void actions.runDelete(options)}
      />

      <RelocateRepositoryModal
        open={actions.relocateRepository != null}
        pending={actions.mutations.relocateRepository.isPending}
        error={actions.mutations.relocateRepository.error}
        storedPath={actions.relocateRepository?.path ?? ""}
        onClose={actions.closeRelocateModal}
        onSubmit={(input) => {
          const repo = actions.relocateRepository;
          if (!repo) return;
          void actions.mutations.relocateRepository
            .mutateAsync({ repositoryId: repo.id, input })
            .then((result) => {
              actions.setAutoReconcileBlocked((prev) => {
                const next = { ...prev };
                delete next[repo.id];
                return next;
              });
              actions.closeRelocateModal();
              toast?.success(formatReconcileSuccess(result));
            });
        }}
      />
    </section>
  );
}
