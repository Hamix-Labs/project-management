import { useEditor } from "@tiptap/react";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { validateRepoRange } from "@/api";
import { useDelayedTrue } from "@/lib/useDelayedTrue";
import {
  looksLikeStoredHtml,
  plainTextToInitialHtml,
} from "@/lib/promptFormat";
import type { RepoFileSuggestionOptions } from "./extensions/repoFileSuggestion";
import { buildRichPromptExtensions } from "./richPromptExtensions";
import {
  computeRepoHintFlags,
  insertRepoFileMentionAt,
  type PendingFileInsert,
} from "./richPromptInsertHelpers";
import { useRepoWorkspaceProbe } from "./useRepoWorkspaceProbe";
import type { RichPromptEditorProps } from "./richPromptEditorTypes";

export function useRichPromptEditorController({
  id,
  value,
  onChange,
  disabled,
  placeholder,
  worktreeId,
}: RichPromptEditorProps) {
  const workspaceProbe = useRepoWorkspaceProbe(worktreeId);
  const [fileSearchUnavailable, setFileSearchUnavailable] = useState(false);
  const [fileSuggestBusy, setFileSuggestBusy] = useState(false);
  const [pendingInsert, setPendingInsert] = useState<PendingFileInsert | null>(
    null,
  );
  const [rangeWarning, setRangeWarning] = useState<string | null>(null);
  const lastEmittedHtml = useRef<string | null>(null);

  const onFilePicked = useCallback(
    (payload: { insertAt: number; path: string }) => {
      setPendingInsert({ insertAt: payload.insertAt, path: payload.path });
      setRangeWarning(null);
    },
    [],
  );

  const worktreeIdRef = useRef(worktreeId);
  useEffect(() => {
    worktreeIdRef.current = worktreeId;
  }, [worktreeId]);

  useEffect(() => {
    setFileSearchUnavailable(false);
  }, [worktreeId]);

  const repoOpts = useMemo<RepoFileSuggestionOptions>(
    () => ({
      onRepoUnavailable: () => setFileSearchUnavailable(true),
      onRepoAvailable: () => setFileSearchUnavailable(false),
      onSuggestFetchChange: setFileSuggestBusy,
      onFilePicked,
      getWorktreeId: () => worktreeIdRef.current,
    }),
    [onFilePicked],
  );

  const extensions = useMemo(
    () => buildRichPromptExtensions(placeholder, repoOpts),
    [placeholder, repoOpts],
  );

  // TipTap 3 + React StrictMode: sync create during render leaves a destroyed
  // editor whose `.commands` getter throws on the remount effect pass.
  const editor = useEditor({
    immediatelyRender: false,
    extensions,
    content: "<p></p>",
    editable: !disabled,
    editorProps: {
      attributes: {
        class: "rich-prompt-editor",
        id,
        "aria-labelledby": `${id}-label`,
      },
    },
    onUpdate: ({ editor: ed }) => {
      const html = ed.getHTML();
      lastEmittedHtml.current = html;
      onChange(html);
    },
  });

  useEffect(() => {
    if (!editor || editor.isDestroyed) return;
    editor.setEditable(!disabled);
  }, [editor, disabled]);

  useEffect(() => {
    if (!editor || editor.isDestroyed) return;
    if (value === lastEmittedHtml.current) return;
    const next = looksLikeStoredHtml(value)
      ? value
      : plainTextToInitialHtml(value);
    editor.commands.setContent(next, { emitUpdate: false });
    lastEmittedHtml.current = next;
    setPendingInsert(null);
  }, [editor, value]);

  const probeDone = workspaceProbe !== "pending";
  const fileSearchLoadingEligible =
    probeDone &&
    workspaceProbe.state === "available" &&
    fileSuggestBusy &&
    !fileSearchUnavailable;
  const showFileSearchSpinner = useDelayedTrue(fileSearchLoadingEligible, 300);

  const repoHints = computeRepoHintFlags(
    workspaceProbe,
    fileSearchUnavailable,
    showFileSearchSpinner,
    worktreeId,
  );

  const insertPathOnly = useCallback(() => {
    if (!editor || editor.isDestroyed || !pendingInsert) return;
    insertRepoFileMentionAt(
      editor,
      pendingInsert.insertAt,
      pendingInsert.path,
    );
    setPendingInsert(null);
  }, [editor, pendingInsert]);

  const insertWithRange = useCallback(
    async (startLine: number, endLine: number) => {
      if (!editor || editor.isDestroyed || !pendingInsert) return;
      const { insertAt, path } = pendingInsert;
      setRangeWarning(null);
      const scopedWorktreeId = worktreeIdRef.current?.trim();
      const res = await validateRepoRange(path, startLine, endLine, {
        worktreeId: scopedWorktreeId || undefined,
      });
      if (res === null) {
        if (editor.isDestroyed) return;
        insertRepoFileMentionAt(editor, insertAt, path, startLine, endLine);
        setPendingInsert(null);
        return;
      }
      if (!res.ok) {
        setRangeWarning(
          res.warning ??
            "That line range is not valid for this file (check line numbers).",
        );
        return;
      }
      if (editor.isDestroyed) return;
      insertRepoFileMentionAt(editor, insertAt, path, startLine, endLine);
      setPendingInsert(null);
    },
    [editor, pendingInsert],
  );

  const dismissPendingInsert = useCallback(() => {
    setPendingInsert(null);
    setRangeWarning(null);
  }, []);

  return {
    editor,
    pendingInsert,
    rangeWarning,
    dismissPendingInsert,
    insertPathOnly,
    insertWithRange,
    repoHints,
  };
}
