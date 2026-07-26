package compose_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/handlertest"
)

func TestHTTP_task_templates_function_instantiate(t *testing.T) {
	srv, st := handlertest.NewCreateServerWithStore(t)
	defer srv.Close()
	ctx := context.Background()
	payload := []byte(handlertest.WithComposeChecklistForURL(srv.URL, `{
		"title":"Function template",
		"priority":"medium",
		"initial_prompt":"Refactor carefully.",
		"function_inputs":[{"id":"scope","kind":"dir","label":"Directory"}]
	}`))
	tmpl, err := st.SaveTemplate(ctx, "", "Function template", payload)
	if err != nil {
		t.Fatal(err)
	}
	if !tmpl.IsFunction {
		t.Fatal("expected is_function on save summary")
	}
	if len(tmpl.InputKinds) != 1 || tmpl.InputKinds[0] != "dir" {
		t.Fatalf("input_kinds = %v", tmpl.InputKinds)
	}

	miss, err := http.Post(srv.URL+"/task-templates/instantiate", "application/json",
		strings.NewReader(`{"items":[{"template_id":"`+tmpl.ID+`","count":1}]}`))
	if err != nil {
		t.Fatal(err)
	}
	missBody, _ := io.ReadAll(miss.Body)
	_ = miss.Body.Close()
	if miss.StatusCode != http.StatusOK {
		t.Fatalf("missing binding status %d body %s", miss.StatusCode, missBody)
	}
	var missResp struct {
		Tasks  []any `json:"tasks"`
		Errors []struct {
			TemplateID string `json:"template_id"`
			Error      string `json:"error"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(missBody, &missResp); err != nil {
		t.Fatal(err)
	}
	if len(missResp.Tasks) != 0 || len(missResp.Errors) != 1 {
		t.Fatalf("want 0 tasks 1 error, got %+v", missResp)
	}

	okRes, err := http.Post(srv.URL+"/task-templates/instantiate", "application/json",
		strings.NewReader(`{"items":[{"template_id":"`+tmpl.ID+`","count":1,"function_bindings":[{"input_id":"scope","paths":["pkgs/repo"]}]}]}`))
	if err != nil {
		t.Fatal(err)
	}
	okBody, _ := io.ReadAll(okRes.Body)
	_ = okRes.Body.Close()
	if okRes.StatusCode != http.StatusOK {
		t.Fatalf("status %d body %s", okRes.StatusCode, okBody)
	}
	var okResp struct {
		Tasks []struct {
			ID            string `json:"id"`
			InitialPrompt string `json:"initial_prompt"`
		} `json:"tasks"`
		Errors []any `json:"errors"`
	}
	if err := json.Unmarshal(okBody, &okResp); err != nil {
		t.Fatal(err)
	}
	if len(okResp.Tasks) != 1 || len(okResp.Errors) != 0 {
		t.Fatalf("resp %+v", okResp)
	}
	if !strings.Contains(okResp.Tasks[0].InitialPrompt, "## Scope (do not expand beyond)") {
		t.Fatalf("prompt missing scope: %q", okResp.Tasks[0].InitialPrompt)
	}
	if !strings.Contains(okResp.Tasks[0].InitialPrompt, "`pkgs/repo`") {
		t.Fatalf("prompt missing dir: %q", okResp.Tasks[0].InitialPrompt)
	}
}
