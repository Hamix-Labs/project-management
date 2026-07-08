import { GitContextMeta } from "../commits/GitContextMeta";
import { useTaskGitBinding } from "@/tasks/hooks/useTaskGitBinding";

type Props = {
  worktreeId?: string;
  projectId?: string;
};

export function TaskDetailGitBinding({ worktreeId, projectId }: Props) {
  const wtId = (worktreeId ?? "").trim();
  const bindingQuery = useTaskGitBinding(worktreeId, projectId);

  if (wtId === "" || bindingQuery.isLoading || !bindingQuery.data) {
    return null;
  }

  return (
    <div className="task-detail-git-binding" data-testid="task-detail-git-binding">
      <GitContextMeta context={bindingQuery.data} />
    </div>
  );
}
