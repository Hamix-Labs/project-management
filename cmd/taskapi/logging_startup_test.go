package main

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

func TestEmitTaskAPIFileLoggingConfig_emitsAtMinLevel(t *testing.T) {
	for _, lv := range []slog.Level{slog.LevelDebug, slog.LevelInfo, slog.LevelWarn, slog.LevelError} {
		var buf bytes.Buffer
		prev := slog.Default()
		t.Cleanup(func() { slog.SetDefault(prev) })
		slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: lv})))
		emitTaskAPIFileLoggingConfig(lv)
		var line string
		for _, candidate := range strings.Split(buf.String(), "\n") {
			candidate = strings.TrimSpace(candidate)
			if candidate == "" {
				continue
			}
			if strings.Contains(candidate, `"msg":"logging config"`) {
				line = candidate
				break
			}
		}
		if line == "" {
			t.Fatalf("level %s: expected logging config line, got %q", lv, buf.String())
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("level %s: %v (line %q)", lv, err, line)
		}
		if m["msg"] != "logging config" {
			t.Fatalf("level %s: msg %v", lv, m["msg"])
		}
		if m["operation"] != "taskapi.logging" {
			t.Fatalf("level %s: operation %v", lv, m["operation"])
		}
		if m["min_level"] != lv.String() {
			t.Fatalf("level %s: min_level field %v", lv, m["min_level"])
		}
		if m["json_file"] != true {
			t.Fatalf("level %s: json_file %v", lv, m["json_file"])
		}
	}
}
