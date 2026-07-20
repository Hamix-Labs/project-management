import { useCallback, useEffect, useMemo, useState } from "react";
import {
  browseWorkspaceDirs,
  fetchWorkspaceRoots,
  type BrowseDirEntry,
  type WorkspaceBrowseRoot,
  type WorkspaceRootsScope,
} from "@/api/settingsBrowse";
import {
  computeCrumbs,
  isBrowseRootPath,
  partitionBrowseRoots,
} from "./workspacePickerPathUtils";

export type WorkspaceDirPickerModalProps = {
  open: boolean;
  onClose: () => void;
  onSelect: (path: string) => void;
  currentPath: string;
  /** Opens above another modal (worktrees register flow). */
  nested?: boolean;
  title?: string;
  lead?: string;
  /** When true, only git checkouts can be confirmed; folders stay visible. */
  requireGitRepository?: boolean;
  /** When set, path must pass async validation before confirm (e.g. repo-scoped worktree probe). */
  validatePath?: (path: string) => Promise<{ ok: boolean; message?: string }>;
  /** Skip the roots screen and list this directory when the modal opens. */
  initialBrowsePath?: string;
  /** When expanded, workspace-roots includes OS bootstrap places alongside registered repos. */
  rootsScope?: WorkspaceRootsScope;
  /** Footer label for the path being confirmed. */
  selectionFooterLabel?: string;
  /** Confirm button label. */
  confirmLabel?: string;
};

type LoadState =
  | { kind: "idle" }
  | { kind: "loading" }
  | { kind: "ready"; roots: WorkspaceBrowseRoot[]; environment: "native" }
  | { kind: "error"; message: string };

type PickerArgs = Pick<
  WorkspaceDirPickerModalProps,
  | "open"
  | "onClose"
  | "onSelect"
  | "requireGitRepository"
  | "validatePath"
  | "initialBrowsePath"
  | "rootsScope"
  | "selectionFooterLabel"
  | "confirmLabel"
  | "lead"
>;

