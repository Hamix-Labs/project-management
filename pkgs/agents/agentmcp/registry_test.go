package agentmcp

import (
	"path/filepath"
	"testing"
)

func TestDefaultTools_andNewServer(t *testing.T) {
	t.Parallel()
	tools := DefaultTools()
	if len(tools) != 2 {
		t.Fatalf("tools=%d", len(tools))
	}
	names := map[string]bool{}
	for _, tool := range tools {
		names[tool.Name()] = true
		if tool.Group() != GroupReports {
			t.Fatalf("group=%s", tool.Group())
		}
		if tool.Description() == "" {
			t.Fatal("empty description")
		}
	}
	if !names[ToolSubmitCriteria] || !names[ToolSubmitVerify] {
		t.Fatalf("names=%v", names)
	}
	sess := &Session{
		TaskID:      "t",
		CycleID:     "c",
		Phase:       PhaseExecute,
		ReportDir:   t.TempDir(),
		SubmitNonce: "n",
	}
	srv := NewServer(sess)
	if srv == nil {
		t.Fatal("nil server")
	}
	reg := NewRegistry()
	reg.Add(DefaultTools()...)
	if len(reg.tools) != 2 {
		t.Fatalf("reg=%d", len(reg.tools))
	}
}

func TestBindPath(t *testing.T) {
	t.Parallel()
	p := BindPath("/tmp/reports", "cycle-9")
	if p == "" || p == "/tmp/reports" {
		t.Fatalf("path=%q", p)
	}
}

func TestLoadBind_rejectsBadPhase(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "bind.json")
	if err := WriteBind(path, BindFile{
		TaskID: "t", CycleID: "c1", Phase: "nope", ReportDir: dir, SubmitNonce: "n",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadBind(path); err == nil {
		t.Fatal("expected invalid phase")
	}
}

func TestSubmitVerify_wrongPhase(t *testing.T) {
	t.Parallel()
	sess := &Session{Phase: PhaseExecute, ActiveCriterionIDs: map[string]struct{}{"a": {}}}
	_, err := submitVerify(sess, submitVerifyInput{})
	if err == nil {
		t.Fatal("expected error")
	}
}
