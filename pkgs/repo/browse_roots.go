package repo

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlexsanderHamir/Hamix/pkgs/gitinventory/domain"
)

const (
	browseRootInstall = "install"
	browseRootHome    = "home"

	// WorkspacePickerScopeDefault returns registered-repo roots in manage mode.
	WorkspacePickerScopeDefault = "default"
	// WorkspacePickerScopeExpanded merges OS bootstrap places with registered roots.
	WorkspacePickerScopeExpanded = "expanded"
)

// BrowseRoot is a top-level directory the workspace picker may start from.
type BrowseRoot struct {
	ID                string        `json:"id"`
	Path              string        `json:"path"`
	Label             string        `json:"label"`
	Category          PlaceCategory `json:"category,omitempty"`
	Available         bool          `json:"available"`
	UnavailableReason string        `json:"unavailable_reason,omitempty"`
}

// BrowseEnvironment is where taskapi runs. Always native on the host OS.
type BrowseEnvironment string

const BrowseEnvNative BrowseEnvironment = "native"

// DetectBrowseEnvironment reports the workspace picker environment.
func DetectBrowseEnvironment() BrowseEnvironment {
	slog.Debug("trace", "operation", "repo.DetectBrowseEnvironment")
	return BrowseEnvNative
}

// ResolveBrowseRoots returns the HAMIX_BROWSE_ROOTS custom paths when the env var is set.
// When HAMIX_BROWSE_ROOTS is not configured it returns an empty slice — workspace roots
// are now sourced from the git_repositories table (Cycle 7).
func ResolveBrowseRoots(startDir string) ([]BrowseRoot, BrowseEnvironment, error) {
	slog.Debug("trace", "operation", "repo.ResolveBrowseRoots")
	env := DetectBrowseEnvironment()
	if !CustomBrowseRootsConfigured() {
		return nil, env, nil
	}
	reg := NewPlaceRegistry(CustomPlaceProvider{})
	places, err := reg.Places(env, startDir)
	if err != nil {
		return nil, env, err
	}
	roots := make([]BrowseRoot, 0, len(places))
	for _, p := range places {
		roots = append(roots, placeToBrowseRoot(p))
	}
	return roots, env, nil
}

// defaultBootstrapPlaceRegistry composes OS place providers for first-time repository
// registration when no git repositories are registered yet.
//
//funclogmeasure:skip category=hot-path reason="Pure constructor; operation trace is emitted by ResolveWorkspacePickerRoots."
func defaultBootstrapPlaceRegistry() *PlaceRegistry {
	return NewPlaceRegistry(
		InstallPlaceProvider{},
		HomePlaceProvider{},
		UserDirsPlaceProvider{},
	)
}

// ResolveWorkspacePickerRoots is the single policy for GET /settings/workspace-roots.
// Custom browse roots override everything; registered repos drive manage mode; otherwise
// bootstrap OS entry points enable the register-repo folder picker.
// When scope is WorkspacePickerScopeExpanded and repos are registered, bootstrap places
// are merged so operators can browse outside registered checkout paths.
func ResolveWorkspacePickerRoots(startDir string, registered []domain.GitRepository, scope string) ([]BrowseRoot, BrowseEnvironment, error) {
	slog.Debug("trace", "operation", "repo.ResolveWorkspacePickerRoots")
	if CustomBrowseRootsConfigured() {
		return ResolveBrowseRoots(startDir)
	}
	env := DetectBrowseEnvironment()
	if len(registered) > 0 {
		roots := make([]BrowseRoot, 0, len(registered))
		for _, gr := range registered {
			roots = append(roots, BrowseRootFromPath(gr.ID, gr.Path, PlaceCategoryRegistered))
		}
		if allBrowseRootsUnavailable(roots) {
			reg := defaultBootstrapPlaceRegistry()
			places, err := reg.Places(env, startDir)
			if err != nil {
				return nil, env, err
			}
			for _, p := range places {
				roots = append(roots, placeToBrowseRoot(p))
			}
		} else if scope == WorkspacePickerScopeExpanded {
			merged, err := mergeBootstrapBrowseRoots(roots, env, startDir)
			if err != nil {
				return nil, env, err
			}
			roots = merged
		}
		return roots, env, nil
	}
	reg := defaultBootstrapPlaceRegistry()
	places, err := reg.Places(env, startDir)
	if err != nil {
		return nil, env, err
	}
	roots := make([]BrowseRoot, 0, len(places))
	for _, p := range places {
		roots = append(roots, placeToBrowseRoot(p))
	}
	return roots, env, nil
}

//funclogmeasure:skip category=hot-path reason="Browse sub-step; operation trace is emitted by ResolveWorkspacePickerRoots."
func mergeBootstrapBrowseRoots(roots []BrowseRoot, env BrowseEnvironment, startDir string) ([]BrowseRoot, error) {
	reg := defaultBootstrapPlaceRegistry()
	places, err := reg.Places(env, startDir)
	if err != nil {
		return nil, err
	}
	extra := make([]BrowseRoot, 0, len(places))
	for _, p := range places {
		extra = append(extra, placeToBrowseRoot(p))
	}
	return mergeBrowseRootsByPath(roots, extra), nil
}

