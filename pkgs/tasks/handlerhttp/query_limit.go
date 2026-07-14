package handlerhttp

import (
	"fmt"
	"strconv"
	"strings"

	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
)

// Max length for ?limit= abuse guards (matches taskcore handler list params).
const maxListIntQueryParamBytes = 32

// FirstQueryValue returns the first value for key, or "" when absent.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func FirstQueryValue(q map[string][]string, key string) string {
	values := q[key]
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

// ParseBoundedLimit parses ?limit= from query values.
// Empty or 0 yields def; other values must be integers in 0..max (inclusive).
// Overlong raw values are rejected before strconv to bound abuse.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ParseBoundedLimit(q map[string][]string, def, max int) (int, error) {
	raw := strings.TrimSpace(FirstQueryValue(q, "limit"))
	if raw == "" {
		return def, nil
	}
	if len(raw) > maxListIntQueryParamBytes {
		return 0, fmt.Errorf("%w: limit value too long", taskcoredomain.ErrInvalidInput)
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 || n > max {
		return 0, fmt.Errorf("%w: limit must be integer 0..%d", taskcoredomain.ErrInvalidInput, max)
	}
	if n == 0 {
		return def, nil
	}
	return n, nil
}
