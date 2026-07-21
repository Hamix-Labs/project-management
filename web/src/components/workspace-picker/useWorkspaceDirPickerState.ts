import { useCallback, useEffect, useMemo, useState } from "react";
import {
  browseWorkspaceDirs,
  fetchWorkspaceRoots,
  probeGitRepository,
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

type ResolvedGitSelection = {
  probedPath: string;
  mainPath: string;
  isMain: boolean;
};

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
      ? "Select any worktree of a Git repository. Repositories are identified by their primary checkout, so all linked worktrees count as one."
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
  const [resolvedSelection, setResolvedSelection] = useState<ResolvedGitSelection | null>(null);
  const [probePending, setProbePending] = useState(false);
  const [probeError, setProbeError] = useState<string | null>(null);
  const [folderFilter, setFolderFilter] = useState("");

  const atRoots = currentBrowsePath.trim() === "";

  const clearResolvedSelection = useCallback(() => {
    setResolvedSelection(null);
    setProbeError(null);
    setProbePending(false);
  }, []);

  const resolveGitSelection = useCallback(async (path: string) => {
    const trimmed = path.trim();
    if (trimmed === "") {
      clearResolvedSelection();
      return;
    }
    setProbePending(true);
    setProbeError(null);
    try {
      const probe = await probeGitRepository(trimmed);
      if (!probe.is_git_repository) {
        setResolvedSelection(null);
        setProbeError("This folder is not a git repository.");
        return;
      }
      const mainPath = (probe.main_path?.trim() || probe.path).trim();
      setResolvedSelection({
        probedPath: probe.path,
        mainPath,
        isMain: probe.is_main === true,
      });
    } catch (err) {
      setResolvedSelection(null);
      setProbeError(err instanceof Error ? err.message : "Could not resolve repository");
    } finally {
      setProbePending(false);
    }
  }, [clearResolvedSelection]);

  const loadListing = useCallback(
    async (path: string) => {
      setListingPending(true);
      setListingError(null);
      setFolderFilter("");
      if (requireGitRepository) {
        clearResolvedSelection();
      }
      try {
        const listing = await browseWorkspaceDirs(path);
        setEntries(listing.entries);
        const listedPath = listing.path ?? path;
        setCurrentBrowsePath(listedPath);
        setParentPath(listing.parent_path ?? "");
        setCurrentPathIsGitRepo(listing.is_git_repo === true);
        if (requireGitRepository && listing.is_git_repo === true) {
          void resolveGitSelection(listedPath);
        }
      } catch (err) {
        setListingError(err instanceof Error ? err.message : "Could not list folders");
        setEntries([]);
      } finally {
        setListingPending(false);
      }
    },
    [clearResolvedSelection, requireGitRepository, resolveGitSelection],
  );

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
    setFolderFilter("");
    clearResolvedSelection();
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
  }, [open, initialBrowsePath, loadListing, rootsScope, clearResolvedSelection]);

  const crumbs = useMemo(() => {
    if (loadState.kind !== "ready") return [];
    return computeCrumbs(loadState.roots, currentBrowsePath);
  }, [loadState, currentBrowsePath]);

  const pathToValidate = requireGitRepository
    ? (resolvedSelection?.mainPath ?? "")
    : currentBrowsePath;

  useEffect(() => {
    if (!open || !validatePath || pathToValidate.trim() === "") {
      setPathValidation(null);
      setValidatingPath(false);
      return;
    }
    let cancelled = false;
    setValidatingPath(true);
    void validatePath(pathToValidate)
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
  }, [open, validatePath, pathToValidate]);

  function goRoots() {
    setEntries([]);
    setCurrentBrowsePath("");
    setParentPath("");
    setCurrentPathIsGitRepo(false);
    setListingError(null);
    setFolderFilter("");
    if (requireGitRepository) {
      clearResolvedSelection();
    }
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
    if (listingPending || probePending) return;
    if (requireGitRepository) {
      if (!resolvedSelection) return;
      onSelect(resolvedSelection.mainPath);
      onClose();
      return;
    }
    if (atRoots || currentBrowsePath.trim() === "") return;
    onSelect(currentBrowsePath);
    onClose();
  }

  const selectionLabel =
    selectionFooterLabel ??
    (requireGitRepository ? "Repository to register" : "Folder to register");
  const confirmButtonLabel =
    confirmLabel ?? (requireGitRepository ? "Use this repository" : "Use this folder");
  const hasOpenFolder = !atRoots && currentBrowsePath.trim() !== "";
  const hasResolvedRepo = resolvedSelection != null && resolvedSelection.mainPath.trim() !== "";
  const remapped =
    resolvedSelection != null &&
    !resolvedSelection.isMain &&
    resolvedSelection.probedPath.trim() !== "" &&
    resolvedSelection.probedPath !== resolvedSelection.mainPath;
  const footerPath = requireGitRepository
    ? (resolvedSelection?.mainPath ?? "")
    : currentBrowsePath;
  const footerEmptyHint = requireGitRepository
    ? "Select a repository above to continue."
    : "Open a folder to register it";
  const customValidationMet = !validatePath || (pathValidation?.ok === true && !validatingPath);
  const canConfirm = requireGitRepository
    ? hasResolvedRepo && !listingPending && !probePending && customValidationMet
    : hasOpenFolder && !listingPending && customValidationMet;

  const filterQuery = folderFilter.trim().toLowerCase();

  const filteredEntries = useMemo(() => {
    if (filterQuery === "") return entries;
    return entries.filter((entry) => entry.name.toLowerCase().includes(filterQuery));
  }, [entries, filterQuery]);

  const rootGroups = useMemo(() => {
    if (loadState.kind !== "ready") return null;
    const partitioned = partitionBrowseRoots(loadState.roots);
    if (filterQuery === "") return partitioned;
    const matchRoot = (root: WorkspaceBrowseRoot) =>
      root.label.toLowerCase().includes(filterQuery) ||
      root.path.toLowerCase().includes(filterQuery);
    return {
      workspace: partitioned.workspace.filter(matchRoot),
      userFolders: partitioned.userFolders.filter(matchRoot),
    };
  }, [loadState, filterQuery]);

  const filterActive = filterQuery !== "";
  const filterEmpty =
    filterActive &&
    ((atRoots &&
      (rootGroups?.workspace.length ?? 0) === 0 &&
      (rootGroups?.userFolders.length ?? 0) === 0) ||
      (!atRoots && filteredEntries.length === 0));

  const backToRoots =
    loadState.kind === "ready" && isBrowseRootPath(loadState.roots, currentBrowsePath);

  return {
    resolvedLead,
    loadState,
    entries: filteredEntries,
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
    resolveGitSelection,
    confirmSelection,
    selectionLabel,
    confirmButtonLabel,
    hasOpenFolder,
    hasResolvedRepo,
    remapped,
    probedPath: resolvedSelection?.probedPath ?? "",
    footerPath,
    footerEmptyHint,
    probePending,
    probeError,
    canConfirm,
    requireGitRepository,
    folderFilter,
    setFolderFilter,
    filterActive,
    filterEmpty,
  };
}
