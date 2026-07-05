import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useEffect, useMemo, useRef } from "react";
import { getTask, listChecklist } from "@/api";
import { useDocumentTitle } from "@/shared/useDocumentTitle";
import { rememberPersistedDetailId } from "@/lib/queryPersist";
import { rumNavigationTiming } from "@/observability";
import type { Task, TaskChecklistResponse } from "@/types";
import type { UseQueryResult } from "@tanstack/react-query";
import { resolveTaskDependencySummaries, taskQueryKeys } from "../task-query";
import { QUERY_POLICY } from "../queryPolicy";

function useTaskDetailNavigationTiming(
  taskId: string,
  taskQuery: UseQueryResult<Task>,
  checklistQuery: UseQueryResult<TaskChecklistResponse>,
) {
  const navigationMountAtRef = useRef(performance.now());
  const taskTimingSentRef = useRef(false);
  const interactiveTimingSentRef = useRef(false);

  useEffect(() => {
    if (taskId) {
      rememberPersistedDetailId(taskId);
    }
    navigationMountAtRef.current = performance.now();
    taskTimingSentRef.current = false;
    interactiveTimingSentRef.current = false;
  }, [taskId]);

  useEffect(() => {
    if (!taskQuery.isSuccess || taskTimingSentRef.current) return;
    taskTimingSentRef.current = true;
    rumNavigationTiming(
      "navigation.task_detail.time_to_task_ms",
      performance.now() - navigationMountAtRef.current,
    );
  }, [taskQuery.isSuccess]);

  useEffect(() => {
    if (
      !taskQuery.isSuccess ||
      !checklistQuery.isSuccess ||
      interactiveTimingSentRef.current
    ) {
      return;
    }
    interactiveTimingSentRef.current = true;
    rumNavigationTiming(
      "navigation.task_detail.time_to_interactive_ms",
      performance.now() - navigationMountAtRef.current,
    );
  }, [taskQuery.isSuccess, checklistQuery.isSuccess]);
}

export function useTaskDetailPageQueries(taskId: string) {
  const queryClient = useQueryClient();

  const taskQuery = useQuery({
    queryKey: taskQueryKeys.detail(taskId),
    queryFn: ({ signal }) => getTask(taskId, { signal }),
    enabled: Boolean(taskId),
    staleTime: QUERY_POLICY.detailStaleTimeMs,
  });

  const checklistQuery = useQuery({
    queryKey: taskQueryKeys.checklist(taskId),
    queryFn: ({ signal }) => listChecklist(taskId, { signal }),
    enabled: Boolean(taskId),
    staleTime: QUERY_POLICY.detailStaleTimeMs,
  });

  useTaskDetailNavigationTiming(taskId, taskQuery, checklistQuery);

  const taskDocTitle =
    taskId && taskQuery.isSuccess && taskQuery.data
      ? taskQuery.data.title.trim() || "Untitled task"
      : null;
  useDocumentTitle(taskDocTitle);

  const dependencySummaries = useMemo(
    () =>
      resolveTaskDependencySummaries(
        queryClient,
        taskQuery.data?.depends_on ?? [],
      ),
    [queryClient, taskQuery.data?.depends_on],
  );

  return { taskQuery, checklistQuery, dependencySummaries };
}
