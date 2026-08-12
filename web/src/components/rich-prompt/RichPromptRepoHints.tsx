type Props = {
  showSelectWorktreeHint: boolean;
  showSelectRepositoryHint?: boolean;
  showRepoMisconfigHint: boolean;
  workspaceBroken: boolean;
  fileSearchFailedWhileAvailable: boolean;
  showRepoUnknownHint: boolean;
  showFileSearchSpinner: boolean;
};

/** Status copy under the rich prompt when repo / @ search is misconfigured or busy. */
export function RichPromptRepoHints({
  showSelectWorktreeHint,
  showSelectRepositoryHint = false,
  showRepoMisconfigHint,
  workspaceBroken,
  fileSearchFailedWhileAvailable,
  showRepoUnknownHint,
  showFileSearchSpinner,
}: Props) {
  return (
    <>
      {showSelectRepositoryHint ? (
        <p className="mention-repo-hint" role="status">
          Select a repository above to enable <code>@file</code> mentions.
        </p>
      ) : null}
      {showSelectWorktreeHint ? (
        <p className="mention-repo-hint" role="status">
          Select a worktree above to enable <code>@file</code> mentions.
        </p>
      ) : null}
      {showRepoMisconfigHint ? (
        <p className="mention-repo-hint" role="status">
          {workspaceBroken ? (
            <>
              The selected worktree path is missing or not a directory.
              Update the repository on the{" "}
              <a href="/repositories" target="_blank" rel="noopener noreferrer">
                Repositories page
              </a>{" "}
              to restore <code>@file</code> mentions.
            </>
          ) : fileSearchFailedWhileAvailable ? (
            <>
              File search failed for the selected worktree. Check the repository
              on the{" "}
              <a href="/repositories" target="_blank" rel="noopener noreferrer">
                Repositories page
              </a>{" "}
              or inspect the server logs.
            </>
          ) : (
            <>
              File search is not available for the selected worktree. Check the
              repository on the{" "}
              <a href="/repositories" target="_blank" rel="noopener noreferrer">
                Repositories page
              </a>{" "}
              to enable <code>@file</code> mentions.
            </>
          )}
        </p>
      ) : null}
      {showRepoUnknownHint ? (
        <p className="mention-repo-hint" role="status">
          Could not verify file search for the selected worktree. Check the
          repository on the{" "}
          <a href="/repositories" target="_blank" rel="noopener noreferrer">
            Repositories page
          </a>
          .
        </p>
      ) : null}
      {showFileSearchSpinner ? (
        <p
          className="mention-repo-hint mention-repo-hint--pending"
          role="status"
          aria-live="polite"
        >
          Searching files…
        </p>
      ) : null}
    </>
  );
}
