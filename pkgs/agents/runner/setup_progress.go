package runner

// ProgressRunStateKind is the ProgressEvent.Kind for worker-authored
// run_state events (setup / handoff progress).
const ProgressRunStateKind = "run_state"

// Harness-owned setup / handoff progress (fills ticker silence before Cursor stdout).
const (
	ProgressToolHarnessSetup = "harness_setup"

	ProgressRunStateSetupStarted  = "setup_started"
	ProgressRunStateSetupGit      = "setup_git"
	ProgressRunStateSetupPlan     = "setup_plan"
	ProgressRunStateSetupInvoke   = "setup_invoke"
	ProgressRunStateSetupSpawn    = "setup_spawn"
	ProgressRunStateSetupIngest   = "setup_ingest"
	ProgressRunStateSetupPrompt   = "setup_prompt"
	ProgressRunStateHandoffVerify = "handoff_verify"
	ProgressRunStateRestartResume = "restart_resume"
)

// SetupProgressEvent builds a worker-authored run_state progress event.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; callers emit operation traces."
func SetupProgressEvent(subtype, message string) ProgressEvent {
	return ProgressEvent{
		Kind:    ProgressRunStateKind,
		Subtype: subtype,
		Message: message,
		Tool:    ProgressToolHarnessSetup,
	}
}
