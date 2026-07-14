package handlertest

import (
	"context"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/realtime"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/AlexsanderHamir/Hamix/internal/taskapi/composition"
	"github.com/AlexsanderHamir/Hamix/internal/tasktestdb"
	settingscontract "github.com/AlexsanderHamir/Hamix/pkgs/settings/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handler"
)

var _ settingscontract.AgentWorkerControl = (*FakeAgentControl)(nil)

// FakeAgentControl is a test double for settings write / probe / cancel routes.
type FakeAgentControl struct {
	CancelResult  atomic.Bool
	CancelCalls   atomic.Int32
	ReloadCalls   atomic.Int32
	ReloadErr     atomic.Pointer[error]
	ProbeCalls    atomic.Int32
	ProbeVersion  atomic.Pointer[string]
	ProbeResolved atomic.Pointer[string]
	ProbeErr      atomic.Pointer[error]
	LastRunner    atomic.Pointer[string]
	LastBinary    atomic.Pointer[string]
}

//funclogmeasure:skip category=tool-required-noop reason="Test double; not part of production trace paths."
func (f *FakeAgentControl) CancelCurrentRun() bool {
	f.CancelCalls.Add(1)
	return f.CancelResult.Load()
}

//funclogmeasure:skip category=tool-required-noop reason="Test double; not part of production trace paths."
func (f *FakeAgentControl) Reload(_ context.Context) error {
	f.ReloadCalls.Add(1)
	if e := f.ReloadErr.Load(); e != nil {
		return *e
	}
	return nil
}

//funclogmeasure:skip category=tool-required-noop reason="Test double; not part of production trace paths."
func (f *FakeAgentControl) ProbeRunner(_ context.Context, runnerID, binaryPath string, _ time.Duration) (string, string, error) {
	f.ProbeCalls.Add(1)
	r := runnerID
	b := binaryPath
	f.LastRunner.Store(&r)
	f.LastBinary.Store(&b)
	resolved := ""
	if rp := f.ProbeResolved.Load(); rp != nil {
		resolved = *rp
	}
	if e := f.ProbeErr.Load(); e != nil {
		return "", resolved, *e
	}
	if v := f.ProbeVersion.Load(); v != nil {
		return *v, resolved, nil
	}
	return "", resolved, nil
}

// NewSettingsTestServer returns an httptest settings-capable server with a
// FakeAgentControl so PATCH /settings and cancel/probe routes succeed.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only HTTP server; not part of production trace paths."
func NewSettingsTestServer(t *testing.T) (*httptest.Server, *composition.API, *realtime.SSEHub, *FakeAgentControl) {
	t.Helper()
	st := composition.NewAPI(tasktestdb.OpenSQLite(t))
	hub := realtime.NewSSEHub()
	ctrl := &FakeAgentControl{}
	h := handler.NewHandler(st, hub, nil, handler.WithAgentWorkerControl(ctrl))
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return srv, st, hub, ctrl
}

// MustSettingsHTTP issues an HTTP request and requires want status.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only HTTP helper; not part of production trace paths."
func MustSettingsHTTP(t *testing.T, method, url, body string, want int) []byte {
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
