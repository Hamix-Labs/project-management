import { useQuery } from "@tanstack/react-query";
import { useMemo } from "react";
import { taskQueryKeys } from "@/lib/taskQueryKeys";
import { useHysteresisBoolean } from "@/lib/useHysteresisBoolean";
import { errorMessage } from "@/lib/errorMessage";
import { TASK_TIMINGS } from "@/constants/tasks";
import { flattenTaskTreeRoots, sortWorktreeFamilyTasks } from "../task-tree";
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
  worktreeFamilyId = "all",
}: UseTasksBoardOptions & { worktreeFamilyId?: string }) {
  const enabled = view === "board" && dataEnabled && bootstrapSettled;
  const familyWorktreeId =
    worktreeFamilyId !== "all" ? worktreeFamilyId.trim() : "";

  const query = useQuery({
    queryKey: taskQueryKeys.board(
      familyWorktreeId ? { worktreeId: familyWorktreeId } : undefined,
    ),
    queryFn: ({ signal }) =>
      fetchActiveTasksForBoard({
        signal,
        ...(familyWorktreeId ? { worktreeId: familyWorktreeId } : {}),
      }),
    enabled,
    staleTime: QUERY_POLICY.listStaleTimeMs,
  });

  const rootTasks = useMemo(() => {
    const rows = query.data?.tasks ?? [];
    if (!familyWorktreeId) return rows;
    return sortWorktreeFamilyTasks(rows);
  }, [query.data?.tasks, familyWorktreeId]);
  const tasks = useMemo(
    () =>
      flattenTaskTreeRoots(rootTasks, {
        worktreeFamilyActive: familyWorktreeId !== "",
      }),
    [rootTasks, familyWorktreeId],
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
