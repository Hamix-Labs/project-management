package storekernel

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"
)

func TestMapPayloadPersistenceError(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		wantMsg string
		wantIn  error
	}{
		{
			name:    "postgres invalid json",
			err:     fmt.Errorf("patch template: ERROR: invalid input syntax for type json (SQLSTATE 22P02)"),
			wantMsg: "payload could not be saved",
			wantIn:  domain.ErrInvalidInput,
		},
		{
			name:    "sqlite malformed json",
			err:     errors.New("malformed JSON"),
			wantMsg: "payload could not be saved",
			wantIn:  domain.ErrInvalidInput,
		},
		{
			name:    "passthrough",
			err:     errors.New("connection reset"),
			wantMsg: "connection reset",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapPayloadPersistenceError(tt.err)
			if tt.wantIn != nil && !errors.Is(got, tt.wantIn) {
				t.Fatalf("errors.Is(%v, %v) = false", got, tt.wantIn)
			}
			if !strings.Contains(got.Error(), tt.wantMsg) {
				t.Fatalf("got %q want substring %q", got.Error(), tt.wantMsg)
			}
		})
	}
}
