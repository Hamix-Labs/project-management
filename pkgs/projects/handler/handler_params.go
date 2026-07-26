package handler

import (
	"strings"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handlerhttp"
)

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func parseProjectListParams(q map[string][]string) (limit int, includeArchived bool, err error) {
	limit, err = handlerhttp.ParseBoundedLimit(q, 50, 100)
	if err != nil {
		return 0, false, err
	}
	includeArchived = strings.EqualFold(strings.TrimSpace(handlerhttp.FirstQueryValue(q, "include_archived")), "true")
	return limit, includeArchived, nil
}
