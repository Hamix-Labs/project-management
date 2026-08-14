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
    <section className="compose-card compose-card--padded compose-destination">
      <ComposeRailSectionTitle icon={<FolderGitIcon />}>
        Destination
      </ComposeRailSectionTitle>
      {showWorktree ? (
        <div className="compose-destination__fields">
          <TaskCreateAssignmentFields
            idsPrefix={idsPrefix}
            repositoryId={repositoryId}
            projectId={projectId}
            worktreeId={worktreeId}
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
        <p className="compose-destination__hint" style={{ marginTop: "1rem" }}>
          <GitBranchIcon />
          <span>Repository binding is fixed for this edit.</span>
        </p>
      )}
    </section>
  );
}

function FolderGitIcon() {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <path d="M4 20h16a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.93a2 2 0 0 1-1.66-.9l-.82-1.2A2 2 0 0 0 7.93 3H4a2 2 0 0 0-2 2v13c0 1.1.9 2 2 2Z" />
      <circle cx="12" cy="13" r="2" />
      <path d="M14 13h4" />
    </svg>
  );
}

function GitBranchIcon() {
  return (
    <svg
      width="14"
      height="14"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
      style={{ flexShrink: 0, marginTop: "0.125rem" }}
    >
      <line x1="6" y1="3" x2="6" y2="15" />
      <circle cx="18" cy="6" r="3" />
      <circle cx="6" cy="18" r="3" />
      <path d="M18 9a9 9 0 0 1-9 9" />
    </svg>
  );
}
