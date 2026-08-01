package prompt

import (
	"strings"
	"testing"

	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

func TestComposeOpenPRDirective(t *testing.T) {
	t.Parallel()
	if got := ComposeOpenPRDirective(nil, nil); got != "" {
		t.Fatalf("nil cycle: got %q", got)
	}
	cycle := &cyclesdomain.TaskCycle{ID: "cyc-1"}
	known := []cyclesdomain.TaskCycleCommit{{SHA: "deadbeefcafe01", Message: "ship it"}}
	got := ComposeOpenPRDirective(cycle, known)
	for _, frag := range []string{
		"Approve and open pull request", "cyc-1", "hamix.create_pull_request",
		"deadbeefcafe", "ship it", "How to verify",
	} {
		if !strings.Contains(got, frag) {
			t.Fatalf("missing %q in %q", frag, got)
		}
	}
}

func TestAppendOpenPRNoticeAndGitPolicy(t *testing.T) {
	t.Parallel()
	cycle := &cyclesdomain.TaskCycle{ID: "cyc-2"}
	base := "BASE_PROMPT"
	got := AppendOpenPRGitPolicy(AppendOpenPRNotice(base, cycle, nil))
	for _, frag := range []string{
		"Approve and open pull request", "cyc-2", "hamix.create_pull_request",
		"open-pr run", "BASE_PROMPT",
	} {
		if !strings.Contains(got, frag) {
			t.Fatalf("missing %q in %q", frag, got)
		}
	}
	if idx := strings.Index(got, "Approve and open"); idx < 0 || idx > strings.Index(got, "BASE_PROMPT") {
		t.Fatalf("open-pr notice should precede base prompt: %q", got)
	}
}
