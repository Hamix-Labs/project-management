package handler

import (
	"context"
	settingscontract "github.com/AlexsanderHamir/Hamix/pkgs/settings/contract"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/AlexsanderHamir/Hamix/internal/taskapi/composition"
	"github.com/AlexsanderHamir/Hamix/internal/tasktestdb"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
)

func TestUserCreatedTaskEnqueuesForAgents(t *testing.T) {
	t.Parallel()
	db := tasktestdb.OpenSQLite(t)
	q := agents.NewMemoryQueue(8)
	st := composition.NewAPI(db)
	st.SetReadyTaskNotifier(q)
	// Disable the global pickup delay; this test exercises the
	// immediate-notify path, not the deferred path. Without this,
	// the post-Stage-0 gate (shouldNotifyReadyNow) correctly defers
	// the task by DefaultAgentPickupDelaySeconds and the Recv loop
	// times out.
	zero := 0
	if _, err := st.UpdateSettings(context.Background(), settingscontract.SettingsPatch{AgentPickupDelaySeconds: &zero}); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}
	srv := newTaskCreateTestServerFromStore(t, st)
	t.Cleanup(srv.Close)

	res, err := http.Post(srv.URL+"/tasks", "application/json", strings.NewReader(withCreateChecklistForURL(srv.URL, `{"title":"from-user","priority":"medium"}`)))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("status %d", res.StatusCode)
	}

	select {
	case got := <-q.Recv():
		q.AckAfterRecv(got.ID)
		if got.Title != "from-user" {
			t.Fatalf("title %q", got.Title)
		}
		if got.Priority != taskcoredomain.PriorityMedium {
			t.Fatalf("priority %s", got.Priority)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for queued task")
	}
}

func TestAgentActorCreateEnqueuesWhenReady(t *testing.T) {
	t.Parallel()
	db := tasktestdb.OpenSQLite(t)
	q := agents.NewMemoryQueue(8)
	st := composition.NewAPI(db)
	st.SetReadyTaskNotifier(q)
	zero := 0
	if _, err := st.UpdateSettings(context.Background(), settingscontract.SettingsPatch{AgentPickupDelaySeconds: &zero}); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}
	srv := newTaskCreateTestServerFromStore(t, st)
	t.Cleanup(srv.Close)

	req, err := http.NewRequest(http.MethodPost, srv.URL+"/tasks", strings.NewReader(withCreateChecklistForURL(srv.URL, `{"title":"from-agent","priority":"medium"}`)))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Actor", "agent")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("status %d", res.StatusCode)
	}

	select {
	case got := <-q.Recv():
		q.AckAfterRecv(got.ID)
		if got.Title != "from-agent" {
			t.Fatalf("title %q", got.Title)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for queued task")
	}
}
