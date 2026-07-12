package store

import "github.com/AlexsanderHamir/Hamix/pkgs/tasks/wire"

// HandlerAPI is the persistence contract required by pkgs/tasks/handler.
// Canonical definition: pkgs/tasks/wire.HandlerAPI.
// *Store implements it; compile-time checks live in handler/handler_store_assert_test.go.
type HandlerAPI = wire.HandlerAPI
