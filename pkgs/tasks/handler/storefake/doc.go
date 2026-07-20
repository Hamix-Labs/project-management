// Package storefake provides recording test doubles for handler store slices
// used by pkgs/tasks/handler. Prefer the focused TaskCRUDFake (and per-BC
// fakes) over HandlerStoreFake when a test only needs task/retry/gate
// surfaces — the composed HandlerStoreFake exists for compile-time coverage
// of wire.HandlerAPI and should not be confused with harness store fakes
// under pkgs/agents/harness (B-44 / F-TEST-13).
//
// Fakes record calls and accept injectable On*/Fail* outcomes for Get,
// Retry, Gate, ListCycles, and ListEvents so handler error-path tests avoid
// SQLite seeding. Remaining HandlerAPI methods stay unimplemented
// compile-time padding.
package storefake
