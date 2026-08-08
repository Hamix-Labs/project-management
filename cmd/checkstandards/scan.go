package main

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

func resolveRepoRoot(flagRoot string) (string, error) {
	slog.Debug("trace", "operation", "checkstandards.resolveRepoRoot")

	start := flagRoot
	if start == "" {
		wd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("getwd: %w", err)
		}
		start = wd
	}
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", fmt.Errorf("abs %q: %w", start, err)
	}
	for {
		_, statErr := os.Stat(filepath.Join(dir, "go.mod"))
		if statErr == nil {
			return dir, nil
		}
		if !errors.Is(statErr, os.ErrNotExist) {
			return "", fmt.Errorf("stat go.mod in %s: %w", dir, statErr)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("go.mod not found (run from inside the repository)")
		}
		dir = parent
	}
}

// collectFiles returns files under dir with one of exts, sorted for stable
// output. A missing directory yields no files, matching the Test-Path guards
// in the PowerShell checker this replaced.
func collectFiles(dir string, recursive bool, exts []string) ([]string, error) {
	slog.Debug("trace", "operation", "checkstandards.collectFiles")

	info, err := os.Stat(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("stat %s: %w", dir, err)
	}
	if !info.IsDir() {
		return nil, nil
	}

	var out []string
	if !recursive {
		entries, readErr := os.ReadDir(dir)
		if readErr != nil {
			return nil, fmt.Errorf("read dir %s: %w", dir, readErr)
		}
		for _, e := range entries {
			if e.IsDir() || !hasExt(e.Name(), exts) {
				continue
			}
			out = append(out, filepath.Join(dir, e.Name()))
		}
		return out, nil
	}

	walkErr := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !hasExt(d.Name(), exts) {
			return nil
		}
		out = append(out, path)
		return nil
	})
	if walkErr != nil {
		return nil, fmt.Errorf("walk %s: %w", dir, walkErr)
	}
	return out, nil
}

func hasExt(name string, exts []string) bool {
	slog.Debug("trace", "operation", "checkstandards.hasExt")

	lower := strings.ToLower(name)
	for _, ext := range exts {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return false
}

func readText(path string) (string, error) {
	slog.Debug("trace", "operation", "checkstandards.readText")

	b, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", path, err)
	}
	return string(b), nil
}

// countLines matches PowerShell's (Get-Content | Measure-Object -Line).Lines,
// which the size thresholds in CODE_STANDARDS.mdc were calibrated against:
// blank lines do not count, and a trailing newline adds no extra line.
// Whitespace-only lines do count, since Measure-Object only discards the
// empty string.
func countLines(text string) int {
	slog.Debug("trace", "operation", "checkstandards.countLines")

	if text == "" {
		return 0
	}
	lines := strings.Split(text, "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	n := 0
	for _, line := range lines {
		if strings.TrimSuffix(line, "\r") != "" {
			n++
		}
	}
	return n
}

func toPosix(path string) string {
	slog.Debug("trace", "operation", "checkstandards.toPosix")

	return filepath.ToSlash(path)
}
