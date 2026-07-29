package prompt

import (
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestComposeRecoveryDelta_verifyImplementation(t *testing.T) {
	t.Parallel()
	delta := ComposeRecoveryDelta(RecoveryContext{
		Kind:       RecoveryVerifyImplementation,
		Phase:      cyclesdomain.PhaseExecute,
		AttemptSeq: 2,
		ReportPath: "/tmp/hamix/cycle-1/criteria-report.json",
		FailedCriteria: []CriterionFailure{{
			ID:        "criterion-a",
			Reasoning: "missing handler",
			Verifier:  "execute_agent",
		}},
		LockedCriteria: []string{"criterion-b"},
	})
	if !strings.Contains(delta, "## Continuation (Hamix attempt 2)") {
		t.Fatalf("missing continuation header: %q", delta)
	}
	if !strings.Contains(delta, "**[criterion-a]**") {
		t.Fatalf("missing structured failure: %q", delta)
	}
	if !strings.Contains(delta, "criterion-b") {
		t.Fatalf("missing locked criteria: %q", delta)
	}
}

func TestComposeRecoveryDelta_criteriaReportInvalid(t *testing.T) {
	t.Parallel()
	delta := ComposeRecoveryDelta(RecoveryContext{
		Kind:           RecoveryCriteriaReportInvalid,
		Phase:          cyclesdomain.PhaseExecute,
		AttemptSeq:     3,
		ReportPath:     "/tmp/report.json",
		ReportParseErr: "criteria report invalid: duplicate criterion id a",
		ExpectedIDs:    []string{"a", "c"},
		LockedCriteria: []string{"b"},
	})
	if !strings.Contains(delta, "Parse error:") {
		t.Fatalf("missing parse error: %q", delta)
	}
	if !strings.Contains(delta, "Expected criterion IDs: a, c") {
		t.Fatalf("missing expected ids: %q", delta)
	}
}

func TestComposeRecoveryDelta_verifyInfra(t *testing.T) {
	t.Parallel()
	delta := ComposeRecoveryDelta(RecoveryContext{
		Kind:       RecoveryVerifyInfra,
		Phase:      cyclesdomain.PhaseVerify,
		AttemptSeq: 1,
		ReportPath: "/tmp/hamix/cycle-1/verify-report.json",
		CommandEvidenceDelta: []CommandEvidenceLine{{
			CriterionID: "lint",
			Command:     "npm test",
			ExitCode:    1,
			Preview:     "FAIL",
		}},
		VerifyContract: VerifyReportContract{
			ReportPath: "/tmp/hamix/cycle-1/verify-report.json",
			Criteria: []VerifyCriterionLine{{
				ID: "lint", Text: "tests pass", Evidence: "npm test ok",
			}},
		},
	})
	if !strings.Contains(delta, "### New command evidence") {
		t.Fatalf("missing command evidence section: %q", delta)
	}
	if !strings.Contains(delta, "npm test") {
		t.Fatalf("missing command: %q", delta)
	}
	if !strings.Contains(delta, "verify-report.json") {
		t.Fatalf("missing verify report path: %q", delta)
	}
	if !strings.Contains(delta, `Schema: {"criteria"`) {
		t.Fatalf("missing verify schema: %q", delta)
	}
	if strings.Contains(delta, "criteria-report.json") {
		t.Fatalf("must not ask for criteria-report.json: %q", delta)
	}
}

func TestComposeRecoveryDelta_humanPolishInstructionsOnly(t *testing.T) {
	t.Parallel()
	delta := ComposeRecoveryDelta(RecoveryContext{
		Kind:           RecoveryHumanPolish,
		Phase:          cyclesdomain.PhaseExecute,
		CycleID:        "polish-cycle-1",
		AttemptSeq:     2,
		ReportPath:     "/tmp/report.json",
		LockedCriteria: []string{"crit-locked"},
		Polish: PolishNoticeInput{
			Instructions: "Create REFACTOR.md explaining the refactor.",
			SkipVerify:   true,
		},
	})
	for _, frag := range []string{
		"Human polish", "Create REFACTOR.md", "skip the verify phase",
		"polishments", "crit-locked", "human polish",
	} {
		if !strings.Contains(delta, frag) {
			t.Fatalf("missing %q in %q", frag, delta)
		}
	}
	for _, bad := range []string{"worker restarted", "Fix the issue described above"} {
		if strings.Contains(delta, bad) {
			t.Fatalf("must not contain %q: %q", bad, delta)
		}
	}
}

func TestComposeRecoveryDelta_humanPolishFlagged(t *testing.T) {
	t.Parallel()
	delta := ComposeRecoveryDelta(RecoveryContext{
		Kind:           RecoveryHumanPolish,
		Phase:          cyclesdomain.PhaseExecute,
		CycleID:        "polish-cycle-2",
		AttemptSeq:     3,
		ReportPath:     "/tmp/report.json",
		LockedCriteria: []string{"crit-ok"},
		Polish: PolishNoticeInput{
			Instructions: "Also document the change.",
			Flagged:      []PolishCriterion{{ID: "crit-a", Text: "Auth works"}},
		},
	})
	for _, frag := range []string{
		"Also document the change", "Human-flagged incorrect criteria",
		"[crit-a] Auth works", "active (flagged/new)", "crit-ok",
	} {
		if !strings.Contains(delta, frag) {
			t.Fatalf("missing %q in %q", frag, delta)
		}
	}
	if strings.Contains(delta, "skip the verify phase") {
		t.Fatalf("flagged polish must not skip verify: %q", delta)
	}
}

func TestComposeRecoveryDelta_humanPolishMixed(t *testing.T) {
	t.Parallel()
	delta := ComposeRecoveryDelta(RecoveryContext{
		Kind:       RecoveryHumanPolish,
		Phase:      cyclesdomain.PhaseExecute,
		CycleID:    "polish-cycle-3",
		AttemptSeq: 4,
		ReportPath: "/tmp/report.json",
		Polish: PolishNoticeInput{
			Instructions: "Write REFACTOR.md",
			Flagged:      []PolishCriterion{{ID: "c1", Text: "Named in report"}},
			New:          []PolishCriterion{{ID: "c2", Text: "Docs updated"}},
		},
	})
	for _, frag := range []string{
		"Write REFACTOR.md",
		"Human-flagged incorrect criteria", "[c1] Named in report",
		"Newly added criteria", "[c2] Docs updated",
		"Only criteria with verify commands will be re-checked",
	} {
		if !strings.Contains(delta, frag) {
			t.Fatalf("missing %q in %q", frag, delta)
		}
	}
	if strings.Contains(delta, "worker restarted") {
		t.Fatalf("mixed polish must not use process restart: %q", delta)
	}
}

func TestComposeRecoveryDelta_humanPolishNewOnly(t *testing.T) {
	t.Parallel()
	delta := ComposeRecoveryDelta(RecoveryContext{
		Kind:       RecoveryHumanPolish,
		Phase:      cyclesdomain.PhaseExecute,
		CycleID:    "polish-cycle-4",
		AttemptSeq: 2,
		Polish: PolishNoticeInput{
			Instructions: "Implement the new requirement.",
			New:          []PolishCriterion{{ID: "c-new", Text: "Ship changelog"}},
		},
	})
	for _, frag := range []string{
		"Implement the new requirement", "Newly added criteria", "[c-new] Ship changelog",
	} {
		if !strings.Contains(delta, frag) {
			t.Fatalf("missing %q in %q", frag, delta)
		}
	}
}

func TestComposeRecoveryDelta_goldenFiles(t *testing.T) {
	cases := map[string]RecoveryContext{
		"verify_implementation_fail": {
			Kind:       RecoveryVerifyImplementation,
			Phase:      cyclesdomain.PhaseExecute,
			AttemptSeq: 2,
			ReportPath: "/tmp/hamix/cycle-1/criteria-report.json",
			FailedCriteria: []CriterionFailure{{
				ID: "criterion-a", Reasoning: "missing handler", Verifier: "execute_agent",
			}},
			LockedCriteria: []string{"criterion-b"},
		},
		"criteria_report_invalid": {
			Kind:           RecoveryCriteriaReportInvalid,
			Phase:          cyclesdomain.PhaseExecute,
			AttemptSeq:     3,
			ReportPath:     "/tmp/report.json",
			ReportParseErr: "criteria report invalid: duplicate criterion id a",
			ExpectedIDs:    []string{"a", "c"},
			LockedCriteria: []string{"b"},
		},
		"criteria_report_missing": {
			Kind:        RecoveryCriteriaReportMissing,
			Phase:       cyclesdomain.PhaseExecute,
			AttemptSeq:  2,
			ReportPath:  "/tmp/report.json",
			ExpectedIDs: []string{"a"},
		},
		"process_restart": {
			Kind:             RecoveryProcessRestart,
			Phase:            cyclesdomain.PhaseExecute,
			AttemptSeq:       1,
			InterruptedPhase: cyclesdomain.PhaseExecute,
			FailureReason:    "shutdown",
		},
		"operator_retry_resume": {
			Kind:           RecoveryOperatorRetryResume,
			Phase:          cyclesdomain.PhaseExecute,
			AttemptSeq:     2,
			FailureClass:   "verify",
			FailureReason:  "verification_failed",
			ScopeFiles:     []string{"src/foo.go"},
			LockedCriteria: []string{"criterion-b"},
			ReportPath:     "/tmp/report.json",
		},
		"verify_infra_retry": {
			Kind:       RecoveryVerifyInfra,
			Phase:      cyclesdomain.PhaseVerify,
			AttemptSeq: 1,
			ReportPath: "/tmp/hamix/cycle-1/verify-report.json",
			CommandEvidenceDelta: []CommandEvidenceLine{{
				CriterionID: "lint", Command: "npm test", ExitCode: 1, Preview: "FAIL",
			}},
			VerifyContract: VerifyReportContract{
				ReportPath: "/tmp/hamix/cycle-1/verify-report.json",
				Criteria: []VerifyCriterionLine{{
					ID: "lint", Text: "tests pass", Evidence: "npm test ok",
				}},
			},
		},
	}
	for name, ctx := range cases {
		name, ctx := name, ctx
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got := ComposeRecoveryDelta(ctx)
			path := filepath.Join("testdata", "recovery", name+".txt")
			if os.Getenv("UPDATE_RECOVERY_GOLDEN") != "" {
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
					t.Fatal(err)
				}
				return
			}
			want, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read golden %s: %v (run with UPDATE_RECOVERY_GOLDEN=1)", path, err)
			}
			norm := func(s string) string { return strings.ReplaceAll(s, "\r\n", "\n") }
			if norm(strings.TrimSpace(got)) != norm(strings.TrimSpace(string(want))) {
				t.Fatalf("golden mismatch for %s\n--- got ---\n%s\n--- want ---\n%s", name, got, want)
			}
		})
	}
}
