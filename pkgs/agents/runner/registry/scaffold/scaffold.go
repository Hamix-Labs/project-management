// Package scaffold opts into scaffold runner adapters that are excluded from
// the default live registry (registry/all). Import for side effects in tests
// or experimental binaries:
//
//	import _ "github.com/AlexsanderHamir/Hamix/pkgs/agents/runner/registry/scaffold"
package scaffold

import "github.com/AlexsanderHamir/Hamix/pkgs/agents/runner/claudecode"

func init() {
	claudecode.Register()
}
