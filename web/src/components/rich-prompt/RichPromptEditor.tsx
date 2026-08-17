import "tippy.js/dist/tippy.css";
import { useCallback, useEffect } from "react";
import { InlineAiComposer } from "./InlineAiComposer";
import { RichPromptFileReferenceModal } from "./RichPromptFileReferenceModal";
import { HamixNotionEditorSurface } from "./HamixNotionEditorSurface";
import { RichPromptRepoHints } from "./RichPromptRepoHints";
import type { RichPromptEditorProps } from "./richPromptEditorTypes";
import { useRichPromptEditorController } from "./useRichPromptEditorController";

/** Rich initial prompt (TipTap Notion-like) with @ file suggestions scoped to the task worktree. */
export function RichPromptEditor(props: RichPromptEditorProps) {
  const { id, disabled, toolbar, menuRight, onAiTrigger, onEditorReady } =
    props;
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

  useEffect(() => {
    onEditorReady?.(editor);
    return () => {
      onEditorReady?.(null);
    };
  }, [editor, onEditorReady]);

  return (
    <div className="rich-prompt-wrap">
      <HamixNotionEditorSurface
        editor={editor}
        toolbar={toolbar}
        menuRight={menuRight}
        onAiTrigger={onAiTrigger}
      />
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
