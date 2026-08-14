import { useEditor } from "@tiptap/react";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { validateRepoRange } from "@/api";
import { useDelayedTrue } from "@/lib/useDelayedTrue";
import {
  looksLikeStoredHtml,
  plainTextToInitialHtml,
} from "@/lib/promptFormat";
import { useResolvedMentionWorktreeId } from "@/hooks/useResolvedMentionWorktreeId";
import { useRepoFileIndex } from "@/hooks/useRepoFileIndex";
import { clearRepoFileIndex } from "@/lib/repoFileIndex";
import type { RepoFileSuggestionOptions } from "./extensions/repoFileSuggestion";
import { buildRichPromptExtensions } from "./richPromptExtensions";
import {
  computeRepoHintFlags,
  insertRepoFileMentionAt,
  type PendingFileInsert,
} from "./richPromptInsertHelpers";
import { useRepoWorkspaceProbe } from "./useRepoWorkspaceProbe";
import type { RichPromptEditorProps } from "./richPromptEditorTypes";

function initialPromptHtml(value: string): string {
  if (looksLikeStoredHtml(value)) {
    return value.trim() === "" ? "<p></p>" : value;
  }
  return plainTextToInitialHtml(value);
}

export type AiComposerState = {
  open: boolean;
  initialValue: string;
};

export function useRichPromptEditorController({
  id,
  value,
  onChange,
  disabled,
  placeholder,
  worktreeId,
  repositoryId,
  preferRepositoryHint = false,
  onAiTrigger,
}: RichPromptEditorProps) {
  const gitScoped = worktreeId !== undefined || repositoryId !== undefined;
  const mentionWorktreeId = useResolvedMentionWorktreeId(
    worktreeId,
    repositoryId,
  );

  useRepoFileIndex(mentionWorktreeId);

  const workspaceProbe = useRepoWorkspaceProbe(
    gitScoped ? mentionWorktreeId : undefined,
  );
  const [fileSearchUnavailable, setFileSearchUnavailable] = useState(false);
  const [fileSuggestBusy, setFileSuggestBusy] = useState(false);
  const [pendingInsert, setPendingInsert] = useState<PendingFileInsert | null>(
    null,
  );
  const [rangeWarning, setRangeWarning] = useState<string | null>(null);
  const initialContentRef = useRef(initialPromptHtml(value));
  const lastEmittedHtml = useRef<string | null>(initialContentRef.current);

  const onFilePicked = useCallback(
    (payload: { insertAt: number; path: string }) => {
      setPendingInsert({ insertAt: payload.insertAt, path: payload.path });
      setRangeWarning(null);
    },
    [],
  );

  const worktreeIdRef = useRef(mentionWorktreeId);
  useEffect(() => {
    worktreeIdRef.current = mentionWorktreeId;
  }, [mentionWorktreeId]);

  useEffect(() => {
    setFileSearchUnavailable(false);
  }, [mentionWorktreeId]);

  useEffect(() => {
    const idToClear = mentionWorktreeId;
    return () => {
      if (idToClear.trim() !== "") {
        clearRepoFileIndex(idToClear);
      }
    };
  }, [mentionWorktreeId]);

  const repoOpts = useMemo<RepoFileSuggestionOptions>(
    () => ({
      onRepoUnavailable: () => setFileSearchUnavailable(true),
      onRepoAvailable: () => setFileSearchUnavailable(false),
      onSuggestFetchChange: setFileSuggestBusy,
      onFilePicked,
      getWorktreeId: () => {
        const id = worktreeIdRef.current.trim();
        return id !== "" ? id : undefined;
      },
    }),
    [onFilePicked],
  );

  const [aiComposer, setAiComposer] = useState<AiComposerState>({
    open: false,
    initialValue: "",
  });

  const onAiTriggerRef = useRef(onAiTrigger);
  useEffect(() => {
    onAiTriggerRef.current = onAiTrigger;
  }, [onAiTrigger]);

  const handleAiTrigger = useCallback((msg: string) => {
    setAiComposer({ open: true, initialValue: msg });
    onAiTriggerRef.current?.(msg);
  }, []);

  const closeAiComposer = useCallback(() => {
    setAiComposer((prev) => (prev.open ? { ...prev, open: false } : prev));
  }, []);

  const extensions = useMemo(
    () =>
      buildRichPromptExtensions(placeholder, repoOpts, {
        onAiTrigger: handleAiTrigger,
      }),
    [placeholder, repoOpts, handleAiTrigger],
  );

  const editor = useEditor({
    immediatelyRender: false,
    extensions,
    content: initialContentRef.current,
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
    {
      gitScoped,
      mentionWorktreeId: gitScoped ? mentionWorktreeId : undefined,
      preferRepositoryHint,
      repositoryId,
    },
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

  const getAiAnchorRect = useCallback((): DOMRect | null => {
    if (!editor || editor.isDestroyed) return null;
    try {
      const { from } = editor.state.selection;
      const coords = editor.view.coordsAtPos(from);
      return new DOMRect(
        coords.left,
        coords.top,
        1,
        coords.bottom - coords.top,
      );
    } catch {
      return null;
    }
  }, [editor]);

  return {
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
  };
}
