// Package storefake provides recording test doubles for handler store slices
// used by pkgs/tasks/handler. Fakes record calls and accept injectable errors
// so handler error-path tests avoid SQLite seeding.
package storefake
