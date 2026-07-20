package gitwork

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"context"
	"log/slog"
	"path/filepath"
	"strings"
)

func (s *DefaultService) OpenRepository(ctx context.Context, path string) (*Repository, error) {
	slog.DebugContext(ctx, "trace", "cmd", calltrace.LogCmd, "operation", "gitwork.OpenRepository")
	abs, err := absPath(path)
	if err != nil {
		return nil, err
	}
	out, err := s.runGit(ctx, abs, "rev-parse", "--show-toplevel", "--git-common-dir")
	if err != nil {
		return nil, mapNotARepository(err)
	}
	lines := splitNonEmptyLines(out)
	if len(lines) < 2 {
		return nil, ErrNotARepository
	}
	root, err := absPath(lines[0])
	if err != nil {
		return nil, err
	}
	common := lines[1]
	if !filepath.IsAbs(common) {
		common = filepath.Join(root, common)
	}
	common, err = absPath(common)
	if err != nil {
		return nil, err
	}
	return &Repository{Root: root, CommonDir: common}, nil
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func splitNonEmptyLines(s string) []string {
	raw := strings.Split(s, "\n")
	out := make([]string, 0, len(raw))
	for _, line := range raw {
		line = strings.TrimSpace(line)
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}
