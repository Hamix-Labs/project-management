package handler

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/AlexsanderHamir/Hamix/internal/taskapi/composition"
	"github.com/AlexsanderHamir/Hamix/internal/tasktestdb"
)

type fakeAgentControl struct {
	cancelResult  atomic.Bool
	cancelCalls   atomic.Int32
	reloadCalls   atomic.Int32
	reloadErr     atomic.Pointer[error]
	probeCalls    atomic.Int32
	probeVersion  atomic.Pointer[string]
	probeResolved atomic.Pointer[string]
	probeErr      atomic.Pointer[error]
	lastRunner    atomic.Pointer[string]
	lastBinary    atomic.Pointer[string]
}

func (f *fakeAgentControl) CancelCurrentRun() bool {
	f.cancelCalls.Add(1)
	return f.cancelResult.Load()
}

func (f *fakeAgentControl) Reload(_ context.Context) error {
	f.reloadCalls.Add(1)
	if e := f.reloadErr.Load(); e != nil {
		return *e
	}
	return nil
}

func (f *fakeAgentControl) ProbeRunner(_ context.Context, runnerID, binaryPath string, _ time.Duration) (string, string, error) {
	f.probeCalls.Add(1)
	r := runnerID
	b := binaryPath
	f.lastRunner.Store(&r)
	f.lastBinary.Store(&b)
	resolved := ""
	if rp := f.probeResolved.Load(); rp != nil {
		resolved = *rp
	}
	if e := f.probeErr.Load(); e != nil {
		return "", resolved, *e
	}
	if v := f.probeVersion.Load(); v != nil {
		return *v, resolved, nil
	}
	return "", resolved, nil
}

func settingsTestServer(t *testing.T) (*httptest.Server, *composition.API, *SSEHub, *fakeAgentControl) {
	t.Helper()
	st := composition.NewAPI(tasktestdb.OpenSQLite(t))
	hub := NewSSEHub()
	ctrl := &fakeAgentControl{}
	h := NewHandler(st, hub, nil, WithAgentWorkerControl(ctrl))
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return srv, st, hub, ctrl
}

func mustSettingsHTTP(t *testing.T, method, url, body string, want int) []byte {
	t.Helper()
	var reqBody io.Reader
	if body != "" {
		reqBody = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, url, err)
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	if res.StatusCode != want {
		t.Fatalf("%s %s: status=%d want=%d body=%s", method, url, res.StatusCode, want, b)
	}
	return b
}
