package repo

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseSearchKinds(t *testing.T) {
	t.Parallel()
	cases := []struct {
		raw     string
		wantOK  bool
		wantFile bool
		wantDir  bool
	}{
		{"", true, true, false},
		{"file", true, true, false},
		{"dir", true, false, true},
		{"file,dir", true, true, true},
		{" dir , FILE ", true, true, true},
		{"bogus", false, false, false},
	}
	for _, tc := range cases {
		got, ok := ParseSearchKinds(tc.raw)
		if ok != tc.wantOK {
			t.Fatalf("ParseSearchKinds(%q) ok=%v want %v", tc.raw, ok, tc.wantOK)
		}
		if !ok {
			continue
		}
		if got.File != tc.wantFile || got.Dir != tc.wantDir {
			t.Fatalf("ParseSearchKinds(%q) = %+v want file=%v dir=%v", tc.raw, got, tc.wantFile, tc.wantDir)
		}
	}
}

func TestRoot_SearchEntries_dirs(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	mustWrite := func(rel, body string) {
		t.Helper()
		p := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mustWrite("pkgs/repo/root.go", "package repo\n")
	mustWrite("pkgs/other/x.go", "package other\n")

	r, err := OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	entries, err := r.SearchEntries("repo", SearchKinds{File: true, Dir: true})
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]EntryKind{}
	for _, e := range entries {
		got[e.Path] = e.Kind
	}
	if got["pkgs/repo"] != EntryKindDir {
		t.Fatalf("missing dir pkgs/repo in %v", entries)
	}
	if got["pkgs/repo/root.go"] != EntryKindFile {
		t.Fatalf("missing file pkgs/repo/root.go in %v", entries)
	}
}

func TestRoot_SearchSymbols(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	goSrc := filepath.Join(dir, "sample.go")
	body := "package sample\n\nfunc OpenRoot() {}\n\nfunc (r *Root) Search() {}\n\ntype Root struct{}\n"
	if err := os.WriteFile(goSrc, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	r, err := OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("empty query returns empty", func(t *testing.T) {
		hits, err := r.SearchSymbols("")
		if err != nil {
			t.Fatal(err)
		}
		if len(hits) != 0 {
			t.Fatalf("hits = %v want []", hits)
		}
	})

	t.Run("matches function name", func(t *testing.T) {
		hits, err := r.SearchSymbols("Open")
		if err != nil {
			t.Fatal(err)
		}
		if len(hits) < 1 {
			t.Fatalf("expected at least one hit, got %v", hits)
		}
		found := false
		for _, h := range hits {
			if h.Name == "OpenRoot" && h.Path == "sample.go" && h.Line > 0 {
				found = true
			}
		}
		if !found {
			t.Fatalf("OpenRoot not found in %v", hits)
		}
	})
}
