package store

import "github.com/AlexsanderHamir/Hamix/pkgs/tasks/contract"

// HandlerAPI is the persistence contract required by pkgs/tasks/handler.
// *Store implements it; tests may pass a narrower fake. Method groups live in
// pkgs/tasks/contract slices for compile-time asserts and incremental fakes.
type HandlerAPI = contract.HandlerStore

// HandlerAPI compliance is checked at compile time in handler/handler_store_assert_test.go
// and pkgs/tasks/contract/*_test.go.
