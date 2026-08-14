import "tippy.js/dist/tippy.css";
import { useEffect } from "react";
import { EditorContent } from "@tiptap/react";
import { RichPromptFileReferenceModal } from "./RichPromptFileReferenceModal";
import { RichPromptMenuBar } from "./RichPromptMenuBar";
import { RichPromptRepoHints } from "./RichPromptRepoHints";
import type { RichPromptEditorProps } from "./richPromptEditorTypes";
import { useRichPromptEditorController } from "./useRichPromptEditorController";

/** Rich initial prompt (TipTap) with @ file suggestions scoped to the task worktree. */
export function RichPromptEditor(props: RichPromptEditorProps) {
  const { id, disabled, menuVariant, menuRight, onEditorReady } = props;
  const {
    editor,
    pendingInsert,
    rangeWarning,
    dismissPendingInsert,
    insertPathOnly,
    insertWithRange,
    repoHints,
    mentionWorktreeId,
  } = useRichPromptEditorController(props);

  useEffect(() => {
    onEditorReady?.(editor);
    return () => {
      onEditorReady?.(null);
    };
  }, [editor, onEditorReady]);

  return (
    <div className="rich-prompt-wrap">
      <RichPromptMenuBar
        editor={editor}
        disabled={disabled}
        variant={menuVariant}
        right={menuRight}
      />
      <EditorContent editor={editor} />
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
