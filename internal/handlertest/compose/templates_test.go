package compose_test

import (
	"context"
	"encoding/json"
	"github.com/AlexsanderHamir/Hamix/internal/handlertest"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func templateSaveBody(baseURL, title string) string {
	return `{"name":"` + title + `","payload":` + handlertest.WithComposeChecklistForURL(baseURL, `{"title":"`+title+`","priority":"medium"}`) + `}`
}

func TestHTTP_task_templates_crud(t *testing.T) {
	srv := handlertest.NewCreateServer(t)
	defer srv.Close()

	saveRes, err := http.Post(srv.URL+"/task-templates", "application/json", strings.NewReader(templateSaveBody(srv.URL, "Template one")))
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(saveRes.Body)
	_ = saveRes.Body.Close()
	if saveRes.StatusCode != http.StatusCreated {
		t.Fatalf("save status %d body %s", saveRes.StatusCode, body)
	}
	var saved struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(body, &saved); err != nil {
		t.Fatal(err)
	}
	if saved.ID == "" {
		t.Fatal("missing template id")
	}
	if saved.Name != "Template one" {
		t.Fatalf("name %q", saved.Name)
	}

	listRes, err := http.Get(srv.URL + "/task-templates")
	if err != nil {
		t.Fatal(err)
	}
	defer listRes.Body.Close()
	if listRes.StatusCode != http.StatusOK {
		t.Fatalf("list status %d", listRes.StatusCode)
	}

	getRes, err := http.Get(srv.URL + "/task-templates/" + saved.ID)
	if err != nil {
		t.Fatal(err)
	}
	getBody, _ := io.ReadAll(getRes.Body)
	_ = getRes.Body.Close()
	if getRes.StatusCode != http.StatusOK {
		t.Fatalf("get status %d body %s", getRes.StatusCode, getBody)
	}

	patchReq, _ := http.NewRequest(http.MethodPatch, srv.URL+"/task-templates/"+saved.ID, strings.NewReader(`{"name":"Renamed"}`))
	patchReq.Header.Set("Content-Type", "application/json")
	patchRes, err := http.DefaultClient.Do(patchReq)
	if err != nil {
		t.Fatal(err)
	}
	patchBody, _ := io.ReadAll(patchRes.Body)
	_ = patchRes.Body.Close()
	if patchRes.StatusCode != http.StatusOK {
		t.Fatalf("patch status %d body %s", patchRes.StatusCode, patchBody)
	}

	delReq, _ := http.NewRequest(http.MethodDelete, srv.URL+"/task-templates/"+saved.ID, nil)
	delRes, err := http.DefaultClient.Do(delReq)
	if err != nil {
		t.Fatal(err)
	}
	_ = delRes.Body.Close()
	if delRes.StatusCode != http.StatusNoContent {
		t.Fatalf("delete status %d", delRes.StatusCode)
	}
}

func TestHTTP_task_templates_patch_full_payload(t *testing.T) {
	srv := handlertest.NewCreateServer(t)
	defer srv.Close()

	saveRes, err := http.Post(srv.URL+"/task-templates", "application/json", strings.NewReader(templateSaveBody(srv.URL, "Patch payload template")))
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(saveRes.Body)
	_ = saveRes.Body.Close()
	if saveRes.StatusCode != http.StatusCreated {
		t.Fatalf("save status %d body %s", saveRes.StatusCode, body)
	}
	var saved struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body, &saved); err != nil {
		t.Fatal(err)
	}

	patchPayload := handlertest.WithComposeChecklistForURL(srv.URL, `{
		"title":"Patch payload template",
		"priority":"high",
		"status":"ready",
		"initial_prompt":"<p>Find a module with poor test coverage.</p>"
	}`)
	patchBody := `{"name":"Patch payload template","payload":` + patchPayload + `}`
	patchReq, _ := http.NewRequest(http.MethodPatch, srv.URL+"/task-templates/"+saved.ID, strings.NewReader(patchBody))
	patchReq.Header.Set("Content-Type", "application/json")
	patchRes, err := http.DefaultClient.Do(patchReq)
	if err != nil {
		t.Fatal(err)
	}
	patchRespBody, _ := io.ReadAll(patchRes.Body)
	_ = patchRes.Body.Close()
	if patchRes.StatusCode != http.StatusOK {
		t.Fatalf("patch status %d body %s", patchRes.StatusCode, patchRespBody)
	}
}

