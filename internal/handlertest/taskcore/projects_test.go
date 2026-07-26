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

func TestHTTP_projectsCRUD(t *testing.T) {
	srv := handlertest.NewCreateServer(t)
	defer srv.Close()
	git := handlertest.MustGitBinding(t, srv.URL)

	res, err := http.Post(srv.URL+"/projects", "application/json", strings.NewReader(`{"name":"Moat","description":"Long work","repository_id":"`+git.RepositoryID+`"}`))
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

	listRes, err := http.Get(srv.URL + "/projects")
	if err != nil {
		t.Fatal(err)
	}
	defer listRes.Body.Close()
	if listRes.StatusCode != http.StatusOK {
		t.Fatalf("list projects status %d", listRes.StatusCode)
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

	getRes, err := http.Get(srv.URL + "/projects/" + project.ID)
	if err != nil {
		t.Fatal(err)
	}
	defer getRes.Body.Close()
	if getRes.StatusCode != http.StatusOK {
		t.Fatalf("get project status %d", getRes.StatusCode)
	}

	// Context memory routes are removed.
	ctxRes, err := http.Get(srv.URL + "/projects/" + project.ID + "/context")
	if err != nil {
		t.Fatal(err)
	}
	defer ctxRes.Body.Close()
	if ctxRes.StatusCode != http.StatusNotFound {
		t.Fatalf("context list status %d, want 404", ctxRes.StatusCode)
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