//funclogmeasure:skip category=hot-path reason="Pure helper; operation trace is emitted by ResolveWorkspacePickerRoots."
func mergeBrowseRootsByPath(existing, extra []BrowseRoot) []BrowseRoot {
	seen := make(map[string]struct{}, len(existing)+len(extra))
	out := make([]BrowseRoot, 0, len(existing)+len(extra))
	for _, root := range existing {
		canon, err := canonicalizePathForContainment(root.Path)
		if err != nil {
			continue
		}
		if _, dup := seen[canon]; dup {
			continue
		}
		seen[canon] = struct{}{}
		out = append(out, root)
	}
	for _, root := range extra {
		canon, err := canonicalizePathForContainment(root.Path)
		if err != nil {
			continue
		}
		if _, dup := seen[canon]; dup {
			continue
		}
		seen[canon] = struct{}{}
		out = append(out, root)
	}
	return out
}

// BrowseRootFromPath builds a BrowseRoot for an absolute path with availability
// checked via os.Stat. Used by handlers that source roots from the git_repositories
// table rather than OS providers.
//
//funclogmeasure:skip category=hot-path reason="Browse sub-step; operation trace is emitted by the calling chokepoint."
func BrowseRootFromPath(id, path string, cat PlaceCategory) BrowseRoot {
	root := BrowseRoot{
		ID:       id,
		Path:     path,
		Label:    browseRootLabel(path),
		Category: cat,
	}
	markBrowseRootAvailable(&root)
	return root
}

// defaultPlaceRegistry returns an empty registry after Cycle 7 retirement.
// Install, Home, and UserDirs providers are no longer registered by default;
// workspace roots are sourced from the git_repositories table via the handler.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by ResolveBrowseRoots."
func defaultPlaceRegistry() *PlaceRegistry {
	return NewPlaceRegistry()
}

//funclogmeasure:skip category=hot-path reason="Browse sub-step; operation trace is emitted by ResolveBrowseRoots."
func parseBrowseRootPaths(raw string) ([]BrowseRoot, error) {
	parts := strings.Split(raw, ",")
	out := make([]BrowseRoot, 0, len(parts))
	for i, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		abs, err := filepath.Abs(p)
		if err != nil {
			return nil, fmt.Errorf("browse root %q: %w", p, err)
		}
		root := BrowseRoot{
			ID:    fmt.Sprintf("custom-%d", i),
			Path:  abs,
			Label: browseRootLabel(abs),
		}
		markBrowseRootAvailable(&root)
		out = append(out, root)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("HAMIX_BROWSE_ROOTS is empty after parsing")
	}
	return out, nil
}

//funclogmeasure:skip category=hot-path reason="Browse sub-step; operation trace is emitted by ResolveBrowseRoots."
func resolveInstallBrowseRoot(startDir string, _ BrowseEnvironment) (BrowseRoot, error) {
	install, err := FindInstallRoot(startDir)
	if err != nil {
		return BrowseRoot{}, err
	}
	root := BrowseRoot{
		ID:    browseRootInstall,
		Path:  install,
		Label: "Hamix checkout",
	}
	markBrowseRootAvailable(&root)
	return root, nil
}

//funclogmeasure:skip category=hot-path reason="Browse sub-step; operation trace is emitted by ResolveBrowseRoots."
func resolveHomeBrowseRoot(_ BrowseEnvironment) (BrowseRoot, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return BrowseRoot{}, err
	}
	path := home
	root := BrowseRoot{
		ID:    browseRootHome,
		Path:  path,
		Label: "Home",
	}
	markBrowseRootAvailable(&root)
	return root, nil
}

//funclogmeasure:skip category=hot-path reason="Browse sub-step; operation trace is emitted by ResolveBrowseRoots."
func markBrowseRootAvailable(root *BrowseRoot) {
	fi, err := os.Stat(root.Path)
	if err != nil {
		root.Available = false
		root.UnavailableReason = "directory is not accessible"
		return
	}
	if !fi.IsDir() {
		root.Available = false
		root.UnavailableReason = "path is not a directory"
		return
	}
	root.Available = true
}

//funclogmeasure:skip category=hot-path reason="Pure helper; operation trace is emitted by ResolveWorkspacePickerRoots."
func allBrowseRootsUnavailable(roots []BrowseRoot) bool {
	if len(roots) == 0 {
		return false
	}
	for _, root := range roots {
		if root.Available {
			return false
		}
	}
	return true
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func browseRootLabel(abs string) string {
	base := filepath.Base(filepath.Clean(abs))
	if base == "" || base == string(filepath.Separator) {
		return abs
	}
	return base
}
