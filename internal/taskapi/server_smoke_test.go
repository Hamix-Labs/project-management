package taskapi_test

import (
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/realtime"

	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/taskapi"
	"github.com/AlexsanderHamir/Hamix/internal/taskapi/composition"
	"github.com/AlexsanderHamir/Hamix/internal/tasktestdb"
	"github.com/AlexsanderHamir/Hamix/pkgs/draftassist/runner"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/postgres"
)

// Smoke test for the production HTTP assembly path (middleware.Stack +
// handler.NewHandler via taskapi.NewHTTPHandler).
func TestNewHTTPHandler_healthAndBootstrap(t *testing.T) {
	db := tasktestdb.OpenSQLite(t)
	st := composition.NewAPI(db)
	hub := realtime.NewSSEHub()
	api, closeHTTP := taskapi.NewHTTPHandler(st, hub, nil, nil, postgres.SchemaDriftReport{
		Status:       postgres.SchemaDriftOK,
		CodeRevision: postgres.SchemaRevision,
		DBRevision:   postgres.SchemaRevision,
	}, taskapi.DraftAssistHost{
		Runner: runner.NewFake(runner.FakeOptions{}),
	})
	srv := httptest.NewServer(api)
	t.Cleanup(func() {
		srv.Close()
		if closeHTTP != nil {
			_ = closeHTTP()
		}
	})

	t.Run("health", func(t *testing.T) {
		res, err := http.Get(srv.URL + "/health")
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			t.Fatalf("status %d", res.StatusCode)
		}
		var body struct {
			Status  string `json:"status"`
			Version string `json:"version"`
		}
		if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body.Status != "ok" || body.Version == "" {
			t.Fatalf("body %+v", body)
		}
	})

	t.Run("bootstrap", func(t *testing.T) {
		res, err := http.Get(srv.URL + "/v1/bootstrap")
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			t.Fatalf("status %d", res.StatusCode)
		}
		var body map[string]json.RawMessage
		if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		for _, key := range []string{"settings", "tasks", "stats", "projects", "drafts"} {
			if _, ok := body[key]; !ok {
				t.Fatalf("bootstrap missing key %q", key)
			}
		}
	})
}
