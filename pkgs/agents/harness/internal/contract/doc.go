// Package contract defines the persistence surface required by harness
// production code and harness test doubles. It lives in internal/contract
// so harness root and internal/{verify,resume,git} can share the type
// without an import cycle.
package contract
