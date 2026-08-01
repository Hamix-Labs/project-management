package desktopconfig

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestResolve_envWinsOverFile(t *testing.T) {
	root := t.TempDir()
	p := WithRoot(root)
	if err := p.Save(File{DatabaseURL: "postgres://file-only"}); err != nil {
		t.Fatal(err)
	}
	t.Setenv(EnvDatabaseURL, "postgres://from-env")
	dsn, src, err := p.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	if src != SourceEnv || dsn != "postgres://from-env" {
		t.Fatalf("got %q %s", dsn, src)
	}
}

func TestResolve_fileWhenEnvEmpty(t *testing.T) {
	root := t.TempDir()
	p := WithRoot(root)
	t.Setenv(EnvDatabaseURL, "")
	if err := p.Save(File{DatabaseURL: "postgres://from-file"}); err != nil {
		t.Fatal(err)
	}
	dsn, src, err := p.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	if src != SourceFile || dsn != "postgres://from-file" {
		t.Fatalf("got %q %s", dsn, src)
	}
}

func TestResolve_noneWhenMissing(t *testing.T) {
	p := WithRoot(t.TempDir())
	t.Setenv(EnvDatabaseURL, "")
	dsn, src, err := p.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	if src != SourceNone || dsn != "" {
		t.Fatalf("got %q %s", dsn, src)
	}
	_, _, err = p.RequireDSN()
	if !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("RequireDSN: %v", err)
	}
}

func TestLoad_missingFileOK(t *testing.T) {
	p := WithRoot(t.TempDir())
	f, err := p.Load()
	if err != nil {
		t.Fatal(err)
	}
	if f.DatabaseURL != "" {
		t.Fatalf("unexpected %#v", f)
	}
}

func TestSave_createsDirAndFile(t *testing.T) {
	root := filepath.Join(t.TempDir(), "nested", "hamix")
	p := WithRoot(root)
	if err := p.Save(File{DatabaseURL: "postgres://x"}); err != nil {
		t.Fatal(err)
	}
	path, err := p.ConfigFilePath()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
	f, err := p.Load()
	if err != nil {
		t.Fatal(err)
	}
	if f.DatabaseURL != "postgres://x" {
		t.Fatalf("got %#v", f)
	}
}

func TestResolve_trimsEnvAndFile(t *testing.T) {
	p := WithRoot(t.TempDir())
	t.Setenv(EnvDatabaseURL, "  postgres://env  ")
	dsn, src, err := p.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	if src != SourceEnv || dsn != "postgres://env" {
		t.Fatalf("got %q %s", dsn, src)
	}
}
