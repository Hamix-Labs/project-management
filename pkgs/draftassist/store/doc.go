// Package store is the in-memory session + event bus implementation for
// pkgs/draftassist. Sessions live only for the modal lifetime; nothing
// touches Postgres. Concurrency-safe under a single sync.Mutex per store.
package store
