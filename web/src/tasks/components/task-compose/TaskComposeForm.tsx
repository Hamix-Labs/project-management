import { useCallback, useRef, type RefObject } from "react";
import { TaskCreateModalActionFooter } from "../task-create-modal/TaskCreateModalActionFooter";
import { TaskCreateModalAdvancedOptions } from "../task-create-modal/TaskCreateModalAdvancedOptions";
import { TaskCreateModalMutationErrors } from "../task-create-modal/TaskCreateModalMutationErrors";
import { TestScenariosPopover } from "../task-create-modal/TestScenariosPopover";
import { TestScenariosTrigger } from "../task-create-modal/TestScenariosTrigger";
import type { TaskCreateModalProps } from "../task-create-modal/taskCreateModalProps";
import type { TaskCreateModalPresentation } from "../task-create-modal/taskCreateModalPresentation";
import type { TestScenario } from "@/tasks/test-scenarios";
import type { DraftAssistSnapshot } from "@/types/draftAssist";
import {
  DraftAssistProvider,
  DraftAssistThread,
  DraftAssistNotReadyBanner,
  useDraftAssistContext,
} from "../draft-assist";
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

/** Routed compose form (ADR-0100) — redesigned page chrome with draft-assist. */
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

  // Refs let DraftAssistProvider read the latest form state without
  // re-instantiating provider callbacks on every keystroke.
  const promptRef = useRef(prompt.prompt);
  promptRef.current = prompt.prompt;
  const titleRef = useRef(essentials.title);
  titleRef.current = essentials.title;
  const priorityRef = useRef(essentials.priority);
  priorityRef.current = essentials.priority;
  const projectIdRef = useRef(git.projectId);
  projectIdRef.current = git.projectId;
  const tagsCsvRef = useRef(criteria.tagsCsv);
  tagsCsvRef.current = criteria.tagsCsv;
  const cursorModelRef = useRef(execution.taskCursorModel);
  cursorModelRef.current = execution.taskCursorModel;

  const getSnapshot = useCallback((): DraftAssistSnapshot => {
    const snap: DraftAssistSnapshot = {
      title: titleRef.current,
      prompt: promptRef.current,
      priority: String(priorityRef.current ?? ""),
    };
    if (projectIdRef.current) snap.project_id = projectIdRef.current;
    const csv = tagsCsvRef.current.trim();
    if (csv !== "") {
      snap.tags = csv
        .split(",")
        .map((t) => t.trim())
        .filter((t) => t !== "");
    }
    if (cursorModelRef.current) snap.cursor_model = cursorModelRef.current;
    return snap;
  }, []);

  const onPromptChangeRef = useRef(prompt.onPromptChange);
  onPromptChangeRef.current = prompt.onPromptChange;
  const applyPromptPatch = useCallback((next: string) => {
    onPromptChangeRef.current(next);
    promptRef.current = next;
  }, []);
  const getPromptSnapshot = useCallback(() => promptRef.current, []);

  return (
    <DraftAssistProvider
      getSnapshot={getSnapshot}
      worktreeId={git.worktreeId.trim() || undefined}
      onApplyPromptPatch={applyPromptPatch}
      getPromptSnapshot={getPromptSnapshot}
    >
      <TaskComposeFormShell
        presentation={presentation}
        draftSubtitle={draftSubtitle}
        backTo={backTo}
        backLabel={backLabel}
        scenariosOpen={scenariosOpen}
        scenariosTriggerRef={scenariosTriggerRef}
        onToggleScenarios={onToggleScenarios}
        onScenarioPicked={onScenarioPicked}
        onCloseScenarios={onCloseScenarios}
        session={session}
        essentials={essentials}
        prompt={prompt}
        criteria={criteria}
        git={git}
        execution={execution}
        actions={actions}
        appTimezone={appTimezone}
      />
    </DraftAssistProvider>
  );
}

type ShellProps = Omit<Props, "presentation" | "backTo" | "backLabel"> & {
  presentation: TaskCreateModalPresentation;
  draftSubtitle: string | null;
  backTo: string;
  backLabel?: string;
};

function TaskComposeFormShell({
  presentation,
  draftSubtitle,
  backTo,
  backLabel,
  scenariosOpen,
  scenariosTriggerRef,
  onToggleScenarios,
  onScenarioPicked,
  onCloseScenarios,
  session,
  essentials,
  prompt,
  criteria,
  git,
  execution,
  actions,
  appTimezone,
}: ShellProps) {
  const assist = useDraftAssistContext();
  const assistNode = assist.active ? <DraftAssistThread /> : null;

  const editingTaskId = session.editingTaskId ?? null;
  const editingTemplateId = session.editingTemplateId ?? null;
  const editingTaskRunner = session.editingTaskRunner ?? "";
  const autonomyDisabled = execution.autonomyDisabled ?? false;
  const createError = session.createError ?? null;
  const createFormError = session.createFormError ?? null;
  const patchError = session.patchError ?? null;
  const formError = session.formError ?? null;

  const editorKey = presentation.isTaskEdit
    ? (editingTaskId ?? "edit-prompt-modal")
    : presentation.isTemplateMode && presentation.isEdit
      ? (editingTemplateId ?? "template-edit-prompt-modal")
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
          assist={assistNode}
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
            <DraftAssistNotReadyBanner />
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
