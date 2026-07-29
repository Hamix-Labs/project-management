import { Suspense, lazy, useState } from "react";
import { useNavigate } from "react-router-dom";
import { RegisterRepositoryFirstPrompt } from "@/components/RegisterRepositoryFirstPrompt";
import { AppErrorBoundary } from "@/shared/AppErrorBoundary";
import { useAppTimezone } from "@/shared/time/appTimezone";
import { DraftResumeModal } from "../components/draft-resume";
import { buildTaskCreateModalProps } from "../components/task-create-modal/buildTaskCreateModalProps";
import { useTasksAppContext } from "../app/TasksAppProvider";
import { CreateModalChunkFallback } from "./CreateModalChunkFallback";

// TipTap / tippy stay off the home cold path until create/edit opens.
const TaskCreateModal = lazy(() =>
  import("../components/task-create-modal").then((m) => ({
    default: m.TaskCreateModal,
  })),
);

const CREATE_MODAL_LAYER_FALLBACK =
  "Something went wrong while opening the task form.";

/**
 * Create/edit overlays sit outside the route outlet boundary, so TipTap /
 * lazy-chunk failures must be contained here — otherwise they wipe the whole
 * SPA via the app-root boundary.
 */
export function TaskCreateModalsLayer() {
  const app = useTasksAppContext();
  const [layerKey, setLayerKey] = useState(0);
  return (
    <AppErrorBoundary
      variant="modal-layer"
      fallbackMessage={CREATE_MODAL_LAYER_FALLBACK}
      onRecover={() => {
        app.closeEdit();
        setLayerKey((k) => k + 1);
      }}
    >
      <TaskCreateModalsLayerBody key={layerKey} />
    </AppErrorBoundary>
  );
}

function TaskCreateModalsLayerBody() {
  const app = useTasksAppContext();
  const navigate = useNavigate();
  const appTimezone = useAppTimezone();

  const isEditing = app.editingTaskId != null;
  const isTemplateMode = app.composeTarget === "template";
  const isTemplateEdit = isTemplateMode && app.composeOperation === "edit";

  const handleResumeDraft = (id: string) => {
    void app.resumeDraftByID(id).catch(() => {
      // Error state is exposed by the hook and rendered in the modal.
    });
  };

  return (
    <>
      {app.createEntryDraftErrorHint ? (
        <div className="err error-banner" role="alert">
          <span className="error-banner__text">
            Saved drafts are unavailable right now, so a fresh task form was opened.
          </span>
          <button
            type="button"
            className="secondary"
            onClick={() => {
              void app.retryCreateEntryDraftLoad();
            }}
          >
            Retry loading drafts
          </button>
        </div>
      ) : null}
      {app.createModalOpen ? (
        <Suspense
          fallback={<CreateModalChunkFallback onClose={app.closeEdit} />}
        >
          <TaskCreateModal
            {...buildTaskCreateModalProps({
              editingTaskId: app.editingTaskId,
              editingTemplateId: app.editingTemplateId,
              composeTarget: app.composeTarget,
              composeOperation: app.composeOperation,
              editingTaskRunner: app.editingTaskRunner,
              composeStatus: app.composeStatus,
              onComposeStatusChange: app.setComposeStatus,
              patchPending: app.patchPending,
              patchError: app.patchError,
              formError: app.editFormError,
              pending: isTemplateMode ? app.templateSavePending : app.createPending,
              saving: app.saving,
              draftSaving: isEditing || isTemplateMode ? false : app.draftSavePending,
              draftSaveLabel: isEditing || isTemplateMode ? null : app.draftSaveLabel,
              draftSaveError: isEditing || isTemplateMode ? false : app.draftSaveError,
              onClose: app.closeEdit,
              title: app.newTitle,
              prompt: app.newPrompt,
              priority: app.newPriority,
              checklistItems: app.newChecklistItems,
              tagsCsv: app.newTagsCsv,
              functionInputs: app.newFunctionInputs,
              onTitleChange: app.setNewTitle,
              onPromptChange: app.setNewPrompt,
              onPriorityChange: app.setNewPriority,
              onAppendChecklistCriterion: app.appendNewChecklistCriterion,
              onUpdateChecklistRow: app.updateNewChecklistRow,
              onRemoveChecklistRow: app.removeNewChecklistRow,
              onFunctionInputsChange: app.setNewFunctionInputs,
              taskRunner: isEditing ? app.editingTaskRunner : app.newTaskRunner,
              taskCursorModel: app.newTaskCursorModel,
              onTaskRunnerChange: app.setNewTaskRunner,
              onTaskCursorModelChange: app.setNewTaskCursorModel,
              schedule: app.newSchedule,
              onScheduleChange: app.setNewSchedule,
              autonomyEnabled: isEditing
                ? app.composeStatus === "ready"
                : app.newAutonomyEnabled,
              onAutonomyChange: app.setNewAutonomyEnabled,
              autonomyDisabled: isEditing,
              milestone: app.newMilestone,
              repositoryId: app.newRepositoryID,
              projectId: app.newProjectID,
              worktreeId: app.newWorktreeID,
              onRepositoryChange: (repositoryId) => {
                app.setNewRepositoryID(repositoryId);
                app.setNewProjectID("");
                app.setNewWorktreeID("");
              },
              onProjectChange: (projectId) => {
                app.setNewProjectID(projectId);
              },
              onWorktreeChange: app.setNewWorktreeID,
              dependsOn: app.newDependsOn,
              onTagsCsvChange: app.setNewTagsCsv,
              onMilestoneChange: app.setNewMilestone,
              onDependsOnChange: app.setNewDependsOn,
              appTimezone,
              onSaveDraft: () => {
                if (!isEditing) void app.saveDraftNow();
              },
              onSubmit: (e) => void app.submitComposeModal(e),
              createError: isEditing
                ? null
                : isTemplateMode
                  ? app.templateSaveError
                  : app.createError,
              createFormError: isEditing ? null : app.createFormError,
              onApplyTestScenario:
                isEditing || isTemplateEdit ? undefined : app.applyTestScenario,
            })}
          />
        </Suspense>
      ) : null}
      {app.draftPickerOpen ? (
        <DraftResumeModal
          drafts={app.taskDrafts}
          onClose={() => app.setDraftPickerOpen(false)}
          onStartFresh={() => void app.startFreshDraft()}
          onResume={handleResumeDraft}
          loading={app.draftListLoading}
          loadError={app.draftListError}
          onRetryLoad={() => {
            void app.retryDraftList();
          }}
          resumePending={app.resumeDraftPending}
          resumeError={app.resumeDraftError}
        />
      ) : null}
      {app.repositorySetupPromptOpen ? (
        <RegisterRepositoryFirstPrompt
          open={app.repositorySetupPromptOpen}
          onClose={() => app.setRepositorySetupPromptOpen(false)}
          onRegister={() => {
            app.setRepositorySetupPromptOpen(false);
            navigate("/repositories?register=1");
          }}
        />
      ) : null}
    </>
  );
}
