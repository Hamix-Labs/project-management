import type { RefObject } from "react";
import { TaskCreateModalActionFooter } from "../task-create-modal/TaskCreateModalActionFooter";
import { TaskCreateModalFormBody } from "../task-create-modal/TaskCreateModalFormBody";
import { TaskCreateModalMutationErrors } from "../task-create-modal/TaskCreateModalMutationErrors";
import { TestScenariosPopover } from "../task-create-modal/TestScenariosPopover";
import { TestScenariosTrigger } from "../task-create-modal/TestScenariosTrigger";
import type { TaskCreateModalProps } from "../task-create-modal/taskCreateModalProps";
import type { TaskCreateModalPresentation } from "../task-create-modal/taskCreateModalPresentation";
import type { TestScenario } from "@/tasks/test-scenarios";
import { TaskComposeLayout } from "./TaskComposeLayout";

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

/** Routed compose form (ADR-0100) — same fields as the former modal, page chrome. */
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

  return (
    <>
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
          <form
            className="task-create-modal-form task-create-form"
            onSubmit={actions.onSubmit}
          >
            <TaskCreateModalFormBody
              presentation={presentation}
              editingTaskId={editingTaskId}
              editingTemplateId={editingTemplateId}
              editingTaskRunner={editingTaskRunner}
              onComposeStatusChange={session.onComposeStatusChange}
              essentials={essentials}
              prompt={prompt}
              criteria={criteria}
              git={git}
              execution={{ ...execution, autonomyDisabled }}
              appTimezone={appTimezone}
            />

            <div className="task-compose-page__errors">
              <TaskCreateModalMutationErrors
                isTaskEdit={presentation.isTaskEdit}
                createFormError={createFormError}
                createError={createError}
                formError={formError}
                patchError={patchError}
              />
            </div>

            <footer className="task-compose-page__footer task-create-modal-footer">
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
            </footer>
          </form>
        </section>
      </TaskComposeLayout>

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
