package runner_test

import (
	"encoding/json"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
	cyclesdomain "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"
)

func TestNewResult_passesSmallValuesThrough(t *testing.T) {
	t.Parallel()

	details := json.RawMessage(`{"ok":true}`)
	res := runner.NewResult(cyclesdomain.PhaseStatusSucceeded, "short summary", details, "small output")

	if res.Truncated {
		t.Errorf("Truncated must be false for under-budget values")
	}
	if res.Summary != "short summary" {
		t.Errorf("Summary mutated: got %q", res.Summary)
	}
	if string(res.Details) != string(details) {
		t.Errorf("Details mutated: got %s want %s", res.Details, details)
	}
	if res.RawOutput != "small output" {
		t.Errorf("RawOutput mutated: got %q", res.RawOutput)
	}
	if res.Status != cyclesdomain.PhaseStatusSucceeded {
		t.Errorf("Status mutated: got %q", res.Status)
	}
}

// TestNewResult_clipsSummaryToMaxRunes asserts rune-correct clipping (no
// mid-codepoint slice on multi-byte input).
func TestNewResult_clipsSummaryToMaxRunes(t *testing.T) {
	t.Parallel()

	long := strings.Repeat("Ã©", runner.MaxSummaryRunes+50)
	res := runner.NewResult(cyclesdomain.PhaseStatusSucceeded, long, nil, "")

	if !res.Truncated {
		t.Errorf("Truncated must be true when Summary is clipped")
	}
	gotRunes := utf8.RuneCountInString(res.Summary)
	if gotRunes != runner.MaxSummaryRunes {
		t.Errorf("Summary rune count: got %d want %d", gotRunes, runner.MaxSummaryRunes)
	}
	if !utf8.ValidString(res.Summary) {
		t.Errorf("clipped Summary is not valid UTF-8")
	}
}

// TestNewResult_clipsRawOutputToCap asserts the trailing-bytes policy and
// that the result is still valid UTF-8 even when the cap lands inside a
// multi-byte sequence.
func TestNewResult_clipsRawOutputToCap(t *testing.T) {
	t.Parallel()

	prefix := strings.Repeat("a", runner.MaxResultRawOutputBytes)
	tail := "TAIL_MARKER"
	full := prefix + tail
	res := runner.NewResult(cyclesdomain.PhaseStatusFailed, "", nil, full)

	if !res.Truncated {
		t.Errorf("Truncated must be true when RawOutput is clipped")
	}
	if len(res.RawOutput) > runner.MaxResultRawOutputBytes {
		t.Errorf("RawOutput byte length %d exceeds cap %d",
			len(res.RawOutput), runner.MaxResultRawOutputBytes)
	}
	if !strings.HasSuffix(res.RawOutput, tail) {
		t.Errorf("trailing bytes lost: tail %q not in clipped output", tail)
	}
	if !utf8.ValidString(res.RawOutput) {
		t.Errorf("clipped RawOutput is not valid UTF-8")
	}
}

// TestNewResult_clipsRawOutputAtUTF8Boundary asserts that a multi-byte
// sequence straddling the cap is dropped cleanly rather than producing a
// half-rune.
func TestNewResult_clipsRawOutputAtUTF8Boundary(t *testing.T) {
	t.Parallel()

	// "Ã©" is 2 bytes. Build a body that pushes a "Ã©" across the boundary.
	body := strings.Repeat("Ã©", runner.MaxResultRawOutputBytes/2+10)
	res := runner.NewResult(cyclesdomain.PhaseStatusFailed, "", nil, body)

	if !res.Truncated {
		t.Errorf("Truncated must be true")
	}
	if !utf8.ValidString(res.RawOutput) {
		t.Errorf("RawOutput is not valid UTF-8 after boundary-snap clip")
	}
}

// TestNewResult_clipsDetailsToSentinel asserts oversized Details are
// replaced with a parseable sentinel, never raw bytes truncated mid-JSON.
func TestNewResult_clipsDetailsToSentinel(t *testing.T) {
	t.Parallel()

	bigPayload := make([]byte, 0, runner.MaxResultDetailsBytes+1024)
	bigPayload = append(bigPayload, '"')
	bigPayload = append(bigPayload, strings.Repeat("x", runner.MaxResultDetailsBytes+1000)...)
	bigPayload = append(bigPayload, '"')
	res := runner.NewResult(cyclesdomain.PhaseStatusFailed, "", json.RawMessage(bigPayload), "")

	if !res.Truncated {
		t.Errorf("Truncated must be true when Details is clipped")
	}
	if len(res.Details) > runner.MaxResultDetailsBytes {
		t.Errorf("clipped Details exceeds cap: %d > %d",
			len(res.Details), runner.MaxResultDetailsBytes)
	}
	var parsed struct {
		Truncated     bool `json:"truncated"`
		OriginalBytes int  `json:"original_bytes"`
	}
	if err := json.Unmarshal(res.Details, &parsed); err != nil {
		t.Fatalf("sentinel must be valid JSON, got %s (err=%v)", res.Details, err)
	}
	if !parsed.Truncated {
		t.Errorf("sentinel must carry truncated=true: %s", res.Details)
	}
	if parsed.OriginalBytes != len(bigPayload) {
		t.Errorf("sentinel original_bytes: got %d want %d",
			parsed.OriginalBytes, len(bigPayload))
	}
}

