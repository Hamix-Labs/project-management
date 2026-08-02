import type { ReactNode } from "react";
import type { ChecklistItemDraft } from "@/types";
import { TaskComposeFields } from "../../task-compose";

type Props = {
  idsPrefix?: string;
  disabled: boolean;
  title: string;
  onTitleChange: (value: string) => void;
  priority: import("@/types").PriorityChoice;
  onPriorityChange: (value: import("@/types").PriorityChoice) => void;
  prompt: string;
  checklistItems: ChecklistItemDraft[];
  hideComposeChecklist: boolean;
  checklistRequirement?: "optional" | "required";
  checklistDisabled?: boolean;
  onOpenPromptEditor: () => void;
  onAppendChecklistCriterion: (item: ChecklistItemDraft | string) => void;
  onUpdateChecklistRow: (index: number, item: ChecklistItemDraft) => void;
  onRemoveChecklistRow: (index: number) => void;
  betweenTitleAndPrompt?: ReactNode;
};

export function TaskCreateModalPrimaryFields({
  idsPrefix = "task-new",
  disabled,
  title,
  onTitleChange,
  priority,
  onPriorityChange,
  prompt,
  checklistItems,
  hideComposeChecklist,
  checklistRequirement = "optional",
  checklistDisabled = false,
  onOpenPromptEditor,
  onAppendChecklistCriterion,
  onUpdateChecklistRow,
  onRemoveChecklistRow,
  betweenTitleAndPrompt,
}: Props) {
  return (
    <TaskComposeFields
      idsPrefix={idsPrefix}
      title={title}
      prompt={prompt}
      priority={priority}
      checklistItems={checklistItems}
      hideChecklist={hideComposeChecklist}
      checklistRequirement={checklistRequirement}
      checklistDisabled={checklistDisabled}
      disabled={disabled}
      onTitleChange={onTitleChange}
      onOpenPromptEditor={onOpenPromptEditor}
      onPriorityChange={onPriorityChange}
      onAppendChecklistCriterion={onAppendChecklistCriterion}
      onUpdateChecklistRow={onUpdateChecklistRow}
      onRemoveChecklistRow={onRemoveChecklistRow}
      betweenTitleAndPrompt={betweenTitleAndPrompt}
    />
  );
}
