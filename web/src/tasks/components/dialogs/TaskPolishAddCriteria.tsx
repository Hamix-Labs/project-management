import type { FormEvent } from "react";
import type { ChecklistItemDraft } from "@/types";
import { ChecklistCriterionModal } from "../task-compose";
import { CloseGlyph } from "./TaskPolishGlyphs";

type Props = {
  headingId: string;
  newCriteria: ChecklistItemDraft[];
  disabled: boolean;
  modalOpen: boolean;
  modalText: string;
  modalCommands: NonNullable<ChecklistItemDraft["verify_commands"]>;
  onOpenModal: () => void;
  onCloseModal: () => void;
  onModalTextChange: (value: string) => void;
  onModalCommandsChange: (
    cmds: NonNullable<ChecklistItemDraft["verify_commands"]>,
  ) => void;
  onSubmitModal: (e: FormEvent) => void;
  onRemove: (index: number) => void;
};

function PlusGlyph() {
  return (
    <svg
      width="16"
      height="16"
      viewBox="0 0 16 16"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.75"
      strokeLinecap="round"
      aria-hidden="true"
    >
      <path d="M8 3.25v9.5M3.25 8h9.5" />
    </svg>
  );
}

function CheckGlyph() {
  return (
    <svg
      width="14"
      height="14"
      viewBox="0 0 16 16"
      fill="none"
      stroke="currentColor"
      strokeWidth="2.5"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <path d="M3.5 8.5L6.5 11.5 12.5 4.5" />
    </svg>
  );
}

export function TaskPolishAddCriteria({
  headingId,
  newCriteria,
  disabled,
  modalOpen,
  modalText,
  modalCommands,
  onOpenModal,
  onCloseModal,
  onModalTextChange,
  onModalCommandsChange,
  onSubmitModal,
  onRemove,
}: Props) {
  return (
    <div className="task-polish-dialog__add-criteria">
      <h3 id={headingId} className="task-polish-dialog__section-title">
        Add criteria
      </h3>
      <p className="task-polish-dialog__section-hint">
        Optional new requirements for this polish attempt.
      </p>
      <button
        type="button"
        className="secondary task-polish-dialog__add-btn"
        disabled={disabled}
        onClick={onOpenModal}
      >
        <PlusGlyph />
        Add criterion
      </button>
      {newCriteria.length > 0 ? (
        <ul
          className="task-polish-dialog__new-list"
          aria-labelledby={headingId}
        >
          {newCriteria.map((item, index) => (
            <li key={`${index}-${item.text}`} className="task-polish-dialog__new-chip">
              <span
                className="task-polish-dialog__check task-polish-dialog__check--on"
                aria-hidden="true"
              >
                <CheckGlyph />
              </span>
              <span className="task-polish-dialog__criteria-text">{item.text}</span>
              <button
                type="button"
                className="task-polish-dialog__new-remove"
                disabled={disabled}
                aria-label={`Remove criterion ${item.text}`}
                onClick={() => onRemove(index)}
              >
                <CloseGlyph />
              </button>
            </li>
          ))}
        </ul>
      ) : null}
      {modalOpen ? (
        <ChecklistCriterionModal
          mode="add"
          pending={false}
          saving={disabled}
          onClose={onCloseModal}
          text={modalText}
          onTextChange={onModalTextChange}
          verifyCommands={modalCommands}
          onVerifyCommandsChange={onModalCommandsChange}
          onSubmit={onSubmitModal}
          modalStack="nested"
          lockBodyScroll={false}
        />
      ) : null}
    </div>
  );
}
