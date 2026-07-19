import { useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { useDocumentTitle } from "@/shared/useDocumentTitle";
import { useOptionalToast } from "@/shared/toast";
import { EmptyState } from "@/shared/EmptyState";
import { useDelayedTrue } from "@/lib/useDelayedTrue";
import { useDebouncedTrimmedValue } from "@/hooks/useDebouncedTrimmedValue";
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
import { RelocateRepositoryModal } from "./modals/RelocateRepositoryModal";
import { formatReconcileSuccess } from "./gitReconcileErrors";

const HEADING_ID = "repository-detail-heading";

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
