package handler

import (
	"fmt"
	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"strconv"
	"strings"
)

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func parseBoundedLimit(q map[string][]string, def, max int) (int, error) {
	raw := strings.TrimSpace(firstQueryValue(q, "limit"))
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

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func firstQueryValue(q map[string][]string, key string) string {
	values := q[key]
	if len(values) == 0 {
		return ""
	}
	return values[0]
}
