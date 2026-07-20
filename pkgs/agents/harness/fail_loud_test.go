package harness_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/taskapi/composition"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/reports"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/storefake"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner/runnerfake"
	settingsdomain "github.com/AlexsanderHamir/Hamix/pkgs/settings/domain"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

type settingsFailStore struct {
	*composition.API
}

func (s *settingsFailStore) GetSettings(context.Context) (settingsdomain.AppSettings, error) {
	return settingsdomain.AppSettings{}, errors.New("settings unavailable")
}

func TestResume_verificationSnapshotLoadFailsTaskAndCycle(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	base := storefake.New(t).API
	tsk, cycle, _ := seedInterruptedExecute(t, base, ctx)
	st := &settingsFailStore{API: base}

	h := harness.New(st, runnerfake.New(), harness.Options{ReportDir: t.TempDir()})
	h.Resume(ctx, tsk, cycle)

	got, err := st.Get(ctx, tsk.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != taskcoredomain.StatusFailed {
		t.Fatalf("task status = %s, want failed", got.Status)
	}
	cycles, err := base.ListCyclesForTask(ctx, tsk.ID, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(cycles) == 0 {
		t.Fatal("expected cycle")
	}
	foundFailed := false
	for _, c := range cycles {
		if c.ID == cycle.ID && c.Status == cyclesdomain.CycleStatusFailed {
			foundFailed = true
			break
		}
	}
	if !foundFailed {
		t.Fatalf("cycle %s not failed; got %+v", cycle.ID, cycles)
	}
}

func TestEnsureReportCycleDir_parentFileFails(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	parentFile := filepath.Join(root, "not-dir")
	if err := os.WriteFile(parentFile, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := reports.EnsureReportCycleDir(parentFile, "c1"); err == nil {
		t.Fatal("expected ensure failure when reportDir is a file")
	}
}

func TestScrubThenEnsure_fileBlocksDir(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	cycleID := "cycle-blocked"
	blocked := reports.ReportCycleDir(root, cycleID)
	if err := os.WriteFile(blocked, []byte("not-a-dir"), 0o600); err != nil {
		t.Fatal(err)
	}
	_ = reports.ScrubCycleArtifacts(root, cycleID)
	if err := os.WriteFile(blocked, []byte("not-a-dir"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := reports.EnsureReportCycleDir(root, cycleID); err == nil {
		t.Fatal("expected EnsureReportCycleDir to fail when path is a file")
	}
}
