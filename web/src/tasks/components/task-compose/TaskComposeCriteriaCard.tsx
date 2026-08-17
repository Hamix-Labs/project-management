import { useCallback, useRef } from "react";
import type { ChecklistItemDraft, TemplateFunctionInputDef } from "@/types";
import { TaskCreateModalCriteriaFields } from "../task-create-modal/fields/TaskCreateModalCriteriaFields";
import { TaskCreateModalFunctionInputsField } from "../task-create-modal/fields/TaskCreateModalFunctionInputsField";

type Props = {
  checklistItems: ChecklistItemDraft[];
  checklistRequirement: "optional" | "required";
  disabled: boolean;
  checklistDisabled?: boolean;
  idsPrefix: string;
  isTemplateMode: boolean;
  functionInputs: TemplateFunctionInputDef[];
  onAppendChecklistCriterion: (item: ChecklistItemDraft | string) => void;
  onUpdateChecklistRow: (index: number, item: ChecklistItemDraft) => void;
  onRemoveChecklistRow: (index: number) => void;
  onFunctionInputsChange: (next: TemplateFunctionInputDef[]) => void;
};

/**
 * Done-criteria card with Add button.
 * List editing still flows through TaskCreateModalCriteriaFields + ChecklistCriterionModal.
 */
export function TaskComposeCriteriaCard({
  checklistItems,
  checklistRequirement,
  disabled,
  checklistDisabled = false,
  idsPrefix,
  isTemplateMode,
  functionInputs,
  onAppendChecklistCriterion,
  onUpdateChecklistRow,
  onRemoveChecklistRow,
  onFunctionInputsChange,
}: Props) {
  const openNewCriterionRef = useRef<(() => void) | null>(null);
  const registerOpenNew = useCallback((open: (() => void) | null) => {
    openNewCriterionRef.current = open;
  }, []);

  const headingId = "task-compose-criteria-heading";

  return (
    <section
      className="compose-card compose-criteria"
      aria-labelledby={headingId}
    >
      <div className="compose-criteria__head">
        <h2 id={headingId} className="compose-criteria__title">
          Done criteria
        </h2>
        <button>
          type="button"
          className="compose-criteria__add"
          data-testid="compose-criteria-add"
          disabled={disabled || checklistDisabled}
          onClick={() => openNewCriterionRef.current?.()}
        >
          <PlusIcon />
          Add
        </button>
      </div>
      <div className="compose-criteria__body">
        <TaskCreateModalCriteriaFields
          checklistHeadingId={headingId}
          checklistItems={checklistItems}
          checklistRequirement={checklistRequirement}
          checklistDisabled={checklistDisabled}
          disabled={disabled}
          onAppendChecklistCriterion={onAppendChecklistCriterion}
          onUpdateChecklistRow={onUpdateChecklistRow}
          onRemoveChecklistRow={onRemoveChecklistRow}
          registerOpenNew={registerOpenNew}
        />
        {isTemplateMode ? (
          <div className="compose-criteria__function-inputs">
            <TaskCreateModalFunctionInputsField
              idsPrefix={idsPrefix}
              inputs={functionInputs}
              disabled={disabled}
              onChange={onFunctionInputsChange}
            />
          </div>
        ) : null}
      </div>
    </section>
  );
}

function PlusIcon() {
  return (
    <svg
      width="16"
      height="16"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <path d="M12 5v14" />
      <path d="M5 12h14" />
    </svg>
  );
}
