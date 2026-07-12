package verify

import (
	"context"
	"errors"
	cyclesstore "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/store"
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/taskapi/composition"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/storefake"
	checklistdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/domain"
)

type failingVerifyUpsertStore struct {
	*composition.API
}

func (f *failingVerifyUpsertStore) UpsertVerifyReports(context.Context, string, int64, []cyclesstore.VerifyReportEntry) error {
	return errors.New("upsert verify reports failed")
}

func TestPersistVerifyReports_failureSetsMirrorDegradedInPipelineOpts(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := &failingVerifyUpsertStore{API: storefake.New(t).API}
	svc := NewService(Deps{Store: st})
	err := svc.persistVerifyReports(ctx, "cycle-1", 1, []Verdict{
		{ID: "c1", Passed: false, Verifier: checklistdomain.VerifierAgentSelf},
	}, nil)
	if err == nil {
		t.Fatal("expected upsert error")
	}
}

func TestEncodePhaseDetails_mirrorDegradedAndRetryCount(t *testing.T) {
	t.Parallel()
	raw := EncodePhaseDetails(1, nil, nil, PhaseDetailsOpts{
		MirrorDegraded:   true,
		VerifyRetryCount: 2,
	})
	if !ParseMirrorDegraded(raw) {
		t.Fatal("expected mirror_degraded true")
	}
	retry, ok := ParseVerifyRetryCount(raw)
	if !ok || retry != 2 {
		t.Fatalf("verify_retry_count = %d ok=%v, want 2 true", retry, ok)
	}
}

func TestParseVerifyRetryCount_absent(t *testing.T) {
	t.Parallel()
	legacy := []byte(`{"verification":{"attempt_seq":1,"passed_count":0,"failed_count":0,"criteria":[]}}`)
	if _, ok := ParseVerifyRetryCount(legacy); ok {
		t.Fatal("expected absent verify_retry_count on legacy payload")
	}
}
