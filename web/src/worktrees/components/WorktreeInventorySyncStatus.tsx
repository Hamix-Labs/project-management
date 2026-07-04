import { worktreeGitCopy } from "../worktreeGitCopy";
import { WorktreeReconcileStatus } from "./WorktreeReconcileStatus";

type Props = {
  pending?: boolean;
};

export function WorktreeInventorySyncStatus({ pending = false }: Props) {
  if (!pending) return null;
  return (
    <WorktreeReconcileStatus
      className="worktrees-form-modal__inventory-sync"
      message={worktreeGitCopy.inventoryRefreshStatus}
    />
  );
}
