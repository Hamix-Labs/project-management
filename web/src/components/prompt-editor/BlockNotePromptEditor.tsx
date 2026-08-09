import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
  SuggestionMenuController,
  useCreateBlockNote,
} from "@blocknote/react";
import { BlockNoteView } from "@blocknote/ariakit";
import "@blocknote/core/style.css";
import "@blocknote/ariakit/style.css";
import { RichPromptFileReferenceModal } from "@/components/rich-prompt/RichPromptFileReferenceModal";
import { promptEditorSchema } from "./blockNoteSchema";
import {
  useEnhanceCodeBlockToolbars,
  type CodeBlockLanguageEditor,
} from "./code/useEnhanceCodeBlockToolbars";
import { PromptEditorRepoContext } from "./context/PromptEditorRepoContext";
import { PromptEditorMentionMenu } from "./mention/PromptEditorMentionMenu";
import { PromptFileMentionHint } from "./mention/PromptFileMentionHint";
import { usePromptFileMentionSearch } from "./mention/usePromptFileMentionSearch";
import { htmlToInitialBlocks } from "./promptEditorHtml";
import { usePromptEditorFileWorktree } from "./usePromptEditorFileWorktree";

export type BlockNotePromptEditorProps = {
  id: string;
  /** Committed snapshot HTML — used once for initialContent (keyed remount). */
  initialHtml: string;
  onChange: (html: string) => void;
  /** Fired once on mount when HTML→blocks used the plain-text fallback. */
  onHydrateFallback?: () => void;
  disabled?: boolean;
  placeholder?: string;
  worktreeId?: string;
  repositoryId?: string;
};

type PendingInsert = { path: string };

export function BlockNotePromptEditor({
  id,
  initialHtml,
  onChange,
  onHydrateFallback,
  disabled = false,
  placeholder = "Write the implementation brief…",
  worktreeId,
  repositoryId,
}: BlockNotePromptEditorProps) {
  const fileWorktree = usePromptEditorFileWorktree({
    worktreeId,
    repositoryId,
  });
  const [pendingInsert, setPendingInsert] = useState<PendingInsert | null>(
    null,
  );
  const [rangeWarning, setRangeWarning] = useState<string | null>(null);
  const hydrateMetaRef = useRef(htmlToInitialBlocks(initialHtml));

  const initialContent = useMemo(
    () => hydrateMetaRef.current.blocks,
    [],
  );

  const editor = useCreateBlockNote({
    schema: promptEditorSchema,
    initialContent,
    placeholders: { default: placeholder },
  });

  useEffect(() => {
    if (hydrateMetaRef.current.usedFallback) {
      onHydrateFallback?.();
    }
  }, [onHydrateFallback]);

  const emitHtml = useCallback(() => {
    const html = editor.blocksToHTMLLossy(editor.document);
    onChange(html);
  }, [editor, onChange]);

  const onSelectPath = useCallback((path: string) => {
    setRangeWarning(null);
    setPendingInsert({ path });
  }, []);

  const mentionSearch = usePromptFileMentionSearch({
    worktree: fileWorktree,
    onSelectPath,
  });

  const insertChip = useCallback(
    (path: string, lineStart?: number, lineEnd?: number) => {
      const props =
        lineStart != null && lineEnd != null
          ? {
              path,
              lineStart: String(lineStart),
              lineEnd: String(lineEnd),
            }
          : { path, lineStart: "", lineEnd: "" };
      editor.insertInlineContent([
        { type: "repoFileMention", props },
        " ",
      ]);
      emitHtml();
    },
    [editor, emitHtml],
  );

  const insertEmbed = useCallback(
    (path: string, lineStart: number, lineEnd: number) => {
      const embed = {
        type: "repoFileEmbed" as const,
        props: {
          path,
          lineStart: String(lineStart),
          lineEnd: String(lineEnd),
        },
      };
      const paragraph = { type: "paragraph" as const };
      // Insert embed at cursor, then a following paragraph for continued typing.
      // eslint-disable-next-line @typescript-eslint/no-explicit-any -- schema-typed insertBlocks
      editor.insertBlocks([embed, paragraph] as any, editor.getTextCursorPosition().block, "after");
      emitHtml();
    },
    [editor, emitHtml],
  );

  const [editorHost, setEditorHost] = useState<HTMLDivElement | null>(null);
  const toolbarEditor = editor as unknown as CodeBlockLanguageEditor;
  useEnhanceCodeBlockToolbars(editorHost, disabled, toolbarEditor);

  return (
    <PromptEditorRepoContext.Provider
      value={{ worktreeId: fileWorktree.worktreeId }}
    >
      <div className="rich-prompt-wrap blocknote-prompt-wrap" id={id}>
        <div
          ref={setEditorHost}
          className={
            disabled
              ? "blocknote-prompt-editor blocknote-prompt-editor--disabled"
              : "blocknote-prompt-editor"
          }
          aria-disabled={disabled || undefined}
        >
          <BlockNoteView
            editor={editor}
            editable={!disabled}
            theme="light"
            onChange={emitHtml}
            slashMenu={true}
          >
            <SuggestionMenuController
              triggerCharacter="@"
              // Must stay referentially stable: BlockNote restarts the load
              // whenever this identity changes.
              getItems={mentionSearch.getItems}
              // Custom menu items include `query` for the search header.
              suggestionMenuComponent={
                PromptEditorMentionMenu as never
              }
            />
          </BlockNoteView>
        </div>
        {pendingInsert ? (
          <RichPromptFileReferenceModal
            id={`${id}-range`}
            pendingInsert={{ insertAt: 0, path: pendingInsert.path }}
            disabled={disabled}
            worktreeId={fileWorktree.worktreeId}
            rangeWarning={rangeWarning}
            onClose={() => {
              setPendingInsert(null);
              setRangeWarning(null);
            }}
            onInsertWithRange={async (start, end) => {
              insertEmbed(pendingInsert.path, start, end);
              setPendingInsert(null);
              setRangeWarning(null);
            }}
            onInsertPathOnly={() => {
              insertChip(pendingInsert.path);
              setPendingInsert(null);
              setRangeWarning(null);
            }}
          />
        ) : null}
        <PromptFileMentionHint status={mentionSearch.status} />
      </div>
    </PromptEditorRepoContext.Provider>
  );
}
