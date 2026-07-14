// Package apijson provides small HTTP helpers for the task JSON API: baseline security
// headers, JSON error responses (including request_id when present on context), and
// cycle-safe error-detail helpers (InvalidInputDetail) that gitinventory can import
// without pkgs/tasks/handlerhttp.
//
// Handlers pass an optional callPath function for debug http.io logs; middleware outside
// pkgs/tasks/handler can pass nil.
package apijson
