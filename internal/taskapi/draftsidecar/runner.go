package draftsidecar

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/draftassist/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/draftassist/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
)

// RunnerName is the stable identifier the runner reports through Name().
const RunnerName = "sdk"

// PortSource resolves the loopback port of the sidecar at call time. The
// supervisor implements it (Port()). Tests substitute a static function.
type PortSource interface {
	Port() int
}

// Runner is the contract.Runner implementation that talks to the
// hamix-draft-agent loopback HTTP + SSE surface. It carries no state per
// run; concurrency is safe as long as the supervisor's port is stable.
type Runner struct {
	port    PortSource
	client  *http.Client
	nowFn   func() time.Time
	logger  *slog.Logger
	baseURL string
}

// RunnerOptions configures the SDK-backed runner.
type RunnerOptions struct {
	// PortSource yields the sidecar's current loopback port; the supervisor
	// implements it. Required.
	PortSource PortSource
	// HTTPClient is used for the POST /runs (streaming) and cancel calls.
	// Defaults to a client with no timeout (streaming) and a per-request
	// context deadline the caller controls.
	HTTPClient *http.Client
	// Logger overrides slog.Default.
	Logger *slog.Logger
	// BaseURLOverride skips PortSource lookup and uses a fixed base URL
	// (test seam; e.g. an httptest.Server URL).
	BaseURLOverride string
	// NowFn is a clock override for tests.
	NowFn func() time.Time
}

// NewRunner returns a Runner backed by the sidecar's loopback endpoints.
//
//funclogmeasure:skip category=tool-required-noop reason="Constructor; Run emits the operation trace."
func NewRunner(opts RunnerOptions) *Runner {
	r := &Runner{
		port:    opts.PortSource,
		client:  opts.HTTPClient,
		logger:  opts.Logger,
		baseURL: opts.BaseURLOverride,
		nowFn:   opts.NowFn,
	}
	if r.client == nil {
		// SSE stream reads may last minutes; no client-side timeout.
		r.client = &http.Client{}
	}
	if r.logger == nil {
		r.logger = slog.Default()
	}
	if r.nowFn == nil {
		r.nowFn = time.Now
	}
	return r
}

// Name returns "sdk" so the ready probe and observability tags identify
// the SDK-backed runner path.
//
//funclogmeasure:skip category=hot-path reason="Constant accessor."
func (r *Runner) Name() string { return RunnerName }

