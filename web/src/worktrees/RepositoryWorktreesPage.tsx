import { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { useDocumentTitle } from "@/shared/useDocumentTitle";
import { useOptionalToast } from "@/shared/toast";
import { EmptyState } from "@/shared/EmptyState";
import { useDelayedTrue } from "@/lib/useDelayedTrue";
import { TASK_TIMINGS } from "@/constants/tasks";
import { TaskDraftsListSkeleton } from "@/components/skeletons/TaskDraftsListSkeleton";
import { repositoryDisplayName } from "./repositoryDisplay";
import { useGlobalRepository } from "./hooks/useGlobalRepository";
import { useRepositoryGitActions } from "./hooks/useRepositoryGitActions";
import { RepositoryDetailCard } from "./components/RepositoryDetailCard";
import { RepositoryDetailHeader } from "./components/RepositoryDetailHeader";
import { RepositoryWorktreesSearch } from "./components/RepositoryWorktreesSearch";
import { RepositoryWorktreesSection } from "./components/RepositoryWorktreesSection";
import { DeleteConfirmDialog } from "./components/DeleteConfirmDialog";
import { RegisterWorktreeModal } from "./modals/RegisterWorktreeModal";
import { CreateWorktreeModal } from "./modals/CreateWorktreeModal";
import { RelocateRepositoryModal } from "./modals/RelocateRepositoryModal";
import { formatReconcileSuccess } from "./gitReconcileErrors";
import { worktreeGitCopy } from "./worktreeGitCopy";

const HEADING_ID = "repository-detail-heading";

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
      <div className="repository-detail-page">
        <RepositoryDetailCard>
          <EmptyState
            title="Missing repository id"
            description="Choose a repository from the list."
            hideIcon
            className="empty-state--task-list-fresh"
          />
        </RepositoryDetailCard>
      </div>
    );
  }

  return (
    <div className="repository-detail-page">
      <RepositoryDetailCard headingId={repository ? HEADING_ID : undefined}>
        {repository ? (
          <>
            <RepositoryDetailHeader
              repository={repository}
              headingId={HEADING_ID}
              reconcilePending={actions.manualReconcilePending}
              onReconcile={() => void actions.handleReconcile(repository)}
              onDeleteRepository={actions.openDeleteRepository}
              addWorktreeMenuItems={[
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
              onAddMenuOpenChange={(open) => {
                if (open) void actions.ensureInventoryFresh(repository);
              }}
            />
            <RepositoryWorktreesSearch value={searchInput} onChange={setSearchInput} />
          </>
        ) : null}

        {repositoryQuery.isError && !repositoryQuery.isLoading ? (
          <div className="repository-detail-card__body err" role="alert">
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

        <div className="repository-detail-card__body">
          {showSkeleton ? <TaskDraftsListSkeleton /> : null}
          {!showSkeleton && repository ? (
            <RepositoryWorktreesSection
              repository={repository}
              searchQuery={debouncedQ}
              onClearSearch={() => setSearchInput("")}
              reconcilePending={actions.manualReconcilePending}
              reconcileError={actions.reconcileError}
              onUnregisterWorktree={actions.openDeleteWorktree}
              onDeleteWorktreeFromDisk={actions.openRemoveWorktreeFromDisk}
            />
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

      </RepositoryDetailCard>

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
    </div>
  );
}
