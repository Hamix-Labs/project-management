package handler_test

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/draftassist/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/draftassist/domain"
	draftassisthandler "github.com/AlexsanderHamir/Hamix/pkgs/draftassist/handler"
	draftassistmetrics "github.com/AlexsanderHamir/Hamix/pkgs/draftassist/metrics"
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
		schemaVersion       int
		err                 error
	}
	ch := make(chan result, 1)
	go func() {
		sawStatus, sawToken, sawDone := false, false, false
		schemaVersion := 0
		sc := bufio.NewScanner(sseRes.Body)
		for sc.Scan() {
			line := sc.Text()
			if strings.HasPrefix(line, "data: ") {
				var envelope struct {
					Kind string          `json:"kind"`
					Data json.RawMessage `json:"data"`
				}
				if err := json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &envelope); err == nil && envelope.Kind == "session" {
					var sessData struct {
						SchemaVersion int `json:"schema_version"`
					}
					_ = json.Unmarshal(envelope.Data, &sessData)
					schemaVersion = sessData.SchemaVersion
				}
			}
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
				ch <- result{status: true, token: true, done: true, schemaVersion: schemaVersion}
				return
			}
		}
		ch <- result{status: sawStatus, token: sawToken, done: sawDone, schemaVersion: schemaVersion, err: sc.Err()}
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
		if got.schemaVersion != domain.DraftAssistSchemaVersion {
			t.Fatalf("schema_version=%d want %d", got.schemaVersion, domain.DraftAssistSchemaVersion)
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

func TestDraftAssist_cancelTwoFrame(t *testing.T) {
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

	type frame struct {
		kind   string
		status string
	}
	frames := make(chan frame, 16)
	go func() {
		sc := bufio.NewScanner(sseRes.Body)
		var kind string
		for sc.Scan() {
			line := sc.Text()
			if strings.HasPrefix(line, "event: ") {
				kind = strings.TrimPrefix(line, "event: ")
				continue
			}
			if !strings.HasPrefix(line, "data: ") || kind == "" {
				continue
			}
			var envelope struct {
				Kind string          `json:"kind"`
				Data json.RawMessage `json:"data"`
			}
			if err := json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &envelope); err != nil {
				continue
			}
			var statusPayload struct {
				Status string `json:"status"`
			}
			_ = json.Unmarshal(envelope.Data, &statusPayload)
			frames <- frame{kind: kind, status: statusPayload.Status}
			kind = ""
		}
	}()

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

	time.Sleep(40 * time.Millisecond)
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

	sawCancelling, sawCancelledDone := false, false
	deadline := time.After(3 * time.Second)
	for !(sawCancelling && sawCancelledDone) {
		select {
		case f := <-frames:
			if f.kind == "status" && f.status == string(domain.RunStatusCancelling) {
				sawCancelling = true
			}
			if f.kind == "done" && f.status == string(domain.RunStatusCancelled) {
				sawCancelledDone = true
			}
		case <-deadline:
			t.Fatalf("cancelling=%v cancelled_done=%v", sawCancelling, sawCancelledDone)
		}
	}
}

func TestDraftAssist_concurrentRun409(t *testing.T) {
	store := draftassiststore.NewMemoryStore()
	fake := runner.NewFake(runner.FakeOptions{SlowAfterToken: 2 * time.Second, TokenDelay: time.Millisecond})
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

	run1, err := http.Post(srv.URL+"/draft-assist/sessions/"+created.ID+"/runs", "application/json",
		strings.NewReader(`{"user_message":"one"}`))
	if err != nil {
		t.Fatal(err)
	}
	if run1.StatusCode != http.StatusAccepted {
		t.Fatalf("first run status=%d", run1.StatusCode)
	}
	run1.Body.Close()

	time.Sleep(20 * time.Millisecond)
	run2, err := http.Post(srv.URL+"/draft-assist/sessions/"+created.ID+"/runs", "application/json",
		strings.NewReader(`{"user_message":"two"}`))
	if err != nil {
		t.Fatal(err)
	}
	defer run2.Body.Close()
	if run2.StatusCode != http.StatusConflict {
		b, _ := io.ReadAll(run2.Body)
		t.Fatalf("second run status=%d body=%s", run2.StatusCode, b)
	}
}