func TestHTTP_task_templates_save_requires_repository_id(t *testing.T) {
	srv := handlertest.NewCreateServer(t)
	defer srv.Close()
	binding, ok := handlertest.GitBindingForURL(srv.URL)
	if !ok {
		t.Fatal("missing git binding")
	}
	res, err := http.Post(srv.URL+"/task-templates", "application/json",
		strings.NewReader(`{"name":"Missing repo","payload":{"title":"Missing repo","priority":"medium","project_id":"`+binding.ProjectID+`","checklist_items":[{"text":"`+handlertest.TestCriterionText+`"}]}}`))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("status %d body %s want 400", res.StatusCode, b)
	}
}

func TestHTTP_task_templates_search_q(t *testing.T) {
	srv, st := handlertest.NewCreateServerWithStore(t)
	defer srv.Close()
	ctx := context.Background()
	payload := []byte(handlertest.WithComposeChecklistForURL(srv.URL, `{"title":"Alpha task","priority":"medium"}`))
	if _, err := st.SaveTemplate(ctx, "", "Alpha task", payload); err != nil {
		t.Fatal(err)
	}
	payload2 := []byte(handlertest.WithComposeChecklistForURL(srv.URL, `{"title":"Beta task","priority":"medium"}`))
	if _, err := st.SaveTemplate(ctx, "", "Beta task", payload2); err != nil {
		t.Fatal(err)
	}

	res, err := http.Get(srv.URL + "/task-templates?q=alpha")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status %d", res.StatusCode)
	}
	var body struct {
		Templates []struct {
			Name string `json:"name"`
		} `json:"templates"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if len(body.Templates) != 1 {
		t.Fatalf("got %d templates want 1", len(body.Templates))
	}
	if body.Templates[0].Name != "Alpha task" {
		t.Fatalf("name %q", body.Templates[0].Name)
	}
}

func TestHTTP_task_templates_instantiate(t *testing.T) {
	srv, st := handlertest.NewCreateServerWithStore(t)
	defer srv.Close()
	ctx := context.Background()
	past := time.Now().UTC().Add(-24 * time.Hour).Format(time.RFC3339)
	payload := []byte(handlertest.WithComposeChecklistForURL(srv.URL, `{
		"title":"From template",
		"priority":"medium",
		"pickup_not_before":"`+past+`",
		"depends_on":[{"task_id":"00000000-0000-4000-8000-000000000001","type":"finish_to_start"}]
	}`))
	tmpl, err := st.SaveTemplate(ctx, "", "From template", payload)
	if err != nil {
		t.Fatal(err)
	}

	res, err := http.Post(srv.URL+"/task-templates/instantiate", "application/json",
		strings.NewReader(`{"template_ids":["`+tmpl.ID+`","missing-id"]}`))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusAccepted {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("status %d body %s", res.StatusCode, b)
	}
	var body struct {
		Accepted bool `json:"accepted"`
		Total    int  `json:"total"`
		Errors   []struct {
			TemplateID string `json:"template_id"`
			Error      string `json:"error"`
		} `json:"errors"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if !body.Accepted || body.Total != 1 {
		t.Fatalf("accepted=%v total=%d want true/1", body.Accepted, body.Total)
	}
	if len(body.Errors) != 1 || body.Errors[0].TemplateID != "missing-id" {
		t.Fatalf("errors %+v", body.Errors)
	}
	tasks := handlertest.WaitForTaskTitleCount(t, st, "From template", 1)
	if tasks[0].PickupNotBefore != nil {
		t.Fatalf("past pickup_not_before should be omitted, got %v", tasks[0].PickupNotBefore)
	}
}

func TestHTTP_task_templates_instantiate_empty_ids(t *testing.T) {
	srv := handlertest.NewCreateServer(t)
	defer srv.Close()
	res, err := http.Post(srv.URL+"/task-templates/instantiate", "application/json",
		strings.NewReader(`{"template_ids":[]}`))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("status %d want 400", res.StatusCode)
	}
}

