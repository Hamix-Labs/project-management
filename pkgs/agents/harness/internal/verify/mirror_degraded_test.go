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

func TestEncodePhaseDetails_mirrorDegraded(t *testing.T) {
	t.Parallel()
	raw := EncodePhaseDetails(1, nil, nil, PhaseDetailsOpts{
		MirrorDegraded: true,
	})
	if !ParseMirrorDegraded(raw) {
		t.Fatal("expected mirror_degraded true")
	}
}
