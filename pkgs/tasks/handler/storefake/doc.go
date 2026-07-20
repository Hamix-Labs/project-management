// Package storefake provides recording test doubles for handler store slices
// used by pkgs/tasks/handler. Fakes record calls and accept injectable
// On*/Fail* outcomes for Get, Retry, Gate, ListCycles, and ListEvents so
// handler error-path tests avoid SQLite seeding. Remaining HandlerStore
// methods stay unimplemented compile-time padding.
package storefake
