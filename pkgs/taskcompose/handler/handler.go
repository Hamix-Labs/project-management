// Package handler registers /task-drafts* and /task-templates* REST routes for taskapi.
package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/AlexsanderHamir/Hamix/pkgs/taskcompose/contract"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"
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

// InstantiateFromTemplateFunc creates one task from a template payload (task create stays in tasks handler).
type InstantiateFromTemplateFunc func(ctx context.Context, r *http.Request, op string, payload json.RawMessage, by domain.Actor) (*domain.Task, error)

// Deps wires compose HTTP handlers into the taskapi mux.
type Deps struct {
	Compose                 contract.ComposeStore
	NormalizeCompose        NormalizeComposeFunc
	InstantiateFromTemplate InstantiateFromTemplateFunc
}

// Handler serves task draft and template REST routes.
type Handler struct {
	compose                 contract.ComposeStore
	normalizeCompose        NormalizeComposeFunc
	instantiateFromTemplate InstantiateFromTemplateFunc
}

// Register mounts /task-drafts* and /task-templates* routes on m.
//
//funclogmeasure:skip category=hot-path reason="Route table wiring only; operation trace is emitted by registered handlers."
func Register(m *http.ServeMux, deps Deps) {
	h := &Handler{
		compose:                 deps.Compose,
		normalizeCompose:        deps.NormalizeCompose,
		instantiateFromTemplate: deps.InstantiateFromTemplate,
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
