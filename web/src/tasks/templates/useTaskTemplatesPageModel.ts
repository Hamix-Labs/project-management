import { useEffect, useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { useNavigate } from "react-router-dom";
import { listTaskTemplates } from "@/api";
import { TASK_TIMINGS } from "@/constants/tasks";
import { useDelayedTrue } from "@/lib/useDelayedTrue";
import { useDebouncedTrimmedValue } from "@/hooks/useDebouncedTrimmedValue";
import type { TaskTemplateSummary } from "@/types";
import { useDeleteWithExitAnimation } from "../hooks/useDeleteWithExitAnimation";
import { taskQueryKeys } from "../task-query";
import type { useTasksAppContext } from "../app/TasksAppProvider";
import { TEMPLATE_CATEGORY_LABELS } from "./templateCategories";
import type { TemplateSortKey } from "./components/TemplateToolbar";
import {
  clampInstanceCount,
  formatInstantiateBatchError,
  sumSelectedInstanceCounts,
} from "./templateUtils";

type TaskTemplatesApp = ReturnType<typeof useTasksAppContext>;

function sortToApiParams(sort: TemplateSortKey): {
  sort: "updated_at" | "name" | "instantiate_count";
  order: "asc" | "desc";
} {
  if (sort === "title") return { sort: "name", order: "asc" };
  if (sort === "runs") return { sort: "instantiate_count", order: "desc" };
  return { sort: "updated_at", order: "desc" };
}

export function useTaskTemplatesPageModel(app: TaskTemplatesApp, navigate: ReturnType<typeof useNavigate>) {
  const [searchInput, setSearchInput] = useState("");
  const debouncedQ = useDebouncedTrimmedValue(searchInput, 300);
  const [sort, setSort] = useState<TemplateSortKey>("recent");
  const [activeTag, setActiveTag] = useState("all");
  const [selectedIds, setSelectedIds] = useState<string[]>([]);
  const [batchDefaultCount, setBatchDefaultCount] = useState(1);
  const [instanceCounts, setInstanceCounts] = useState<Record<string, number>>({});
  const [batchError, setBatchError] = useState<string | null>(null);

  const apiSort = sortToApiParams(sort);
  const queryParams = useMemo(() => {
    const params: {
      q?: string;
      sort: "updated_at" | "name" | "instantiate_count";
      order: "asc" | "desc";
      tag?: string;
    } = { ...apiSort };
    if (debouncedQ) params.q = debouncedQ;
    if (activeTag !== "all") params.tag = activeTag;
    return params;
  }, [debouncedQ, activeTag, apiSort]);

  const templatesQuery = useQuery({
    queryKey: taskQueryKeys.templates(queryParams),
    queryFn: ({ signal }) => listTaskTemplates({ ...queryParams, signal }),
  });

  const templates = useMemo(() => templatesQuery.data ?? [], [templatesQuery.data]);
  const loading = templatesQuery.isPending;
  const error = templatesQuery.isError
    ? templatesQuery.error instanceof Error
      ? templatesQuery.error.message
      : "Could not load templates."
    : null;
  const showSkeleton = useDelayedTrue(loading, TASK_TIMINGS.draftResumeMinLoadingMs);
  const renderNow = new Date();

  const dynamicTags = useMemo(() => {
    const presets = new Set<string>(TEMPLATE_CATEGORY_LABELS);
    const tags = new Set<string>();
    for (const template of templates) {
      if (template.primary_tag && !presets.has(template.primary_tag)) {
        tags.add(template.primary_tag);
      }
    }
    return [...tags].sort((a, b) => a.localeCompare(b));
  }, [templates]);

  const { deletingId: deletingTemplateId, exitingIds: exitingTemplateIds, deleteWithExit } =
    useDeleteWithExitAnimation({
      entityIds: templates.map((t) => t.id),
      onDelete: async (templateId) => {
        await app.deleteTemplateByID(templateId);
        setSelectedIds((current) => current.filter((id) => id !== templateId));
      },
    });

  useEffect(() => {
    const ids = new Set(templates.map((t) => t.id));
    setSelectedIds((current) => current.filter((id) => ids.has(id)));
    setInstanceCounts((current) => {
      const next: Record<string, number> = {};
      for (const template of templates) {
        next[template.id] = current[template.id] ?? 1;
      }
      return next;
    });
  }, [templates]);

  const allSelected = templates.length > 0 && selectedIds.length === templates.length;
  const someSelected = selectedIds.length > 0 && !allSelected;
  const selectedCount = selectedIds.length;
  const totalTaskCount = sumSelectedInstanceCounts(selectedIds, instanceCounts);
  const hasFilters = debouncedQ !== "" || activeTag !== "all";

  const toggleSelected = (id: string) => {
    setSelectedIds((current) =>
      current.includes(id) ? current.filter((value) => value !== id) : [...current, id],
    );
  };

  const toggleSelectAll = () => {
    if (allSelected) {
      setSelectedIds([]);
      return;
    }
    setSelectedIds(templates.map((t) => t.id));
  };

  const setInstanceCountForTemplate = (id: string, count: number) => {
    setInstanceCounts((current) => ({
      ...current,
      [id]: clampInstanceCount(count),
    }));
  };

  const applyBatchDefaultToSelected = () => {
    setInstanceCounts((current) => {
      const next = { ...current };
      for (const id of selectedIds) {
        next[id] = batchDefaultCount;
      }
      return next;
    });
  };

  const clearSelection = () => setSelectedIds([]);

  const clearFilters = () => {
    setSearchInput("");
    setActiveTag("all");
  };

  const runBatchCreate = async () => {
    if (selectedIds.length === 0) return;
    setBatchError(null);
    try {
      const items = selectedIds.map((id) => ({
        template_id: id,
        count: instanceCounts[id] ?? 1,
      }));
      const result = await app.instantiateTemplates(items);
      const batchMessage = formatInstantiateBatchError(result);
      if (batchMessage !== null) {
        setBatchError(batchMessage);
        if (result.errors.length > 0 && result.tasks.length > 0) {
          setSelectedIds([...new Set(result.errors.map((entry) => entry.template_id))]);
        }
        return;
      }
      setSelectedIds([]);
      navigate("/");
    } catch (err) {
      setBatchError(err instanceof Error ? err.message : "Could not create tasks from templates.");
    }
  };

  return {
    searchInput,
    setSearchInput,
    debouncedQ,
    sort,
    setSort,
    activeTag,
    setActiveTag,
    dynamicTags,
    selectedIds,
    batchDefaultCount,
    setBatchDefaultCount,
    instanceCounts,
    deletingTemplateId,
    exitingTemplateIds,
    batchError,
    templatesQuery,
    templates,
    loading,
    error,
    showSkeleton,
    renderNow,
    allSelected,
    someSelected,
    selectedCount,
    totalTaskCount,
    hasFilters,
    toggleSelected,
    toggleSelectAll,
    setInstanceCountForTemplate,
    applyBatchDefaultToSelected,
    clearSelection,
    clearFilters,
    deleteTemplate: deleteWithExit,
    runBatchCreate,
  };
}

export type TaskTemplatesPageModel = ReturnType<typeof useTaskTemplatesPageModel>;

export type { TaskTemplateSummary };
