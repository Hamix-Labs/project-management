// Command checkstandards enforces the layout guardrails in
// .cursor/rules/CODE_STANDARDS.mdc and .cursor/rules/web-layout.mdc: import
// boundaries between web verticals, handler and policy package purity, and
// per-zone file-size limits with a burn-down baseline.
//
// Exit 0 when clean; exit 1 when a rule is violated.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
)

// errViolations reports that at least one guardrail failed. The offending
// files are already printed, so main exits without adding a second message.
var errViolations = errors.New("code standards violated")

func main() {
	root := flag.String("root", "", "repository root (default: nearest ancestor with go.mod)")
	flag.Parse()

	if err := run(*root, os.Stdout); err != nil {
		if !errors.Is(err, errViolations) {
			fmt.Fprintf(os.Stderr, "check-code-standards: %v\n", err)
		}
		os.Exit(1)
	}
}

func run(root string, out io.Writer) error {
	slog.Debug("trace", "operation", "checkstandards.run")

	repoRoot, err := resolveRepoRoot(root)
	if err != nil {
		return err
	}

	failed, err := checkContentRules(repoRoot, out)
	if err != nil {
		return err
	}

	size, err := checkFileSizes(repoRoot, out)
	if err != nil {
		return err
	}
	if size.newRed > 0 {
		failed = true
	}

	if failed {
		if size.newRed > 0 {
			fmt.Fprintln(out, "Tip: split the file, or if intentional legacy debt add it to scripts/code-standards-size-baseline.txt (prefer split).")
		}
		return errViolations
	}

	if size.warnings > 0 {
		fmt.Fprintf(out, "check-code-standards: OK (%d file-size warning(s); new red files fail CI)\n", size.warnings)
		return nil
	}
	fmt.Fprintln(out, "check-code-standards: OK")
	return nil
}
