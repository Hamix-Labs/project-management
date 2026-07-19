package prompt

import (
	"strings"

	"github.com/AlexsanderHamir/Hamix/pkgs/projects/domain"
)

// ProjectContextInput carries project context rows assembled by the harness
// before prompt injection.
type ProjectContextInput struct {
	Project domain.Project
	Items   []domain.ProjectContextItem
}

// BuildProjectContextSection renders the XML-tagged project context block.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func BuildProjectContextSection(in ProjectContextInput) string {
	var b strings.Builder
	b.WriteString("<project_context>\n")
	b.WriteString("Project: ")
	b.WriteString(in.Project.Name)
	b.WriteString("\n")
	if strings.TrimSpace(in.Project.ContextSummary) != "" {
		b.WriteString("Summary: ")
		b.WriteString(strings.TrimSpace(in.Project.ContextSummary))
		b.WriteString("\n")
	}
	for _, item := range in.Items {
		b.WriteString("\n[")
		b.WriteString(string(item.Kind))
		b.WriteString("] ")
		b.WriteString(item.Title)
		b.WriteString("\n")
		b.WriteString(item.Body)
		b.WriteString("\n")
	}
	b.WriteString("</project_context>")
	return b.String()
}

// EstimateTokens returns a coarse token estimate for rendered context text.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func EstimateTokens(s string) int {
	if strings.TrimSpace(s) == "" {
		return 0
	}
	return (len([]rune(s)) + 3) / 4
}

// WrapWithProjectContext wraps the task prompt in project context when present.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func WrapWithProjectContext(prompt string, projectContext string) string {
	if strings.TrimSpace(projectContext) == "" {
		return prompt
	}
	return projectContext + "\n\n<task_prompt>\n" + prompt + "\n</task_prompt>"
}
