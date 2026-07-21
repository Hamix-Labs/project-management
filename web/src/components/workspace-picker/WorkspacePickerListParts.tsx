import { Fragment } from "react";
import type { WorkspaceBrowseRoot } from "@/api/settingsBrowse";
import type { Crumb } from "./workspacePickerPathUtils";
import {
  PickerCheckIcon,
  PickerChevronLeftIcon,
  PickerChevronRightIcon,
  PickerFolderGitIcon,
  PickerFolderIcon,
  PickerGitBranchIcon,
} from "./WorkspacePickerIcons";

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
        <PickerChevronLeftIcon />
        <span>Back</span>
      </button>
      <ol className="workspace-picker-crumb-path">
        {crumbs.map((crumb, idx) => {
          const isLast = idx === crumbs.length - 1;
          return (
            <Fragment key={crumb.path}>
              {idx > 0 ? (
                <li aria-hidden="true" className="workspace-picker-crumb-sep">
                  <PickerChevronRightIcon />
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
  /** When set, shows git status chrome (subtitle + badge). */
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
  const resolvedSublabel =
    sublabel ??
    (gitRepoStatus === true
      ? "Git repository"
      : gitRepoStatus === false
        ? "Not a repository"
        : undefined);

  const body = (
    <>
      <span
        className={
          selected
            ? "workspace-picker-row-tile workspace-picker-row-tile--selected"
            : gitRepoStatus === true
              ? "workspace-picker-row-tile workspace-picker-row-tile--git"
              : "workspace-picker-row-tile"
        }
      >
        {selected ? (
          <PickerCheckIcon />
        ) : gitRepoStatus === true ? (
          <PickerFolderGitIcon />
        ) : (
          <PickerFolderIcon />
        )}
      </span>
      <span className="workspace-picker-row-main">
        <span className="workspace-picker-row-name">{name}</span>
        {resolvedSublabel ? (
          <span className="workspace-picker-row-sub">{resolvedSublabel}</span>
        ) : null}
      </span>
      {gitRepoStatus === true ? (
        <span className="workspace-picker-git-badge" aria-label="Git repository">
          <PickerGitBranchIcon />
          git
        </span>
      ) : null}
    </>
  );

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
          {body}
        </button>
        <button
          type="button"
          className="workspace-picker-row-open"
          onClick={onOpen}
          disabled={disabled}
          aria-label={`Open ${name}`}
        >
          <PickerChevronRightIcon className="workspace-picker-row-chevron" />
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
      {body}
      <PickerChevronRightIcon className="workspace-picker-row-chevron" />
    </button>
  );
}
