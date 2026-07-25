package verify

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/cursorresume"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
)

func TestEffectiveVerifyModel(t *testing.T) {
	t.Parallel()
	task := &taskcoredomain.Task{CursorModel: "task-model"}
	if got := EffectiveVerifyModel(task, Snapshot{VerifyModel: "verify-pin"}); got != "verify-pin" {
		t.Fatalf("verify pin: got %q", got)
	}
	if got := EffectiveVerifyModel(task, Snapshot{}); got != "task-model" {
		t.Fatalf("inherit task: got %q", got)
	}
	if got := EffectiveVerifyModel(&taskcoredomain.Task{}, Snapshot{}); got != "" {
		t.Fatalf("empty: got %q", got)
	}
}

func TestMergeFailureDetailsIntoPhaseJSON(t *testing.T) {
	t.Parallel()
	base := EncodePhaseDetails(1, nil, nil, PhaseDetailsOpts{})
	merged := MergeFailureDetailsIntoPhaseJSON(base, cursorresume.FailureKindMissingSessionID, cursorresume.MsgMissingSessionForVerify)
	var m map[string]any
	if err := json.Unmarshal(merged, &m); err != nil {
		t.Fatal(err)
	}
	if m["failure_kind"] != cursorresume.FailureKindMissingSessionID {
		t.Fatalf("failure_kind: %#v", m["failure_kind"])
	}
	msg, _ := m["standardized_message"].(string)
	if msg == "" {
		t.Fatal("expected standardized_message")
	}
}

func TestResumeSessionFailedUnwraps(t *testing.T) {
	t.Parallel()
	err := cursorresume.ResumeSessionFailed(runner.ErrResumeSession)
	if !errors.Is(err, runner.ErrResumeSession) {
		t.Fatal("expected unwrap to ErrResumeSession")
	}
	hf, ok := cursorresume.AsHardFail(err)
	if !ok || hf.Kind != cursorresume.FailureKindResumeSession {
		t.Fatalf("AsHardFail: %#v ok=%v", hf, ok)
	}
}
