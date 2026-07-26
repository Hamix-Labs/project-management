package taskcore_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/AlexsanderHamir/Hamix/internal/handlertest"
)

func TestHTTP_projectRoutesPublishSSE(t *testing.T) {
	srv, _, hub := handlertest.NewSSETriggerServer(t)
	defer srv.Close()

	ch, cancel := hub.Subscribe()
	defer cancel()
	git := handlertest.MustGitBinding(t, srv.URL)
	project := postProjectJSON(t, srv, `{"name":"SSE project","repository_id":"`+git.RepositoryID+`"}`, http.StatusCreated)
	got := handlertest.SummarizeSSEEvents(handlertest.DrainSSE(t, ch, 1, 2*time.Second))
	handlertest.MustEqualEvents(t, "POST /projects", got, []string{"project_created:" + project.ID})

	handlertest.MustDoJSON(t, http.MethodPost, srv.URL+"/projects/"+project.ID+"/context",
		`{"tag":"General","title":"Pinned","body":"Context","pinned":true}`, "", http.StatusCreated)
	got = handlertest.SummarizeSSEEvents(handlertest.DrainSSE(t, ch, 1, 2*time.Second))
	handlertest.MustEqualEvents(t, "POST /projects/{id}/context", got, []string{"project_updated:" + project.ID})

	handlertest.MustDoJSON(t, http.MethodPatch, srv.URL+"/projects/"+project.ID,
		`{"description":"Updated"}`, "", http.StatusOK)
	got = handlertest.SummarizeSSEEvents(handlertest.DrainSSE(t, ch, 1, 2*time.Second))
	handlertest.MustEqualEvents(t, "PATCH /projects/{id}", got, []string{"project_updated:" + project.ID})
}
