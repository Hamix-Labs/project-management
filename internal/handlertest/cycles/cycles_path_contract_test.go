package cycles_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/handlertest"
)

func TestHTTP_cycle_path_segment_caps(t *testing.T) {
	srv := handlertest.NewCreateServer(t)
	defer srv.Close()
	long := strings.Repeat("a", 129)

	t.Run("taskIdTooLong", func(t *testing.T) {
		res, raw := handlertest.DoCyclesRequest(t, http.MethodGet, srv.URL+"/tasks/"+long+"/cycles", "")
		if res.StatusCode != http.StatusBadRequest {
			t.Fatalf("status %d body=%s", res.StatusCode, raw)
		}
		var errBody handlertest.JSONErrorBody
		if err := json.Unmarshal(raw, &errBody); err != nil {
			t.Fatal(err)
		}
		if errBody.Error != "id too long" {
			t.Fatalf("error=%q", errBody.Error)
		}
	})

	t.Run("cycleIdTooLong", func(t *testing.T) {
		taskID := handlertest.MustCreateTaskForCycles(t, srv.URL)
		res, raw := handlertest.DoCyclesRequest(t, http.MethodGet, srv.URL+"/tasks/"+taskID+"/cycles/"+long, "")
		if res.StatusCode != http.StatusBadRequest {
			t.Fatalf("status %d body=%s", res.StatusCode, raw)
		}
		var errBody handlertest.JSONErrorBody
		if err := json.Unmarshal(raw, &errBody); err != nil {
			t.Fatal(err)
		}
		if errBody.Error != "cycle id too long" {
			t.Fatalf("error=%q", errBody.Error)
		}
	})
}

// TestHTTP_getTaskCycle_phase_response_shape pins the envelope and per-phase
// keys returned by GET /tasks/{id}/cycles/{cycleId}.
