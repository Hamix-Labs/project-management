import type { RichPromptEditorProjectContextProps } from "@/components/rich-prompt";
import { TaskCreateModalPromptFields } from "./fields/TaskCreateModalPromptFields";
import { TaskCreateModalSection } from "./fields/TaskCreateModalSection";
import type { TaskCreateModalPresentation } from "./taskCreateModalPresentation";

type Props = {
  presentation: TaskCreateModalPresentation;
  editorKey: string;
  prompt: string;
  worktreeId: string;
  onPromptChange: (v: string) => void;
  promptProjectContext?: RichPromptEditorProjectContextProps;
};

export function TaskCreateModalPromptSection({
  presentation,
  editorKey,
  prompt,
  worktreeId,
  onPromptChange,
  promptProjectContext,
}: Props) {
  return (
    <TaskCreateModalSection
      variant="prompt"
      title="Initial prompt"
      lede="The full brief the agent starts from. Supports Markdown."
    >
      <TaskCreateModalPromptFields
        idsPrefix={presentation.idsPrefix}
        editorKey={editorKey}
        prompt={prompt}
        disabled={presentation.disabled}
        onPromptChange={onPromptChange}
        projectContext={promptProjectContext}
        worktreeId={worktreeId.trim() || undefined}
      />
    </TaskCreateModalSection>
  );
}
