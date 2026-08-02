import { useQuery } from "@tanstack/react-query";
import { useCallback, useEffect, useMemo, useState } from "react";
import { getTaskStats, listTasks } from "../../api";
import { flattenTaskTreeRoots, sortWorktreeFamilyTasks } from "../task-tree";
import { TASK_LIST_PAGE_SIZE } from "../task-paging";
import { taskQueryKeys } from "../task-query";
import { useHysteresisBoolean } from "@/lib/useHysteresisBoolean";
import { TASK_TIMINGS } from "@/constants/tasks";
import { QUERY_POLICY } from "../queryPolicy";
import { errorMessage } from "@/lib/errorMessage";

const LIST_REFRESH_SHOW_MS = TASK_TIMINGS.listRefreshShowMs;
const LIST_REFRESH_HIDE_MS = TASK_TIMINGS.listRefreshHideMs;

export type UseTasksHomeListOptions = {
  /** When false, list/stats stay suspended (cache kept). */
  dataEnabled?: boolean;
  /** When false, wait for bootstrap before enabling queries. */
  bootstrapSettled?: boolean;
};

/**
 * Home list + stats queries and pagination. Composed by `useTasksApp`.
 */
export function useTasksHomeList({
  dataEnabled = true,
  bootstrapSettled = true,
  worktreeFamilyId = "all",
}: UseTasksHomeListOptions & { worktreeFamilyId?: string } = {}) {
  const [taskListPage, setTaskListPage] = useState(0);
  const homeDataReady = dataEnabled && bootstrapSettled;
  const familyWorktreeId =
    worktreeFamilyId !== "all" ? worktreeFamilyId.trim() : "";

  const tasksQuery = useQuery({
    queryKey: taskQueryKeys.list({
      limit: TASK_LIST_PAGE_SIZE,
      offset: taskListPage * TASK_LIST_PAGE_SIZE,
      ...(familyWorktreeId ? { worktreeId: familyWorktreeId } : {}),
    }),
    queryFn: ({ signal }) =>
      listTasks(
        TASK_LIST_PAGE_SIZE,
        taskListPage * TASK_LIST_PAGE_SIZE,
        {
          signal,
          ...(familyWorktreeId ? { worktreeId: familyWorktreeId } : {}),
        },
      ),
    enabled: homeDataReady,
    staleTime: QUERY_POLICY.listStaleTimeMs,
  });
  const taskStatsQuery = useQuery({
    queryKey: taskQueryKeys.stats(),
    queryFn: async ({ signal }) => {
      try {
        return await getTaskStats({ signal });
      } catch {
        return null;
      }
    },
    enabled: homeDataReady,
    staleTime: QUERY_POLICY.listStaleTimeMs,
  });

  const resetTaskListPage = useCallback(() => {
    setTaskListPage(0);
  }, []);

  const rootTaskTrees = useMemo(() => {
    const rows = tasksQuery.data?.tasks ?? [];
    if (!familyWorktreeId) return rows;
    return sortWorktreeFamilyTasks(rows);
  }, [tasksQuery.data?.tasks, familyWorktreeId]);
  const tasks = useMemo(
    () =>
      flattenTaskTreeRoots(rootTaskTrees, {
        worktreeFamilyActive: familyWorktreeId !== "",
      }),
    [rootTaskTrees, familyWorktreeId],
  );

  const loading = tasksQuery.isPending;
  const rawListRefreshing =
    tasksQuery.isFetching && !tasksQuery.isPending;
  const listRefreshing = useHysteresisBoolean(
    rawListRefreshing,
    LIST_REFRESH_SHOW_MS,
    LIST_REFRESH_HIDE_MS,
  );

  useEffect(() => {
    if (!tasksQuery.isPending && rootTaskTrees.length === 0 && taskListPage > 0) {
      setTaskListPage(0);
    }
  }, [tasksQuery.isPending, rootTaskTrees.length, taskListPage]);

  const hasNextTaskPage = rootTaskTrees.length === TASK_LIST_PAGE_SIZE;
  const hasPrevTaskPage = taskListPage > 0;
  const listError = tasksQuery.isError ? errorMessage(tasksQuery.error) : null;

  return {
    tasks,
    rootTasksOnPage: rootTaskTrees.length,
    loading,
    listRefreshing,
    listError,
    taskStats: taskStatsQuery.data,
    taskStatsLoading: taskStatsQuery.isPending,
    taskListPage,
    setTaskListPage,
    resetTaskListPage,
    taskListPageSize: TASK_LIST_PAGE_SIZE,
    hasNextTaskPage,
    hasPrevTaskPage,
  };
}
