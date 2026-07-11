// Package handler registers /runners* REST routes for taskapi.
package handler

import (
	"net/http"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/contract"
)

// Deps wires runner registry HTTP handlers into the taskapi mux.
type Deps struct {
	Settings contract.SettingsStore
}

// Handler serves /runners* REST routes.
type Handler struct {
	settings contract.SettingsStore
}

// Register mounts /runners* routes on m.
//
//funclogmeasure:skip category=hot-path reason="Route table wiring only; operation trace is emitted by registered handlers."
func Register(m *http.ServeMux, deps Deps) {
	h := &Handler{settings: deps.Settings}
	m.Handle("GET /runners", http.HandlerFunc(h.listRunners))
	m.Handle("GET /runners/{id}/config-schema", http.HandlerFunc(h.runnerConfigSchema))
	m.Handle("POST /runners/{id}/probe", http.HandlerFunc(h.probeRunner))
	m.Handle("POST /runners/{id}/list-models", http.HandlerFunc(h.listRunnerModels))
	m.Handle("POST /runners/{id}/validate-config", http.HandlerFunc(h.validateRunnerConfig))
}
