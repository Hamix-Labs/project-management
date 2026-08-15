package taskapiruntime

import (
	"context"
	"strings"
	"testing"

	"github.com/AlexsanderHamir/Hamix/internal/taskapi/draftsidecar"
)

func TestDraftAssistHost_failsWithoutAPIKey(t *testing.T) {
	t.Setenv(draftsidecar.APIKeyEnv, "")
	_, err := draftAssistHost(context.Background())
	if err == nil {
		t.Fatal("expected error without CURSOR_API_KEY")
	}
	if !strings.Contains(err.Error(), draftsidecar.APIKeyEnv) {
		t.Fatalf("error %q should name %s", err, draftsidecar.APIKeyEnv)
	}
}
