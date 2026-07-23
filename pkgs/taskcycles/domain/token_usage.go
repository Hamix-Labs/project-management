package domain

import (
	"encoding/json"
)

// TokenUsage is the Cursor CLI usage object persisted under details_json.usage.
// Wire keys stay camelCase to match Cursor and existing SPA parsers.
type TokenUsage struct {
	InputTokens      int64 `json:"inputTokens,omitempty"`
	OutputTokens     int64 `json:"outputTokens,omitempty"`
	CacheReadTokens  int64 `json:"cacheReadTokens,omitempty"`
	CacheWriteTokens int64 `json:"cacheWriteTokens,omitempty"`
	TotalTokens      int64 `json:"totalTokens,omitempty"`
}

// Known reports whether any token field is present with a non-zero value
// or TotalTokens was explicitly set. Zero-valued structs are unknown and
// must be omitted from aggregates unless ParseTokenUsage said otherwise.
func (u TokenUsage) Known() bool {
	return u.TotalTokens != 0 ||
		u.InputTokens != 0 ||
		u.OutputTokens != 0 ||
		u.CacheReadTokens != 0 ||
		u.CacheWriteTokens != 0
}

// Consumed is the billed/context metric for operators.
// Prefer TotalTokens when present and > 0; otherwise sum the four parts.
func (u TokenUsage) Consumed() int64 {
	if u.TotalTokens > 0 {
		return u.TotalTokens
	}
	return u.InputTokens + u.OutputTokens + u.CacheReadTokens + u.CacheWriteTokens
}

// AddTokenUsage sums two usages field-wise. Unknown (zero) operands are fine.
func AddTokenUsage(a, b TokenUsage) TokenUsage {
	return TokenUsage{
		InputTokens:      a.InputTokens + b.InputTokens,
		OutputTokens:     a.OutputTokens + b.OutputTokens,
		CacheReadTokens:  a.CacheReadTokens + b.CacheReadTokens,
		CacheWriteTokens: a.CacheWriteTokens + b.CacheWriteTokens,
		TotalTokens:      a.TotalTokens + b.TotalTokens,
	}
}

// ParseTokenUsage decodes a Cursor usage JSON object.
// The second return is false when raw is empty, not an object, or has no
// recognizable token fields (all omitted / null).
func ParseTokenUsage(raw json.RawMessage) (TokenUsage, bool) {
	raw = json.RawMessage(trimSpaceBytes(raw))
	if len(raw) == 0 || string(raw) == "null" {
		return TokenUsage{}, false
	}
	var u TokenUsage
	if err := json.Unmarshal(raw, &u); err != nil {
		return TokenUsage{}, false
	}
	if !u.Known() {
		var probe map[string]json.RawMessage
		if err := json.Unmarshal(raw, &probe); err != nil || len(probe) == 0 {
			return TokenUsage{}, false
		}
		if hasTokenKey(probe) {
			return u, true
		}
		return TokenUsage{}, false
	}
	return u, true
}

// TokenUsageFromDetailsJSON reads the top-level "usage" key from phase details_json.
func TokenUsageFromDetailsJSON(details json.RawMessage) (TokenUsage, bool) {
	details = json.RawMessage(trimSpaceBytes(details))
	if len(details) == 0 || string(details) == "null" || string(details) == "{}" {
		return TokenUsage{}, false
	}
	var envelope struct {
		Usage json.RawMessage `json:"usage"`
	}
	if err := json.Unmarshal(details, &envelope); err != nil {
		return TokenUsage{}, false
	}
	return ParseTokenUsage(envelope.Usage)
}

// PhaseUsageRow is one phase with parseable Cursor usage under details_json.
type PhaseUsageRow struct {
	CycleID    string
	AttemptSeq int64
	Phase      Phase
	Usage      TokenUsage
}

// SumPhaseUsageByKind aggregates usage rows into execute vs verify totals.
func SumPhaseUsageByKind(rows []PhaseUsageRow) (execute, verify TokenUsage) {
	for _, r := range rows {
		switch r.Phase {
		case PhaseExecute:
			execute = AddTokenUsage(execute, r.Usage)
		case PhaseVerify:
			verify = AddTokenUsage(verify, r.Usage)
		}
	}
	return execute, verify
}

// SumPhaseUsageByCycleID aggregates usage rows per cycle id.
func SumPhaseUsageByCycleID(rows []PhaseUsageRow) map[string]TokenUsage {
	out := make(map[string]TokenUsage, len(rows))
	for _, r := range rows {
		out[r.CycleID] = AddTokenUsage(out[r.CycleID], r.Usage)
	}
	return out
}

func hasTokenKey(m map[string]json.RawMessage) bool {
	for _, k := range []string{
		"inputTokens", "outputTokens", "cacheReadTokens", "cacheWriteTokens", "totalTokens",
	} {
		if _, ok := m[k]; ok {
			return true
		}
	}
	return false
}

func trimSpaceBytes(b []byte) []byte {
	i, j := 0, len(b)
	for i < j && (b[i] == ' ' || b[i] == '\n' || b[i] == '\r' || b[i] == '\t') {
		i++
	}
	for j > i && (b[j-1] == ' ' || b[j-1] == '\n' || b[j-1] == '\r' || b[j-1] == '\t') {
		j--
	}
	return b[i:j]
}
