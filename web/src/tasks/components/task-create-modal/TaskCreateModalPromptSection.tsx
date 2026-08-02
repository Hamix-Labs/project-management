import { PromptEditorEntry } from "@/components/prompt-editor";
import { TaskCreateModalSection } from "./fields/TaskCreateModalSection";
import type { TaskCreateModalPresentation } from "./taskCreateModalPresentation";

type Props = {
  presentation: TaskCreateModalPresentation;
  prompt: string;
  onOpenPromptEditor: () => void;
};

export function TaskCreateModalPromptSection({
  presentation,
  prompt,
  onOpenPromptEditor,
}: Props) {
  return (
    <TaskCreateModalSection
      variant="prompt"
      title="Initial prompt"
      lede="Write the full brief in the Prompt Editor — headings, lists, and @ file references."
    >
      <PromptEditorEntry
        promptHtml={prompt}
        disabled={presentation.disabled}
        onOpen={onOpenPromptEditor}
        openLabel="Open Prompt Editor"
      />
    </TaskCreateModalSection>
  );
}
