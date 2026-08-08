package main

import (
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

// Every pattern is case-insensitive because PowerShell's -match operator is,
// and these rules were ported from scripts/check-code-standards.ps1 without a
// behaviour change.
var (
	tsExts  = []string{".ts", ".tsx"}
	cssExts = []string{".css"}
	goExts  = []string{".go"}

	skipTSTests  = regexp.MustCompile(`(?i)\.test\.(ts|tsx)$`)
	skipTSOnly   = regexp.MustCompile(`(?i)\.test\.ts$`)
	skipGoTests  = regexp.MustCompile(`(?i)_test\.go$`)
	underWebAPI  = regexp.MustCompile(`(?i)/web/src/api/`)
	underTestDir = regexp.MustCompile(`(?i)/test/`)
	underTokens  = regexp.MustCompile(`(?i)/web/src/app/styles/tokens/`)

	patFetch      = regexp.MustCompile(`(?i)(?:^|[^\w.])fetch\s*\(`)
	patTransport  = regexp.MustCompile(`(?i)(?:^|[^\w.])(?:EventSource|WebSocket)\s*\(`)
	patRawColor   = regexp.MustCompile(`(?i)#[0-9a-fA-F]{3,8}\b|rgba?\(|hsla?\(`)
	patTinyFont   = regexp.MustCompile(`(?i)font-size:\s*0\.[0-6][0-9]*rem`)
	patDBStack    = regexp.MustCompile(`(?i)database/sql|jackc/pgx|gorm\.io/gorm`)
	patDBOrHTTP   = regexp.MustCompile(`(?i)database/sql|jackc/pgx|gorm\.io/gorm|net/http`)
	patReact      = regexp.MustCompile(`(?i)from\s+["']react["']|from\s+["']react/`)
	patCreateModa = regexp.MustCompile(`(?i)task-create-modal`)
	patFeatImport = regexp.MustCompile(`(?i)from\s+["']@/(tasks|projects|settings|worktrees)/`)
	patInvalidate = regexp.MustCompile(`(?i)invalidateQueries`)
)

// simpleRule flags any file under dir whose contents match pattern.
type simpleRule struct {
	dir       string
	recursive bool
	exts      []string
	skipName  *regexp.Regexp
	excludes  []*regexp.Regexp
	pattern   *regexp.Regexp
	message   string
}

var simpleRules = []simpleRule{
	{
		dir: "web/src", recursive: true, exts: tsExts,
		skipName: skipTSTests,
		excludes: []*regexp.Regexp{underWebAPI, underTestDir},
		pattern:  patFetch,
		message:  "VIOLATION: fetch( outside web/src/api/: %s",
	},
	{
		dir: "web/src", recursive: true, exts: cssExts,
		excludes: []*regexp.Regexp{underTokens},
		pattern:  patRawColor,
		message:  "VIOLATION: raw color outside web style tokens: %s",
	},
	{
		dir: "web/src", recursive: true, exts: cssExts,
		excludes: []*regexp.Regexp{underTokens},
		pattern:  patTinyFont,
		message:  "VIOLATION: font-size below --text-xs in component CSS: %s",
	},
	{
		dir: "pkgs/tasks/handler", recursive: true, exts: goExts,
		skipName: skipGoTests,
		pattern:  patDBStack,
		message:  "VIOLATION: handler imports DB stack: %s",
	},
	{
		dir: "pkgs/tasks/handler/readpolicy", exts: goExts,
		skipName: skipGoTests,
		pattern:  patDBOrHTTP,
		message:  "VIOLATION: handler policy subpackage imports HTTP/DB: %s",
	},
	{
		dir: "pkgs/tasks/handler/writepolicy", exts: goExts,
		skipName: skipGoTests,
		pattern:  patDBOrHTTP,
		message:  "VIOLATION: handler policy subpackage imports HTTP/DB: %s",
	},
	{
		dir: "web/src/tasks/mutations", exts: []string{".ts"},
		skipName: skipTSOnly,
		pattern:  patReact,
		message:  "VIOLATION: mutations pure module imports react: %s",
	},
	{
		dir: "web/src/tasks/create", exts: []string{".ts"},
		skipName: skipTSOnly,
		pattern:  patReact,
		message:  "VIOLATION: create pure module imports react: %s",
	},
	{
		dir: "web/src/tasks/create", exts: []string{".ts"},
		skipName: skipTSOnly,
		pattern:  patCreateModa,
		message:  "VIOLATION: create pure module imports task-create-modal: %s",
	},
	{
		dir: "web/src/components", recursive: true, exts: tsExts,
		pattern: patFeatImport,
		message: "VIOLATION: inner-ring components/ imports vertical module: %s",
	},
	{
		dir: "web/src/hooks", recursive: true, exts: tsExts,
		pattern: patFeatImport,
		message: "VIOLATION: inner-ring hooks/ imports vertical module: %s",
	},
	{
		dir: "web/src/lib", recursive: true, exts: tsExts,
		pattern: patFeatImport,
		message: "VIOLATION: inner-ring lib/ imports vertical module: %s",
	},
	{
		dir: "web/src/shared", recursive: true, exts: tsExts,
		pattern: patFeatImport,
		message: "VIOLATION: inner-ring shared/ imports vertical module: %s",
	},
	{
		dir: "web/src", recursive: true, exts: tsExts,
		skipName: skipTSTests,
		excludes: []*regexp.Regexp{underWebAPI, underTestDir},
		pattern:  patTransport,
		message:  "VIOLATION: EventSource/WebSocket outside web/src/api/: %s",
	},
}

type verticalRule struct {
	name      string
	forbidden []string
}

var crossImportRules = []verticalRule{
	{name: "tasks", forbidden: []string{"projects", "worktrees"}},
	{name: "projects", forbidden: []string{"tasks", "settings", "worktrees"}},
	{name: "settings", forbidden: []string{"tasks", "projects", "worktrees"}},
	{name: "worktrees", forbidden: []string{"tasks", "projects", "settings"}},
}

var invalidateRules = []struct {
	name  string
	allow []string
}{
	{name: "projects", allow: []string{"mutations"}},
	{name: "worktrees", allow: []string{"mutations"}},
	{name: "tasks", allow: []string{"mutations", "sync"}},
}

func checkContentRules(repoRoot string, out io.Writer) (bool, error) {
	slog.Debug("trace", "operation", "checkstandards.checkContentRules")

	failed := false
	for _, rule := range simpleRules {
		hit, err := runSimpleRule(repoRoot, rule, out)
		if err != nil {
			return failed, err
		}
		failed = failed || hit
	}

	crossHit, err := checkCrossVerticalImports(repoRoot, out)
	if err != nil {
		return failed, err
	}
	failed = failed || crossHit

	invalidateHit, err := checkInvalidateQueries(repoRoot, out)
	if err != nil {
		return failed, err
	}
	return failed || invalidateHit, nil
}

func runSimpleRule(repoRoot string, rule simpleRule, out io.Writer) (bool, error) {
	slog.Debug("trace", "operation", "checkstandards.runSimpleRule")

	files, err := collectFiles(filepath.Join(repoRoot, filepath.FromSlash(rule.dir)), rule.recursive, rule.exts)
	if err != nil {
		return false, err
	}

	failed := false
	for _, path := range files {
		if skipPath(path, rule.skipName, rule.excludes) {
			continue
		}
		text, readErr := readText(path)
		if readErr != nil {
			return failed, readErr
		}
		if rule.pattern.MatchString(text) {
			fmt.Fprintf(out, rule.message+"\n", path)
			failed = true
		}
	}
	return failed, nil
}

func skipPath(path string, skipName *regexp.Regexp, excludes []*regexp.Regexp) bool {
	slog.Debug("trace", "operation", "checkstandards.skipPath")

	if skipName != nil && skipName.MatchString(filepath.Base(path)) {
		return true
	}
	posix := toPosix(path)
	for _, ex := range excludes {
		if ex.MatchString(posix) {
			return true
		}
	}
	return false
}

func checkCrossVerticalImports(repoRoot string, out io.Writer) (bool, error) {
	slog.Debug("trace", "operation", "checkstandards.checkCrossVerticalImports")

	failed := false
	for _, rule := range crossImportRules {
		dir := filepath.Join(repoRoot, "web", "src", rule.name)
		files, err := collectFiles(dir, true, tsExts)
		if err != nil {
			return failed, err
		}
		for _, path := range files {
			text, readErr := readText(path)
			if readErr != nil {
				return failed, readErr
			}
			for _, m := range patFeatImport.FindAllStringSubmatch(text, -1) {
				if !slices.Contains(rule.forbidden, strings.ToLower(m[1])) {
					continue
				}
				fmt.Fprintf(out, "VIOLATION: %s feature imports @%s/: %s\n", rule.name, m[1], path)
				failed = true
			}
		}
	}
	return failed, nil
}

func checkInvalidateQueries(repoRoot string, out io.Writer) (bool, error) {
	slog.Debug("trace", "operation", "checkstandards.checkInvalidateQueries")

	failed := false
	for _, rule := range invalidateRules {
		dir := filepath.Join(repoRoot, "web", "src", rule.name)
		files, err := collectFiles(dir, true, tsExts)
		if err != nil {
			return failed, err
		}
		for _, path := range files {
			if skipTSTests.MatchString(filepath.Base(path)) || inAllowedSubdir(path, rule.allow) {
				continue
			}
			text, readErr := readText(path)
			if readErr != nil {
				return failed, readErr
			}
			if patInvalidate.MatchString(text) {
				fmt.Fprintf(out, "VIOLATION: invalidateQueries outside allowed paths in %s/: %s\n", rule.name, path)
				failed = true
			}
		}
	}
	return failed, nil
}

func inAllowedSubdir(path string, allow []string) bool {
	slog.Debug("trace", "operation", "checkstandards.inAllowedSubdir")

	posix := strings.ToLower(toPosix(path))
	for _, sub := range allow {
		if strings.Contains(posix, "/"+sub+"/") {
			return true
		}
	}
	return false
}
