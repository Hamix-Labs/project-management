import type { RefObject } from "react";
import { TaskCreateModalActionFooter } from "../task-create-modal/TaskCreateModalActionFooter";
import { TaskCreateModalAdvancedOptions } from "../task-create-modal/TaskCreateModalAdvancedOptions";
import { TaskCreateModalMutationErrors } from "../task-create-modal/TaskCreateModalMutationErrors";
import { TestScenariosPopover } from "../task-create-modal/TestScenariosPopover";
import { TestScenariosTrigger } from "../task-create-modal/TestScenariosTrigger";
import type { TaskCreateModalProps } from "../task-create-modal/taskCreateModalProps";
import type { TaskCreateModalPresentation } from "../task-create-modal/taskCreateModalPresentation";
import type { TestScenario } from "@/tasks/test-scenarios";
import { TaskComposeBriefCard } from "./TaskComposeBriefCard";
import { TaskComposeCriteriaCard } from "./TaskComposeCriteriaCard";
import { TaskComposeLayout } from "./TaskComposeLayout";
import { TaskComposeReadinessBar } from "./TaskComposeReadinessBar";
import { TaskComposeAgentCard } from "./rail/TaskComposeAgentCard";
import { TaskComposeDestinationCard } from "./rail/TaskComposeDestinationCard";
import { TaskComposePriorityCard } from "./rail/TaskComposePriorityCard";
import { TaskComposeTagsCard } from "./rail/TaskComposeTagsCard";

type Props = TaskCreateModalProps & {
  presentation: TaskCreateModalPresentation;
  backTo: string;
  backLabel?: string;
  scenariosOpen: boolean;
  scenariosTriggerRef: RefObject<HTMLButtonElement>;
  onToggleScenarios: () => void;
  onScenarioPicked: (scenario: TestScenario) => void;
  onCloseScenarios: () => void;
};

