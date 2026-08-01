// Package desktopconfig resolves the Postgres DSN for cmd/hamix-desktop.
// Precedence: DATABASE_URL env → {UserConfigDir}/hamix/desktop.json → none.
// The DSN is never stored in Postgres AppSettings (chicken-and-egg).
package desktopconfig

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	// EnvDatabaseURL is the process override (dev / CI / power users).
	EnvDatabaseURL = "DATABASE_URL"
	// FileName is the config file under the Hamix user config root.
	FileName = "desktop.json"
	// hamixDirName matches ADR-0081 managed worktree root basename.
	hamixDirName = "hamix"
)

// Source identifies where Resolve found the DSN.
type Source string

const (
	SourceNone Source = "none"
	SourceEnv  Source = "env"
	SourceFile Source = "file"
)

// ErrNotConfigured means no env override and no usable desktop.json URL.
var ErrNotConfigured = errors.New("desktop database URL not configured")

// File is the on-disk shape of desktop.json.
type File struct {
	DatabaseURL string `json:"database_url"`
}

// Paths locates the Hamix user config directory. Tests inject a temp root
// via WithRoot; production uses os.UserConfigDir()/hamix.
type Paths struct {
	root string // absolute hamix config dir; empty → resolve from UserConfigDir
}

// WithRoot returns Paths rooted at root (must be the hamix config directory,
// e.g. a temp dir in tests). Empty root uses the OS default.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func WithRoot(root string) Paths {
	return Paths{root: strings.TrimSpace(root)}
}

// Default returns Paths using os.UserConfigDir()/hamix.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func Default() Paths {
	return Paths{}
}

// ConfigDir returns the absolute Hamix user config directory.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (p Paths) ConfigDir() (string, error) {
	if p.root != "" {
		return p.root, nil
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("user config dir: %w", err)
	}
	return filepath.Join(base, hamixDirName), nil
}

// ConfigFilePath returns the absolute path to desktop.json.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (p Paths) ConfigFilePath() (string, error) {
	dir, err := p.ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, FileName), nil
}

// Load reads desktop.json. Missing file returns empty File and nil error.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (p Paths) Load() (File, error) {
	path, err := p.ConfigFilePath()
	if err != nil {
		return File{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return File{}, nil
		}
		return File{}, fmt.Errorf("read desktop config: %w", err)
	}
	var f File
	if err := json.Unmarshal(data, &f); err != nil {
		return File{}, fmt.Errorf("parse desktop config: %w", err)
	}
	f.DatabaseURL = strings.TrimSpace(f.DatabaseURL)
	return f, nil
}

// Save writes desktop.json (creates the config directory if needed).
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (p Paths) Save(f File) error {
	f.DatabaseURL = strings.TrimSpace(f.DatabaseURL)
	path, err := p.ConfigFilePath()
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("mkdir desktop config: %w", err)
	}
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write desktop config: %w", err)
	}
	return nil
}

// Resolve returns the DSN and its source.
// Env DATABASE_URL wins when non-empty after trim; else the file URL; else SourceNone.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (p Paths) Resolve() (dsn string, source Source, err error) {
	if env := strings.TrimSpace(os.Getenv(EnvDatabaseURL)); env != "" {
		return env, SourceEnv, nil
	}
	f, err := p.Load()
	if err != nil {
		return "", SourceNone, err
	}
	if f.DatabaseURL != "" {
		return f.DatabaseURL, SourceFile, nil
	}
	return "", SourceNone, nil
}

// RequireDSN is Resolve that returns ErrNotConfigured when source is none.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (p Paths) RequireDSN() (dsn string, source Source, err error) {
	dsn, source, err = p.Resolve()
	if err != nil {
		return "", source, err
	}
	if source == SourceNone || dsn == "" {
		return "", SourceNone, ErrNotConfigured
	}
	return dsn, source, nil
}
