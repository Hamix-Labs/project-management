import { GitContextMeta } from "../commits/GitContextMeta";
import { useCopyToClipboard } from "../commits/useCopyToClipboard";
import { useTaskGitBinding } from "@/tasks/hooks/useTaskGitBinding";
import { useGlobalRepositories } from "@/hooks/useGlobalRepositories";
import { useProject } from "@/hooks/useProject";
import {
  predictedTaskWorktreeName,
  taskBranchName,
} from "@/tasks/task-git/taskWorktreeNames";
import {
  TaskDetailCheckGlyph,
  TaskDetailCopyGlyph,
} from "./TaskDetailActionGlyphs";
import { OpenInEditorMenu } from "./OpenInEditorMenu";

type Props = {
  taskId: string;
  worktreeId?: string;
  projectId?: string;
  repositoryId?: string;
};

function GitBindingActions({ openPath }: { openPath: string }) {
  const pathCopy = useCopyToClipboard("Copy path");
  const hasPath = openPath.trim() !== "";
  if (!hasPath) {
    return null;
  }
  return (
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
        {pathCopy.copied ? (
          <TaskDetailCheckGlyph className="task-detail-action-glyph" />
        ) : (
          <TaskDetailCopyGlyph className="task-detail-action-glyph" />
        )}
        {pathCopy.copyLabel}
      </button>
      <OpenInEditorMenu openPath={openPath} />
    </div>
  );
}

export function TaskDetailGitBinding({
  taskId,
  worktreeId,
  projectId,
  repositoryId,
}: Props) {
  const wtId = (worktreeId ?? "").trim();
  const bindingQuery = useTaskGitBinding(worktreeId);

  const projectKey = (projectId ?? "").trim();
  const projectQuery = useProject(projectKey, {
    enabled: wtId === "" && projectKey !== "",
  });
  const repositoriesQuery = useGlobalRepositories({
    enabled:
      wtId === "" &&
      (projectKey !== "" || (repositoryId ?? "").trim() !== ""),
  });

  if (wtId !== "") {
    if (bindingQuery.isLoading || !bindingQuery.data) {
      return null;
    }
    const context = bindingQuery.data;
    const openPath = (context.openPath ?? context.worktree).trim();

    return (
      <div
        className="task-detail-git-binding"
        data-testid="task-detail-git-binding"
      >
        <GitContextMeta context={context} />
        <GitBindingActions openPath={openPath} />
      </div>
    );
  }

  const tid = taskId.trim();
  if (tid === "") {
    return null;
  }

  const branch = taskBranchName(tid);
  const worktreeName = predictedTaskWorktreeName(tid);
  const repoId =
    (repositoryId ?? "").trim() ||
    projectQuery.data?.repository_id?.trim() ||
    "";
  const repoPath =
    repositoriesQuery.data?.find((r) => r.id === repoId)?.path?.trim() ?? "";

  return (
    <div
      className="task-detail-git-binding"
      data-testid="task-detail-git-binding"
      data-provisioning="true"
    >
      <GitContextMeta
        context={{
          branch,
          worktree: worktreeName,
          repo: repoPath,
        }}
      />
      <p
        className="task-detail-git-binding-pending"
        data-testid="task-detail-git-binding-pending"
      >
        Preparing workspace…
      </p>
    </div>
  );
}