/** Routed compose form (ADR-0100) — same fields as the former modal, redesigned page chrome. */
export function TaskComposeForm({
  presentation,
  session,
  essentials,
  prompt,
  criteria,
  git,
  execution,
  actions,
  appTimezone,
  backTo,
  backLabel,
  scenariosOpen,
  scenariosTriggerRef,
  onToggleScenarios,
  onScenarioPicked,
  onCloseScenarios,
}: Props) {
  const editingTaskId = session.editingTaskId ?? null;
  const editingTemplateId = session.editingTemplateId ?? null;
  const editingTaskRunner = session.editingTaskRunner ?? "";
  const autonomyDisabled = execution.autonomyDisabled ?? false;
  const createError = session.createError ?? null;
  const createFormError = session.createFormError ?? null;
  const patchError = session.patchError ?? null;
  const formError = session.formError ?? null;

  const createHint =
    !presentation.isEdit && !presentation.isTemplateMode
      ? "Define the work, then hand it off to your agent."
      : null;
  const draftSubtitle =
    presentation.showDraftStatus && session.draftSaveLabel
      ? session.draftSaveLabel
      : session.draftSaveError
        ? "Draft autosave failed"
        : createHint;

  const editorKey = presentation.isTaskEdit
    ? editingTaskId ?? "edit-prompt-modal"
    : presentation.isTemplateMode && presentation.isEdit
      ? editingTemplateId ?? "template-edit-prompt-modal"
      : presentation.isTemplateMode
        ? "template-prompt-modal"
        : "create-prompt-modal";
  const checklistRequirement = presentation.isTaskEdit
    ? "optional"
    : "required";
  const agentRunner = presentation.isTaskEdit
    ? editingTaskRunner
    : execution.taskRunner;

  const showTagsCard = presentation.tagsUiEnabled;
  const showMoreOptions =
    presentation.scheduleUiEnabled ||
    presentation.dependenciesUiEnabled ||
    Boolean(presentation.isTaskEdit && session.onComposeStatusChange);

  return (
    <>
      <form
        className="task-create-modal-form task-create-form task-compose-page__form-shell"
        onSubmit={actions.onSubmit}
      >
        <TaskComposeLayout
          title={presentation.modalTitle}
          subtitle={draftSubtitle}
          backTo={backTo}
          backLabel={backLabel}
          topActions={
            presentation.showTestScenarios ? (
              <TestScenariosTrigger
                ref={scenariosTriggerRef}
                open={scenariosOpen}
                disabled={presentation.disabled}
                onToggle={onToggleScenarios}
              />
            ) : null
          }
          rightRail={
            <>
              <TaskComposeDestinationCard
                idsPrefix={presentation.idsPrefix}
                repositoryId={git.repositoryId}
                projectId={git.projectId}
                worktreeId={git.worktreeId}
                assignmentLocked={git.assignmentLocked === true}
                disabled={presentation.disabled}
                showWorktree={!presentation.isTaskEdit}
                onRepositoryChange={git.onRepositoryChange}
                onProjectChange={git.onProjectChange}
                onWorktreeChange={git.onWorktreeChange}
              />
              <TaskComposePriorityCard
                value={essentials.priority}
                disabled={presentation.disabled}
                onChange={essentials.onPriorityChange}
              />
              <TaskComposeAgentCard
                disabled={presentation.disabled}
                lockRunner={presentation.isTaskEdit}
                runner={agentRunner}
                cursorModel={execution.taskCursorModel}
                autonomyEnabled={execution.autonomyEnabled}
                autonomyDisabled={autonomyDisabled}
                onRunnerChange={execution.onTaskRunnerChange}
                onCursorModelChange={execution.onTaskCursorModelChange}
                onAutonomyChange={execution.onAutonomyChange}
              />
              {showTagsCard ? (
                <TaskComposeTagsCard
                  tagsCsv={criteria.tagsCsv}
                  disabled={presentation.disabled}
                  onTagsCsvChange={criteria.onTagsCsvChange}
                />
              ) : null}
              {showMoreOptions ? (
                <div className="compose-more-options">
                  <TaskCreateModalAdvancedOptions
                    presentation={presentation}
                    editingTaskRunner={editingTaskRunner}
                    taskRunner={execution.taskRunner}
                    taskCursorModel={execution.taskCursorModel}
                    onTaskRunnerChange={execution.onTaskRunnerChange}
                    onTaskCursorModelChange={execution.onTaskCursorModelChange}
                    onComposeStatusChange={session.onComposeStatusChange}
                    schedule={execution.schedule}
                    onScheduleChange={execution.onScheduleChange}
                    appTimezone={appTimezone}
                    tagsCsv={criteria.tagsCsv}
                    milestone={execution.milestone}
                    projectId={git.projectId}
                    dependsOn={execution.dependsOn}
                    onTagsCsvChange={criteria.onTagsCsvChange}
                    onMilestoneChange={execution.onMilestoneChange}
                    onDependsOnChange={execution.onDependsOnChange}
                    omitAgent
                    omitTags
                  />
                </div>
              ) : null}
            </>
          }
          stickyFooter={
            <>
              <TaskComposeReadinessBar
                title={essentials.title}
                brief={prompt.prompt}
                repositoryId={git.repositoryId}
                checklistItems={criteria.checklistItems}
              />
              <TaskCreateModalActionFooter
                presentation={presentation}
                title={essentials.title}
                priority={essentials.priority}
                checklistItems={criteria.checklistItems}
                repositoryId={git.repositoryId}
                draftSaving={session.draftSaving}
                onClose={actions.onClose}
                onSaveDraft={actions.onSaveDraft}
              />
            </>
          }
          errors={
            <TaskCreateModalMutationErrors
              isTaskEdit={presentation.isTaskEdit}
              createFormError={createFormError}
              createError={createError}
              formError={formError}
              patchError={patchError}
            />
          }
        >
          <section
            className="panel modal-sheet modal-sheet--edit task-create-modal-sheet task-create"
            aria-labelledby="task-compose-page-title"
          >
            {presentation.isTaskEdit && editingTaskId ? (
              <p
                className="muted stack-tight-zero task-create-modal-task-id"
                id="task-edit-modal-description"
              >
                <code>{editingTaskId}</code>
              </p>
            ) : null}

            <TaskComposeBriefCard
              idsPrefix={presentation.idsPrefix}
              editorKey={editorKey}
              title={essentials.title}
              prompt={prompt.prompt}
              disabled={presentation.disabled}
              onTitleChange={essentials.onTitleChange}
              onPromptChange={prompt.onPromptChange}
              worktreeId={git.worktreeId.trim() || undefined}
              repositoryId={
                presentation.isTaskEdit
                  ? git.repositoryId.trim() || undefined
                  : git.repositoryId
              }
              preferRepositoryHint={!presentation.isTaskEdit}
            />

            <TaskComposeCriteriaCard
              checklistItems={criteria.checklistItems}
              checklistRequirement={checklistRequirement}
              disabled={presentation.disabled}
              checklistDisabled={presentation.isTaskEdit}
              idsPrefix={presentation.idsPrefix}
              isTemplateMode={presentation.isTemplateMode}
              functionInputs={criteria.functionInputs}
              onAppendChecklistCriterion={criteria.onAppendChecklistCriterion}
              onUpdateChecklistRow={criteria.onUpdateChecklistRow}
              onRemoveChecklistRow={criteria.onRemoveChecklistRow}
              onFunctionInputsChange={criteria.onFunctionInputsChange}
            />
          </section>
        </TaskComposeLayout>
      </form>

      {scenariosOpen && presentation.showTestScenarios ? (
        <TestScenariosPopover
          anchor={scenariosTriggerRef.current}
          onPick={onScenarioPicked}
          onClose={onCloseScenarios}
        />
      ) : null}
    </>
  );
}
