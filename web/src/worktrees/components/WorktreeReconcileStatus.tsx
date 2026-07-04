import { worktreeGitCopy } from "../worktreeGitCopy";

type Props = {
  className?: string;
  message?: string;
};

export function WorktreeReconcileStatus({ className, message }: Props) {
  return (
    <div
      className={["worktrees-reconcile-status", className].filter(Boolean).join(" ")}
      role="status"
      aria-live="polite"
    >
      <span className="worktrees-reconcile-status__spinner" aria-hidden />
      <span>{message ?? worktreeGitCopy.reconcilingStatus}</span>
    </div>
  );
}
