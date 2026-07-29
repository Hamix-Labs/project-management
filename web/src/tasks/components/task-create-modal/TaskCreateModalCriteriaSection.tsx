import { useCallback, useRef } from "react";
import type { ChecklistItemDraft, TemplateFunctionInputDef } from "@/types";
import { TaskCreateModalCriteriaFields } from "./fields/TaskCreateModalCriteriaFields";
import { TaskCreateModalFunctionInputsField } from "./fields/TaskCreateModalFunctionInputsField";
import { TaskCreateModalSection } from "./fields/TaskCreateModalSection";
import type { TaskCreateModalPresentation } from "./taskCreateModalPresentation";

type Props = {
  presentation: TaskCreateModalPresentation;
  checklistItems: ChecklistItemDraft[];
  checklistRequirement: "optional" | "required";
  functionInputs: TemplateFunctionInputDef[];
  onAppendChecklistCriterion: (item: ChecklistItemDraft | string) => void;
  onUpdateChecklistRow: (index: number, item: ChecklistItemDraft) => void;
  onRemoveChecklistRow: (index: number) => void;
  onFunctionInputsChange: (next: TemplateFunctionInputDef[]) => void;
};

export function TaskCreateModalCriteriaSection({
  presentation,
  checklistItems,
  checklistRequirement,
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

  return (
    <TaskCreateModalSection
      variant="criteria"
      title="Done criteria"
      lede="Clear, checkable conditions that define when this task is complete."
      requirement={checklistRequirement}
      action={
        <button
          type="button"
          className="task-detail-add-checklist-btn"
          disabled={presentation.disabled || presentation.isTaskEdit}
          onClick={() => openNewCriterionRef.current?.()}
        >
          New criterion
        </button>
      }
    >
      <TaskCreateModalCriteriaFields
        checklistItems={checklistItems}
        checklistRequirement={checklistRequirement}
        checklistDisabled={presentation.isTaskEdit}
        disabled={presentation.disabled}
        onAppendChecklistCriterion={onAppendChecklistCriterion}
        onUpdateChecklistRow={onUpdateChecklistRow}
        onRemoveChecklistRow={onRemoveChecklistRow}
        registerOpenNew={registerOpenNew}
      />
      {presentation.isTemplateMode ? (
        <TaskCreateModalFunctionInputsField
          idsPrefix={presentation.idsPrefix}
          inputs={functionInputs}
          disabled={presentation.disabled}
          onChange={onFunctionInputsChange}
        />
      ) : null}
    </TaskCreateModalSection>
  );
}
