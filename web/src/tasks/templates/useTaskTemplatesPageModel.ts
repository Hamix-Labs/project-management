import { useEffect, useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { useNavigate } from "react-router-dom";
import { getTaskTemplate, listGlobalGitWorktrees, listTaskTemplates } from "@/api";
import { TASK_TIMINGS } from "@/constants/tasks";
import { useDelayedTrue } from "@/lib/useDelayedTrue";
import { useDebouncedTrimmedValue } from "@/hooks/useDebouncedTrimmedValue";
import { useOptionalToast } from "@/shared/toast/ToastProvider";
import type { TaskTemplateSummary, TemplateFunctionBinding } from "@/types";
import { useDeleteWithExitAnimation } from "../hooks/useDeleteWithExitAnimation";
import { taskQueryKeys } from "../task-query";
import type { useTasksAppContext } from "../app/TasksAppProvider";
import type { TemplateSortKey } from "./components/TemplateToolbar";
import {
  bindingsFromDrafts,
  buildBindDraftsFromDetails,
  type TemplateBindDraft,
  validateBindDrafts,
} from "./components/TemplateFunctionBindModal";
import {
  clampInstanceCount,
  formatInstantiateBatchError,
  sumSelectedInstanceCounts,
} from "./templateUtils";

async function resolveWorktreeIdForTemplate(detail: {
  payload: { worktree_id?: string; repository_id?: string };
}): Promise<string | null> {
  const fromPayload = detail.payload.worktree_id?.trim();
  if (fromPayload) return fromPayload;
  const repoId = detail.payload.repository_id?.trim();
  if (!repoId) return null;
  try {
    const trees = await listGlobalGitWorktrees(repoId);
    const main = trees.find((wt) => wt.is_main);
    return main?.id ?? null;
  } catch {
    return null;
  }
}

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
  const toast = useOptionalToast();
  const [searchInput, setSearchInput] = useState("");
  const debouncedQ = useDebouncedTrimmedValue(searchInput, 300);
  const [sort, setSort] = useState<TemplateSortKey>("recent");
  const [selectedIds, setSelectedIds] = useState<string[]>([]);
  const [batchDefaultCount, setBatchDefaultCount] = useState(1);
  const [instanceCounts, setInstanceCounts] = useState<Record<string, number>>({});
  const [batchError, setBatchError] = useState<string | null>(null);
  const [bindDrafts, setBindDrafts] = useState<TemplateBindDraft[] | null>(null);
  const [bindError, setBindError] = useState<string | null>(null);

  const apiSort = sortToApiParams(sort);
  const queryParams = useMemo(() => {
    const params: {
      q?: string;
      sort: "updated_at" | "name" | "instantiate_count";
      order: "asc" | "desc";
    } = { ...apiSort };
    if (debouncedQ) params.q = debouncedQ;
    return params;
  }, [debouncedQ, apiSort]);

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
  const hasFilters = debouncedQ !== "";

  const toggleSelected = (id: string) => {
    setSelectedIds((current) => {
      if (current.includes(id)) {
        return current.filter((value) => value !== id);
      }
      setInstanceCounts((counts) => ({
        ...counts,
        [id]: batchDefaultCount,
      }));
      return [...current, id];
    });
  };

  const toggleSelectAll = () => {
    if (allSelected) {
      setSelectedIds([]);
      return;
    }
    const ids = templates.map((t) => t.id);
    setSelectedIds(ids);
    setInstanceCounts((current) => {
      const next = { ...current };
      for (const id of ids) {
        next[id] = batchDefaultCount;
      }
      return next;
    });
  };

  const setInstanceCountForTemplate = (id: string, count: number) => {
    setInstanceCounts((current) => ({
      ...current,
      [id]: clampInstanceCount(count),
    }));
  };

  const setBatchDefaultCountAndApply = (count: number) => {
    const clamped = clampInstanceCount(count);
    setBatchDefaultCount(clamped);
    setInstanceCounts((current) => {
      const next = { ...current };
      for (const id of selectedIds) {
        next[id] = clamped;
      }
      return next;
    });
  };

  const clearSelection = () => setSelectedIds([]);

  const clearFilters = () => {
    setSearchInput("");
  };

  const executeInstantiate = async (
    bindingByTemplate: Record<string, TemplateFunctionBinding[]>,
  ) => {
    setBatchError(null);
    setBindError(null);
    try {
      const items = selectedIds.map((id) => ({
        template_id: id,
        count: instanceCounts[id] ?? 1,
        ...(bindingByTemplate[id]?.length
          ? { function_bindings: bindingByTemplate[id] }
          : {}),
      }));
      const result = await app.instantiateTemplates(items);
      const batchMessage = formatInstantiateBatchError(result);
      if (batchMessage !== null && !result.accepted) {
        setBatchError(batchMessage);
        setBindDrafts(null);
        if (result.errors.length > 0) {
          setSelectedIds([...new Set(result.errors.map((entry) => entry.template_id))]);
        }
        return;
      }
      setBindDrafts(null);
      setSelectedIds([]);
      if (result.errors.length > 0) {
        toast.info(batchMessage ?? `Creating ${result.total} tasks…`);
      } else {
        toast.success(
          result.total === 1
            ? "Creating 1 task…"
            : `Creating ${result.total} tasks…`,
        );
      }
      navigate("/");
    } catch (err) {
      const message =
        err instanceof Error ? err.message : "Could not create tasks from templates.";
      if (bindDrafts) {
        setBindError(message);
      } else {
        setBatchError(message);
      }
    }
  };

  const runBatchCreate = async () => {
    if (selectedIds.length === 0) return;
    setBatchError(null);
    const selected = templates.filter((t) => selectedIds.includes(t.id));
    const functionTemplates = selected.filter((t) => t.is_function);
    if (functionTemplates.length === 0) {
      await executeInstantiate({});
      return;
    }
    try {
      const details = await Promise.all(functionTemplates.map((t) => getTaskTemplate(t.id)));
      const worktreeByTemplate: Record<string, string | null> = {};
      await Promise.all(
        details.map(async (d) => {
          worktreeByTemplate[d.id] = await resolveWorktreeIdForTemplate(d);
        }),
      );
      setBindDrafts(buildBindDraftsFromDetails(details, worktreeByTemplate));
      setBindError(null);
    } catch (err) {
      setBatchError(
        err instanceof Error ? err.message : "Could not load template function inputs.",
      );
    }
  };

  const confirmBindAndCreate = async () => {
    if (!bindDrafts) return;
    const msg = validateBindDrafts(bindDrafts);
    if (msg) {
      setBindError(msg);
      return;
    }
    await executeInstantiate(bindingsFromDrafts(bindDrafts));
  };

  return {
    searchInput,
    setSearchInput,
    debouncedQ,
    sort,
    setSort,
    selectedIds,
    batchDefaultCount,
    setBatchDefaultCountAndApply,
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
    clearSelection,
    clearFilters,
    deleteTemplate: deleteWithExit,
    runBatchCreate,
    bindDrafts,
    setBindDrafts,
    bindError,
    confirmBindAndCreate,
    closeBindModal: () => {
      setBindDrafts(null);
      setBindError(null);
    },
  };
}

export type TaskTemplatesPageModel = ReturnType<typeof useTaskTemplatesPageModel>;

export type { TaskTemplateSummary };
