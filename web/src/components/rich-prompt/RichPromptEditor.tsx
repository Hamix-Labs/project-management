import "tippy.js/dist/tippy.css";
import { EditorContent } from "@tiptap/react";
import { useCallback } from "react";
import { InlineAiComposer } from "./InlineAiComposer";
import { RichPromptFileReferenceModal } from "./RichPromptFileReferenceModal";
import { RichPromptMenuBar } from "./RichPromptMenuBar";
import { RichPromptRepoHints } from "./RichPromptRepoHints";
import type { RichPromptEditorProps } from "./richPromptEditorTypes";
import { useRichPromptEditorController } from "./useRichPromptEditorController";

/** Rich initial prompt (TipTap) with @ file suggestions scoped to the task worktree. */
export function RichPromptEditor(props: RichPromptEditorProps) {
  const { id, disabled, onAiTrigger } = props;
  const {
    editor,
    pendingInsert,
    rangeWarning,
    dismissPendingInsert,
    insertPathOnly,
    insertWithRange,
    repoHints,
    mentionWorktreeId,
    aiComposer,
    closeAiComposer,
    getAiAnchorRect,
  } = useRichPromptEditorController(props);

  const handleSubmit = useCallback(
    (msg: string) => {
      onAiTrigger?.(msg);
      closeAiComposer();
    },
    [onAiTrigger, closeAiComposer],
  );

  return (
    <div className="rich-prompt-wrap">
      <RichPromptMenuBar editor={editor} disabled={disabled} />
      <EditorContent editor={editor} />
      <InlineAiComposer
        open={aiComposer.open}
        initialValue={aiComposer.initialValue}
        getAnchorRect={getAiAnchorRect}
        onClose={closeAiComposer}
        onSubmit={handleSubmit}
      />
      {pendingInsert ? (
        <RichPromptFileReferenceModal
          id={id}
          pendingInsert={pendingInsert}
          disabled={disabled}
          worktreeId={mentionWorktreeId || undefined}
          rangeWarning={rangeWarning}
          onClose={dismissPendingInsert}
          onInsertWithRange={insertWithRange}
          onInsertPathOnly={insertPathOnly}
        />
      ) : null}
      <RichPromptRepoHints
        showSelectWorktreeHint={repoHints.showSelectWorktreeHint}
        showSelectRepositoryHint={repoHints.showSelectRepositoryHint}
        showRepoMisconfigHint={repoHints.showRepoMisconfigHint}
        workspaceBroken={repoHints.workspaceBroken}
        fileSearchFailedWhileAvailable={
          repoHints.fileSearchFailedWhileAvailable
        }
        showRepoUnknownHint={repoHints.showRepoUnknownHint}
        showFileSearchSpinner={repoHints.showFileSearchSpinner}
      />
    </div>
  );
}