// TestNewResult_clipsInvalidJSONDetailsToSentinel pins the
// JSON-validity contract on Result.Details. The doc on
// runner.NewResult says: "consumers always see well-formed JSON".
// Before the fix, clipDetails only checked size: an under-cap but
// MALFORMED Details payload (e.g. a third-party adapter that hands
// in a `}` short of a valid object, a payload with a trailing comma,
// a half-finished string literal â€” all things that can happen when
// an adapter assembles a payload from substring concatenation rather
// than encoding/json.Marshal) flowed through unchanged. Downstream
// consumers (worker dual-write into TaskCyclePhase.MetaJSON, the
// SPA reading /tasks/{id}/cycles/{cycleId}/phases/{seq}, any future
// log shipper that re-decodes the audit row) then either crashed
// with a JSON decode error or silently passed garbage through.
//
// The fix routes any input that is not json.Valid through the same
// sentinel as oversized payloads (truncated=true, original_bytes=N)
// so consumers can distinguish "no details" (nil) from "had details
// but they were lost to size or invalidity" (sentinel) without ever
// having to handle malformed JSON.
func TestNewResult_clipsInvalidJSONDetailsToSentinel(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		input string
	}{
		{"truncated_object", `{"a":1`},
		{"trailing_comma", `{"a":1,}`},
		{"unterminated_string", `{"a":"hello`},
		{"raw_garbage", `not json at all`},
		{"single_brace", `}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			res := runner.NewResult(cyclesdomain.PhaseStatusFailed, "",
				json.RawMessage(tc.input), "")
			if !res.Truncated {
				t.Errorf("Truncated must be true when invalid Details is replaced with sentinel; got Details=%q", res.Details)
			}
			if !json.Valid(res.Details) {
				t.Errorf("Details after NewResult must be valid JSON; got %q", res.Details)
			}
			var parsed struct {
				Truncated     bool `json:"truncated"`
				OriginalBytes int  `json:"original_bytes"`
			}
			if err := json.Unmarshal(res.Details, &parsed); err != nil {
				t.Fatalf("sentinel must unmarshal: %v body=%s", err, res.Details)
			}
			if !parsed.Truncated {
				t.Errorf("sentinel must carry truncated=true: %s", res.Details)
			}
			if parsed.OriginalBytes != len(tc.input) {
				t.Errorf("sentinel original_bytes: got %d want %d (input=%q)",
					parsed.OriginalBytes, len(tc.input), tc.input)
			}
		})
	}
}

// TestNewResult_passesValidNonObjectDetailsThrough confirms the
// validity check intentionally permits valid non-object JSON values
// (top-level arrays, strings, numbers, null, true/false). The
// Details field is documented as "JSON-safe" / "well-formed JSON",
// not "always an object" â€” the per-field "object only" invariant
// belongs to handler.normalizeJSONObjectForResponse, not the runner.
// This test guards against an over-eager fix that would coerce
// `["a","b"]` or `null` to the sentinel.
func TestNewResult_passesValidNonObjectDetailsThrough(t *testing.T) {
	t.Parallel()

	cases := []string{
		`["a","b"]`,
		`null`,
		`42`,
		`"a string"`,
		`true`,
	}
	for _, in := range cases {
		t.Run(in, func(t *testing.T) {
			t.Parallel()
			res := runner.NewResult(cyclesdomain.PhaseStatusSucceeded, "",
				json.RawMessage(in), "")
			if res.Truncated {
				t.Errorf("Truncated must be false for valid JSON %q", in)
			}
			if string(res.Details) != in {
				t.Errorf("Details mutated for valid JSON: got %q want %q",
					res.Details, in)
			}
		})
	}
}

// TestNewResult_atCapBoundary asserts a value of exactly MaxBytes bytes is
// NOT clipped. (Cap is inclusive of the boundary.)
func TestNewResult_atCapBoundary(t *testing.T) {
	t.Parallel()

	body := strings.Repeat("a", runner.MaxResultRawOutputBytes)
	res := runner.NewResult(cyclesdomain.PhaseStatusSucceeded, "", nil, body)
	if res.Truncated {
		t.Errorf("value of exactly MaxBytes must not be clipped")
	}
	if res.RawOutput != body {
		t.Errorf("RawOutput mutated at cap boundary")
	}
}

// TestErrSentinels_distinct guards against accidental sentinel collapse so
// errors.Is in the worker keeps differentiating timeout / non-zero / invalid.
