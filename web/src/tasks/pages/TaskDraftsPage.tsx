import { TASK_TIMINGS } from "@/constants/tasks";
import { useDelayedTrue } from "@/lib/useDelayedTrue";
import { EmptyState } from "@/shared/EmptyState";
import { useDocumentTitle } from "@/shared/useDocumentTitle";
import { useNavigate } from "react-router-dom";
import { TaskDraftsListSkeleton } from "../components/skeletons";
import { SavedEntityRow } from "../components/saved-entities/SavedEntityRow";
import { useDeleteWithExitAnimation } from "../hooks/useDeleteWithExitAnimation";
import { useTasksAppContext } from "../app/TasksAppProvider";
import { tasksNewPath } from "../composeRoutes";

export function TaskDraftsPage() {
  const app = useTasksAppContext();
  useDocumentTitle("Task drafts");
  const navigate = useNavigate();
  const openDraftInCreateForm = async (draftId: string) => {
    navigate(tasksNewPath({ draft: draftId }));
  };
  const loading = app.draftListLoading;
  const error = app.draftListError;
  const drafts = app.taskDrafts;
  const resumePending = app.resumeDraftPending;
  const resumeError = app.resumeDraftError;
  const deletePending = app.deleteDraftPending;
  const deleteError = app.deleteDraftError;
  const { deletingId, exitingIds, deleteWithExit } = useDeleteWithExitAnimation({
    entityIds: drafts.map((d) => d.id),
    onDelete: app.deleteDraftByID,
  });
  const showDraftsSkeleton = useDelayedTrue(
    loading,
    TASK_TIMINGS.draftResumeMinLoadingMs,
  );
  /* Reference "now" for relative-time rendering, captured once per render
     so every row in a paint shows times computed against the same instant.
     React Query refetches the draft list periodically; on each refetch
     the page re-renders and `now` updates naturally — no manual interval
     required. */
  const renderNow = new Date();

  const draftCount = drafts.length;
  const hasDrafts = draftCount > 0;

  return (
    <section className="panel task-list-section-panel task-detail-content--enter">
      <header className="task-list-section-head">
        <div className="task-list-section-head__text">
          <h2 id="task-drafts-heading" className="task-list-section-title">
            Task drafts
          </h2>
        </div>
        {hasDrafts ? (
          <div className="task-list-section-actions">
            <span className="draft-count-pill" aria-live="polite">
              <strong>{draftCount}</strong>{" "}
              {draftCount === 1 ? "saved draft" : "saved drafts"}
            </span>
          </div>
        ) : null}
      </header>

      {resumeError ? (
        <div className="err" role="alert">
          <p>{resumeError}</p>
        </div>
      ) : null}
      {deleteError ? (
        <div className="err" role="alert">
          <p>{deleteError}</p>
        </div>
      ) : null}

      <div className="stack">
        {loading && showDraftsSkeleton ? <TaskDraftsListSkeleton /> : null}
        {!loading ? (
          <div className="stack task-list-content task-list-content--enter">
            {error ? (
              <div className="err" role="alert">
                <p>{error}</p>
                <div className="task-detail-error-actions">
                  <button
                    type="button"
                    className="secondary"
                    onClick={() => {
                      void app.retryDraftList();
                    }}
                  >
                    Try again
                  </button>
                </div>
              </div>
            ) : !hasDrafts ? (
              <EmptyState
                title="No saved drafts"
                description="Autosaved while you work."
                className="empty-state--task-list-fresh"
                action={{
                  label: "Create a task",
                  onClick: () => {
                    navigate(tasksNewPath());
                  },
                }}
              />
            ) : (
              <ul className="draft-row-list" aria-label="Saved drafts">
                {drafts.map((d) => {
                  const lastEdited = d.updated_at || d.created_at;
                  const isDeleting = deletingId === d.id;
                  const rowDisabled =
                    resumePending ||
                    deletePending ||
                    exitingIds.includes(d.id);
                  return (
                    <SavedEntityRow
                      key={d.id}
                      name={d.name}
                      lastEdited={lastEdited}
                      renderNow={renderNow}
                      isDeleting={isDeleting}
                      rowDisabled={rowDisabled}
                      isExiting={exitingIds.includes(d.id)}
                      resumeLabel={`Resume draft: ${d.name}`}
                      deleteLabel={`Delete draft "${d.name}"`}
                      deletingLabel={`Deleting draft "${d.name}"`}
                      onOpen={() => void openDraftInCreateForm(d.id)}
                      onDelete={() => void deleteWithExit(d.id)}
                    />
                  );
                })}
              </ul>
            )}
          </div>
        ) : null}
      </div>
    </section>
  );
}
