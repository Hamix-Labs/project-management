import { Link } from "react-router-dom";
import { TaskCreateAssignmentFields } from "../../task-create-modal/fields/TaskCreateAssignmentFields";
import { ComposeRailSectionTitle } from "../TaskComposeBriefCard";

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
      <ComposeRailSectionTitle icon={<FolderGit2Icon />}>
        Destination
      </ComposeRailSectionTitle>
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
          <p className="compose-destination__hint">
            <GitBranchIcon />
            <span>
              {worktreeId.trim() !== "" && (disabled || assignmentLocked)
                ? "This task will reuse the selected worktree and branch. "
                : "A worktree and branch are allocated when the task is created. "}
              <Link to="/repositories" target="_blank" rel="noopener noreferrer">
                Inspect repositories
              </Link>
            </span>
          </p>
        </div>
      ) : (
        <p className="compose-destination__hint">
          <GitBranchIcon />
          <span>Repository binding is fixed for this edit.</span>
        </p>
      )}
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

function GitBranchIcon() {
  return (
    <svg
      className="compose-destination__hint-icon"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <line x1="6" y1="3" x2="6" y2="15" />
      <circle cx="18" cy="6" r="3" />
      <circle cx="6" cy="18" r="3" />
      <path d="M18 9a9 9 0 0 1-9 9" />
    </svg>
  );
}
