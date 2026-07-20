// Package all imports production runner adapters so their init() functions
// register with the global registry. Binaries that need runner support import
// this package for the side effect:
//
//	import _ "github.com/AlexsanderHamir/Hamix/pkgs/agents/runner/registry/all"
//
// Scaffold adapters (e.g. claude-code) are intentionally excluded; import
// pkgs/agents/runner/registry/scaffold to opt in.
package all

import (
	_ "github.com/AlexsanderHamir/Hamix/pkgs/agents/runner/cursor"
)
