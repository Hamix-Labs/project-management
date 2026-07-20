import { useQuery } from "@tanstack/react-query";
import { listChecklist } from "@/api";
import type { Status } from "@/types";
import { canMutateTaskCriteria } from "../../../task-display/canMutateTaskCriteria";
import { useTaskDetailChecklist } from "../../../checklist/hooks/useTaskDetailChecklist";
import { taskQueryKeys } from "../../../task-query";
import { QUERY_POLICY } from "../../../queryPolicy";
import { TaskDetailChecklistSection } from "./TaskDetailChecklistSection";

type Props = {
  taskId: string;
  saving: boolean;
  taskStatus: Status;
};

export function TaskDetailChecklistContainer({
  taskId,
  saving,
  taskStatus,
}: Props) {
  const checklist = useTaskDetailChecklist(taskId);
  const checklistQuery = useQuery({
    queryKey: taskQueryKeys.checklist(taskId),
    queryFn: ({ signal }) => listChecklist(taskId, { signal }),
    enabled: Boolean(taskId),
    staleTime: QUERY_POLICY.detailStaleTimeMs,
  });

  const items = checklistQuery.data?.items ?? [];
  const doneCount = items.filter((item) => item.done).length;

  return (
    <TaskDetailChecklistSection
      saving={saving}
      taskStatus={taskStatus}
      canAddCriterion={canMutateTaskCriteria(taskStatus)}
      checklistQuery={checklistQuery}
      doneCount={doneCount}
      totalCount={items.length}
      checklist={checklist}
    />
  );
}
