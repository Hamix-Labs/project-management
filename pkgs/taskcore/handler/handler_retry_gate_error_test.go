package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/taskcore/handler"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handler/storefake"
	"github.com/google/uuid"
)

// retryGateFake scripts RequestTaskRetry / ApplyTaskGateAction / Get for thin
// handler error→status mapping tests without SQLite.
type retryGateFake struct {
	*storefake.TaskCRUDFake

	retryErr error
	retryOut *domain.Task

	gateErr error
	gateOut *domain.Task
}

func (f *retryGateFake) RequestTaskRetry(ctx context.Context, in contract.RequestRetryInput, by domain.Actor) (*domain.Task, error) {
	if f.retryErr != nil {
		return nil, f.retryErr
	}
	if f.retryOut != nil {
		return f.retryOut, nil
	}
	return &domain.Task{ID: in.TaskID, Status: domain.StatusReady}, nil
}

func (f *retryGateFake) ApplyTaskGateAction(ctx context.Context, taskID, action string, by domain.Actor) (*domain.Task, error) {
	if f.gateErr != nil {
		return nil, f.gateErr
	}
	if f.gateOut != nil {
		return f.gateOut, nil
	}
	return &domain.Task{ID: taskID}, nil
}

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
			fake := &retryGateFake{TaskCRUDFake: storefake.NewTaskCRUD(), retryErr: tc.retryErr}
			mux := http.NewServeMux()
			handler.Register(mux, handler.Deps{Tasks: fake})

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
	fake := &retryGateFake{TaskCRUDFake: storefake.NewTaskCRUD()}
	mux := http.NewServeMux()
	handler.Register(mux, handler.Deps{Tasks: fake})

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
			fake := &retryGateFake{TaskCRUDFake: storefake.NewTaskCRUD(), gateErr: tc.gateErr}
			mux := http.NewServeMux()
			handler.Register(mux, handler.Deps{Tasks: fake})

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
	fake := &retryGateFake{TaskCRUDFake: storefake.NewTaskCRUD(), gateOut: task}
	fake.OnGet(task)

	mux := http.NewServeMux()
	handler.Register(mux, handler.Deps{Tasks: fake})

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
