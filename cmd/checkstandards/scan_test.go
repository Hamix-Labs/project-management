package main

import "testing"

func TestCountLines_skipsBlankLines(t *testing.T) {
	for _, tt := range []struct {
		name string
		text string
		want int
	}{
		{name: "empty file", text: "", want: 0},
		{name: "single newline", text: "\n", want: 0},
		{name: "trailing newline not counted twice", text: "a\nb\n", want: 2},
		{name: "no trailing newline", text: "a\nb", want: 2},
		{name: "blank lines excluded", text: "a\n\nb\n", want: 2},
		{name: "whitespace-only line counted", text: "a\n \nb\n", want: 3},
		{name: "crlf blank lines excluded", text: "a\r\n\r\nb\r\n", want: 2},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := countLines(tt.text); got != tt.want {
				t.Fatalf("countLines(%q) = %d, want %d", tt.text, got, tt.want)
			}
		})
	}
}

func TestZoneFor_selectsExpectedZone(t *testing.T) {
	for _, tt := range []struct {
		name     string
		rel      string
		file     string
		wantZone string
		wantRed  int
		wantOK   bool
	}{
		{
			name: "go test file", rel: "pkgs/tasks/store_x_test.go", file: "store_x_test.go",
			wantZone: "Go *_test.go", wantRed: 600, wantOK: true,
		},
		{
			name: "domain beats general", rel: "pkgs/taskcore/domain/task.go", file: "task.go",
			wantZone: "Go domain", wantRed: 350, wantOK: true,
		},
		{
			name: "handler prefix", rel: "pkgs/taskcore/handler/handler_tasks.go", file: "handler_tasks.go",
			wantZone: "Go handler_*.go", wantRed: 500, wantOK: true,
		},
		{
			name: "middleware", rel: "pkgs/tasks/middleware/metrics_http.go", file: "metrics_http.go",
			wantZone: "Go middleware", wantRed: 250, wantOK: true,
		},
		{
			name: "page component", rel: "web/src/tasks/pages/TaskListPage.tsx", file: "TaskListPage.tsx",
			wantZone: "TS *Page.tsx", wantRed: 150, wantOK: true,
		},
		{
			name: "hook", rel: "web/src/tasks/hooks/useTaskCycles.ts", file: "useTaskCycles.ts",
			wantZone: "TS use*.ts hook", wantRed: 150, wantOK: true,
		},
		{
			name: "container component", rel: "web/src/tasks/TaskListSection.tsx", file: "TaskListSection.tsx",
			wantZone: "TS container component", wantRed: 200, wantOK: true,
		},
		{
			name: "component css", rel: "web/src/app/styles/base/app-shell.css", file: "app-shell.css",
			wantZone: "TS component CSS", wantRed: 350, wantOK: true,
		},
		{
			name: "unzoned extension", rel: "docs/api.md", file: "api.md", wantOK: false,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			zone, ok := zoneFor(tt.rel, tt.file)
			if ok != tt.wantOK {
				t.Fatalf("zoneFor(%q) ok = %v, want %v", tt.rel, ok, tt.wantOK)
			}
			if !tt.wantOK {
				return
			}
			if zone.name != tt.wantZone || zone.red != tt.wantRed {
				t.Fatalf("zoneFor(%q) = %q/red %d, want %q/red %d", tt.rel, zone.name, zone.red, tt.wantZone, tt.wantRed)
			}
		})
	}
}

// The cmd and api zone patterns require a leading slash that repo-relative
// paths never carry, so both fall through. Pinned so a future fix is a
// deliberate change with a visible diff rather than a silent CI tightening.
func TestZoneFor_leadingSlashZonesFallThrough(t *testing.T) {
	zone, ok := zoneFor("cmd/taskapi/main.go", "main.go")
	if !ok || zone.name != "Go general" {
		t.Fatalf("cmd main zone = %q (ok=%v), want %q", zone.name, ok, "Go general")
	}

	zone, ok = zoneFor("web/src/api/tasks.ts", "tasks.ts")
	if !ok || zone.name != "TS general" {
		t.Fatalf("api zone = %q (ok=%v), want %q", zone.name, ok, "TS general")
	}
}
