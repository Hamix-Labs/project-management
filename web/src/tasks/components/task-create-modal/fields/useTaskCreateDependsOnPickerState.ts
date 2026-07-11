import { useQuery } from "@tanstack/react-query";
import {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
  type KeyboardEvent,
} from "react";
import { listTasks } from "@/api";
import { taskQueryKeys } from "@/tasks/task-query";
import type { Task } from "@/types";
import {
  MAX_TYPEAHEAD_RESULTS,
  TYPEAHEAD_BLUR_DELAY_MS,
  buildDependsOnHelperCopy,
  filterBrowseCandidates,
  filterTypeaheadCandidates,
} from "./taskCreateDependsOnUtils";

export type TaskCreateDependsOnPickerProps = {
  /**
   * Project the new task is being scoped to. The picker only surfaces
   * tasks that share this `project_id` — picking dependencies across
   * project boundaries is not a use case we support today (cross-project
   * dependency wiring belongs in the detail page after both tasks exist).
   * Empty string means "no project chosen" and the picker reads as
   * disabled chrome with a one-line nudge.
   */
  projectId: string;
  selected: string[];
  onChange: (next: string[]) => void;
  disabled: boolean;
};

export function useTaskCreateDependsOnPickerState({
  projectId,
  selected,
  onChange,
  disabled,
}: TaskCreateDependsOnPickerProps) {
  const hasProject = projectId.trim().length > 0;
  const [query, setQuery] = useState("");
  const [listOpen, setListOpen] = useState(false);
  const [browseOpen, setBrowseOpen] = useState(false);
  const [browseQuery, setBrowseQuery] = useState("");
  const blurTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    return () => {
      if (blurTimerRef.current) clearTimeout(blurTimerRef.current);
    };
  }, []);

  const tasksQuery = useQuery({
    queryKey: taskQueryKeys.list({ limit: 200, offset: 0 }),
    queryFn: ({ signal }) => listTasks(200, 0, { signal }),
    enabled: hasProject,
    staleTime: 30_000,
  });

  const projectTasks = useMemo(() => {
    if (!hasProject) return [] as Task[];
    return (tasksQuery.data?.tasks ?? []).filter(
      (t) => t.project_id === projectId,
    );
  }, [hasProject, projectId, tasksQuery.data?.tasks]);

  const labelLookup = useMemo(() => {
    const m = new Map<string, string>();
    for (const t of projectTasks) m.set(t.id, t.title);
    return m;
  }, [projectTasks]);

  const selectedSet = useMemo(() => new Set(selected), [selected]);

  const typeaheadResults = useMemo(
    () =>
      filterTypeaheadCandidates(
        projectTasks,
        query,
        selectedSet,
        MAX_TYPEAHEAD_RESULTS,
      ),
    [projectTasks, query, selectedSet],
  );

  const browseResults = useMemo(
    () => filterBrowseCandidates(projectTasks, browseQuery),
    [projectTasks, browseQuery],
  );

  const inputDisabled = disabled || !hasProject;

  const addId = useCallback(
    (id: string) => {
      if (selectedSet.has(id)) return;
      onChange([...selected, id]);
    },
    [onChange, selected, selectedSet],
  );

  const removeId = useCallback(
    (id: string) => {
      if (!selectedSet.has(id)) return;
      onChange(selected.filter((s) => s !== id));
    },
    [onChange, selected, selectedSet],
  );

  const toggleId = useCallback(
    (id: string) => {
      if (selectedSet.has(id)) removeId(id);
      else addId(id);
    },
    [addId, removeId, selectedSet],
  );

  const handleSelectFromTypeahead = useCallback(
    (id: string) => {
      addId(id);
      setQuery("");
      setListOpen(true);
    },
    [addId],
  );

  const handleInputFocus = useCallback(() => {
    if (blurTimerRef.current) {
      clearTimeout(blurTimerRef.current);
      blurTimerRef.current = null;
    }
    setListOpen(true);
  }, []);

  const handleInputBlur = useCallback(() => {
    // Defer closing the listbox so a click on a result still fires its
    // `mousedown -> blur -> click` sequence before the listbox unmounts.
    blurTimerRef.current = setTimeout(() => {
      setListOpen(false);
      blurTimerRef.current = null;
    }, TYPEAHEAD_BLUR_DELAY_MS);
  }, []);

  const handleInputKeyDown = useCallback(
    (e: KeyboardEvent<HTMLInputElement>) => {
      if (e.key === "Escape" && listOpen) {
        e.preventDefault();
        setListOpen(false);
        return;
      }
      if (e.key === "Enter" && listOpen && typeaheadResults.length > 0) {
        e.preventDefault();
        handleSelectFromTypeahead(typeaheadResults[0].id);
      }
    },
    [handleSelectFromTypeahead, listOpen, typeaheadResults],
  );

  const handleQueryChange = useCallback((value: string) => {
    setQuery(value);
    setListOpen(true);
  }, []);

  const helperCopy = buildDependsOnHelperCopy(
    hasProject,
    tasksQuery.isLoading,
    projectTasks.length,
  );

  return {
    hasProject,
    query,
    listOpen,
    browseOpen,
    browseQuery,
    projectTasks,
    labelLookup,
    selectedSet,
    typeaheadResults,
    browseResults,
    inputDisabled,
    helperCopy,
    removeId,
    toggleId,
    handleSelectFromTypeahead,
    handleInputFocus,
    handleInputBlur,
    handleInputKeyDown,
    handleQueryChange,
    setBrowseOpen,
    setBrowseQuery,
  };
}
