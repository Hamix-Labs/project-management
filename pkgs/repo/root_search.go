package repo

import (
	"io/fs"
	"log/slog"
	"path/filepath"
	"strings"
)

// EntryKind is a filesystem entry kind returned by SearchEntries.
type EntryKind string

const (
	EntryKindFile EntryKind = "file"
	EntryKindDir  EntryKind = "dir"
)

// SearchEntry is one repo-relative path hit from SearchEntries.
type SearchEntry struct {
	Path string    `json:"path"`
	Kind EntryKind `json:"kind"`
}

// SearchKinds selects which entry kinds SearchEntries returns.
type SearchKinds struct {
	File bool
	Dir  bool
}

// SearchKindsFilesOnly is the default for @-mention file browse.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func SearchKindsFilesOnly() SearchKinds {
	return SearchKinds{File: true}
}

// ParseSearchKinds parses a comma-separated kinds query (file, dir).
// Empty or whitespace-only defaults to files only. Unknown tokens return false.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ParseSearchKinds(raw string) (SearchKinds, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return SearchKindsFilesOnly(), true
	}
	var k SearchKinds
	for _, part := range strings.Split(raw, ",") {
		part = strings.ToLower(strings.TrimSpace(part))
		if part == "" {
			continue
		}
		switch part {
		case "file":
			k.File = true
		case "dir":
			k.Dir = true
		default:
			return SearchKinds{}, false
		}
	}
	if !k.File && !k.Dir {
		return SearchKindsFilesOnly(), true
	}
	return k, true
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func shouldSkipSearchDir(name string) bool {
	switch name {
	case ".git", "node_modules", "vendor":
		return true
	case "dist", "build", "out", "target", "coverage", ".next", ".nuxt", ".turbo",
		"__pycache__", ".pytest_cache", ".venv", "venv", ".mypy_cache", ".tox":
		return true
	default:
		return false
	}
}

// Search returns repo-relative file paths matching query (substring, case-insensitive).
// Empty query lists up to maxSearchResultsBrowse files (walk order); non-empty query up to maxSearchResultsFilter matches.
func (r *Root) Search(query string) ([]string, error) {
	slog.Debug("trace", "operation", "repo.Root.Search")
	entries, err := r.SearchEntries(query, SearchKindsFilesOnly())
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.Kind == EntryKindFile {
			out = append(out, e.Path)
		}
	}
	return out, nil
}

// SearchEntries returns repo-relative file and/or directory paths matching query.
func (r *Root) SearchEntries(query string, kinds SearchKinds) ([]SearchEntry, error) {
	slog.Debug("trace", "operation", "repo.Root.SearchEntries")
	if !kinds.File && !kinds.Dir {
		kinds = SearchKindsFilesOnly()
	}
	q := strings.ToLower(strings.TrimSpace(query))
	limit := maxSearchResultsFilter
	if q == "" {
		limit = maxSearchResultsBrowse
	}
	var out []SearchEntry
	err := filepath.WalkDir(r.abs, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// One unreadable entry used to fail the whole request with a 500.
			// A permission-denied directory deep in a repository should cost
			// its own subtree, not every result the user was looking for. An
			// unreadable root is still fatal — there is nothing to return.
			if path == r.abs {
				return err
			}
			slog.Debug("repo search skipped unreadable entry", "operation", "repo.Root.SearchEntries", "err", err)
			if d != nil && d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if shouldSkipSearchDir(name) {
				return filepath.SkipDir
			}
			if path == r.abs {
				return nil
			}
			if !kinds.Dir {
				return nil
			}
			rel, relErr := filepath.Rel(r.abs, path)
			if relErr != nil {
				return nil
			}
			rel = filepath.ToSlash(rel)
			if q == "" || strings.Contains(strings.ToLower(rel), q) {
				out = append(out, SearchEntry{Path: rel, Kind: EntryKindDir})
				if len(out) >= limit {
					return fs.SkipAll
				}
			}
			return nil
		}
		if !kinds.File {
			return nil
		}
		rel, relErr := filepath.Rel(r.abs, path)
		if relErr != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		if q == "" || strings.Contains(strings.ToLower(rel), q) {
			out = append(out, SearchEntry{Path: rel, Kind: EntryKindFile})
			if len(out) >= limit {
				return fs.SkipAll
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}