export function useWorkspaceDirPickerState({
  open,
  onClose,
  onSelect,
  requireGitRepository = false,
  validatePath,
  initialBrowsePath,
  rootsScope = "default",
  selectionFooterLabel,
  confirmLabel,
  lead,
}: PickerArgs) {
  const resolvedLead =
    lead ??
    (requireGitRepository
      ? "Navigate to a git repository checkout. Hamix needs a .git folder at the path you register."
      : "Open a folder to browse inside it. Confirm the folder you're in to register it.");

  const [loadState, setLoadState] = useState<LoadState>({ kind: "idle" });
  const [entries, setEntries] = useState<BrowseDirEntry[]>([]);
  const [currentBrowsePath, setCurrentBrowsePath] = useState("");
  const [parentPath, setParentPath] = useState("");
  const [currentPathIsGitRepo, setCurrentPathIsGitRepo] = useState(false);
  const [listingError, setListingError] = useState<string | null>(null);
  const [listingPending, setListingPending] = useState(false);
  const [pathValidation, setPathValidation] = useState<{ ok: boolean; message?: string } | null>(
    null,
  );
  const [validatingPath, setValidatingPath] = useState(false);

  const atRoots = currentBrowsePath.trim() === "";

  const loadListing = useCallback(async (path: string) => {
    setListingPending(true);
    setListingError(null);
    try {
      const listing = await browseWorkspaceDirs(path);
      setEntries(listing.entries);
      setCurrentBrowsePath(listing.path ?? path);
      setParentPath(listing.parent_path ?? "");
      setCurrentPathIsGitRepo(listing.is_git_repo === true);
    } catch (err) {
      setListingError(err instanceof Error ? err.message : "Could not list folders");
      setEntries([]);
    } finally {
      setListingPending(false);
    }
  }, []);

  useEffect(() => {
    if (!open) return;
    let cancelled = false;
    setLoadState({ kind: "loading" });
    setEntries([]);
    setCurrentBrowsePath("");
    setParentPath("");
    setCurrentPathIsGitRepo(false);
    setListingError(null);
    setPathValidation(null);
    setValidatingPath(false);
    void fetchWorkspaceRoots({ scope: rootsScope })
      .then((roots) => {
        if (cancelled) return;
        setLoadState({
          kind: "ready",
          roots: roots.roots,
          environment: roots.environment,
        });
        const startPath = initialBrowsePath?.trim() ?? "";
        if (startPath !== "") {
          void loadListing(startPath);
        }
      })
      .catch((err) => {
        if (cancelled) return;
        setLoadState({
          kind: "error",
          message: err instanceof Error ? err.message : "Could not load locations",
        });
      });
    return () => {
      cancelled = true;
    };
  }, [open, initialBrowsePath, loadListing, rootsScope]);

  const crumbs = useMemo(() => {
    if (loadState.kind !== "ready") return [];
    return computeCrumbs(loadState.roots, currentBrowsePath);
  }, [loadState, currentBrowsePath]);

  useEffect(() => {
    if (!open || !validatePath || currentBrowsePath.trim() === "") {
      setPathValidation(null);
      setValidatingPath(false);
      return;
    }
    let cancelled = false;
    setValidatingPath(true);
    void validatePath(currentBrowsePath)
      .then((result) => {
        if (!cancelled) {
          setPathValidation(result);
          setValidatingPath(false);
        }
      })
      .catch((err) => {
        if (!cancelled) {
          setPathValidation({
            ok: false,
            message: err instanceof Error ? err.message : "Could not validate folder",
          });
          setValidatingPath(false);
        }
      });
    return () => {
      cancelled = true;
    };
  }, [open, validatePath, currentBrowsePath]);

  function goRoots() {
    setEntries([]);
    setCurrentBrowsePath("");
    setParentPath("");
    setCurrentPathIsGitRepo(false);
    setListingError(null);
  }

  function goBack() {
    if (atRoots || listingPending) return;
    // Entered from the starting-locations screen — Back returns there, not
    // to the filesystem parent (e.g. Documents → OneDrive → Users).
    if (loadState.kind === "ready" && isBrowseRootPath(loadState.roots, currentBrowsePath)) {
      goRoots();
      return;
    }
    if (parentPath.trim() === "") {
      goRoots();
      return;
    }
    void loadListing(parentPath);
  }

  function confirmSelection() {
    if (atRoots || listingPending || currentBrowsePath.trim() === "") return;
    if (requireGitRepository && !currentPathIsGitRepo) return;
    onSelect(currentBrowsePath);
    onClose();
  }

  const selectionLabel =
    selectionFooterLabel ??
    (requireGitRepository ? "Repository checkout" : "Folder to register");
  const confirmButtonLabel = confirmLabel ?? "Use this folder";
  const hasOpenFolder = !atRoots && currentBrowsePath.trim() !== "";
  const gitRequirementMet = !requireGitRepository || currentPathIsGitRepo;
  const customValidationMet = !validatePath || (pathValidation?.ok === true && !validatingPath);
  const canConfirm = hasOpenFolder && !listingPending && gitRequirementMet && customValidationMet;

  const rootGroups =
    loadState.kind === "ready" ? partitionBrowseRoots(loadState.roots) : null;

  const backToRoots =
    loadState.kind === "ready" && isBrowseRootPath(loadState.roots, currentBrowsePath);

  return {
    resolvedLead,
    loadState,
    entries,
    currentBrowsePath,
    currentPathIsGitRepo,
    listingError,
    listingPending,
    pathValidation,
    validatingPath,
    atRoots,
    crumbs,
    rootGroups,
    backToRoots,
    goBack,
    goRoots,
    loadListing,
    confirmSelection,
    selectionLabel,
    confirmButtonLabel,
    hasOpenFolder,
    canConfirm,
    requireGitRepository,
  };
}
