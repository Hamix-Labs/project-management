import { PromptEditorEntry } from "@/components/prompt-editor";

type Props = {
  idsPrefix: string;
  prompt: string;
  disabled: boolean;
  onOpenPromptEditor: () => void;
};

/** Compose prompt summary + Open Prompt Editor CTA (no in-place editor). */
export function TaskComposePromptField({
  idsPrefix,
  prompt,
  disabled,
  onOpenPromptEditor,
}: Props) {
  return (
    <div className="field grow stack-tight prompt-field-full task-create-prompt">
      <label htmlFor={`${idsPrefix}-prompt-open`}>Initial prompt</label>
      <div className="task-create-editor-shell" id={`${idsPrefix}-prompt-open`}>
        <PromptEditorEntry
          promptHtml={prompt}
          disabled={disabled}
          onOpen={onOpenPromptEditor}
        />
      </div>
    </div>
  );
}
