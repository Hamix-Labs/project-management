package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/AlexsanderHamir/Hamix/internal/taskapi/composition"
	"github.com/AlexsanderHamir/Hamix/internal/tasktestdb"
	"github.com/AlexsanderHamir/Hamix/pkgs/gitwork"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/realtime"
)

func mustEqualEvents(t *testing.T, route string, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s: events len=%d want=%d\ngot=%v\nwant=%v", route, len(got), len(want), got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("%s: event[%d]=%q want=%q", route, i, got[i], want[i])
		}
	}
}

type sseNotifyCapture struct {
	ch chan string
}

func newSSENotifyCapture() (*sseNotifyCapture, NotifyFunc) {
	c := &sseNotifyCapture{ch: make(chan string, 8)}
	return c, func(typ realtime.ChangeType) {
		c.ch <- string(typ) + ":"
	}
}

func drainNotifyEvents(t *testing.T, ch <-chan string, want int, timeout time.Duration) []string {
	t.Helper()
	out := make([]string, 0, want)
	deadline := time.After(timeout)
	for len(out) < want {
		select {
		case s, ok := <-ch:
			if !ok {
				return out
			}
			out = append(out, s)
		case <-deadline:
			return out
		}
	}
	select {
	case s := <-ch:
		out = append(out, s)
	case <-time.After(50 * time.Millisecond):
	}
	return out
}

func summarizeNotifyEvents(events []string) []string {
	out := append([]string(nil), events...)
	sort.Strings(out)
	return out
}

func newSettingsHTTPServer(t *testing.T, st *composition.API, deps Deps) *httptest.Server {
	t.Helper()
	if deps.Settings == nil {
		deps.Settings = st
	}
	if deps.GitRead == nil {
		deps.GitRead = st
	}
	if deps.Git == nil {
		deps.Git = gitwork.New()
	}
	mux := http.NewServeMux()
	Register(mux, deps)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func settingsTestServer(t *testing.T) (*httptest.Server, *composition.API, *sseNotifyCapture, *fakeAgentControl) {
	t.Helper()
	st := composition.NewAPI(tasktestdb.OpenSQLite(t))
	capture, notify := newSSENotifyCapture()
	ctrl := &fakeAgentControl{}
	srv := newSettingsHTTPServer(t, st, Deps{
		Settings: st,
		GitRead:  st,
		Agent:    ctrl,
		Notify:   notify,
	})
	return srv, st, capture, ctrl
}

func settingsTestServerNoAgent(t *testing.T) (*httptest.Server, *composition.API) {
	t.Helper()
	st := composition.NewAPI(tasktestdb.OpenSQLite(t))
	srv := newSettingsHTTPServer(t, st, Deps{Settings: st, GitRead: st})
	return srv, st
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

func mustPatchSettingsJSON(t *testing.T, url, body string, want int) []byte {
	t.Helper()
	return mustSettingsHTTP(t, http.MethodPatch, url, body, want)
}

func mustGetSettingsJSON(t *testing.T, url string, want int) []byte {
	t.Helper()
	res, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	if res.StatusCode != want {
		t.Fatalf("GET %s status=%d want=%d body=%s", url, res.StatusCode, want, b)
	}
	return b
}
