import type { RefObject } from "react";
import { TaskCreateModalActionFooter } from "../task-create-modal/TaskCreateModalActionFooter";
import { TaskCreateModalMutationErrors } from "../task-create-modal/TaskCreateModalMutationErrors";
import { TestScenariosPopover } from "../task-create-modal/TestScenariosPopover";
import { TestScenariosTrigger } from "../task-create-modal/TestScenariosTrigger";
import type { TaskCreateModalProps } from "../task-create-modal/taskCreateModalProps";
import type { TaskCreateModalPresentation } from "../task-create-modal/taskCreateModalPresentation";
import type { TestScenario } from "@/tasks/test-scenarios";
import {
  DraftAssistThread,
  useDraftAssistContext,
} from "../draft-assist";
import { TaskComposeBriefCard } from "./TaskComposeBriefCard";
import { TaskComposeCriteriaCard } from "./TaskComposeCriteriaCard";
import { TaskComposeHandoffRail } from "./TaskComposeHandoffRail";
import { TaskComposeLayout } from "./TaskComposeLayout";
import { TaskComposeReadinessBar } from "./TaskComposeReadinessBar";

export type TaskComposeFormShellProps = Omit<
  TaskCreateModalProps,
  never
> & {
  presentation: TaskCreateModalPresentation;
  draftSubtitle: string | null;
  backTo: string;
  backLabel?: string;
  scenariosOpen: boolean;
  scenariosTriggerRef: RefObject<HTMLButtonElement>;
  onToggleScenarios: () => void;
  onScenarioPicked: (scenario: TestScenario) => void;
  onCloseScenarios: () => void;
};

/** Inner compose UI; must sit under DraftAssistProvider. */
export function TaskComposeFormShell({
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
}: TaskComposeFormShellProps) {
  const assist = useDraftAssistContext();
  const assistNode = assist.active ? <DraftAssistThread /> : null;

  const editingTaskId = session.editingTaskId ?? null;
  const editingTemplateId = session.editingTemplateId ?? null;
  const editingTaskRunner = session.editingTaskRunner ?? "";
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

  return (
    <>
      <form
        id="task-compose-form"
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
            <TaskComposeHandoffRail
              presentation={presentation}
              session={session}
              essentials={essentials}
              criteria={criteria}
              git={git}
              execution={execution}
              appTimezone={appTimezone}
              editingTaskRunner={editingTaskRunner}
            />
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
                form="task-compose-form"
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
