package calltrace

// LogCmd is the slog "cmd" field value shared with pkgs/tasks/handler for taskapi HTTP logs.
const LogCmd = "taskapi"

// helper.io observability field values (obs_category / phase on helper.io log lines).
const (
	ObsCategoryHelperIO = "helper_io"
	PhaseHelperIn       = "helper_in"
	PhaseHelperOut      = "helper_out"
)
