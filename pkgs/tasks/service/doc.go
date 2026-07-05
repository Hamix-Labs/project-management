// Package service holds HTTP-agnostic use-case orchestration for the tasks
// domain. Handlers translate transport concerns; service composes store
// facades (and gitwork where needed) for multi-step flows reused by API
// and future binaries.
package service
