package draftsidecar

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestMustHost_missingKey(t *testing.T) {
	t.Setenv(APIKeyEnv, "")
	t.Setenv(BinEnv, filepath.Join(t.TempDir(), "unused"))
	_, err := mustHost(context.Background(), Options{})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), APIKeyEnv) {
		t.Fatalf("error %q should name %s", err, APIKeyEnv)
	}
}

func TestMustHost_missingBinary(t *testing.T) {
	t.Setenv(APIKeyEnv, "test-key")
	missing := filepath.Join(t.TempDir(), "no-such-launcher")
	t.Setenv(BinEnv, missing)
	_, err := mustHost(context.Background(), Options{APIKey: "test-key"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), BinEnv) {
		t.Fatalf("error %q should name %s", err, BinEnv)
	}
}

func TestMustHost_ready(t *testing.T) {
	srv := makeReadyzServer(t, `{"ready":true}`)
	port := extractPort(t, srv.URL)
	stdout, stdin := pipePair()
	child := newFakeChild(stdout)
	script := newSpawnScript(child)
	go writeListenLine(stdin, port)

	host, err := mustHost(context.Background(), Options{
		BinaryPath:     writeLauncher(t, t.TempDir(), BinaryName),
		APIKey:         "test-key",
		Env:            []string{"PATH=/x"},
		StartupTimeout: time.Second,
		ProbeInterval:  20 * time.Millisecond,
		ReadyTimeout:   100 * time.Millisecond,
		spawn:          script.spawn,
	})
	if err != nil {
		t.Fatalf("MustHost: %v", err)
	}
	t.Cleanup(func() { _ = host.Close() })
	ready, name, reason := host.Ready()
	if !ready || name != RunnerName || reason != "" {
		t.Fatalf("ready=%v name=%q reason=%q", ready, name, reason)
	}
	if host.Runner == nil {
		t.Fatal("Runner is nil")
	}
}

func TestMustHost_notReady(t *testing.T) {
	srv := makeReadyzServer(t, `{"ready":false,"reason":"sidecar_down"}`)
	port := extractPort(t, srv.URL)
	stdout, stdin := pipePair()
	child := newFakeChild(stdout)
	script := newSpawnScript(child)
	go writeListenLine(stdin, port)

	_, err := mustHost(context.Background(), Options{
		BinaryPath:     writeLauncher(t, t.TempDir(), BinaryName),
		APIKey:         "test-key",
		Env:            []string{"PATH=/x"},
		StartupTimeout: 150 * time.Millisecond,
		ProbeInterval:  20 * time.Millisecond,
		ReadyTimeout:   50 * time.Millisecond,
		spawn:          script.spawn,
	})
	if err == nil {
		t.Fatal("expected not-ready error")
	}
	if !strings.Contains(err.Error(), "not ready") {
		t.Fatalf("error %q should say not ready", err)
	}
}

func TestMustHost_clearsKeyFromEnv(t *testing.T) {
	// Guard: production MustHost must not invent a key.
	t.Setenv(APIKeyEnv, "")
	if os.Getenv(APIKeyEnv) != "" {
		t.Fatal("setup")
	}
	_, err := MustHost(context.Background())
	if err == nil || !strings.Contains(err.Error(), APIKeyEnv) {
		t.Fatalf("MustHost without key: %v", err)
	}
}
