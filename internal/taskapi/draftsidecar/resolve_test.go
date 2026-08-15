package draftsidecar

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeLauncher(t *testing.T, dir, name string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestResolveBinary_envWins(t *testing.T) {
	dir := t.TempDir()
	want := writeLauncher(t, dir, "pinned-agent")
	got, err := resolveBinary(want, "unused-exe-dir", func(string) (string, error) {
		t.Fatal("lookPath must not run when env is set")
		return "", errors.New("unreachable")
	})
	if err != nil {
		t.Fatal(err)
	}
	abs, err := filepath.Abs(want)
	if err != nil {
		t.Fatal(err)
	}
	if got != abs {
		t.Fatalf("got %q want %q", got, abs)
	}
}

func TestResolveBinary_envMissingFile(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "no-such-launcher")
	_, err := resolveBinary(missing, t.TempDir(), func(string) (string, error) {
		t.Fatal("must not fall through when env is set")
		return "", errors.New("unreachable")
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), BinEnv) {
		t.Fatalf("error %q should name %s", err, BinEnv)
	}
}

func TestResolveBinary_sibling(t *testing.T) {
	dir := t.TempDir()
	want := writeLauncher(t, dir, BinaryName)
	got, err := resolveBinary("", dir, func(string) (string, error) {
		t.Fatal("lookPath must not run when sibling exists")
		return "", errors.New("unreachable")
	})
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestResolveBinary_siblingCmd(t *testing.T) {
	dir := t.TempDir()
	want := writeLauncher(t, dir, BinaryName+".cmd")
	got, err := resolveBinary("", dir, func(string) (string, error) {
		return "", errors.New("not on PATH")
	})
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestResolveBinary_lookPath(t *testing.T) {
	want := filepath.Join(t.TempDir(), "from-path", BinaryName)
	got, err := resolveBinary("", t.TempDir(), func(name string) (string, error) {
		if name != BinaryName {
			t.Fatalf("LookPath name %q", name)
		}
		return want, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestResolveBinary_missing(t *testing.T) {
	_, err := resolveBinary("", t.TempDir(), func(string) (string, error) {
		return "", errors.New("not found")
	})
	if err == nil {
		t.Fatal("expected error")
	}
	msg := err.Error()
	if !strings.Contains(msg, BinaryName) || !strings.Contains(msg, BinEnv) {
		t.Fatalf("error %q should name binary and %s", err, BinEnv)
	}
}
