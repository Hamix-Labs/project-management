import { EditorContent } from "@tiptap/react";
import { ProjectContextChoiceDialog } from "@/components/project-context";
import { ProjectReferencesBlock } from "./ProjectReferencesBlock";
import { RichPromptFileReferenceModal } from "./RichPromptFileReferenceModal";
import { RichPromptMenuBar } from "./RichPromptMenuBar";
import { RichPromptRepoHints } from "./RichPromptRepoHints";
import type { RichPromptEditorProps } from "./richPromptEditorTypes";
import { useRichPromptEditorController } from "./useRichPromptEditorController";

export type { RichPromptEditorProjectContextProps } from "./richPromptEditorTypes";

/** Rich initial prompt (TipTap) with @ file suggestions scoped to the task worktree, plus optional `#` project-context mentions. */
export function RichPromptEditor(props: RichPromptEditorProps) {
  const { id, disabled } = props;
  const {
    editor,
    projectContextEnabled,
    referencesItems,
    onProjectIdsChange,
    removeSelectedProjectId,
    pendingInsert,
    rangeWarning,
    dismissPendingInsert,
    insertPathOnly,
    insertWithRange,
    pendingProjectChoice,
    projectEdges,
    selectedProjectIds,
    cancelProjectContextChoice,
    confirmProjectContextChoice,
    repoHints,
  } = useRichPromptEditorController(props);

  return (
    <div className="rich-prompt-wrap">
      <RichPromptMenuBar editor={editor} disabled={disabled} />
      {projectContextEnabled ? (
        <ProjectReferencesBlock
          items={referencesItems}
          disabled={disabled}
          onRemove={onProjectIdsChange ? removeSelectedProjectId : undefined}
        />
      ) : null}
      <EditorContent editor={editor} />
      {pendingInsert ? (
        <RichPromptFileReferenceModal
          id={id}
          pendingInsert={pendingInsert}
          disabled={disabled}
          rangeWarning={rangeWarning}
          onClose={dismissPendingInsert}
          onInsertWithRange={insertWithRange}
          onInsertPathOnly={insertPathOnly}
        />
      ) : null}
      {pendingProjectChoice ? (
        <ProjectContextChoiceDialog
          item={pendingProjectChoice.item}
          edges={projectEdges}
          selectedIds={selectedProjectIds}
          onClose={cancelProjectContextChoice}
          onConfirm={confirmProjectContextChoice}
        />
      ) : null}
      <RichPromptRepoHints
        showSelectWorktreeHint={repoHints.showSelectWorktreeHint}
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
