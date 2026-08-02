import { useQuery } from "@tanstack/react-query";
import { listTasks } from "@/api/tasks";
import { ConfirmDialog } from "@/components/feedback/ConfirmDialog";

type Props = {
  saving: boolean;
  pending: boolean;
  error?: string | null;
  worktreeId?: string | null;
  onCancel: () => void;
  onConfirm: () => void;
};

export function TaskOpenPRConfirmDialog({
  saving,
  pending,
  error = null,
  worktreeId,
  onCancel,
  onConfirm,
}: Props) {
  const wt = worktreeId?.trim() ?? "";
  const familyQuery = useQuery({
    queryKey: ["tasks", "worktree-family", wt],
    queryFn: ({ signal }) => listTasks(50, 0, { signal, worktreeId: wt }),
    enabled: wt.length > 0,
  });
  const familySize = familyQuery.data?.tasks?.length ?? 1;
  const publishesStack = familySize > 1;
  const description = publishesStack
    ? `This worktree has ${familySize} tasks. Approving publishes the entire GitHub stack (every layer PR), not only this task.`
    : "Resumes the same conversation to publish the GitHub stack (or open the PR).";

  return (
    <ConfirmDialog
      title="Approve and Open a pull request?"
      description={description}
      confirmLabel="Approve & Open PR"
      confirmVariant="primary"
      busy={pending}
      cancelDisabled={saving}
      confirmDisabled={saving}
      error={error}
      onCancel={onCancel}
      onConfirm={onConfirm}
      titleId="task-open-pr-dialog-title"
      descriptionId="task-open-pr-dialog-description"
    />
  );
}
