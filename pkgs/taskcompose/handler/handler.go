// Package handler registers /task-drafts* and /task-templates* REST routes for taskapi.
package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/AlexsanderHamir/Hamix/pkgs/taskcompose/contract"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
)

const (
	maxTemplateInstantiateCountPerItem = 25
	maxTemplateInstantiateTotalCreates = 100
)

// NormalizeComposeResult is the validated compose payload for template save/patch.
type NormalizeComposeResult struct {
	Payload json.RawMessage
	Title   string
}

// NormalizeComposeFunc validates and normalizes a template compose payload.
type NormalizeComposeFunc func(ctx context.Context, raw json.RawMessage) (NormalizeComposeResult, error)

// InstantiateJobItem is one prepared template create unit for the async worker.
type InstantiateJobItem struct {
	TemplateID string
	Count      int
	Payload    json.RawMessage
}

// InstantiateJob is enqueued after POST /task-templates/instantiate accepts.
type InstantiateJob struct {
	Items []InstantiateJobItem
	Actor taskcoredomain.Actor
}

// EnqueueInstantiateFunc schedules async template creates; returns false if rejected.
type EnqueueInstantiateFunc func(job InstantiateJob) bool

// Deps wires compose HTTP handlers into the taskapi mux.
type Deps struct {
	Compose            contract.ComposeStore
	NormalizeCompose   NormalizeComposeFunc
	EnqueueInstantiate EnqueueInstantiateFunc
}

// Handler serves task draft and template REST routes.
type Handler struct {
	compose            contract.ComposeStore
	normalizeCompose   NormalizeComposeFunc
	enqueueInstantiate EnqueueInstantiateFunc
}

// Register mounts /task-drafts* and /task-templates* routes on m.
//
//funclogmeasure:skip category=hot-path reason="Route table wiring only; operation trace is emitted by registered handlers."
func Register(m *http.ServeMux, deps Deps) {
	h := &Handler{
		compose:            deps.Compose,
		normalizeCompose:   deps.NormalizeCompose,
		enqueueInstantiate: deps.EnqueueInstantiate,
	}
	m.Handle("GET /task-drafts", http.HandlerFunc(h.listTaskDrafts))
	m.Handle("POST /task-drafts", http.HandlerFunc(h.saveTaskDraft))
	m.Handle("GET /task-drafts/{id}", http.HandlerFunc(h.getTaskDraft))
	m.Handle("DELETE /task-drafts/{id}", http.HandlerFunc(h.deleteTaskDraft))
	m.Handle("GET /task-templates", http.HandlerFunc(h.listTaskTemplates))
	m.Handle("POST /task-templates", http.HandlerFunc(h.saveTaskTemplate))
	m.Handle("GET /task-templates/{id}", http.HandlerFunc(h.getTaskTemplate))
	m.Handle("PATCH /task-templates/{id}", http.HandlerFunc(h.patchTaskTemplate))
	m.Handle("DELETE /task-templates/{id}", http.HandlerFunc(h.deleteTaskTemplate))
	m.Handle("POST /task-templates/instantiate", http.HandlerFunc(h.instantiateTaskTemplates))
}
