import { TaskCreateAssignmentFields } from "../../task-create-modal/fields/TaskCreateAssignmentFields";

type Props = {
  idsPrefix: string;
  repositoryId: string;
  projectId: string;
  worktreeId: string;
  disabled: boolean;
  assignmentLocked?: boolean;
  showWorktree: boolean;
  onRepositoryChange: (repositoryId: string) => void;
  onProjectChange: (projectId: string) => void;
  onWorktreeChange: (worktreeId: string) => void;
};

export function TaskComposeDestinationCard({
  idsPrefix,
  repositoryId,
  projectId,
  worktreeId,
  disabled,
  assignmentLocked = false,
  showWorktree,
  onRepositoryChange,
  onProjectChange,
  onWorktreeChange,
}: Props) {
  return (
    <section className="compose-handoff__section compose-destination">
      <h2 className="compose-handoff__title">Destination</h2>
      {showWorktree ? (
        <div className="compose-destination__fields">
          <TaskCreateAssignmentFields
            idsPrefix={idsPrefix}
            repositoryId={repositoryId}
            projectId={projectId}
            worktreeId={worktreeId}
            repositoryLeadingIcon={
              <FolderGit2Icon className="worktrees-git-selector__icon" />
            }
            onAssignmentChange={(next) => {
              if (next.repositoryId !== repositoryId) {
                onRepositoryChange(next.repositoryId);
              }
              if (next.projectId !== projectId) {
                onProjectChange(next.projectId);
              }
              if (next.worktreeId !== worktreeId) {
                onWorktreeChange(next.worktreeId);
              }
            }}
            disabled={disabled || assignmentLocked}
          />
        </div>
      ) : null}
    </section>
  );
}

/** Lucide FolderGit2 — folder with a branch node, not a keyhole or plus. */
function FolderGit2Icon({ className }: { className?: string }) {
  return (
    <svg
      className={className}
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <path d="M9 20H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h3.9a2 2 0 0 1 1.69.9l.81 1.2a2 2 0 0 0 1.67.9H20a2 2 0 0 1 2 2v5" />
      <circle cx="13" cy="12" r="2" />
      <path d="M18 19c-2.8 0-5-2.2-5-5v8" />
      <circle cx="20" cy="19" r="2" />
    </svg>
  );
}
