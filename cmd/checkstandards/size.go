package main

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type sizeZone struct {
	name  string
	green int
	red   int
}

type sizeResult struct {
	warnings int
	newRed   int
}

var (
	patGenerated = regexp.MustCompile(`(?m)^// Code generated\b`)
	patVendored  = regexp.MustCompile(`(?i)/node_modules/|/dist/`)

	zoneGoTest     = regexp.MustCompile(`(?i)_test\.go$`)
	zoneCmdMain    = regexp.MustCompile(`(?i)/cmd/[^/]+/main\.go$`)
	zoneDomain     = regexp.MustCompile(`(?i)/domain/`)
	zoneStoreDir   = regexp.MustCompile(`(?i)/store/internal/`)
	zoneStoreFile  = regexp.MustCompile(`(?i)^store_`)
	zoneHandler    = regexp.MustCompile(`(?i)^handler_.*\.go$`)
	zoneJSON       = regexp.MustCompile(`(?i)_json\.go$`)
	zoneMiddleware = regexp.MustCompile(`(?i)/middleware/`)
	zoneGo         = regexp.MustCompile(`(?i)\.go$`)
	zonePage       = regexp.MustCompile(`(?i)page\.tsx$`)
	zoneHook       = regexp.MustCompile(`(?i)^use.+\.ts$`)
	zoneAPI        = regexp.MustCompile(`(?i)/web/src/api/`)
	zoneUtils      = regexp.MustCompile(`(?i)/utils/`)
	zoneTestTSX    = regexp.MustCompile(`(?i)\.test\.tsx$`)
	zoneCSS        = regexp.MustCompile(`(?i)\.css$`)
	zoneContainer  = regexp.MustCompile(`(?i)(section|panel|view|layout)\.tsx$`)
	zoneTSX        = regexp.MustCompile(`(?i)\.tsx$`)
	zoneTS         = regexp.MustCompile(`(?i)\.ts$`)
)

// zoneFor mirrors Get-CodeStandardsSizeZone from the PowerShell original,
// including its match order. relPath is repo-relative with forward slashes.
//
// Note: zoneCmdMain and zoneAPI both require a leading slash that a
// repo-relative path never has, so cmd/*/main.go and web/src/api/*.ts fall
// through to the general Go and TS zones. That quirk is preserved
// deliberately — tightening it would newly fail existing files and belongs in
// its own change.
func zoneFor(relPath, fileName string) (sizeZone, bool) {
	slog.Debug("trace", "operation", "checkstandards.zoneFor")

	n := strings.ToLower(relPath)
	fn := strings.ToLower(fileName)

	switch {
	case zoneGoTest.MatchString(fn):
		return sizeZone{"Go *_test.go", 400, 600}, true
	case zoneCmdMain.MatchString(n):
		return sizeZone{"Go cmd/main.go", 80, 120}, true
	case zoneDomain.MatchString(n):
		return sizeZone{"Go domain", 200, 350}, true
	case zoneStoreDir.MatchString(n) || zoneStoreFile.MatchString(fn):
		return sizeZone{"Go store", 300, 500}, true
	case zoneHandler.MatchString(fn):
		return sizeZone{"Go handler_*.go", 300, 500}, true
	case zoneJSON.MatchString(fn):
		return sizeZone{"Go *_json.go", 200, 350}, true
	case zoneMiddleware.MatchString(n):
		return sizeZone{"Go middleware", 150, 250}, true
	case zoneGo.MatchString(fn):
		return sizeZone{"Go general", 400, 800}, true
	case zonePage.MatchString(fn):
		return sizeZone{"TS *Page.tsx", 80, 150}, true
	case zoneHook.MatchString(fn):
		return sizeZone{"TS use*.ts hook", 80, 150}, true
	case zoneAPI.MatchString(n):
		return sizeZone{"TS api/*.ts", 200, 350}, true
	case zoneUtils.MatchString(n):
		return sizeZone{"TS utils/*.ts", 150, 250}, true
	case zoneTestTSX.MatchString(fn):
		return sizeZone{"TS *.test.tsx", 300, 500}, true
	case zoneCSS.MatchString(fn) && !underTokens.MatchString(n):
		return sizeZone{"TS component CSS", 200, 350}, true
	case zoneContainer.MatchString(fn):
		return sizeZone{"TS container component", 120, 200}, true
	case zoneTSX.MatchString(fn):
		return sizeZone{"TS presentational component", 150, 250}, true
	case zoneTS.MatchString(fn):
		return sizeZone{"TS general", 200, 400}, true
	}
	return sizeZone{}, false
}

// loadSizeBaseline reads the burn-down list of files allowed to sit in the red
// zone. Keys are lower-cased because the PowerShell hashtable it replaces was
// case-insensitive.
func loadSizeBaseline(repoRoot string) (map[string]bool, error) {
	slog.Debug("trace", "operation", "checkstandards.loadSizeBaseline")

	path := filepath.Join(repoRoot, "scripts", "code-standards-size-baseline.txt")
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return map[string]bool{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	baseline := map[string]bool{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		baseline[strings.ToLower(strings.ReplaceAll(line, `\`, "/"))] = true
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	return baseline, nil
}

func collectSizeTargets(repoRoot string) ([]string, error) {
	slog.Debug("trace", "operation", "checkstandards.collectSizeTargets")

	var targets []string
	for _, dir := range []string{"pkgs", "cmd", "internal"} {
		found, err := collectFiles(filepath.Join(repoRoot, dir), true, goExts)
		if err != nil {
			return nil, err
		}
		targets = append(targets, found...)
	}

	webFiles, err := collectFiles(filepath.Join(repoRoot, "web", "src"), true, []string{".ts", ".tsx", ".css"})
	if err != nil {
		return nil, err
	}
	for _, path := range webFiles {
		if patVendored.MatchString(toPosix(path)) {
			continue
		}
		targets = append(targets, path)
	}
	return targets, nil
}

func checkFileSizes(repoRoot string, out io.Writer) (sizeResult, error) {
	slog.Debug("trace", "operation", "checkstandards.checkFileSizes")

	var result sizeResult
	baseline, err := loadSizeBaseline(repoRoot)
	if err != nil {
		return result, err
	}
	targets, err := collectSizeTargets(repoRoot)
	if err != nil {
		return result, err
	}

	for _, path := range targets {
		rel := strings.TrimLeft(strings.TrimPrefix(path, repoRoot), `\/`)
		relPosix := toPosix(rel)
		zone, ok := zoneFor(relPosix, filepath.Base(path))
		if !ok {
			continue
		}

		text, readErr := readText(path)
		if readErr != nil {
			return result, readErr
		}
		if strings.EqualFold(filepath.Ext(path), ".go") && patGenerated.MatchString(text) {
			continue
		}

		lines := countLines(text)
		if lines <= zone.green {
			continue
		}

		result.warnings++
		if lines <= zone.red {
			fmt.Fprintf(out, "SIZE (yellow): %d lines [%s] %s\n", lines, zone.name, relPosix)
			continue
		}

		fmt.Fprintf(out, "SIZE (red): %d lines [%s] %s\n", lines, zone.name, relPosix)
		if !baseline[strings.ToLower(relPosix)] {
			fmt.Fprintf(out, "VIOLATION: new red-zone file (not in size baseline): %s (%d > %d)\n", relPosix, lines, zone.red)
			result.newRed++
		}
	}
	return result, nil
}
