// Package httperr maps domain/store errors to HTTP JSON responses without
// importing pkgs/tasks/handlerhttp or BC handler packages.
//
// It exists to break the gitinventory ↔ handlerhttp import cycle: handlerhttp
// delegates git-coded errors here, and gitinventory handlers call the same
// helpers instead of maintaining a local JSON mirror.
package httperr
