package domain

import (
	"encoding/json"
	"testing"
)

func TestTokenUsage_Consumed_prefersTotalTokens(t *testing.T) {
	t.Parallel()
	u := TokenUsage{
		InputTokens:  10,
		OutputTokens: 5,
		TotalTokens:  100,
	}
	if got := u.Consumed(); got != 100 {
		t.Fatalf("Consumed() = %d, want 100", got)
	}
}

func TestTokenUsage_Consumed_sumsPartsWhenTotalMissing(t *testing.T) {
	t.Parallel()
	u := TokenUsage{
		InputTokens:      10,
		OutputTokens:     5,
		CacheReadTokens:  3,
		CacheWriteTokens: 2,
	}
	if got := u.Consumed(); got != 20 {
		t.Fatalf("Consumed() = %d, want 20", got)
	}
}

func TestTokenUsage_Consumed_sumsPartsWhenTotalZero(t *testing.T) {
	t.Parallel()
	u := TokenUsage{
		InputTokens:  7,
		OutputTokens: 3,
		TotalTokens:  0,
	}
	if got := u.Consumed(); got != 10 {
		t.Fatalf("Consumed() = %d, want 10", got)
	}
}

func TestAddTokenUsage_sumsFieldWise(t *testing.T) {
	t.Parallel()
	a := TokenUsage{InputTokens: 10, OutputTokens: 5, TotalTokens: 15}
	b := TokenUsage{InputTokens: 3, OutputTokens: 2, CacheReadTokens: 1}
	got := AddTokenUsage(a, b)
	if got.InputTokens != 13 || got.OutputTokens != 7 || got.CacheReadTokens != 1 || got.TotalTokens != 15 {
		t.Fatalf("AddTokenUsage = %+v", got)
	}
}

func TestParseTokenUsage(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		raw    string
		want   TokenUsage
		wantOK bool
	}{
		{
			name:   "empty rejected",
			raw:    "",
			wantOK: false,
		},
		{
			name:   "null rejected",
			raw:    "null",
			wantOK: false,
		},
		{
			name:   "non-object rejected",
			raw:    `"nope"`,
			wantOK: false,
		},
		{
			name: "full object",
			raw:  `{"inputTokens":10,"outputTokens":5,"cacheReadTokens":1,"cacheWriteTokens":2,"totalTokens":18}`,
			want: TokenUsage{
				InputTokens: 10, OutputTokens: 5, CacheReadTokens: 1, CacheWriteTokens: 2, TotalTokens: 18,
			},
			wantOK: true,
		},
		{
			name:   "explicit zero keys still known",
			raw:    `{"inputTokens":0,"outputTokens":0,"cacheReadTokens":0,"cacheWriteTokens":0,"totalTokens":0}`,
			want:   TokenUsage{},
			wantOK: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, ok := ParseTokenUsage(json.RawMessage(tc.raw))
			if ok != tc.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tc.wantOK)
			}
			if ok && got != tc.want {
				t.Fatalf("usage = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestTokenUsageFromDetailsJSON(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		raw    string
		want   TokenUsage
		wantOK bool
	}{
		{
			name:   "empty rejected",
			raw:    `{}`,
			wantOK: false,
		},
		{
			name:   "missing usage rejected",
			raw:    `{"verification":{"attempt_seq":1}}`,
			wantOK: false,
		},
		{
			name:   "top-level usage",
			raw:    `{"usage":{"inputTokens":4,"outputTokens":6}}`,
			want:   TokenUsage{InputTokens: 4, OutputTokens: 6},
			wantOK: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, ok := TokenUsageFromDetailsJSON(json.RawMessage(tc.raw))
			if ok != tc.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tc.wantOK)
			}
			if ok && got != tc.want {
				t.Fatalf("usage = %+v, want %+v", got, tc.want)
			}
		})
	}
}
