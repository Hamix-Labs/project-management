import { useCallback, useEffect, useRef, useState } from "react";
import {
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
import { looksLikeStoredHtml, plainTextToInitialHtml } from "@/lib/promptFormat";
import { promptEditorSchema } from "./blockNoteSchema";
import { PromptEditorRepoContext } from "./context/PromptEditorRepoContext";
import {
  PromptEditorMentionMenu,
  type PromptFileMentionItem,
} from "./mention/PromptEditorMentionMenu";

export type BlockNotePromptEditorProps = {
  id: string;
  value: string;
  onChange: (html: string) => void;
  disabled?: boolean;
  placeholder?: string;
  worktreeId?: string;
};

type PendingInsert = { path: string };

function htmlForEditor(value: string): string {
  if (!value.trim()) return "<p></p>";
  return looksLikeStoredHtml(value) ? value : plainTextToInitialHtml(value);
}

export function BlockNotePromptEditor({
  id,
  value,
  onChange,
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
  const skipEmitRef = useRef(true);
  const lastEmittedRef = useRef(value);
  const seededRef = useRef(false);

  const editor = useCreateBlockNote({
    schema: promptEditorSchema,
    placeholders: { default: placeholder },
  });

  useEffect(() => {
    const html = htmlForEditor(value);
    if (seededRef.current && html === lastEmittedRef.current) {
      return;
    }
    try {
      const blocks = editor.tryParseHTMLToBlocks(html);
      skipEmitRef.current = true;
      editor.replaceBlocks(editor.document, blocks);
      lastEmittedRef.current = html;
      seededRef.current = true;
    } catch {
      // leave current document
    }
  }, [value, editor]);

  const emitHtml = useCallback(() => {
    if (skipEmitRef.current) {
      skipEmitRef.current = false;
      return;
    }
    const html = editor.blocksToHTMLLossy(editor.document);
    lastEmittedRef.current = html;
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

  return (
    <PromptEditorRepoContext.Provider value={{ worktreeId }}>
      <div className="rich-prompt-wrap blocknote-prompt-wrap" id={id}>
        <div
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
