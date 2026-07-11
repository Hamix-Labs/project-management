import { useState } from "react";
import { Modal } from "../../../../shared/Modal";
import { ChecklistCriterionModalForm } from "./ChecklistCriterionModalForm";
import {
  checklistCriterionModalSheetClass,
  createVerifyCommandHandlers,
  resolveChecklistCriterionModalCopy,
  type ChecklistCriterionModalProps,
} from "./checklistCriterionModalCopy";

export function ChecklistCriterionModal({
  mode,
  readOnly = false,
  pending,
  saving,
  onClose,
  text,
  onTextChange,
  verifyCommands,
  onVerifyCommandsChange,
  onSubmit,
  modalStack = "default",
  lockBodyScroll = true,
  dismissibleWhileBusy = false,
  error = null,
  errorFallback,
}: ChecklistCriterionModalProps) {
  const controlsDisabled = pending || saving;
  const copy = resolveChecklistCriterionModalCopy(mode, readOnly);
  const [verifySectionOpen, setVerifySectionOpen] = useState(
    () => readOnly || verifyCommands.length > 0,
  );

  const verifyHandlers = createVerifyCommandHandlers(
    verifyCommands,
    onVerifyCommandsChange,
    setVerifySectionOpen,
  );

  return (
    <Modal
      onClose={onClose}
      labelledBy={copy.titleId}
      describedBy="checklist-criterion-modal-description"
      busy={pending}
      busyLabel={copy.busyLabel}
      stack={modalStack}
      lockBodyScroll={lockBodyScroll}
      dismissibleWhileBusy={dismissibleWhileBusy}
    >
      <section className={checklistCriterionModalSheetClass(readOnly)}>
        <ChecklistCriterionModalForm
          copy={copy}
          readOnly={readOnly}
          mode={mode}
          text={text}
          controlsDisabled={controlsDisabled}
          verifyCommands={verifyCommands}
          verifySectionOpen={verifySectionOpen}
          verifyHandlers={verifyHandlers}
          error={error}
          errorFallback={errorFallback}
          onTextChange={onTextChange}
          onSubmit={onSubmit}
          onClose={onClose}
        />
      </section>
    </Modal>
  );
}
