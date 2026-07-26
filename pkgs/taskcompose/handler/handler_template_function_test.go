package handler

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestApplyFunctionBindingsToPayload_softScope(t *testing.T) {
	t.Parallel()
	raw := json.RawMessage(`{
		"title":"Scoped",
		"priority":"medium",
		"status":"ready",
		"initial_prompt":"Do the work.",
		"checklist_items":[{"text":"done","verifier":"human"}],
		"function_inputs":[{"id":"scope","kind":"dir","label":"Dir"},{"id":"entry","kind":"function","label":"Fn"}]
	}`)
	out, err := applyFunctionBindingsToPayload(raw, []functionBindingJSON{
		{InputID: "scope", Paths: []string{"pkgs/repo"}},
		{InputID: "entry", Functions: []functionRefBindingJSON{{Path: "pkgs/repo/root.go", Name: "OpenRoot", Line: 28}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	if err := json.Unmarshal(out, &payload); err != nil {
		t.Fatal(err)
	}
	if _, ok := payload["function_inputs"]; ok {
		t.Fatalf("function_inputs should be stripped, got %v", payload["function_inputs"])
	}
	prompt, _ := payload["initial_prompt"].(string)
	if !strings.Contains(prompt, "## Scope (do not expand beyond)") {
		t.Fatalf("missing scope block in %q", prompt)
	}
	if !strings.Contains(prompt, "`pkgs/repo`") {
		t.Fatalf("missing dir in %q", prompt)
	}
	if !strings.Contains(prompt, "@pkgs/repo/root.go(28-28)") {
		t.Fatalf("missing function mention in %q", prompt)
	}
}

func TestApplyFunctionBindingsToPayload_missingRequired(t *testing.T) {
	t.Parallel()
	raw := json.RawMessage(`{
		"title":"Scoped",
		"priority":"medium",
		"status":"ready",
		"initial_prompt":"x",
		"checklist_items":[{"text":"done","verifier":"human"}],
		"function_inputs":[{"id":"scope","kind":"file","label":"File"}]
	}`)
	_, err := applyFunctionBindingsToPayload(raw, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "missing binding") {
		t.Fatalf("err = %v", err)
	}
}
