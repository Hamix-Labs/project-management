package repo

import (
	"context"
	"errors"
	"io/fs"
	"log/slog"
	"path/filepath"
	"sort"

	"github.com/AlexsanderHamir/Hamix/pkgs/gitexec"
)

// MaxFileListPaths caps one file listing. The list is sent to the browser as a
// single JSON array and held in memory there, so the ceiling is sized for the
// client (roughly 2 MiB of JSON), not for the filesystem. Repositories past it
// need a server-side index.
const MaxFileListPaths = 50000

// FileListSource records how a listing was produced, because the two sources
// have different semantics: only the git one honors gitignore.
type FileListSource string

const (
	FileListSourceGit  FileListSource = "git"
	FileListSourceWalk FileListSource = "walk"
)

// FileListing is the full set of referenceable files under a Root.
type FileListing struct {
	Paths     []string       `json:"paths"`
	Truncated bool           `json:"truncated"`
	Source    FileListSource `json:"source"`
}

// Files lists every file under the root, gitignore-aware when the root is a git
// work tree. Falls back to a directory walk otherwise, which cannot apply
// ignore rules and so uses a fixed skip list instead.
func (r *Root) Files(ctx context.Context) (FileListing, error) {
	slog.Debug("trace", "operation", "repo.Root.Files")
	paths, truncated, err := gitexec.ListFiles(ctx, r.abs, MaxFileListPaths)
	if err == nil {
		sort.Strings(paths)
		return FileListing{Paths: paths, Truncated: truncated, Source: FileListSourceGit}, nil
	}
	if !errors.Is(err, gitexec.ErrNoWorkTree) {
		return FileListing{}, err
	}
	return r.walkFiles()
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
