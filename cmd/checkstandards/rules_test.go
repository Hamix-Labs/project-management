package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeRepoFile(t *testing.T, root, rel, content string) string {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir for %s: %v", rel, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", rel, err)
	}
	return path
}

func newRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeRepoFile(t, root, "go.mod", "module example.com/fixture\n\ngo 1.25.0\n")
	return root
}

func runCheck(t *testing.T, root string) (string, error) {
	t.Helper()
	var buf bytes.Buffer
	err := run(root, &buf)
	return buf.String(), err
}

func TestRun_cleanRepoReportsOK(t *testing.T) {
	root := newRepo(t)
	writeRepoFile(t, root, "web/src/api/tasks.ts", "export const get = () => fetch(\"/tasks\");\n")

	out, err := runCheck(t, root)
	if err != nil {
		t.Fatalf("run returned %v, want nil\noutput:\n%s", err, out)
	}
	if !strings.Contains(out, "check-code-standards: OK") {
		t.Fatalf("output missing OK line:\n%s", out)
	}
}

func TestRun_fetchOutsideAPIFails(t *testing.T) {
	root := newRepo(t)
	writeRepoFile(t, root, "web/src/tasks/loadTasks.ts", "export const load = () => fetch(\"/tasks\");\n")

	out, err := runCheck(t, root)
	if !errors.Is(err, errViolations) {
		t.Fatalf("run returned %v, want errViolations\noutput:\n%s", err, out)
	}
	if !strings.Contains(out, "VIOLATION: fetch( outside web/src/api/") {
		t.Fatalf("output missing fetch violation:\n%s", out)
	}
}

func TestRun_testFilesExemptFromFetchRule(t *testing.T) {
	root := newRepo(t)
	writeRepoFile(t, root, "web/src/tasks/loadTasks.test.ts", "it(\"x\", () => fetch(\"/tasks\"));\n")

	if _, err := runCheck(t, root); err != nil {
		t.Fatalf("run returned %v, want nil for a .test.ts file", err)
	}
}

func TestRun_crossVerticalImportFails(t *testing.T) {
	root := newRepo(t)
	writeRepoFile(t, root, "web/src/projects/ProjectCard.tsx", "import { Task } from \"@/tasks/types\";\n")

	out, err := runCheck(t, root)
	if !errors.Is(err, errViolations) {
		t.Fatalf("run returned %v, want errViolations\noutput:\n%s", err, out)
	}
	if !strings.Contains(out, "VIOLATION: projects feature imports @tasks/") {
		t.Fatalf("output missing cross-import violation:\n%s", out)
	}
}

// tasks/ may import settings/ — only projects and worktrees are forbidden.
func TestRun_allowedCrossVerticalImportPasses(t *testing.T) {
	root := newRepo(t)
	writeRepoFile(t, root, "web/src/tasks/TaskCard.tsx", "import { z } from \"@/settings/tz\";\n")

	if _, err := runCheck(t, root); err != nil {
		t.Fatalf("run returned %v, want nil for tasks -> settings", err)
	}
}

func TestRun_invalidateQueriesAllowedOnlyInMutations(t *testing.T) {
	root := newRepo(t)
	writeRepoFile(t, root, "web/src/projects/mutations/useRename.ts", "queryClient.invalidateQueries();\n")

	if _, err := runCheck(t, root); err != nil {
		t.Fatalf("run returned %v, want nil inside mutations/", err)
	}

	writeRepoFile(t, root, "web/src/projects/ProjectPanel.tsx", "queryClient.invalidateQueries();\n")
	out, err := runCheck(t, root)
	if !errors.Is(err, errViolations) {
		t.Fatalf("run returned %v, want errViolations\noutput:\n%s", err, out)
	}
	if !strings.Contains(out, "VIOLATION: invalidateQueries outside allowed paths in projects/") {
		t.Fatalf("output missing invalidateQueries violation:\n%s", out)
	}
}

func TestRun_handlerImportingDBStackFails(t *testing.T) {
	root := newRepo(t)
	writeRepoFile(t, root, "pkgs/tasks/handler/handler_tasks.go", "package handler\n\nimport \"gorm.io/gorm\"\n")

	out, err := runCheck(t, root)
	if !errors.Is(err, errViolations) {
		t.Fatalf("run returned %v, want errViolations\noutput:\n%s", err, out)
	}
	if !strings.Contains(out, "VIOLATION: handler imports DB stack") {
		t.Fatalf("output missing handler DB violation:\n%s", out)
	}
}

func TestRun_policyPackageImportingHTTPFails(t *testing.T) {
	root := newRepo(t)
	writeRepoFile(t, root, "pkgs/tasks/handler/readpolicy/policy.go", "package readpolicy\n\nimport \"net/http\"\n")

	out, err := runCheck(t, root)
	if !errors.Is(err, errViolations) {
		t.Fatalf("run returned %v, want errViolations\noutput:\n%s", err, out)
	}
	if !strings.Contains(out, "VIOLATION: handler policy subpackage imports HTTP/DB") {
		t.Fatalf("output missing policy purity violation:\n%s", out)
	}
}

func TestCheckFileSizes_newRedFileFailsUnlessBaselined(t *testing.T) {
	root := newRepo(t)
	rel := "pkgs/oversized/huge.go"
	writeRepoFile(t, root, rel, "package oversized\n"+strings.Repeat("var _ = 1\n", 900))

	out, err := runCheck(t, root)
	if !errors.Is(err, errViolations) {
		t.Fatalf("run returned %v, want errViolations\noutput:\n%s", err, out)
	}
	if !strings.Contains(out, "VIOLATION: new red-zone file (not in size baseline): "+rel) {
		t.Fatalf("output missing new red-zone violation:\n%s", out)
	}

	writeRepoFile(t, root, "scripts/code-standards-size-baseline.txt", "# legacy debt\n"+rel+"\n")
	out, err = runCheck(t, root)
	if err != nil {
		t.Fatalf("run returned %v after baselining, want nil\noutput:\n%s", err, out)
	}
	if !strings.Contains(out, "SIZE (red): ") {
		t.Fatalf("baselined file should still print a red size line:\n%s", out)
	}
}

func TestCheckFileSizes_generatedGoExempt(t *testing.T) {
	root := newRepo(t)
	writeRepoFile(t, root, "pkgs/gen/models.go",
		"// Code generated by tool. DO NOT EDIT.\n\npackage gen\n"+strings.Repeat("var _ = 1\n", 900))

	out, err := runCheck(t, root)
	if err != nil {
		t.Fatalf("run returned %v, want nil for generated file\noutput:\n%s", err, out)
	}
	if strings.Contains(out, "SIZE (red)") {
		t.Fatalf("generated file should be skipped entirely:\n%s", out)
	}
}
