import { Fragment } from "react";
import type { WorkspaceBrowseRoot } from "@/api/settingsBrowse";
import type { Crumb } from "./workspacePickerPathUtils";

type RootGroupProps = {
  label: string;
  roots: WorkspaceBrowseRoot[];
  listingPending: boolean;
  onOpen: (path: string) => void;
};

export function RootGroup({ label, roots, listingPending, onOpen }: RootGroupProps) {
  return (
    <li className="workspace-picker-root-group">
      <p className="workspace-picker-section-label">{label}</p>
      <ul className="workspace-picker-root-group-list">
        {roots.map((root) => (
          <li key={root.id}>
            <FolderRow
              name={root.label}
              sublabel={root.path}
              disabled={listingPending || !root.available}
              onClick={() => onOpen(root.path)}
            />
            {!root.available && root.unavailable_reason ? (
              <p className="workspace-picker-row-note">{root.unavailable_reason}</p>
            ) : null}
          </li>
        ))}
      </ul>
    </li>
  );
}

type PickerBreadcrumbProps = {
  crumbs: Crumb[];
  listingPending: boolean;
  backToRoots: boolean;
  onBack: () => void;
  onJump: (path: string) => void;
};

export function PickerBreadcrumb({
  crumbs,
  listingPending,
  backToRoots,
  onBack,
  onJump,
}: PickerBreadcrumbProps) {
  return (
    <nav className="workspace-picker-crumbs" aria-label="Folder location">
      <button
        type="button"
        className="workspace-picker-back"
        onClick={onBack}
        disabled={listingPending}
        aria-label={backToRoots ? "Back to starting locations" : "Go up one folder"}
      >
        <BackIcon />
        <span>Back</span>
      </button>
      <ol className="workspace-picker-crumb-path">
        {crumbs.map((crumb, idx) => {
          const isLast = idx === crumbs.length - 1;
          return (
            <Fragment key={crumb.path}>
              {idx > 0 ? (
                <li aria-hidden="true" className="workspace-picker-crumb-sep">
                  /
                </li>
              ) : null}
              <li>
                <button
                  type="button"
                  className="workspace-picker-crumb"
                  onClick={() => onJump(crumb.path)}
                  disabled={isLast || listingPending}
                  aria-current={isLast ? "location" : undefined}
                  title={crumb.path}
                >
                  {crumb.label}
                </button>
              </li>
            </Fragment>
          );
        })}
      </ol>
    </nav>
  );
}

type FolderRowProps = {
  name: string;
  sublabel?: string;
  /** When set, shows git status before the chevron. */
  gitRepoStatus?: boolean;
  disabled?: boolean;
  selected?: boolean;
  onClick: () => void;
  /** When set, chevron opens the folder instead of selecting (git row select vs open). */
  onOpen?: () => void;
};

export function FolderRow({
  name,
  sublabel,
  gitRepoStatus,
  disabled,
  selected,
  onClick,
  onOpen,
}: FolderRowProps) {
  if (onOpen) {
    return (
      <div
        className={
          selected
            ? "workspace-picker-row-split workspace-picker-row-split--selected"
            : "workspace-picker-row-split"
        }
      >
        <button
          type="button"
          className="workspace-picker-row-select"
          onClick={onClick}
          disabled={disabled}
          aria-pressed={selected}
        >
          <FolderIcon />
          <span className="workspace-picker-row-main">
            <span className="workspace-picker-row-name">{name}</span>
            {sublabel ? (
              <span className="workspace-picker-row-sub">{sublabel}</span>
            ) : null}
          </span>
          {gitRepoStatus !== undefined ? (
            <GitRepoStatusIcon isGitRepo={gitRepoStatus} />
          ) : null}
        </button>
        <button
          type="button"
          className="workspace-picker-row-open"
          onClick={onOpen}
          disabled={disabled}
          aria-label={`Open ${name}`}
        >
          <ChevronIcon />
        </button>
      </div>
    );
  }

  return (
    <button
      type="button"
      className={
        selected ? "workspace-picker-row workspace-picker-row--selected" : "workspace-picker-row"
      }
      onClick={onClick}
      disabled={disabled}
      aria-pressed={selected || undefined}
    >
      <FolderIcon />
      <span className="workspace-picker-row-main">
        <span className="workspace-picker-row-name">{name}</span>
        {sublabel ? (
          <span className="workspace-picker-row-sub">{sublabel}</span>
        ) : null}
      </span>
      {gitRepoStatus !== undefined ? (
        <GitRepoStatusIcon isGitRepo={gitRepoStatus} />
      ) : null}
      <ChevronIcon />
    </button>
  );
}

function GitRepoStatusIcon({ isGitRepo }: { isGitRepo: boolean }) {
  return (
    <span
      className={
        isGitRepo
          ? "workspace-picker-git-icon workspace-picker-git-icon--yes"
          : "workspace-picker-git-icon workspace-picker-git-icon--no"
      }
      title={isGitRepo ? "Git repository" : "Not a git repository"}
      aria-label={isGitRepo ? "Git repository" : "Not a git repository"}
    >
      {isGitRepo ? <GitRepoBadge /> : <NoGitRepoIcon />}
    </span>
  );
}

function FolderIcon() {
  return (
    <svg
      className="workspace-picker-row-icon"
      viewBox="0 0 20 20"
      width="18"
      height="18"
      aria-hidden="true"
    >
      <path
        d="M2.75 5.5A1.75 1.75 0 0 1 4.5 3.75h3.13c.46 0 .9.18 1.23.5l1.12 1.06H15.5c.97 0 1.75.78 1.75 1.75v7c0 .97-.78 1.75-1.75 1.75h-11A1.75 1.75 0 0 1 2.75 14V5.5Z"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.4"
        strokeLinejoin="round"
      />
    </svg>
  );
}

function ChevronIcon() {
  return (
    <svg
      className="workspace-picker-row-chevron"
      viewBox="0 0 16 16"
      width="14"
      height="14"
      aria-hidden="true"
    >
      <path
        d="m6 4 4 4-4 4"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.6"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

function BackIcon() {
  return (
    <svg viewBox="0 0 16 16" width="13" height="13" aria-hidden="true">
      <path
        d="M10 3 5 8l5 5"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.6"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

function GitRepoBadge() {
  return <span className="workspace-picker-git-badge">git</span>;
}

function NoGitRepoIcon() {
  return (
    <svg viewBox="0 0 16 16" width="16" height="16" aria-hidden="true">
      <circle
        cx="8"
        cy="8"
        r="5.75"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.25"
      />
      <path
        d="M5.25 5.25 10.75 10.75"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.25"
        strokeLinecap="round"
      />
    </svg>
  );
}
