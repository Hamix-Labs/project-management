package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/handler"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handler/storefake"
	"github.com/google/uuid"
)

func TestPostTaskRetry_mapsStoreErrorsToHTTPStatus(t *testing.T) {
	t.Parallel()
	taskID := uuid.NewString()

	cases := []struct {
		name       string
		retryErr   error
		wantStatus int
	}{
		{name: "not found", retryErr: domain.ErrNotFound, wantStatus: http.StatusNotFound},
		{name: "invalid input", retryErr: domain.ErrInvalidInput, wantStatus: http.StatusBadRequest},
		{name: "conflict", retryErr: domain.ErrConflict, wantStatus: http.StatusConflict},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			fake := storefake.NewTaskCRUD()
			fake.FailRetry(tc.retryErr)
			mux := http.NewServeMux()
			handler.Register(mux, testDeps(fake))

			req := httptest.NewRequest(http.MethodPost, "/tasks/"+taskID+"/retry", bytes.NewReader([]byte(`{"mode":"fresh","parent_cycle_id":"c1"}`)))
			req.Header.Set("Content-Type", "application/json")
			req.SetPathValue("id", taskID)
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d body=%s", rec.Code, tc.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestPostTaskRetry_rejectsNonUserActor(t *testing.T) {
	t.Parallel()
	taskID := uuid.NewString()
	fake := storefake.NewTaskCRUD()
	mux := http.NewServeMux()
	handler.Register(mux, testDeps(fake))

	req := httptest.NewRequest(http.MethodPost, "/tasks/"+taskID+"/retry", bytes.NewReader([]byte(`{"mode":"fresh","parent_cycle_id":"c1"}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Actor", "agent")
	req.SetPathValue("id", taskID)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 body=%s", rec.Code, rec.Body.String())
	}
}

func TestPatchTaskGate_mapsStoreErrorsToHTTPStatus(t *testing.T) {
	t.Parallel()
	taskID := uuid.NewString()

	cases := []struct {
		name       string
		gateErr    error
		wantStatus int
	}{
		{name: "not found", gateErr: domain.ErrNotFound, wantStatus: http.StatusNotFound},
		{name: "invalid input", gateErr: domain.ErrInvalidInput, wantStatus: http.StatusBadRequest},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			fake := storefake.NewTaskCRUD()
			fake.FailGate(tc.gateErr)
			mux := http.NewServeMux()
			handler.Register(mux, testDeps(fake))

			body, _ := json.Marshal(map[string]string{"action": "release"})
			req := httptest.NewRequest(http.MethodPatch, "/tasks/"+taskID+"/gate", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.SetPathValue("id", taskID)
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d body=%s", rec.Code, tc.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestPatchTaskGate_okReturnsTask(t *testing.T) {
	t.Parallel()
	taskID := uuid.NewString()
	task := &domain.Task{
		ID:     taskID,
		Status: domain.StatusReady,
		Gate: &domain.TaskGate{
			Kind:   domain.GateKindManualApproval,
			Status: domain.GateStatusReleased,
		},
	}
	fake := storefake.NewTaskCRUD()
	fake.OnGate(task)
	fake.OnGet(task)

	mux := http.NewServeMux()
	handler.Register(mux, testDeps(fake))

	body, _ := json.Marshal(map[string]string{"action": "release"})
	req := httptest.NewRequest(http.MethodPatch, "/tasks/"+taskID+"/gate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", taskID)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	var got domain.Task
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.ID != taskID || got.Gate == nil || got.Gate.Status != domain.GateStatusReleased {
		t.Fatalf("body = %+v", got)
	}
}