func TestHTTP_task_templates_instantiate_count(t *testing.T) {
	srv, st := handlertest.NewCreateServerWithStore(t)
	defer srv.Close()
	ctx := context.Background()
	payload := []byte(handlertest.WithComposeChecklistForURL(srv.URL, `{"title":"Repeated template","priority":"medium"}`))
	tmpl, err := st.SaveTemplate(ctx, "", "Repeated template", payload)
	if err != nil {
		t.Fatal(err)
	}

	res, err := http.Post(srv.URL+"/task-templates/instantiate", "application/json",
		strings.NewReader(`{"template_ids":["`+tmpl.ID+`"],"count":5}`))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusAccepted {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("status %d body %s", res.StatusCode, b)
	}
	var body struct {
		Accepted bool  `json:"accepted"`
		Total    int   `json:"total"`
		Errors   []any `json:"errors"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if !body.Accepted || body.Total != 5 || len(body.Errors) != 0 {
		t.Fatalf("body %+v", body)
	}
	handlertest.WaitForTaskTitleCount(t, st, "Repeated template", 5)
}

func TestHTTP_task_templates_instantiate_items_mixed_counts(t *testing.T) {
	srv, st := handlertest.NewCreateServerWithStore(t)
	defer srv.Close()
	ctx := context.Background()
	payloadA := []byte(handlertest.WithComposeChecklistForURL(srv.URL, `{"title":"Template A","priority":"medium"}`))
	tmplA, err := st.SaveTemplate(ctx, "", "Template A", payloadA)
	if err != nil {
		t.Fatal(err)
	}
	payloadB := []byte(handlertest.WithComposeChecklistForURL(srv.URL, `{"title":"Template B","priority":"medium"}`))
	tmplB, err := st.SaveTemplate(ctx, "", "Template B", payloadB)
	if err != nil {
		t.Fatal(err)
	}

	res, err := http.Post(srv.URL+"/task-templates/instantiate", "application/json",
		strings.NewReader(`{"items":[{"template_id":"`+tmplA.ID+`","count":3},{"template_id":"`+tmplB.ID+`","count":2}]}`))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusAccepted {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("status %d body %s", res.StatusCode, b)
	}
	var body struct {
		Accepted bool `json:"accepted"`
		Total    int  `json:"total"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if !body.Accepted || body.Total != 5 {
		t.Fatalf("accepted=%v total=%d", body.Accepted, body.Total)
	}
	handlertest.WaitForTaskTitleCount(t, st, "Template A", 3)
	handlertest.WaitForTaskTitleCount(t, st, "Template B", 2)
}

func TestHTTP_task_templates_instantiate_invalid_count(t *testing.T) {
	srv, st := handlertest.NewCreateServerWithStore(t)
	defer srv.Close()
	ctx := context.Background()
	payload := []byte(handlertest.WithComposeChecklistForURL(srv.URL, `{"title":"T","priority":"medium"}`))
	tmpl, err := st.SaveTemplate(ctx, "", "T", payload)
	if err != nil {
		t.Fatal(err)
	}

	for _, bodyJSON := range []string{
		`{"template_ids":["` + tmpl.ID + `"],"count":0}`,
		`{"template_ids":["` + tmpl.ID + `"],"count":26}`,
	} {
		res, err := http.Post(srv.URL+"/task-templates/instantiate", "application/json", strings.NewReader(bodyJSON))
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
		if res.StatusCode != http.StatusBadRequest {
			t.Fatalf("body %s status %d want 400", bodyJSON, res.StatusCode)
		}
	}
}

func TestHTTP_task_templates_instantiate_total_cap_exceeded(t *testing.T) {
	srv, st := handlertest.NewCreateServerWithStore(t)
	defer srv.Close()
	ctx := context.Background()
	payload := []byte(handlertest.WithComposeChecklistForURL(srv.URL, `{"title":"T","priority":"medium"}`))
	tmpl, err := st.SaveTemplate(ctx, "", "T", payload)
	if err != nil {
		t.Fatal(err)
	}

	res, err := http.Post(srv.URL+"/task-templates/instantiate", "application/json",
		strings.NewReader(`{"template_ids":["`+tmpl.ID+`"],"count":101}`))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("status %d want 400", res.StatusCode)
	}
}

func TestHTTP_task_templates_instantiate_duplicate_items(t *testing.T) {
	srv, st := handlertest.NewCreateServerWithStore(t)
	defer srv.Close()
	ctx := context.Background()
	payload := []byte(handlertest.WithComposeChecklistForURL(srv.URL, `{"title":"T","priority":"medium"}`))
	tmpl, err := st.SaveTemplate(ctx, "", "T", payload)
	if err != nil {
		t.Fatal(err)
	}

	res, err := http.Post(srv.URL+"/task-templates/instantiate", "application/json",
		strings.NewReader(`{"items":[{"template_id":"`+tmpl.ID+`","count":2},{"template_id":"`+tmpl.ID+`","count":1}]}`))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("status %d want 400", res.StatusCode)
	}
}

func TestHTTP_task_templates_save_requires_valid_payload(t *testing.T) {
	srv := handlertest.NewCreateServer(t)
	defer srv.Close()
	res, err := http.Post(srv.URL+"/task-templates", "application/json",
		strings.NewReader(`{"payload":{"title":"   ","priority":"medium","checklist_items":[{"text":"x"}]}}`))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("status %d body %s", res.StatusCode, b)
	}
}

