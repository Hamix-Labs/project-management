import { useEditor } from "@tiptap/react";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { validateRepoRange } from "@/api";
import { useDelayedTrue } from "@/lib/useDelayedTrue";
import {
  looksLikeStoredHtml,
  plainTextToInitialHtml,
} from "@/lib/promptFormat";
import {
  expandProjectContextSelection,
  mergeProjectContextSelection,
  selectedProjectContextItems,
  type ProjectContextAddMode,
} from "@/lib/projectContextRefs";
import type { ProjectContextEdge, ProjectContextItem } from "@/types";
import type { ProjectContextPickedPayload } from "./extensions/projectContextSuggestion";
import type { RepoFileSuggestionOptions } from "./extensions/repoFileSuggestion";
import type { ProjectContextSuggestionOptions } from "./extensions/projectContextSuggestion";
import { buildRichPromptExtensions } from "./richPromptExtensions";
import {
  computeRepoHintFlags,
  insertProjectContextChipAt,
  insertRepoFileMentionAt,
  type PendingFileInsert,
  type PendingProjectChoice,
} from "./richPromptInsertHelpers";
import { useRepoWorkspaceProbe } from "./useRepoWorkspaceProbe";
import type { RichPromptEditorProps } from "./richPromptEditorTypes";

const EMPTY_CONTEXT_ITEMS: ProjectContextItem[] = [];
const EMPTY_CONTEXT_EDGES: ProjectContextEdge[] = [];
const EMPTY_SELECTED_IDS: string[] = [];

export function useRichPromptEditorController({
  id,
  value,
  onChange,
  disabled,
  placeholder,
  projectContext,
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

  const projectItems = projectContext?.items ?? EMPTY_CONTEXT_ITEMS;
  const projectEdges = projectContext?.edges ?? EMPTY_CONTEXT_EDGES;
  const selectedProjectIds = projectContext?.selectedIds ?? EMPTY_SELECTED_IDS;
  const onProjectIdsChange = projectContext?.onSelectedIdsChange;

  const projectItemsRef = useRef<ProjectContextItem[] | null>(
    projectContext != null ? projectItems : null,
  );
  useEffect(() => {
    projectItemsRef.current = projectContext != null ? projectItems : null;
  }, [projectContext, projectItems]);

  const [pendingProjectChoice, setPendingProjectChoice] =
    useState<PendingProjectChoice | null>(null);

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

  const projectContextEnabled = projectContext != null;

  const onProjectContextPicked = useCallback(
    (payload: ProjectContextPickedPayload) => {
      setPendingProjectChoice({
        item: payload.item,
        insertAt: payload.insertAt,
      });
    },
    [],
  );

  const projectSuggestionOpts = useMemo<ProjectContextSuggestionOptions>(
    () => ({
      getItems: () => projectItemsRef.current,
      onContextPicked: onProjectContextPicked,
    }),
    [onProjectContextPicked],
  );

  const extensions = useMemo(
    () =>
      buildRichPromptExtensions(placeholder, repoOpts, projectSuggestionOpts),
    [placeholder, repoOpts, projectSuggestionOpts],
  );

  const editor = useEditor({
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
    editor?.setEditable(!disabled);
  }, [editor, disabled]);

  useEffect(() => {
    if (!editor) return;
    if (value === lastEmittedHtml.current) return;
    const next = looksLikeStoredHtml(value)
      ? value
      : plainTextToInitialHtml(value);
    editor.commands.setContent(next, { emitUpdate: false });
    lastEmittedHtml.current = next;
    setPendingInsert(null);
    setPendingProjectChoice(null);
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
    if (!editor || !pendingInsert) return;
    insertRepoFileMentionAt(
      editor,
      pendingInsert.insertAt,
      pendingInsert.path,
    );
    setPendingInsert(null);
  }, [editor, pendingInsert]);

  const insertWithRange = useCallback(
    async (startLine: number, endLine: number) => {
      if (!editor || !pendingInsert) return;
      const { insertAt, path } = pendingInsert;
      setRangeWarning(null);
      const res = await validateRepoRange(path, startLine, endLine);
      if (res === null) {
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
      insertRepoFileMentionAt(editor, insertAt, path, startLine, endLine);
      setPendingInsert(null);
    },
    [editor, pendingInsert],
  );

  const insertProjectContextChip = useCallback(
    (item: ProjectContextItem, insertAt: number | null) => {
      if (!editor) return;
      insertProjectContextChipAt(editor, item, insertAt);
    },
    [editor],
  );

  const confirmProjectContextChoice = useCallback(
    (mode: ProjectContextAddMode) => {
      if (!pendingProjectChoice) return;
      const { item, insertAt } = pendingProjectChoice;
      const expanded = expandProjectContextSelection(
        item.id,
        mode,
        projectEdges,
      );
      const merged = mergeProjectContextSelection(
        selectedProjectIds,
        expanded,
      );
      onProjectIdsChange?.(merged);
      insertProjectContextChip(item, insertAt);
      setPendingProjectChoice(null);
    },
    [
      pendingProjectChoice,
      projectEdges,
      selectedProjectIds,
      onProjectIdsChange,
      insertProjectContextChip,
    ],
  );

  const cancelProjectContextChoice = useCallback(() => {
    setPendingProjectChoice(null);
  }, []);

  const removeSelectedProjectId = useCallback(
    (contextId: string) => {
      if (!onProjectIdsChange) return;
      const next = selectedProjectIds.filter(
        (existing) => existing !== contextId,
      );
      if (next.length === selectedProjectIds.length) return;
      onProjectIdsChange(next);
    },
    [onProjectIdsChange, selectedProjectIds],
  );

  const referencesItems = useMemo(
    () => selectedProjectContextItems(projectItems, selectedProjectIds),
    [projectItems, selectedProjectIds],
  );

  const dismissPendingInsert = useCallback(() => {
    setPendingInsert(null);
    setRangeWarning(null);
  }, []);

  return {
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
  };
}
