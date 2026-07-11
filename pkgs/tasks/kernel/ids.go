package kernel

import (
	"strings"

	"github.com/google/uuid"
)

// ResolveID returns trimmed id, or a new UUID when id is empty.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ResolveID(id string) string {
	id = strings.TrimSpace(id)
	if id == "" {
		return uuid.NewString()
	}
	return id
}
