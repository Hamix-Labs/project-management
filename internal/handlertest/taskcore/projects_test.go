package taskcore_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/handlertest"

	projectsdomain "github.com/AlexsanderHamir/Hamix/pkgs/projects/domain"
)

func TestHTTP_projectsCRUDAndContext(t *testing.T) {
	srv := handlertest.NewCreateServer(t)
	defer srv.Close()
	git := handlertest.MustGitBinding(t, srv.URL)

	res, err := http.Post(srv.URL+"/projects", "application/json", strings.NewReader(`{"name":"Moat","description":"Long work","context_summary":"Shared memory","repository_id":"`+git.RepositoryID+`"}`))
	if err != nil {
		t.Fatal(err)
	}
	projectBytes, err := io.ReadAll(res.Body)
	if cerr := res.Body.Close(); cerr != nil {
		t.Fatal(cerr)
	}
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("create project status %d body %s", res.StatusCode, projectBytes)
	}
	var project projectsdomain.Project
	if err := json.Unmarshal(projectBytes, &project); err != nil {
		t.Fatal(err)
	}
	if project.ID == "" || project.Status != projectsdomain.ProjectStatusActive {
		t.Fatalf("project = %#v", project)
	}
	if project.RepositoryID == nil || *project.RepositoryID != git.RepositoryID {
		t.Fatalf("project repository_id = %#v, want %s", project.RepositoryID, git.RepositoryID)
	}

	itemRes, err := http.Post(srv.URL+"/projects/"+project.ID+"/context", "application/json", strings.NewReader(`{"tag":"requirement","title":"Use relational context","description":"Pick when deciding memory architecture","body":"No vector store in v1","pinned":true}`))
	if err != nil {
		t.Fatal(err)
	}
	itemBytes, err := io.ReadAll(itemRes.Body)
	if cerr := itemRes.Body.Close(); cerr != nil {
		t.Fatal(cerr)
	}
	if err != nil {
		t.Fatal(err)
	}
	if itemRes.StatusCode != http.StatusCreated {
		t.Fatalf("create context status %d body %s", itemRes.StatusCode, itemBytes)
	}
	var item projectsdomain.ProjectContextItem
	if err := json.Unmarshal(itemBytes, &item); err != nil {
		t.Fatal(err)
	}
	if item.ProjectID != project.ID || item.Tag != "requirement" || item.Description != "Pick when deciding memory architecture" || !item.Pinned {
		t.Fatalf("context item = %#v", item)
	}
	secondItemRes, err := http.Post(srv.URL+"/projects/"+project.ID+"/context", "application/json", strings.NewReader(`{"tag":"Payment rules","title":"Explicit selection","body":"Tasks choose context nodes."}`))
	if err != nil {
		t.Fatal(err)
	}
	secondItemBytes, err := io.ReadAll(secondItemRes.Body)
	if cerr := secondItemRes.Body.Close(); cerr != nil {
		t.Fatal(cerr)
	}
	if err != nil {
		t.Fatal(err)
	}
	if secondItemRes.StatusCode != http.StatusCreated {
		t.Fatalf("create second context status %d body %s", secondItemRes.StatusCode, secondItemBytes)
	}
	var secondItem projectsdomain.ProjectContextItem
	if err := json.Unmarshal(secondItemBytes, &secondItem); err != nil {
		t.Fatal(err)
	}

	listRes, err := http.Get(srv.URL + "/projects/" + project.ID + "/context")
	if err != nil {
		t.Fatal(err)
	}
	defer listRes.Body.Close()
	if listRes.StatusCode != http.StatusOK {
		t.Fatalf("context list status %d", listRes.StatusCode)
	}
	var list struct {
		Items []projectsdomain.ProjectContextItem `json:"items"`
		Edges []projectsdomain.ProjectContextEdge `json:"edges"`
	}
	if err := json.NewDecoder(listRes.Body).Decode(&list); err != nil {
		t.Fatal(err)
	}
	if len(list.Items) != 2 || list.Items[0].ID != item.ID || list.Items[1].ID != secondItem.ID {
		t.Fatalf("items = %#v", list.Items)
	}
	if list.Edges == nil || len(list.Edges) != 0 {
		t.Fatalf("edges = %#v, want empty array", list.Edges)
	}

	req, err := http.NewRequest(http.MethodPatch, srv.URL+"/projects/"+project.ID, strings.NewReader(`{"status":"archived"}`))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	patchRes, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer patchRes.Body.Close()
	if patchRes.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(patchRes.Body)
		t.Fatalf("patch project status %d body %s", patchRes.StatusCode, b)
	}
	var updated projectsdomain.Project
	if err := json.NewDecoder(patchRes.Body).Decode(&updated); err != nil {
		t.Fatal(err)
	}
	if updated.Status != projectsdomain.ProjectStatusArchived {
		t.Fatalf("updated status = %q", updated.Status)
	}
}

func TestHTTP_projectContextListEmptyUsesJSONArrays(t *testing.T) {
	srv := handlertest.NewCreateServer(t)
	defer srv.Close()
	git := handlertest.MustGitBinding(t, srv.URL)

	project := postProjectJSON(t, srv, `{"name":"Empty context","repository_id":"`+git.RepositoryID+`"}`, http.StatusCreated)

	listRes, err := http.Get(srv.URL + "/projects/" + project.ID + "/context")
	if err != nil {
		t.Fatal(err)
	}
	defer listRes.Body.Close()
	body, err := io.ReadAll(listRes.Body)
	if err != nil {
		t.Fatal(err)
	}
	if listRes.StatusCode != http.StatusOK {
		t.Fatalf("context list status %d body %s", listRes.StatusCode, body)
	}

	var top map[string]json.RawMessage
	if err := json.Unmarshal(body, &top); err != nil {
		t.Fatal(err)
	}
	if string(top["items"]) != "[]" {
		t.Fatalf("items = %s, want []", top["items"])
	}
	if string(top["edges"]) != "[]" {
		t.Fatalf("edges = %s, want []", top["edges"])
	}
}

func TestHTTP_taskProjectIDCreatePatchAndClear(t *testing.T) {
	srv := handlertest.NewCreateServer(t)
	defer srv.Close()
	git := handlertest.MustGitBinding(t, srv.URL)

	project := postProjectJSON(t, srv, `{"name":"Project owner","repository_id":"`+git.RepositoryID+`"}`, http.StatusCreated)
	task := handlertest.PostTaskJSON(t, srv, `{"title":"linked","priority":"medium","project_id":"`+project.ID+`","repository_id":"`+git.RepositoryID+`"}`, http.StatusCreated)
	if task.ProjectID == nil || *task.ProjectID != project.ID {
		t.Fatalf("created task project_id = %#v, want %s", task.ProjectID, project.ID)
	}
	if task.Number == nil || *task.Number < 1 {
		t.Fatalf("created task number = %#v, want >= 1", task.Number)
	}

	// Numbered tasks cannot clear project_id (immutable #N identity).
	req, err := http.NewRequest(http.MethodPatch, srv.URL+"/tasks/"+task.ID, strings.NewReader(`{"project_id":null}`))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("patch clear project status %d body %s, want 400", res.StatusCode, b)
	}
}

func postProjectJSON(t *testing.T, srv *httptest.Server, body string, want int) projectsdomain.Project {
	t.Helper()
	res, err := http.Post(srv.URL+"/projects", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != want {
		t.Fatalf("POST /projects status %d body %s", res.StatusCode, b)
	}
	var project projectsdomain.Project
	if err := json.Unmarshal(b, &project); err != nil {
		t.Fatal(err)
	}
	return project
}
