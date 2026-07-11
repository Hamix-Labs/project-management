import type { FormEvent } from "react";
import type { ChecklistVerifyCommandInput } from "@/types";
import { FieldLabel } from "@/shared/FieldLabel";
import { MutationErrorBanner } from "../../../../shared/MutationErrorBanner";
import { ChecklistVerifyCommandsSection } from "./ChecklistVerifyCommandsSection";
import type {
  ChecklistCriterionModalCopy,
  VerifyCommandHandlers,
} from "./checklistCriterionModalCopy";
import { handleChecklistCriterionFormSubmit } from "./checklistCriterionModalCopy";

type ChecklistCriterionModalHeaderProps = {
  titleId: string;
  title: string;
  lead: string;
};

function ChecklistCriterionModalHeader({
  titleId,
  title,
  lead,
}: ChecklistCriterionModalHeaderProps) {
  return (
    <>
      <h2 id={titleId}>{title}</h2>
      <p
        className="muted task-checklist-criterion-modal-lead"
        id="checklist-criterion-modal-description"
      >
        {lead}
      </p>
    </>
  );
}

type ChecklistCriterionTextFieldProps = {
  text: string;
  readOnly: boolean;
  controlsDisabled: boolean;
  onTextChange: (v: string) => void;
};

function ChecklistCriterionTextField({
  text,
  readOnly,
  controlsDisabled,
  onTextChange,
}: ChecklistCriterionTextFieldProps) {
  return (
    <div className="field">
      <FieldLabel
        htmlFor="checklist-criterion-text"
        requirement={readOnly ? undefined : "required"}
      >
        Criterion
      </FieldLabel>
      <textarea
        id="checklist-criterion-text"
        className="task-checklist-criterion-text-input"
        value={text}
        onChange={(ev) => onTextChange(ev.target.value)}
        placeholder="e.g. All subtasks marked done"
        disabled={controlsDisabled}
        readOnly={readOnly}
        autoFocus={!readOnly}
        required={!readOnly}
        aria-required={readOnly ? undefined : "true"}
        rows={3}
      />
    </div>
  );
}

type ChecklistCriterionModalActionsProps = {
  readOnly: boolean;
  mode: "add" | "edit";
  text: string;
  controlsDisabled: boolean;
  onClose: () => void;
};

function ChecklistCriterionModalActions({
  readOnly,
  mode,
  text,
  controlsDisabled,
  onClose,
}: ChecklistCriterionModalActionsProps) {
  return (
    <div className="row stack-row-actions task-checklist-criterion-modal-actions">
      {readOnly ? (
        <button type="button" className="secondary" onClick={onClose}>
          Close
        </button>
      ) : (
        <>
          <button
            type="button"
            className="secondary"
            disabled={controlsDisabled}
            onClick={onClose}
          >
            Cancel
          </button>
          <button
            type="submit"
            className="task-create-submit"
            disabled={!text.trim() || controlsDisabled}
          >
            {mode === "add" ? "Add criterion" : "Save changes"}
          </button>
        </>
      )}
    </div>
  );
}

type ChecklistCriterionModalFormProps = {
  copy: ChecklistCriterionModalCopy;
  readOnly: boolean;
  mode: "add" | "edit";
  text: string;
  controlsDisabled: boolean;
  verifyCommands: ChecklistVerifyCommandInput[];
  verifySectionOpen: boolean;
  verifyHandlers: VerifyCommandHandlers;
  error: unknown;
  errorFallback?: string;
  onTextChange: (v: string) => void;
  onSubmit: (e: FormEvent) => void;
  onClose: () => void;
};

export function ChecklistCriterionModalForm({
  copy,
  readOnly,
  mode,
  text,
  controlsDisabled,
  verifyCommands,
  verifySectionOpen,
  verifyHandlers,
  error,
  errorFallback,
  onTextChange,
  onSubmit,
  onClose,
}: ChecklistCriterionModalFormProps) {
  return (
    <>
      <ChecklistCriterionModalHeader
        titleId={copy.titleId}
        title={copy.title}
        lead={copy.lead}
      />
      <form
        className="task-checklist-criterion-modal-form task-create-form"
        onSubmit={(e) =>
          handleChecklistCriterionFormSubmit(e, readOnly, onSubmit)
        }
      >
        <ChecklistCriterionTextField
          text={text}
          readOnly={readOnly}
          controlsDisabled={controlsDisabled}
          onTextChange={onTextChange}
        />
        <ChecklistVerifyCommandsSection
          readOnly={readOnly}
          controlsDisabled={controlsDisabled}
          verifyCommands={verifyCommands}
          verifySectionOpen={verifySectionOpen}
          handlers={verifyHandlers}
        />
        <MutationErrorBanner
          error={error}
          fallback={errorFallback ?? copy.defaultErrorFallback}
          className="task-checklist-criterion-modal-err"
        />
        <ChecklistCriterionModalActions
          readOnly={readOnly}
          mode={mode}
          text={text}
          controlsDisabled={controlsDisabled}
          onClose={onClose}
        />
      </form>
    </>
  );
}
