package mcp

import (
	"context"

	"github.com/AlexsanderHamir/Hamix/pkgs/draftassist/domain"
)

// WritePromptForTest exposes the shared write path so package-external tests
// can assert the validator fires before the HTTP client is touched. This is
// test-only; production callers reach the MCP tool bodies directly.
func (h *ToolHost) WritePromptForTest(ctx context.Context, prompt string) error {
	return h.writePrompt(ctx, prompt, domain.PatchEventData{
		Op:      domain.PatchOpSet,
		Value:   prompt,
		Summary: "test",
	})
}
