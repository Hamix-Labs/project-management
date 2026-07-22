import { GitContextMeta } from "../commits/GitContextMeta";
import { useCopyToClipboard } from "../commits/useCopyToClipboard";
import { useTaskGitBinding } from "@/tasks/hooks/useTaskGitBinding";
import { OpenInEditorMenu } from "./OpenInEditorMenu";

type Props = {
  worktreeId?: string;
  projectId?: string;
};

export function TaskDetailGitBinding({ worktreeId, projectId }: Props) {
  const wtId = (worktreeId ?? "").trim();
  const bindingQuery = useTaskGitBinding(worktreeId, projectId);
  const pathCopy = useCopyToClipboard("Copy path");

  if (wtId === "" || bindingQuery.isLoading || !bindingQuery.data) {
    return null;
  }

  const context = bindingQuery.data;
  const openPath = (context.openPath ?? context.worktree).trim();

  return (
    <div className="task-detail-git-binding" data-testid="task-detail-git-binding">
      <GitContextMeta context={context} />
      {openPath !== "" ? (
        <div
          className="task-detail-git-binding-actions"
          data-testid="task-detail-git-binding-actions"
        >
          <button
            type="button"
            className="btn-utility task-detail-git-binding-copy"
            onClick={() => pathCopy.copy(openPath)}
            aria-label={
              pathCopy.copied ? "Copied worktree path" : "Copy worktree path"
            }
          >
            {pathCopy.copyLabel}
          </button>
          <OpenInEditorMenu openPath={openPath} />
        </div>
      ) : null}
    </div>
  );
}
