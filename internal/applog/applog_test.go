package applog

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

func TestOpenJSONL_createsFileUnderDir(t *testing.T) {
	t.Setenv("HAMIX_LOG_DIR", "")
	base := t.TempDir()
	f, path, err := OpenJSONL("taskapi", base, slog.LevelDebug)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(path, base) {
		t.Fatalf("path %q not under %q", path, base)
	}
	baseName := filepath.Base(path)
	if !strings.HasPrefix(baseName, "taskapi-") || !strings.HasSuffix(baseName, ".jsonl") {
		t.Fatalf("unexpected file name: %s", baseName)
	}
	h := slog.New(slog.NewJSONHandler(f, &slog.HandlerOptions{Level: slog.LevelInfo}))
	h.Info("probe", "ok", true)
	if err := f.Sync(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	s := string(raw)
	if !strings.Contains(s, "applog.OpenJSONL") {
		t.Fatalf("expected open trace in jsonl, got %q", s)
	}
	if !strings.Contains(s, `"msg":"probe"`) {
		t.Fatalf("expected JSON log line with probe, got %q", s)
	}
}

func TestOpenJSONL_skipsBootstrapDebugWhenMinInfo(t *testing.T) {
	t.Setenv("HAMIX_LOG_DIR", "")
	base := t.TempDir()
	f, path, err := OpenJSONL("taskapi", base, slog.LevelInfo)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "applog.OpenJSONL") {
		t.Fatalf("bootstrap debug should be suppressed at info level, got %q", string(raw))
	}
}

func TestOpenJSONL_prefersFlagOverEnv(t *testing.T) {
	flagDir := t.TempDir()
	envDir := t.TempDir()
	t.Setenv("HAMIX_LOG_DIR", envDir)
	f, path, err := OpenJSONL("taskapi", flagDir, slog.LevelDebug)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(path, flagDir) {
		t.Fatalf("want log under flag dir %q, got %q", flagDir, path)
	}
}

func TestInstall_fanOutFileAndStderrLevels(t *testing.T) {
	prev := slog.Default()
	t.Cleanup(func() { slog.SetDefault(prev) })

	dir := t.TempDir()
	f, path, err := OpenJSONL("hamix-desktop", dir, slog.LevelInfo)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = f.Close() })

	stderrLevel := slog.LevelWarn
	var seq atomic.Uint64
	Install(InstallConfig{
		File:          f,
		FileLevel:     slog.LevelInfo,
		StderrLevel:   &stderrLevel,
		ProcessLogSeq: &seq,
	})

	slog.Info("info_only_in_file", "k", 1)
	slog.Warn("warn_both", "k", 2)

	if err := f.Sync(); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	s := string(raw)
	if !strings.Contains(s, "info_only_in_file") {
		t.Fatalf("file missing info line: %q", s)
	}
	if !strings.Contains(s, "warn_both") {
		t.Fatalf("file missing warn line: %q", s)
	}
}

func TestInstall_minimizedErrorsOnly(t *testing.T) {
	prev := slog.Default()
	t.Cleanup(func() { slog.SetDefault(prev) })

	var seq atomic.Uint64
	Install(InstallConfig{Minimized: true, ProcessLogSeq: &seq})

	var buf bytes.Buffer
	// Capture via replacing default again would lose minimized — just ensure Install does not panic
	// and Error is enabled.
	if !slog.Default().Enabled(nil, slog.LevelError) {
		t.Fatal("error should be enabled when minimized")
	}
	if slog.Default().Enabled(nil, slog.LevelInfo) {
		t.Fatal("info should be disabled when minimized")
	}
	_ = buf
}

func TestEmitFileLoggingConfig(t *testing.T) {
	var buf bytes.Buffer
	prev := slog.Default()
	t.Cleanup(func() { slog.SetDefault(prev) })
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})))
	EmitFileLoggingConfig("taskapi", slog.LevelInfo)
	line := strings.TrimSpace(buf.String())
	var m map[string]any
	if err := json.Unmarshal([]byte(line), &m); err != nil {
		t.Fatal(err)
	}
	if m["msg"] != "logging config" {
		t.Fatalf("msg %v", m["msg"])
	}
	if m["operation"] != "taskapi.logging" {
		t.Fatalf("operation %v", m["operation"])
	}
}
