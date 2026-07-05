package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func templateSaveBody(title string) string {
	return `{"name":"` + title + `","payload":` + withCreateChecklist(`{"title":"`+title+`","priority":"medium"}`) + `}`
}

func TestHTTP_task_templates_crud(t *testing.T) {
	srv := newTaskTestServer(t)
	defer srv.Close()

	saveRes, err := http.Post(srv.URL+"/task-templates", "application/json", strings.NewReader(templateSaveBody("Template one")))
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

func TestHTTP_task_templates_search_q(t *testing.T) {
	srv, st := newTaskTestServerWithStore(t)
	defer srv.Close()
	ctx := context.Background()
	payload := []byte(withCreateChecklist(`{"title":"Alpha task","priority":"medium"}`))
	if _, err := st.SaveTemplate(ctx, "", "Alpha task", payload); err != nil {
		t.Fatal(err)
	}
	payload2 := []byte(withCreateChecklist(`{"title":"Beta task","priority":"medium"}`))
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
	srv, st := newTaskTestServerWithStore(t)
	defer srv.Close()
	ctx := context.Background()
	past := time.Now().UTC().Add(-24 * time.Hour).Format(time.RFC3339)
	payload := []byte(withCreateChecklist(`{
		"title":"From template",
		"priority":"medium",
		"pickup_not_before":"` + past + `",
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
	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("status %d body %s", res.StatusCode, b)
	}
	var body struct {
		Tasks []struct {
			Title           string  `json:"title"`
			PickupNotBefore *string `json:"pickup_not_before"`
		} `json:"tasks"`
		Errors []struct {
			TemplateID string `json:"template_id"`
			Error      string `json:"error"`
		} `json:"errors"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if len(body.Tasks) != 1 {
		t.Fatalf("tasks %d want 1", len(body.Tasks))
	}
	if body.Tasks[0].Title != "From template" {
		t.Fatalf("title %q", body.Tasks[0].Title)
	}
	if body.Tasks[0].PickupNotBefore != nil {
		t.Fatalf("past pickup_not_before should be omitted, got %v", body.Tasks[0].PickupNotBefore)
	}
	if len(body.Errors) != 1 || body.Errors[0].TemplateID != "missing-id" {
		t.Fatalf("errors %+v", body.Errors)
	}
}

func TestHTTP_task_templates_instantiate_empty_ids(t *testing.T) {
	srv := newTaskTestServer(t)
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
	srv, st := newTaskTestServerWithStore(t)
	defer srv.Close()
	ctx := context.Background()
	payload := []byte(withCreateChecklist(`{"title":"Repeated template","priority":"medium"}`))
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
	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("status %d body %s", res.StatusCode, b)
	}
	var body struct {
		Tasks []struct {
			Title string `json:"title"`
		} `json:"tasks"`
		Errors []any `json:"errors"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if len(body.Tasks) != 5 {
		t.Fatalf("tasks %d want 5", len(body.Tasks))
	}
	for _, task := range body.Tasks {
		if task.Title != "Repeated template" {
			t.Fatalf("title %q", task.Title)
		}
	}
	if len(body.Errors) != 0 {
		t.Fatalf("errors %+v", body.Errors)
	}
}

func TestHTTP_task_templates_instantiate_items_mixed_counts(t *testing.T) {
	srv, st := newTaskTestServerWithStore(t)
	defer srv.Close()
	ctx := context.Background()
	payloadA := []byte(withCreateChecklist(`{"title":"Template A","priority":"medium"}`))
	tmplA, err := st.SaveTemplate(ctx, "", "Template A", payloadA)
	if err != nil {
		t.Fatal(err)
	}
	payloadB := []byte(withCreateChecklist(`{"title":"Template B","priority":"medium"}`))
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
	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("status %d body %s", res.StatusCode, b)
	}
	var body struct {
		Tasks []struct {
			Title string `json:"title"`
		} `json:"tasks"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if len(body.Tasks) != 5 {
		t.Fatalf("tasks %d want 5", len(body.Tasks))
	}
	if body.Tasks[0].Title != "Template A" || body.Tasks[2].Title != "Template A" {
		t.Fatalf("first three titles want Template A, got %q %q %q", body.Tasks[0].Title, body.Tasks[1].Title, body.Tasks[2].Title)
	}
	if body.Tasks[3].Title != "Template B" || body.Tasks[4].Title != "Template B" {
		t.Fatalf("last two titles want Template B, got %q %q", body.Tasks[3].Title, body.Tasks[4].Title)
	}
}

func TestHTTP_task_templates_instantiate_invalid_count(t *testing.T) {
	srv, st := newTaskTestServerWithStore(t)
	defer srv.Close()
	ctx := context.Background()
	payload := []byte(withCreateChecklist(`{"title":"T","priority":"medium"}`))
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
	srv, st := newTaskTestServerWithStore(t)
	defer srv.Close()
	ctx := context.Background()
	payload := []byte(withCreateChecklist(`{"title":"T","priority":"medium"}`))
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
	srv, st := newTaskTestServerWithStore(t)
	defer srv.Close()
	ctx := context.Background()
	payload := []byte(withCreateChecklist(`{"title":"T","priority":"medium"}`))
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
	srv := newTaskTestServer(t)
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
	srv, st := newTaskTestServerWithStore(t)
	defer srv.Close()
	ctx := context.Background()
	payload := []byte(withCreateChecklist(`{"title":"Tagged","priority":"medium","tags":["Refactor","Docs"]}`))
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
	srv, st := newTaskTestServerWithStore(t)
	defer srv.Close()
	ctx := context.Background()
	for _, name := range []string{"Zebra tpl", "Alpha tpl"} {
		payload := []byte(withCreateChecklist(`{"title":"` + name + `","priority":"medium"}`))
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
	srv, st := newTaskTestServerWithStore(t)
	defer srv.Close()
	ctx := context.Background()
	refactorPayload := []byte(withCreateChecklist(`{"title":"Refactor task","priority":"medium","tags":["Refactor"]}`))
	if _, err := st.SaveTemplate(ctx, "", "Refactor task", refactorPayload); err != nil {
		t.Fatal(err)
	}
	bugPayload := []byte(withCreateChecklist(`{"title":"Bug task","priority":"medium","tags":["Bugfix"]}`))
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
	srv := newTaskTestServer(t)
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
	srv, st := newTaskTestServerWithStore(t)
	defer srv.Close()
	ctx := context.Background()
	payload := []byte(withCreateChecklist(`{"title":"Counter tpl","priority":"medium"}`))
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
	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("status %d body %s", res.StatusCode, b)
	}

	listRes, err := http.Get(srv.URL + "/task-templates")
	if err != nil {
		t.Fatal(err)
	}
	defer listRes.Body.Close()
	var body struct {
		Templates []struct {
			ID               string `json:"id"`
			InstantiateCount int    `json:"instantiate_count"`
		} `json:"templates"`
	}
	if err := json.NewDecoder(listRes.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if len(body.Templates) != 1 {
		t.Fatalf("got %d templates want 1", len(body.Templates))
	}
	if body.Templates[0].ID != tmpl.ID {
		t.Fatalf("id %q want %q", body.Templates[0].ID, tmpl.ID)
	}
	if body.Templates[0].InstantiateCount != 3 {
		t.Fatalf("instantiate_count %d want 3", body.Templates[0].InstantiateCount)
	}
}
