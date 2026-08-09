import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
  FormattingToolbarController,
  SuggestionMenuController,
  useCreateBlockNote,
} from "@blocknote/react";
import { BlockNoteView } from "@blocknote/ariakit";
import "@blocknote/core/style.css";
import "@blocknote/ariakit/style.css";
import { searchRepoFiles } from "@/api";
import { RichPromptFileReferenceModal } from "@/components/rich-prompt/RichPromptFileReferenceModal";
import { RichPromptRepoHints } from "@/components/rich-prompt/RichPromptRepoHints";
import { computeRepoHintFlags } from "@/components/rich-prompt/richPromptInsertHelpers";
import { useRepoWorkspaceProbe } from "@/components/rich-prompt/useRepoWorkspaceProbe";
import { promptEditorSchema } from "./blockNoteSchema";
import {
  useEnhanceCodeBlockToolbars,
  type CodeBlockLanguageEditor,
} from "./code/useEnhanceCodeBlockToolbars";
import { PromptEditorRepoContext } from "./context/PromptEditorRepoContext";
import {
  PromptEditorMentionMenu,
  type PromptFileMentionItem,
} from "./mention/PromptEditorMentionMenu";
import { htmlToInitialBlocks } from "./promptEditorHtml";
import { PromptEditorSelectionToolbar } from "./toolbar/PromptEditorSelectionToolbar";

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
}: BlockNotePromptEditorProps) {
  const worktreeRef = useRef(worktreeId);
  worktreeRef.current = worktreeId;
  const [pendingInsert, setPendingInsert] = useState<PendingInsert | null>(
    null,
  );
  const [rangeWarning, setRangeWarning] = useState<string | null>(null);
  const [fileSearchUnavailable, setFileSearchUnavailable] = useState(false);
  const [fileSearchBusy, setFileSearchBusy] = useState(false);
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

  const workspaceProbe = useRepoWorkspaceProbe(worktreeId);
  const repoHints = computeRepoHintFlags(
    workspaceProbe,
    fileSearchUnavailable,
    fileSearchBusy,
    worktreeId,
  );

  const getMentionItems = useCallback(
    async (query: string): Promise<PromptFileMentionItem[]> => {
      const wt = worktreeRef.current?.trim();
      if (!wt) {
        setFileSearchUnavailable(true);
        return [];
      }
      setFileSearchBusy(true);
      try {
        const paths = await searchRepoFiles(query, { worktreeId: wt });
        if (paths == null) {
          setFileSearchUnavailable(true);
          return [];
        }
        setFileSearchUnavailable(false);
        return paths.slice(0, 20).map((pathRaw) => {
          const path = pathRaw.replace(/\\/g, "/");
          return {
            title: path,
            query,
            onItemClick: () => {
              setRangeWarning(null);
              setPendingInsert({ path });
            },
          };
        });
      } catch {
        setFileSearchUnavailable(true);
        return [];
      } finally {
        setFileSearchBusy(false);
      }
    },
    [],
  );

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
    <PromptEditorRepoContext.Provider value={{ worktreeId }}>
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
            formattingToolbar={false}
          >
            <FormattingToolbarController
              formattingToolbar={PromptEditorSelectionToolbar}
            />
            <SuggestionMenuController
              triggerCharacter="@"
              getItems={async (q) => getMentionItems(q)}
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
            worktreeId={worktreeId}
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
    </PromptEditorRepoContext.Provider>
  );
}
