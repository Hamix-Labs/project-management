import { useQuery } from "@tanstack/react-query";
import { useMemo } from "react";
import { taskQueryKeys } from "@/lib/taskQueryKeys";
import { useHysteresisBoolean } from "@/lib/useHysteresisBoolean";
import { errorMessage } from "@/lib/errorMessage";
import { TASK_TIMINGS } from "@/constants/tasks";
import { flattenTaskTreeRoots } from "../task-tree";
import { QUERY_POLICY } from "../queryPolicy";
import { fetchActiveTasksForBoard } from "../components/task-board/fetchActiveTasksForBoard";
import type { TaskHomeView } from "../pages/taskHomeView";

export type UseTasksBoardOptions = {
  view: TaskHomeView;
  /** When false, board query stays suspended (same gate as home list). */
  dataEnabled?: boolean;
  /** When false, wait for bootstrap before enabling. */
  bootstrapSettled?: boolean;
};

/**
 * Board-only query: page-walks active tasks when `view === "board"`.
 * Cache key sits under `listRoot()` for SSE / optimistic updates.
 */
export function useTasksBoard({
  view,
  dataEnabled = true,
  bootstrapSettled = true,
}: UseTasksBoardOptions) {
  const enabled = view === "board" && dataEnabled && bootstrapSettled;

  const query = useQuery({
    queryKey: taskQueryKeys.board(),
    queryFn: ({ signal }) => fetchActiveTasksForBoard({ signal }),
    enabled,
    staleTime: QUERY_POLICY.listStaleTimeMs,
  });

  const rootTasks = useMemo(
    () => query.data?.tasks ?? [],
    [query.data?.tasks],
  );
  const tasks = useMemo(
    () => flattenTaskTreeRoots(rootTasks),
    [rootTasks],
  );

  const loading = enabled && query.isPending;
  const rawRefreshing = query.isFetching && !query.isPending;
  const refreshing = useHysteresisBoolean(
    rawRefreshing,
    TASK_TIMINGS.listRefreshShowMs,
    TASK_TIMINGS.listRefreshHideMs,
  );

  return {
    tasks,
    loading,
    refreshing,
    error: query.isError ? errorMessage(query.error) : null,
    truncated: query.data?.has_more === true,
    refetch: query.refetch,
  };
}
