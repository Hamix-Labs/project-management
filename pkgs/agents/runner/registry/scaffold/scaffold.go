// Package scaffold opts into scaffold runner adapters that are excluded from
// the default live registry (registry/all). Import for side effects in tests
// or experimental binaries:
//
//	import _ "github.com/AlexsanderHamir/Hamix/pkgs/agents/runner/registry/scaffold"
package scaffold

import "github.com/AlexsanderHamir/Hamix/pkgs/agents/runner/claudecode"

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func init() {
	claudecode.Register()
}