// Run POSTs the run to the sidecar and pumps SSE frames into h.Emit until
// a terminal (done) frame arrives, the context is cancelled, or the
// stream errors. Cancellation triggers a fire-and-forget POST to
// /runs/{run_id}/cancel; the sidecar then emits status=cancelling and
// done{cancelled} on the open stream.
//
//funclogmeasure:skip category=tool-required-noop reason="High-cardinality per-run driver; SSE frames emitted downstream carry per-event traces."
func (r *Runner) Run(ctx context.Context, sessionID, runID string, in contract.RunInput, h contract.RunHandle) error {
	base, err := r.resolveBaseURL()
	if err != nil {
		return err
	}
	body := runRequestBody{
		SessionID:   sessionID,
		RunID:       runID,
		UserMessage: in.UserMessage,
		Snapshot:    snapshotToWire(in.Snapshot),
		WorktreeCwd: in.WorktreeCwd,
	}
	if in.Snapshot.CursorModel != "" {
		body.Model = in.Snapshot.CursorModel
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("draftsidecar: marshal run body: %w", err)
	}

	// Request context is independent of the caller's cancel: caller cancel
	// posts /cancel and we keep reading the SSE stream until done/EOF.
	reqCtx, reqCancel := context.WithCancel(context.Background())
	defer reqCancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, base+"/runs", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("draftsidecar: build run request: %w", err)
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("accept", "text/event-stream")

	resp, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("draftsidecar: POST /runs: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 8*1024))
		return fmt.Errorf("draftsidecar: /runs returned %d: %s", resp.StatusCode, strings.TrimSpace(string(msg)))
	}

	// Kick off cancel forwarder that fires when ctx is cancelled. Runs
	// in the background so we can still drain the SSE stream, which is
	// expected to end with a done{cancelled} frame.
	cancelDone := make(chan struct{})
	stopCancelWatch := make(chan struct{})
	go func() {
		defer close(cancelDone)
		select {
		case <-ctx.Done():
			bgCtx, bgCancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer bgCancel()
			r.postCancel(bgCtx, base, sessionID, runID)
		case <-stopCancelWatch:
			return
		}
	}()

	streamErr := r.consumeSSE(ctx, resp.Body, sessionID, runID, h)
	close(stopCancelWatch)
	<-cancelDone
	reqCancel()
	if streamErr != nil && !errors.Is(streamErr, context.Canceled) {
		return streamErr
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return nil
}

// runRequestBody mirrors sidecars/hamix-draft-agent/src/types.ts.RunRequestBody
// so wire compat stays trivially verifiable.
type runRequestBody struct {
	SessionID   string       `json:"session_id"`
	RunID       string       `json:"run_id,omitempty"`
	UserMessage string       `json:"user_message"`
	Snapshot    wireSnapshot `json:"snapshot,omitempty"`
	WorktreeCwd string       `json:"worktree_cwd"`
	Model       string       `json:"model,omitempty"`
	AgentID     string       `json:"agent_id,omitempty"`
}

type wireSnapshot struct {
	Title       string   `json:"title,omitempty"`
	Prompt      string   `json:"prompt,omitempty"`
	Priority    string   `json:"priority,omitempty"`
	ProjectID   string   `json:"project_id,omitempty"`
	Criteria    []string `json:"criteria,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	CursorModel string   `json:"cursor_model,omitempty"`
}

func snapshotToWire(s domain.FormSnapshot) wireSnapshot {
	return wireSnapshot{
		Title:       s.Title,
		Prompt:      s.Prompt,
		Priority:    s.Priority,
		ProjectID:   s.ProjectID,
		Criteria:    s.Criteria,
		Tags:        s.Tags,
		CursorModel: s.CursorModel,
	}
}

// consumeSSE walks the response body, decoding named SSE frames and
// forwarding them onto RunHandle.Emit. It stops on the terminal `done`
// frame or when the body closes. Context cancel does NOT abort the
// reader — the cancel forwarder posts to the sidecar and the stream is
// expected to deliver a terminal cancelled frame; aborting the reader
// would drop those frames.
func (r *Runner) consumeSSE(ctx context.Context, body io.Reader, sessionID, runID string, h contract.RunHandle) error {
	reader := bufio.NewReader(body)
	var event, data string
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return fmt.Errorf("draftsidecar: read SSE: %w", err)
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			if event != "" {
				// Prefer a live ctx for Emit so cancelled runs still
				// publish the terminal frames; fall back to Background
				// once the caller's ctx is done.
				emitCtx := ctx
				if ctx.Err() != nil {
					emitCtx = context.Background()
				}
				done, dispatchErr := r.dispatchFrame(emitCtx, sessionID, runID, event, data, h)
				event, data = "", ""
				if dispatchErr != nil {
					return dispatchErr
				}
				if done {
					return nil
				}
			}
			continue
		}
		if strings.HasPrefix(line, ":") {
			// heartbeat or other comment; ignore.
			continue
		}
		if strings.HasPrefix(line, "event:") {
			event = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
			continue
		}
		if strings.HasPrefix(line, "data:") {
			frag := strings.TrimPrefix(line, "data:")
			if strings.HasPrefix(frag, " ") {
				frag = frag[1:]
			}
			if data == "" {
				data = frag
			} else {
				data = data + "\n" + frag
			}
			continue
		}
		// id:, retry:, or unknown — ignored.
	}
}

// dispatchFrame decodes one SSE frame's data payload into a domain event
// payload and emits it via h. Returns done=true after the terminal
// `done` frame so the caller can stop reading.
func (r *Runner) dispatchFrame(ctx context.Context, sessionID, runID, event, data string, h contract.RunHandle) (bool, error) {
	switch event {
	case string(domain.EventSession):
		var payload domain.SessionEventData
		if data != "" {
			if err := json.Unmarshal([]byte(data), &payload); err != nil {
				return false, fmt.Errorf("draftsidecar: decode session frame: %w", err)
			}
		}
		return false, h.Emit(ctx, sessionID, runID, domain.EventSession, payload)
	case string(domain.EventStatus):
		var payload domain.StatusEventData
		if err := json.Unmarshal([]byte(data), &payload); err != nil {
			return false, fmt.Errorf("draftsidecar: decode status frame: %w", err)
		}
		return false, h.Emit(ctx, sessionID, runID, domain.EventStatus, payload)
	case string(domain.EventToken):
		var payload domain.TokenEventData
		if err := json.Unmarshal([]byte(data), &payload); err != nil {
			return false, fmt.Errorf("draftsidecar: decode token frame: %w", err)
		}
		return false, h.Emit(ctx, sessionID, runID, domain.EventToken, payload)
	case string(domain.EventTool):
		var payload domain.ToolEventData
		if err := json.Unmarshal([]byte(data), &payload); err != nil {
			return false, fmt.Errorf("draftsidecar: decode tool frame: %w", err)
		}
		return false, h.Emit(ctx, sessionID, runID, domain.EventTool, payload)
	case string(domain.EventPatch):
		var payload domain.PatchEventData
		if err := json.Unmarshal([]byte(data), &payload); err != nil {
			return false, fmt.Errorf("draftsidecar: decode patch frame: %w", err)
		}
		return false, h.Emit(ctx, sessionID, runID, domain.EventPatch, payload)
	case string(domain.EventError):
		var payload domain.ErrorEventData
		if err := json.Unmarshal([]byte(data), &payload); err != nil {
			return false, fmt.Errorf("draftsidecar: decode error frame: %w", err)
		}
		return false, h.Emit(ctx, sessionID, runID, domain.EventError, payload)
	case string(domain.EventDone):
		var payload domain.DoneEventData
		if err := json.Unmarshal([]byte(data), &payload); err != nil {
			return false, fmt.Errorf("draftsidecar: decode done frame: %w", err)
		}
		return true, h.Emit(ctx, sessionID, runID, domain.EventDone, payload)
	default:
		// Unknown event kind; ignore rather than fail the whole stream.
		r.logger.Debug("draftsidecar unknown SSE event",
			"cmd", calltrace.LogCmd,
			"operation", "draftsidecar.Runner.dispatchFrame",
			"event", event,
		)
		return false, nil
	}
}

// postCancel best-effort POSTs the cancel to the sidecar. Errors are
// logged (without the API key) but never surfaced; the SSE stream will
// deliver the terminal cancelled frame either way.
func (r *Runner) postCancel(ctx context.Context, base, sessionID, runID string) {
	body, _ := json.Marshal(map[string]string{"session_id": sessionID})
	url := fmt.Sprintf("%s/runs/%s/cancel", base, runID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("content-type", "application/json")
	resp, err := r.client.Do(req)
	if err != nil {
		r.logger.Debug("draftsidecar cancel dispatch failed",
			"cmd", calltrace.LogCmd,
			"operation", "draftsidecar.Runner.postCancel",
			"err", err,
			"session_id", sessionID,
			"run_id", runID,
		)
		return
	}
	_ = resp.Body.Close()
}

func (r *Runner) resolveBaseURL() (string, error) {
	if r.baseURL != "" {
		return strings.TrimRight(r.baseURL, "/"), nil
	}
	if r.port == nil {
		return "", errors.New("draftsidecar: runner has no PortSource")
	}
	p := r.port.Port()
	if p == 0 {
		return "", errors.New("draftsidecar: sidecar port not yet known")
	}
	return fmt.Sprintf("http://127.0.0.1:%d", p), nil
}

var _ contract.Runner = (*Runner)(nil)