func TestDraftAssist_lastEventIDReplay(t *testing.T) {
	store := draftassiststore.NewMemoryStore()
	fake := runner.NewFake(runner.FakeOptions{TokenDelay: time.Millisecond})
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

	ctx1, cancel1 := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel1()
	req1, _ := http.NewRequestWithContext(ctx1, http.MethodGet, srv.URL+"/draft-assist/sessions/"+created.ID+"/events", nil)
	sse1, err := http.DefaultClient.Do(req1)
	if err != nil {
		t.Fatal(err)
	}
	lastID := uint64(0)
	done := make(chan struct{})
	go func() {
		defer close(done)
		sc := bufio.NewScanner(sse1.Body)
		for sc.Scan() {
			line := sc.Text()
			if strings.HasPrefix(line, "id: ") {
				id, err := strconv.ParseUint(strings.TrimPrefix(line, "id: "), 10, 64)
				if err == nil && id > lastID {
					lastID = id
				}
			}
			if strings.HasPrefix(line, "event: done") {
				return
			}
		}
	}()

	runRes, err := http.Post(srv.URL+"/draft-assist/sessions/"+created.ID+"/runs", "application/json",
		strings.NewReader(`{"user_message":"replay me"}`))
	if err != nil {
		t.Fatal(err)
	}
	runRes.Body.Close()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("first sse timed out")
	}
	sse1.Body.Close()
	if lastID == 0 {
		t.Fatal("expected last event id")
	}

	// Publish an extra event after disconnect so replay has something beyond lastID.
	_ = store.Publish(context.Background(), created.ID, domain.Event{
		Kind: domain.EventStatus,
		Data: domain.StatusEventData{Status: domain.RunStatusIdle, Reason: "replay-marker"},
	})

	ctx2, cancel2 := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel2()
	req2, _ := http.NewRequestWithContext(ctx2, http.MethodGet, srv.URL+"/draft-assist/sessions/"+created.ID+"/events", nil)
	req2.Header.Set("Last-Event-ID", strconv.FormatUint(lastID, 10))
	sse2, err := http.DefaultClient.Do(req2)
	if err != nil {
		t.Fatal(err)
	}
	defer sse2.Body.Close()

	sawReplay := false
	sc := bufio.NewScanner(sse2.Body)
	deadline := time.After(2 * time.Second)
	for !sawReplay {
		select {
		case <-deadline:
			t.Fatal("did not replay marker")
		default:
		}
		if !sc.Scan() {
			break
		}
		line := sc.Text()
		if strings.Contains(line, "replay-marker") {
			sawReplay = true
		}
	}
	if !sawReplay {
		t.Fatal("expected replay of post-disconnect event")
	}
}

func TestDraftAssist_readyMissing(t *testing.T) {
	store := draftassiststore.NewMemoryStore()
	mux := http.NewServeMux()
	draftassisthandler.Register(mux, draftassisthandler.Deps{Store: store, Runner: nil})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	res, err := http.Get(srv.URL + "/draft-assist/ready")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var body struct {
		Ready  bool   `json:"ready"`
		Runner string `json:"runner"`
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(res.Body).Decode(&body)
	if body.Ready || body.Runner != "missing" || body.Reason != draftassistmetrics.ReasonNoRunner {
		t.Fatalf("ready=%v runner=%q reason=%q", body.Ready, body.Runner, body.Reason)
	}
}

type probeStub struct {
	ready  bool
	runner string
	reason string
}

func (p probeStub) Ready() (bool, string, string) {
	return p.ready, p.runner, p.reason
}

func TestDraftAssist_readyProbeOverride(t *testing.T) {
	store := draftassiststore.NewMemoryStore()
	fake := runner.NewFake(runner.FakeOptions{})
	mux := http.NewServeMux()
	draftassisthandler.Register(mux, draftassisthandler.Deps{
		Store:  store,
		Runner: fake,
		Ready:  probeStub{ready: false, runner: "sdk", reason: draftassistmetrics.ReasonMissingKey},
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	res, err := http.Get(srv.URL + "/draft-assist/ready")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var body struct {
		Ready  bool   `json:"ready"`
		Runner string `json:"runner"`
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(res.Body).Decode(&body)
	if body.Ready || body.Runner != "sdk" || body.Reason != draftassistmetrics.ReasonMissingKey {
		t.Fatalf("ready=%v runner=%q reason=%q", body.Ready, body.Runner, body.Reason)
	}
}

func TestDraftAssist_heartbeatComment(t *testing.T) {
	store := draftassiststore.NewMemoryStore()
	fake := runner.NewFake(runner.FakeOptions{SlowAfterToken: 4 * time.Second, TokenDelay: time.Millisecond})
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

	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/draft-assist/sessions/"+created.ID+"/events", nil)
	sseRes, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer sseRes.Body.Close()

	gotHeartbeat := make(chan struct{}, 1)
	go func() {
		sc := bufio.NewScanner(sseRes.Body)
		for sc.Scan() {
			if strings.HasPrefix(sc.Text(), ": heartbeat") {
				select {
				case gotHeartbeat <- struct{}{}:
				default:
				}
				return
			}
		}
	}()

	runRes, err := http.Post(srv.URL+"/draft-assist/sessions/"+created.ID+"/runs", "application/json",
		strings.NewReader(`{"user_message":"hold"}`))
	if err != nil {
		t.Fatal(err)
	}
	runRes.Body.Close()

	select {
	case <-gotHeartbeat:
	case <-time.After(5 * time.Second):
		t.Fatal("expected heartbeat comment while run active")
	}
}
