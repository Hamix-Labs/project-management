package draftsidecar

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// BinEnv is the optional absolute (or cwd-relative) path to the
// hamix-draft-agent launcher. When set, ResolveBinary does not fall
// through to sibling or PATH lookup if the file is missing.
const BinEnv = "HAMIX_DRAFT_AGENT_BIN"

// ResolveBinary locates the hamix-draft-agent launcher. First hit wins:
// HAMIX_DRAFT_AGENT_BIN, a launcher next to the current executable, then
// exec.LookPath(BinaryName).
//
//funclogmeasure:skip category=tool-required-noop reason="Boot-time path lookup; no production operation boundary."
func ResolveBinary() (string, error) {
	exeDir := ""
	if exe, err := os.Executable(); err == nil {
		exeDir = filepath.Dir(exe)
	}
	return resolveBinary(strings.TrimSpace(os.Getenv(BinEnv)), exeDir, exec.LookPath)
}

type lookPathFunc func(file string) (string, error)

func resolveBinary(envPath, exeDir string, lookPath lookPathFunc) (string, error) {
	if envPath != "" {
		abs, err := filepath.Abs(envPath)
		if err != nil {
			return "", fmt.Errorf("draftsidecar: %s: %w", BinEnv, err)
		}
		if err := checkLauncher(abs); err != nil {
			return "", fmt.Errorf("draftsidecar: %s=%s: %w (set %s to a built launcher or start via scripts/dev.*)", BinEnv, abs, err, BinEnv)
		}
		return abs, nil
	}
	if exeDir != "" {
		if sibling := findSibling(exeDir); sibling != "" {
			return sibling, nil
		}
	}
	if lookPath != nil {
		if found, err := lookPath(BinaryName); err == nil && found != "" {
			return found, nil
		}
	}
	return "", fmt.Errorf("draftsidecar: %s not found: set %s or start via scripts/dev.* so the launcher is built", BinaryName, BinEnv)
}

func findSibling(dir string) string {
	for _, name := range launcherNames() {
		p := filepath.Join(dir, name)
		if checkLauncher(p) == nil {
			return p
		}
	}
	return ""
}

func launcherNames() []string {
	return []string{BinaryName, BinaryName + ".cmd", BinaryName + ".exe"}
}

func checkLauncher(path string) error {
	st, err := os.Stat(path)
	if err != nil {
		return err
	}
	if st.IsDir() {
		return fmt.Errorf("is a directory")
	}
	return nil
}
