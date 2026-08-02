import type { RefObject } from "react";
import { Modal } from "../../../shared/Modal";
import { TaskCreateModalActionFooter } from "./TaskCreateModalActionFooter";
import { TaskCreateModalFormBody } from "./TaskCreateModalFormBody";
import { TaskCreateModalHeader } from "./TaskCreateModalHeader";
import { TaskCreateModalMutationErrors } from "./TaskCreateModalMutationErrors";
import { TestScenariosPopover } from "./TestScenariosPopover";
import type { TaskCreateModalProps } from "./taskCreateModalProps";
import type { TaskCreateModalPresentation } from "./taskCreateModalPresentation";
import type { TestScenario } from "@/tasks/test-scenarios";

type Props = TaskCreateModalProps & {
  presentation: TaskCreateModalPresentation;
  busyLabel: string;
  scenariosOpen: boolean;
  scenariosTriggerRef: RefObject<HTMLButtonElement | null>;
  onToggleScenarios: () => void;
  onScenarioPicked: (scenario: TestScenario) => void;
  onCloseScenarios: () => void;
};

export function TaskCreateModalShell({
  presentation,
  session,
  essentials,
  prompt,
  criteria,
  git,
  execution,
  actions,
  appTimezone,
  busyLabel,
  scenariosOpen,
  scenariosTriggerRef,
  onToggleScenarios,
  onScenarioPicked,
  onCloseScenarios,
}: Props) {
  const editingTaskId = session.editingTaskId ?? null;
  const editingTaskRunner = session.editingTaskRunner ?? "";
  const autonomyDisabled = execution.autonomyDisabled ?? false;
  const createError = session.createError ?? null;
  const createFormError = session.createFormError ?? null;
  const patchError = session.patchError ?? null;
  const formError = session.formError ?? null;

  return (
    <>
      <Modal
        onClose={actions.onClose}
        labelledBy={presentation.modalTitleId}
        describedBy={presentation.modalDescribedBy}
        size="wide"
        busy={presentation.modalBusy}
        busyLabel={presentation.isEdit ? undefined : busyLabel}
        dismissibleWhileBusy
      >
        <section className="panel modal-sheet modal-sheet--edit task-create-modal-sheet task-create">
          <TaskCreateModalHeader
            presentation={presentation}
            editingTaskId={editingTaskId}
            draftSaveLabel={session.draftSaveLabel}
            draftSaveError={session.draftSaveError}
            disabled={presentation.disabled}
            scenariosOpen={scenariosOpen}
            scenariosTriggerRef={scenariosTriggerRef}
            onToggleScenarios={onToggleScenarios}
            onClose={actions.onClose}
          />

          <form
            className="task-create-modal-form task-create-form"
            onSubmit={actions.onSubmit}
          >
            <TaskCreateModalFormBody
              presentation={presentation}
              editingTaskRunner={editingTaskRunner}
              onComposeStatusChange={session.onComposeStatusChange}
              essentials={essentials}
              prompt={prompt}
              criteria={criteria}
              git={git}
              execution={{ ...execution, autonomyDisabled }}
              appTimezone={appTimezone}
            />

            <TaskCreateModalMutationErrors
              isTaskEdit={presentation.isTaskEdit}
              createFormError={createFormError}
              createError={createError}
              formError={formError}
              patchError={patchError}
            />

            <footer className="task-create-modal-footer">
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
      </Modal>

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
