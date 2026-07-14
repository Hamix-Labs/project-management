package sse_test

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/AlexsanderHamir/Hamix/internal/handlertest"
	"github.com/AlexsanderHamir/Hamix/internal/taskapi/composition"
	"github.com/AlexsanderHamir/Hamix/internal/tasktestdb"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handler"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/realtime"
)

func TestHTTP_SSE_responseHeaders(t *testing.T) {
	db := tasktestdb.OpenSQLite(t)
	h := handler.NewHandler(composition.NewAPI(db), realtime.NewSSEHub(), nil)
	srv := httptest.NewServer(h)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/events", nil)
	if err != nil {
		t.Fatal(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status %d", res.StatusCode)
	}
	if got := res.Header.Get("Content-Type"); got != "text/event-stream" {
		t.Errorf("Content-Type = %q want text/event-stream", got)
	}
	if got := res.Header.Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q want no-store (docs/api.md)", got)
	}
	if got := res.Header.Get("Connection"); got != "keep-alive" {
		t.Errorf("Connection = %q want keep-alive", got)
	}
	if got := res.Header.Get("X-Accel-Buffering"); got != "no" {
		t.Errorf("X-Accel-Buffering = %q want no", got)
	}
	_, _ = io.Copy(io.Discard, res.Body)
}

func TestHTTP_SSE_receivesEventAfterCreate(t *testing.T) {
	srv := handlertest.NewCreateServer(t)
	defer srv.Close()

	streamReady := make(chan struct{})
	payload := make(chan string, 1)
	go func() {
		res, err := http.Get(srv.URL + "/events")
		if err != nil {
			t.Error(err)
			return
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			t.Errorf("sse status %d", res.StatusCode)
			return
		}
		br := bufio.NewReader(res.Body)
		line1, err := br.ReadString('\n')
		if err != nil {
			t.Error(err)
			return
		}
		if !strings.HasPrefix(strings.TrimSpace(line1), "retry:") {
			t.Errorf("want retry line, got %q", line1)
		}
		if _, err := br.ReadString('\n'); err != nil {
			t.Error(err)
			return
		}
		close(streamReady)
		// Each event is now framed as `id: N\ndata: …\n\n` so the
		// browser EventSource captures the id for Last-Event-ID
		// resume. Skip past `id:` lines until the `data:` line shows
		// up; tolerate stray `:heartbeat` comments interleaved with
		// real frames.
		var payloadLine string
		for {
			line, err := br.ReadString('\n')
			if err != nil {
				t.Error(err)
				return
			}
			s := strings.TrimSpace(line)
			if s == "" {
				continue
			}
			if strings.HasPrefix(s, ":") || strings.HasPrefix(s, "id:") {
				continue
			}
			if strings.HasPrefix(s, "data:") {
				payloadLine = strings.TrimSpace(strings.TrimPrefix(s, "data:"))
				break
			}
		}
		payload <- payloadLine
	}()

	<-streamReady
	res, err := http.Post(srv.URL+"/tasks", "application/json", strings.NewReader(handlertest.WithCreateChecklistForURL(srv.URL, `{"title":"sse","priority":"medium"}`)))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("create status %d", res.StatusCode)
	}

	select {
	case p := <-payload:
		var ev realtime.Event
		if err := json.Unmarshal([]byte(p), &ev); err != nil {
			t.Fatal(err)
		}
		if ev.Type != realtime.TaskCreated {
			t.Fatalf("type %q", ev.Type)
		}
		if ev.ID == "" {
			t.Fatal("empty id")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for SSE payload")
	}
}
