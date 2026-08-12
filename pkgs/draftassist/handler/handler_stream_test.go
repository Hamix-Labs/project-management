package handler_test

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

	"github.com/AlexsanderHamir/Hamix/pkgs/draftassist/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/draftassist/domain"
	draftassisthandler "github.com/AlexsanderHamir/Hamix/pkgs/draftassist/handler"
	"github.com/AlexsanderHamir/Hamix/pkgs/draftassist/runner"
	draftassiststore "github.com/AlexsanderHamir/Hamix/pkgs/draftassist/store"
)

func TestDraftAssist_streamContract(t *testing.T) {
	store := draftassiststore.NewMemoryStore()
	fake := runner.NewFake(runner.FakeOptions{TokenDelay: 2 * time.Millisecond})
	mux := http.NewServeMux()
	draftassisthandler.Register(mux, draftassisthandler.Deps{Store: store, Runner: fake})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	res, err := http.Post(srv.URL+"/draft-assist/sessions", "application/json",
		strings.NewReader(`{"worktree_id":"wt-1","snapshot":{"title":"T","prompt":"<p>hi</p>"}}`))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("create status=%d", res.StatusCode)
	}
	var created struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(res.Body).Decode(&created); err != nil {
		t.Fatal(err)
	}

	ready, err := http.Get(srv.URL + "/draft-assist/ready")
	if err != nil {
		t.Fatal(err)
	}
	defer ready.Body.Close()
	var readyBody struct {
		Ready  bool   `json:"ready"`
		Runner string `json:"runner"`
	}
	_ = json.NewDecoder(ready.Body).Decode(&readyBody)
	if !readyBody.Ready || readyBody.Runner != "fake" {
		t.Fatalf("ready=%v runner=%q", readyBody.Ready, readyBody.Runner)
	}

	// Subscribe before the run so we don't miss early frames.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/draft-assist/sessions/"+created.ID+"/events", nil)
	if err != nil {
		t.Fatal(err)
	}
	sseRes, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer sseRes.Body.Close()
	if sseRes.StatusCode != http.StatusOK {
		t.Fatalf("sse status=%d", sseRes.StatusCode)
	}

	type result struct {
		status, token, done bool
		err                 error
	}
	ch := make(chan result, 1)
	go func() {
		sawStatus, sawToken, sawDone := false, false, false
		sc := bufio.NewScanner(sseRes.Body)
		for sc.Scan() {
			line := sc.Text()
			if !strings.HasPrefix(line, "event: ") {
				continue
			}
			switch strings.TrimPrefix(line, "event: ") {
			case "status":
				sawStatus = true
			case "token":
				sawToken = true
			case "done":
				sawDone = true
			}
			if sawStatus && sawToken && sawDone {
				ch <- result{status: true, token: true, done: true}
				return
			}
		}
		ch <- result{status: sawStatus, token: sawToken, done: sawDone, err: sc.Err()}
	}()

	runRes, err := http.Post(srv.URL+"/draft-assist/sessions/"+created.ID+"/runs", "application/json",
		strings.NewReader(`{"user_message":"Tighten this brief"}`))
	if err != nil {
		t.Fatal(err)
	}
	defer runRes.Body.Close()
	if runRes.StatusCode != http.StatusAccepted {
		b, _ := io.ReadAll(runRes.Body)
		t.Fatalf("run status=%d body=%s", runRes.StatusCode, b)
	}
	var runAccepted struct {
		RunID string `json:"run_id"`
	}
	_ = json.NewDecoder(runRes.Body).Decode(&runAccepted)
	if runAccepted.RunID == "" {
		t.Fatal("expected run_id")
	}

	select {
	case got := <-ch:
		if !got.status || !got.token || !got.done {
			t.Fatalf("saw status=%v token=%v done=%v err=%v", got.status, got.token, got.done, got.err)
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for status/token/done")
	}
}

func TestDraftAssist_nonceMismatch(t *testing.T) {
	store := draftassiststore.NewMemoryStore()
	sess, err := store.CreateSession(context.Background(), contract.CreateSessionInput{
		Snapshot: domain.FormSnapshot{Prompt: "<p>a</p>"},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.UpdatePrompt(context.Background(), sess.ID, "wrong-nonce", "<p>x</p>")
	if err == nil {
		t.Fatal("expected nonce mismatch")
	}
	if !strings.Contains(err.Error(), "nonce") {
		t.Fatalf("err=%v", err)
	}
}

func TestDraftAssist_cancel(t *testing.T) {
	store := draftassiststore.NewMemoryStore()
	fake := runner.NewFake(runner.FakeOptions{SlowAfterToken: 5 * time.Second, TokenDelay: time.Millisecond})
	mux := http.NewServeMux()
	draftassisthandler.Register(mux, draftassisthandler.Deps{Store: store, Runner: fake})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	res, err := http.Post(srv.URL+"/draft-assist/sessions", "application/json",
		strings.NewReader(`{"snapshot":{"prompt":"p"}}`))
	if err != nil {
		t.Fatal(err)
	}
	var created struct {
		ID string `json:"id"`
	}
	_ = json.NewDecoder(res.Body).Decode(&created)
	res.Body.Close()

	runRes, err := http.Post(srv.URL+"/draft-assist/sessions/"+created.ID+"/runs", "application/json",
		strings.NewReader(`{"user_message":"go"}`))
	if err != nil {
		t.Fatal(err)
	}
	var runAccepted struct {
		RunID string `json:"run_id"`
	}
	_ = json.NewDecoder(runRes.Body).Decode(&runAccepted)
	runRes.Body.Close()

	time.Sleep(30 * time.Millisecond)
	cancelReq, err := http.NewRequest(http.MethodPost,
		srv.URL+"/draft-assist/sessions/"+created.ID+"/runs/"+runAccepted.RunID+"/cancel", nil)
	if err != nil {
		t.Fatal(err)
	}
	cancelRes, err := http.DefaultClient.Do(cancelReq)
	if err != nil {
		t.Fatal(err)
	}
	defer cancelRes.Body.Close()
	if cancelRes.StatusCode != http.StatusAccepted {
		t.Fatalf("cancel status=%d", cancelRes.StatusCode)
	}
}