func TestHTTP_task_templates_list_primary_tag(t *testing.T) {
	srv, st := handlertest.NewCreateServerWithStore(t)
	defer srv.Close()
	ctx := context.Background()
	payload := []byte(handlertest.WithComposeChecklistForURL(srv.URL, `{"title":"Tagged","priority":"medium","tags":["Refactor","Docs"]}`))
	if _, err := st.SaveTemplate(ctx, "", "Tagged", payload); err != nil {
		t.Fatal(err)
	}

	res, err := http.Get(srv.URL + "/task-templates")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status %d", res.StatusCode)
	}
	var body struct {
		Templates []struct {
			PrimaryTag       string `json:"primary_tag"`
			InstantiateCount int    `json:"instantiate_count"`
		} `json:"templates"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if len(body.Templates) != 1 {
		t.Fatalf("got %d templates want 1", len(body.Templates))
	}
	if body.Templates[0].PrimaryTag != "Refactor" {
		t.Fatalf("primary_tag %q want Refactor", body.Templates[0].PrimaryTag)
	}
	if body.Templates[0].InstantiateCount != 0 {
		t.Fatalf("instantiate_count %d want 0", body.Templates[0].InstantiateCount)
	}
}

func TestHTTP_task_templates_list_sort_by_name(t *testing.T) {
	srv, st := handlertest.NewCreateServerWithStore(t)
	defer srv.Close()
	ctx := context.Background()
	for _, name := range []string{"Zebra tpl", "Alpha tpl"} {
		payload := []byte(handlertest.WithComposeChecklistForURL(srv.URL, `{"title":"`+name+`","priority":"medium"}`))
		if _, err := st.SaveTemplate(ctx, "", name, payload); err != nil {
			t.Fatal(err)
		}
	}

	res, err := http.Get(srv.URL + "/task-templates?sort=name&order=asc")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status %d", res.StatusCode)
	}
	var body struct {
		Templates []struct {
			Name string `json:"name"`
		} `json:"templates"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if len(body.Templates) != 2 {
		t.Fatalf("got %d templates want 2", len(body.Templates))
	}
	if body.Templates[0].Name != "Alpha tpl" || body.Templates[1].Name != "Zebra tpl" {
		t.Fatalf("order %q then %q want Alpha tpl then Zebra tpl", body.Templates[0].Name, body.Templates[1].Name)
	}
}

func TestHTTP_task_templates_list_tag_filter(t *testing.T) {
	srv, st := handlertest.NewCreateServerWithStore(t)
	defer srv.Close()
	ctx := context.Background()
	refactorPayload := []byte(handlertest.WithComposeChecklistForURL(srv.URL, `{"title":"Refactor task","priority":"medium","tags":["Refactor"]}`))
	if _, err := st.SaveTemplate(ctx, "", "Refactor task", refactorPayload); err != nil {
		t.Fatal(err)
	}
	bugPayload := []byte(handlertest.WithComposeChecklistForURL(srv.URL, `{"title":"Bug task","priority":"medium","tags":["Bugfix"]}`))
	if _, err := st.SaveTemplate(ctx, "", "Bug task", bugPayload); err != nil {
		t.Fatal(err)
	}

	res, err := http.Get(srv.URL + "/task-templates?tag=refactor")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status %d", res.StatusCode)
	}
	var body struct {
		Templates []struct {
			Name string `json:"name"`
		} `json:"templates"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if len(body.Templates) != 1 {
		t.Fatalf("got %d templates want 1", len(body.Templates))
	}
	if body.Templates[0].Name != "Refactor task" {
		t.Fatalf("name %q want Refactor task", body.Templates[0].Name)
	}
}

func TestHTTP_task_templates_list_invalid_sort(t *testing.T) {
	srv := handlertest.NewCreateServer(t)
	defer srv.Close()
	res, err := http.Get(srv.URL + "/task-templates?sort=title")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("status %d want 400", res.StatusCode)
	}
}

func TestHTTP_task_templates_instantiate_count_increment(t *testing.T) {
	srv, st := handlertest.NewCreateServerWithStore(t)
	defer srv.Close()
	ctx := context.Background()
	payload := []byte(handlertest.WithComposeChecklistForURL(srv.URL, `{"title":"Counter tpl","priority":"medium"}`))
	tmpl, err := st.SaveTemplate(ctx, "", "Counter tpl", payload)
	if err != nil {
		t.Fatal(err)
	}

	res, err := http.Post(srv.URL+"/task-templates/instantiate", "application/json",
		strings.NewReader(`{"template_ids":["`+tmpl.ID+`"],"count":3}`))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusAccepted {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("status %d body %s", res.StatusCode, b)
	}
	handlertest.WaitForTaskTitleCount(t, st, "Counter tpl", 3)
	handlertest.WaitForInstantiateCount(t, srv.URL, tmpl.ID, 3)
}
