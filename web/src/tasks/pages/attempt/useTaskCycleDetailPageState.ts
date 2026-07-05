import { useQuery } from "@tanstack/react-query";
import { useEffect, useId, useState, type Dispatch, type SetStateAction } from "react";
import { useParams } from "react-router-dom";
import { listTaskEvents } from "@/api";
import { useDocumentTitle } from "@/shared/useDocumentTitle";
import { useNow } from "@/shared/useNow";
import type { TaskCycleDetail } from "@/types";
import type { UseQueryResult } from "@tanstack/react-query";
import { useTaskCycle, useTaskCycleStream } from "@/tasks/hooks/useTaskCycles";
import type { UseTaskCycleStreamResult } from "@/tasks/hooks/useTaskCycles";
import { taskQueryKeys } from "@/tasks/task-query";

const STREAM_VISIBLE_INITIAL = 6;
const AUDIT_VISIBLE_INITIAL = 6;

type ActivityTab = "cursor" | "audit";

export type TaskCycleDetailPageState = {
  taskId: string;
  cycleId: string;
  paramsValid: boolean;
  activityTab: ActivityTab;
  setActivityTab: (tab: ActivityTab) => void;
  visibleStreamCount: number;
  setVisibleStreamCount: Dispatch<SetStateAction<number>>;
  visibleAuditCount: number;
  setVisibleAuditCount: Dispatch<SetStateAction<number>>;
  cursorTabId: string;
  auditTabId: string;
  cursorPanelId: string;
  auditPanelId: string;
  cycleQuery: UseQueryResult<TaskCycleDetail, Error>;
  streamQuery: UseTaskCycleStreamResult;
  auditQuery: UseQueryResult<
    Awaited<ReturnType<typeof listTaskEvents>>,
    Error
  >;
  now: number;
};

export function useTaskCycleDetailPageState(): TaskCycleDetailPageState {
  const { taskId = "", cycleId = "" } = useParams<{
    taskId: string;
    cycleId: string;
  }>();
  const paramsValid = Boolean(taskId) && Boolean(cycleId);
  const [activityTab, setActivityTab] = useState<ActivityTab>("cursor");
  const [visibleStreamCount, setVisibleStreamCount] = useState(
    STREAM_VISIBLE_INITIAL,
  );
  const [visibleAuditCount, setVisibleAuditCount] =
    useState(AUDIT_VISIBLE_INITIAL);
  const cursorTabId = useId();
  const auditTabId = useId();
  const cursorPanelId = useId();
  const auditPanelId = useId();

  const cycleQuery = useTaskCycle(taskId, cycleId, { enabled: paramsValid });
  const streamQuery = useTaskCycleStream(taskId, cycleId, {
    enabled: paramsValid,
    limit: 500,
  });
  const auditQuery = useQuery({
    queryKey: taskQueryKeys.events(taskId, { k: "head" }),
    queryFn: ({ signal }) => listTaskEvents(taskId, { signal, limit: 200 }),
    enabled: Boolean(taskId),
  });

  useEffect(() => {
    setVisibleStreamCount(STREAM_VISIBLE_INITIAL);
    setVisibleAuditCount(AUDIT_VISIBLE_INITIAL);
    setActivityTab("cursor");
  }, [cycleId]);

  useDocumentTitle(
    cycleQuery.data
      ? `Attempt #${cycleQuery.data.attempt_seq}`
      : paramsValid
        ? "Attempt"
        : "Invalid attempt",
  );
  const now = useNow({
    enabled: cycleQuery.data?.status === "running" && !cycleQuery.data?.ended_at,
  });

  return {
    taskId,
    cycleId,
    paramsValid,
    activityTab,
    setActivityTab,
    visibleStreamCount,
    setVisibleStreamCount,
    visibleAuditCount,
    setVisibleAuditCount,
    cursorTabId,
    auditTabId,
    cursorPanelId,
    auditPanelId,
    cycleQuery,
    streamQuery,
    auditQuery,
    now,
  };
}
