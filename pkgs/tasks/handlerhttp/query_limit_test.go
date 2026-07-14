package handlerhttp_test

import (
	"errors"
	"strings"
	"testing"

	taskcoredomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handlerhttp"
)

func TestFirstQueryValue(t *testing.T) {
	t.Parallel()

	if got := handlerhttp.FirstQueryValue(nil, "limit"); got != "" {
		t.Fatalf("nil map: got %q", got)
	}
	if got := handlerhttp.FirstQueryValue(map[string][]string{}, "limit"); got != "" {
		t.Fatalf("missing key: got %q", got)
	}
	if got := handlerhttp.FirstQueryValue(map[string][]string{"limit": {}}, "limit"); got != "" {
		t.Fatalf("empty values: got %q", got)
	}
	q := map[string][]string{"limit": {"10", "20"}}
	if got := handlerhttp.FirstQueryValue(q, "limit"); got != "10" {
		t.Fatalf("first value: got %q want 10", got)
	}
}

func TestParseBoundedLimit(t *testing.T) {
	t.Parallel()

	const def, max = 50, 100

	t.Run("emptyUsesDefault", func(t *testing.T) {
		t.Parallel()
		got, err := handlerhttp.ParseBoundedLimit(map[string][]string{}, def, max)
		if err != nil || got != def {
			t.Fatalf("got=%d err=%v want def=%d", got, err, def)
		}
	})

	t.Run("zeroUsesDefault", func(t *testing.T) {
		t.Parallel()
		got, err := handlerhttp.ParseBoundedLimit(map[string][]string{"limit": {"0"}}, def, max)
		if err != nil || got != def {
			t.Fatalf("got=%d err=%v want def=%d", got, err, def)
		}
	})

	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		got, err := handlerhttp.ParseBoundedLimit(map[string][]string{"limit": {" 25 "}}, def, max)
		if err != nil || got != 25 {
			t.Fatalf("got=%d err=%v want 25", got, err)
		}
	})

	t.Run("atMax", func(t *testing.T) {
		t.Parallel()
		got, err := handlerhttp.ParseBoundedLimit(map[string][]string{"limit": {"100"}}, def, max)
		if err != nil || got != 100 {
			t.Fatalf("got=%d err=%v want 100", got, err)
		}
	})

	cases := []struct {
		name string
		raw  string
		want string
	}{
		{"nonNumeric", "nope", "limit must be integer 0..100"},
		{"negative", "-1", "limit must be integer 0..100"},
		{"overMax", "101", "limit must be integer 0..100"},
		{"overlong", strings.Repeat("1", 33), "limit value too long"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := handlerhttp.ParseBoundedLimit(map[string][]string{"limit": {tc.raw}}, def, max)
			if !errors.Is(err, taskcoredomain.ErrInvalidInput) {
				t.Fatalf("err=%v want wrap of ErrInvalidInput", err)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err=%q want substring %q", err, tc.want)
			}
		})
	}
}
