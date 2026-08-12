package repo

import (
	"context"
	"errors"
	"io/fs"
	"log/slog"
	"path/filepath"
	"sort"
	"strings"

	"github.com/AlexsanderHamir/Hamix/pkgs/gitexec"
)

// DefaultFilePageLimit is the warm-batch size for GET /repo/files when limit is omitted.
const DefaultFilePageLimit = 500

// MaxFilePageLimit caps one HTTP page (not the total index size).
const MaxFilePageLimit = 2000

// MaxFileListPaths is a safety ceiling when building the in-process index from a
// non-git walk fallback. Git listings are uncapped (limit 0 to gitexec.ListFiles).
const MaxFileListPaths = 500_000

// FileListSource records how a listing was produced, because the two sources
// have different semantics: only the git one honors gitignore.
type FileListSource string

const (
	FileListSourceGit  FileListSource = "git"
	FileListSourceWalk FileListSource = "walk"
)

// FileListing is the full set of referenceable files under a Root (server index).
type FileListing struct {
	Paths     []string       `json:"paths"`
	Truncated bool           `json:"truncated"`
	Source    FileListSource `json:"source"`
}

// FilePage is one cursor page of referenceable paths for client index warm / fallback search.
type FilePage struct {
	Paths     []string       `json:"paths"`
	NextAfter string         `json:"next_after,omitempty"`
	HasMore   bool           `json:"has_more"`
	Source    FileListSource `json:"source"`
	Truncated bool           `json:"truncated,omitempty"`
}

// Files lists every file under the root, gitignore-aware when the root is a git
// work tree. Falls back to a directory walk otherwise, which cannot apply
// ignore rules and so uses a fixed skip list instead.
func (r *Root) Files(ctx context.Context) (FileListing, error) {
	slog.Debug("trace", "operation", "repo.Root.Files")
	paths, truncated, err := gitexec.ListFiles(ctx, r.abs, 0)
	if err == nil {
		sort.Strings(paths)
		return FileListing{Paths: paths, Truncated: truncated, Source: FileListSourceGit}, nil
	}
	if !errors.Is(err, gitexec.ErrNoWorkTree) {
		return FileListing{}, err
	}
	return r.walkFiles()
}

// FilesPage returns a cursor page over the cached full listing.
// q, when non-empty, filters paths with a case-insensitive substring match
// (basename matches sort ahead of path matches in the page order among equals
// is not re-ranked here — clients rank locally once warm; q is a warm-incomplete fallback).
func (r *Root) FilesPage(ctx context.Context, q, after string, limit int) (FilePage, error) {
	slog.Debug("trace", "operation", "repo.Root.FilesPage")
	if limit <= 0 {
		limit = DefaultFilePageLimit
	}
	if limit > MaxFilePageLimit {
		limit = MaxFilePageLimit
	}
	listing, err := r.cachedFiles(ctx)
	if err != nil {
		return FilePage{}, err
	}
	paths := listing.Paths
	if q = strings.TrimSpace(q); q != "" {
		paths = filterPathsSubstring(paths, q)
	}
	start := 0
	if after != "" {
		start = pathCursorStart(paths, after)
	}
	if start > len(paths) {
		start = len(paths)
	}
	end := start + limit
	if end > len(paths) {
		end = len(paths)
	}
	page := paths[start:end]
	hasMore := end < len(paths)
	nextAfter := ""
	if hasMore && len(page) > 0 {
		nextAfter = page[len(page)-1]
	}
	if page == nil {
		page = []string{}
	}
	return FilePage{
		Paths:     page,
		NextAfter: nextAfter,
		HasMore:   hasMore,
		Source:    listing.Source,
		Truncated: listing.Truncated,
	}, nil
}

//funclogmeasure:skip category=hot-path reason="Pure filter; ListFiles emits the operation-level trace."
func filterPathsSubstring(paths []string, q string) []string {
	needle := strings.ToLower(q)
	out := make([]string, 0, len(paths)/8+1)
	for _, p := range paths {
		if strings.Contains(strings.ToLower(p), needle) {
			out = append(out, p)
		}
	}
	return out
}

// pathCursorStart returns the index of the first path strictly after `after`
// in a sorted slice (lexicographic). If after is missing, starts at 0 so a
// stale cursor still returns a stable page rather than erroring.
//
//funclogmeasure:skip category=hot-path reason="Pure cursor helper; ListFiles emits the operation-level trace."
func pathCursorStart(paths []string, after string) int {
	i := sort.Search(len(paths), func(i int) bool {
		return paths[i] > after
	})
	return i
}

// walkFiles is the non-git fallback. Without ignore rules it would otherwise
// drown the list in dependency and build directories, so it keeps the fixed
// skip list that /repo/search uses.
func (r *Root) walkFiles() (FileListing, error) {
	slog.Debug("trace", "operation", "repo.Root.walkFiles")
	paths := make([]string, 0, 1024)
	truncated := false
	err := filepath.WalkDir(r.abs, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if path == r.abs {
				return err
			}
			slog.Debug("repo file listing skipped unreadable entry", "operation", "repo.Root.walkFiles", "err", err)
			if d != nil && d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			if path != r.abs && shouldSkipSearchDir(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		rel, relErr := filepath.Rel(r.abs, path)
		if relErr != nil {
			return nil
		}
		if len(paths) >= MaxFileListPaths {
			truncated = true
			return fs.SkipAll
		}
		paths = append(paths, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return FileListing{}, err
	}
	sort.Strings(paths)
	return FileListing{Paths: paths, Truncated: truncated, Source: FileListSourceWalk}, nil
}
