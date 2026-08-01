package handlertest

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/AlexsanderHamir/Hamix/internal/taskapi/composition"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
)

// WaitForTaskTitleCount polls ListFlat until at least want tasks match title
// or the timeout elapses (async instantiate).
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only async wait helper."
func WaitForTaskTitleCount(t *testing.T, st *composition.API, title string, want int) []taskcoredomain.Task {
	t.Helper()
	ctx := context.Background()
	deadline := time.Now().Add(5 * time.Second)
	var matched []taskcoredomain.Task
	for time.Now().Before(deadline) {
		tasks, err := st.ListFlat(ctx, 200, 0, nil)
		if err != nil {
			t.Fatal(err)
		}
		matched = matched[:0]
		for _, task := range tasks {
			if task.Title == title {
				matched = append(matched, task)
			}
		}
		if len(matched) >= want {
			return matched
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %d tasks titled %q (got %d)", want, title, len(matched))
	return nil
}

// WaitForInstantiateCount polls template list until instantiate_count reaches want.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only async wait helper."
func WaitForInstantiateCount(t *testing.T, baseURL, templateID string, want int) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		res, err := http.Get(baseURL + "/task-templates")
		if err != nil {
			t.Fatal(err)
		}
		raw, _ := io.ReadAll(res.Body)
		_ = res.Body.Close()
		var body struct {
			Templates []struct {
				ID               string `json:"id"`
				InstantiateCount int    `json:"instantiate_count"`
			} `json:"templates"`
		}
		if err := json.Unmarshal(raw, &body); err != nil {
			t.Fatal(err)
		}
		for _, tmpl := range body.Templates {
			if tmpl.ID == templateID && tmpl.InstantiateCount >= want {
				return
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for instantiate_count >= %d on %s", want, templateID)
}
