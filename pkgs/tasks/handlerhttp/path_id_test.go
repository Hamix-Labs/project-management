package handlerhttp

import (
	"errors"
	"strings"
	"testing"

	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
)

func TestParsePathID(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		_, err := ParsePathID("   ")
		if err == nil || !errors.Is(err, taskcoredomain.ErrInvalidInput) {
			t.Fatalf("got %v", err)
		}
	})
	t.Run("too_long", func(t *testing.T) {
		_, err := ParsePathID(strings.Repeat("a", MaxPathIDBytes+1))
		if err == nil || !errors.Is(err, taskcoredomain.ErrInvalidInput) {
			t.Fatalf("got %v", err)
		}
	})
	t.Run("ok_trimmed", func(t *testing.T) {
		id, err := ParsePathID("  " + strings.Repeat("c", 36) + "  ")
		if err != nil || id != strings.Repeat("c", 36) {
			t.Fatalf("got %q %v", id, err)
		}
	})
}

func TestParseBoundedPathID(t *testing.T) {
	t.Run("custom_field_empty", func(t *testing.T) {
		_, err := ParseBoundedPathID("test", "  ", "item id")
		if err == nil || !errors.Is(err, taskcoredomain.ErrInvalidInput) {
			t.Fatalf("got %v", err)
		}
		if !strings.Contains(err.Error(), "item id") {
			t.Fatalf("expected field in error, got %v", err)
		}
	})
	t.Run("custom_field_too_long", func(t *testing.T) {
		_, err := ParseBoundedPathID("test", strings.Repeat("x", MaxPathIDBytes+1), "item id")
		if err == nil || !errors.Is(err, taskcoredomain.ErrInvalidInput) {
			t.Fatalf("got %v", err)
		}
		if !strings.Contains(err.Error(), "item id too long") {
			t.Fatalf("expected too-long detail, got %v", err)
		}
	})
}
