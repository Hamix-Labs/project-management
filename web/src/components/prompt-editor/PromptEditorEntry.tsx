import { previewTextFromPrompt, promptHasVisibleContent } from "@/lib/promptFormat";

export type PromptEditorEntryProps = {
  promptHtml: string;
  disabled?: boolean;
  onOpen: () => void;
  openLabel?: string;
  emptyHint?: string;
};

/** Summary + CTA that replaces an in-place rich prompt field. */
export function PromptEditorEntry({
  promptHtml,
  disabled = false,
  onOpen,
  openLabel = "Open Prompt Editor",
  emptyHint = "No prompt yet. Open the editor to write a rich implementation brief.",
}: PromptEditorEntryProps) {
  const hasContent = promptHasVisibleContent(promptHtml);
  const preview = hasContent ? previewTextFromPrompt(promptHtml) : "";

  return (
    <div className="prompt-editor-entry">
      <div
        className={
          hasContent
            ? "prompt-editor-entry__preview"
            : "prompt-editor-entry__preview prompt-editor-entry__preview--empty"
        }
        aria-live="polite"
      >
        {hasContent ? (
          <p className="prompt-editor-entry__preview-text">{preview}</p>
        ) : (
          <p className="prompt-editor-entry__empty">{emptyHint}</p>
        )}
      </div>
      <button
        type="button"
        className="primary prompt-editor-entry__open"
        disabled={disabled}
        onClick={onOpen}
      >
        {openLabel}
      </button>
    </div>
  );
}
