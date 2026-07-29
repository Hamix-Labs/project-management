package prompt

import (
	"strings"
	"testing"
)

func TestInjectCriteria_ToolOnly_OmitsPathAndSchema(t *testing.T) {
	t.Parallel()
	items := []ChecklistItem{{ID: "a", Text: "do a"}}
	reportPath := "/tmp/hamix-worker/cycle-1/criteria-report.json"
	out := InjectCriteria("base", items, reportPath, nil, true)
	if strings.Contains(out, reportPath) {
		t.Fatalf("tool-only prompt must not include report path; got=%s", out)
	}
	if !strings.Contains(out, "hamix.submit_criteria_report") {
		t.Fatalf("tool-only prompt missing submit tool; got=%s", out)
	}
	if strings.Contains(out, "Schema:") {
		t.Fatalf("tool-only prompt must not dump schema; got=%s", out)
	}
}

func TestInjectCriteria_NoAlreadyVerified_RendersAllItems(t *testing.T) {
	t.Parallel()
	items := []ChecklistItem{
		{ID: "c1", Text: "first criterion"},
		{ID: "c2", Text: "second criterion"},
	}
	reportPath := "/tmp/hamix-worker/cycle-1/criteria-report.json"
	out := InjectCriteria("base", items, reportPath, nil, false)
	if !strings.Contains(out, "first criterion") || !strings.Contains(out, "second criterion") {
		t.Fatalf("expected all items in prompt, got:\n%s", out)
	}
	if strings.Contains(out, "Already verified") {
		t.Errorf("unexpected Already-verified header when no locked passes; out=%q", out)
	}
	if !strings.Contains(out, reportPath) {
		t.Fatalf("absolute report path missing from prompt; want=%q got=%s", reportPath, out)
	}
	if !strings.Contains(out, "schema_version") {
		t.Fatalf("schema_version missing from prompt:\n%s", out)
	}
}

func TestInjectCriteria_LockedItem_OmittedFromActiveChecklist(t *testing.T) {
	t.Parallel()
	items := []ChecklistItem{
		{ID: "c1", Text: "first criterion"},
		{ID: "c2", Text: "second criterion"},
	}
	already := map[string]struct{}{"c1": {}}
	out := InjectCriteria("base", items, "/tmp/hamix-worker/cycle-1/criteria-report.json", already, false)
	if !strings.Contains(out, "Already verified") {
		t.Fatalf("missing Already verified header:\n%s", out)
	}
	if !strings.Contains(out, "[c1]") {
		t.Fatalf("locked id missing from Already verified:\n%s", out)
	}
	// Active section should list c2 but the "do not include already-verified" note applies.
	if !strings.Contains(out, "[c2]") {
		t.Fatalf("active criterion missing:\n%s", out)
	}
}

func TestInjectCriteria_AllLocked_NoActiveSchema(t *testing.T) {
	t.Parallel()
	items := []ChecklistItem{
		{ID: "c1", Text: "first criterion"},
	}
	already := map[string]struct{}{"c1": {}}
	out := InjectCriteria("base", items, "/tmp/hamix-worker/cycle-1/criteria-report.json", already, false)
	if !strings.Contains(out, "already verified") && !strings.Contains(out, "Already verified") {
		t.Fatalf("expected all-locked messaging:\n%s", out)
	}
	if strings.Contains(out, "schema_version") {
		t.Fatalf("schema should be omitted when all locked:\n%s", out)
	}
}
