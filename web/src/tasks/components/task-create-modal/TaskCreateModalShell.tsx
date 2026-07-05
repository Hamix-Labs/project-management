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
  editingTaskId: string | null;
  editingTaskRunner: string;
  autonomyDisabled: boolean;
  createError: Error | null;
  createFormError: string | null;
  patchError: string | null;
  formError: string | null;
  busyLabel: string;
  scenariosOpen: boolean;
  scenariosTriggerRef: RefObject<HTMLButtonElement | null>;
  onToggleScenarios: () => void;
  onScenarioPicked: (scenario: TestScenario) => void;
  onCloseScenarios: () => void;
};

export function TaskCreateModalShell({
  presentation,
  editingTaskId,
  editingTaskRunner,
  draftSaveLabel,
  draftSaveError,
  onClose,
  onSubmit,
  autonomyDisabled,
  createError,
  createFormError,
  patchError,
  formError,
  busyLabel,
  scenariosOpen,
  scenariosTriggerRef,
  onToggleScenarios,
  onScenarioPicked,
  onCloseScenarios,
  ...formProps
}: Props) {
  return (
    <>
      <Modal
        onClose={onClose}
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
            draftSaveLabel={draftSaveLabel}
            draftSaveError={draftSaveError}
            disabled={presentation.disabled}
            scenariosOpen={scenariosOpen}
            scenariosTriggerRef={scenariosTriggerRef}
            onToggleScenarios={onToggleScenarios}
            onClose={onClose}
          />

          <form
            className="task-create-modal-form task-create-form"
            onSubmit={onSubmit}
          >
            <TaskCreateModalFormBody
              presentation={presentation}
              editingTaskId={editingTaskId}
              editingTaskRunner={editingTaskRunner}
              autonomyDisabled={autonomyDisabled}
              {...formProps}
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
                title={formProps.title}
                priority={formProps.priority}
                checklistItems={formProps.checklistItems}
                worktreeId={formProps.worktreeId}
                draftSaving={formProps.draftSaving}
                onClose={onClose}
                onSaveDraft={formProps.onSaveDraft}
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
