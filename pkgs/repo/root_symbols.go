package repo

import (
	"bufio"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"
)

const (
	maxSymbolSearchResults = 50
	maxSymbolFileBytes     = 1 << 20 // 1 MiB per file when scanning for symbols
)

// SymbolKind is a best-effort declaration kind from SearchSymbols.
type SymbolKind string

const (
	SymbolKindFunction SymbolKind = "function"
	SymbolKindMethod   SymbolKind = "method"
	SymbolKindClass    SymbolKind = "class"
)

// SymbolHit is one declaration match from SearchSymbols.
type SymbolHit struct {
	Path string     `json:"path"`
	Name string     `json:"name"`
	Line int        `json:"line"`
	Kind SymbolKind `json:"kind"`
}

type symbolPattern struct {
	re   *regexp.Regexp
	kind SymbolKind
	// nameGroup is the capture group index for the symbol name (1-based).
	nameGroup int
}

var symbolPatterns = []symbolPattern{
	// Go
	{regexp.MustCompile(`^\s*func\s+\((?:[^)]+)\)\s+([A-Za-z_][\w]*)\s*\(`), SymbolKindMethod, 1},
	{regexp.MustCompile(`^\s*func\s+([A-Za-z_][\w]*)\s*\(`), SymbolKindFunction, 1},
	{regexp.MustCompile(`^\s*type\s+([A-Za-z_][\w]*)\s+struct\b`), SymbolKindClass, 1},
	{regexp.MustCompile(`^\s*type\s+([A-Za-z_][\w]*)\s+interface\b`), SymbolKindClass, 1},
	// TypeScript / JavaScript
	{regexp.MustCompile(`^\s*(?:export\s+)?(?:async\s+)?function\s+\*?\s*([A-Za-z_$][\w$]*)\s*\(`), SymbolKindFunction, 1},
	{regexp.MustCompile(`^\s*(?:export\s+)?(?:const|let|var)\s+([A-Za-z_$][\w$]*)\s*=\s*(?:async\s*)?\(`), SymbolKindFunction, 1},
	{regexp.MustCompile(`^\s*(?:export\s+)?(?:const|let|var)\s+([A-Za-z_$][\w$]*)\s*=\s*(?:async\s*)?[A-Za-z_$][\w$]*\s*=>`), SymbolKindFunction, 1},
	{regexp.MustCompile(`^\s*(?:export\s+)?class\s+([A-Za-z_$][\w$]*)\b`), SymbolKindClass, 1},
	{regexp.MustCompile(`^\s*(?:public|private|protected|static|async|\s)*([A-Za-z_$][\w$]*)\s*\([^;]*\)\s*\{`), SymbolKindMethod, 1},
	// Python
	{regexp.MustCompile(`^\s*def\s+([A-Za-z_][\w]*)\s*\(`), SymbolKindFunction, 1},
	{regexp.MustCompile(`^\s*async\s+def\s+([A-Za-z_][\w]*)\s*\(`), SymbolKindFunction, 1},
	{regexp.MustCompile(`^\s*class\s+([A-Za-z_][\w]*)\s*[:(]`), SymbolKindClass, 1},
	// Rust
	{regexp.MustCompile(`^\s*(?:pub(?:\([^)]*\))?\s+)?(?:async\s+)?fn\s+([A-Za-z_][\w]*)\s*[<(]`), SymbolKindFunction, 1},
	{regexp.MustCompile(`^\s*(?:pub(?:\([^)]*\))?\s+)?struct\s+([A-Za-z_][\w]*)\b`), SymbolKindClass, 1},
	{regexp.MustCompile(`^\s*(?:pub(?:\([^)]*\))?\s+)?enum\s+([A-Za-z_][\w]*)\b`), SymbolKindClass, 1},
	{regexp.MustCompile(`^\s*(?:pub(?:\([^)]*\))?\s+)?trait\s+([A-Za-z_][\w]*)\b`), SymbolKindClass, 1},
}

var symbolScanExts = map[string]struct{}{
	".go": {}, ".ts": {}, ".tsx": {}, ".js": {}, ".jsx": {}, ".mjs": {}, ".cjs": {},
	".py": {}, ".rs": {},
}

// SearchSymbols returns best-effort declaration hits whose name contains q (case-insensitive).
// Empty q returns an empty slice (no full-repo dump).
func (r *Root) SearchSymbols(query string) ([]SymbolHit, error) {
	slog.Debug("trace", "operation", "repo.Root.SearchSymbols")
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return nil, nil
	}
	var out []SymbolHit
	err := filepath.WalkDir(r.abs, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if shouldSkipSearchDir(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if _, ok := symbolScanExts[ext]; !ok {
			return nil
		}
		rel, relErr := filepath.Rel(r.abs, path)
		if relErr != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		hits, scanErr := scanFileSymbols(path, rel, q)
		if scanErr != nil {
			return nil
		}
		for _, h := range hits {
			out = append(out, h)
			if len(out) >= maxSymbolSearchResults {
				return fs.SkipAll
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func scanFileSymbols(absPath, relPath, qLower string) ([]SymbolHit, error) {
	f, err := os.Open(absPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	limited := io.LimitReader(f, maxSymbolFileBytes)
	sc := bufio.NewScanner(limited)
	// Allow long lines but keep memory bounded.
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var out []SymbolHit
	lineNo := 0
	for sc.Scan() {
		lineNo++
		line := sc.Text()
		if !utf8.ValidString(line) {
			continue
		}
		for _, p := range symbolPatterns {
			m := p.re.FindStringSubmatch(line)
			if m == nil || p.nameGroup >= len(m) {
				continue
			}
			name := m[p.nameGroup]
			if !strings.Contains(strings.ToLower(name), qLower) {
				continue
			}
			out = append(out, SymbolHit{
				Path: relPath,
				Name: name,
				Line: lineNo,
				Kind: p.kind,
			})
			break
		}
	}
	if err := sc.Err(); err != nil {
		return out, err
	}
	return out, nil
}
