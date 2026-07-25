package verify

import "github.com/AlexsanderHamir/Hamix/pkgs/obs/calltrace"
import (
	"context"
	"encoding/json"
	"fmt"
	cyclesstore "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/store"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/AlexsanderHamir/Hamix/pkgs/agents/harness/internal/reports"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner"
	"github.com/AlexsanderHamir/Hamix/pkgs/agents/runner/adapterkit"
	settingsdomain "github.com/AlexsanderHamir/Hamix/pkgs/settings/domain"
)

const maxCommandOutputBytes = 256 * 1024

const inlineCommandPreviewLines = 40

// ProgressToolVerifyCommand is the ProgressEvent.Tool value for worker-executed
// checklist verify commands (live ticker + durable stream).
const ProgressToolVerifyCommand = "verify_command"

const maxProgressCommandChars = 120

// commandProgressHeartbeat is the interval between "running" progress events
// while a verify shell command blocks. Tests may shorten this.
var commandProgressHeartbeat = 5 * time.Second

// CommandEvidence captures one verify-phase command run for LLM prompt and audit.
type CommandEvidence struct {
	CriterionID     string
	CommandSeq      int
	Command         string
	ExpectedOutcome string
	ExitCode        int
	DurationMS      int64
	StdoutPath      string
	StderrPath      string
	MetaPath        string
	Truncated       bool
	RunError        string
	StdoutPreview   string
}

type commandMetaFile struct {
	CriterionID     string `json:"criterion_id"`
	Seq             int    `json:"seq"`
	Command         string `json:"command"`
	ExpectedOutcome string `json:"expected_outcome"`
	ExitCode        int    `json:"exit_code"`
	DurationMS      int64  `json:"duration_ms"`
	StdoutBytes     int    `json:"stdout_bytes"`
	StderrBytes     int    `json:"stderr_bytes"`
	Truncated       bool   `json:"truncated"`
	Error           string `json:"error,omitempty"`
}

type shellExecFunc func(ctx context.Context, dir string, command string) (stdout, stderr []byte, exitCode int, err error)

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func commandEvidenceDir(reportDir, cycleID, criterionID string) string {
	return filepath.Join(reportDir, cycleID, "checks", criterionID)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func commandArtifactBase(reportDir, cycleID, criterionID string, seq int) string {
	return filepath.Join(commandEvidenceDir(reportDir, cycleID, criterionID), fmt.Sprintf("%d", seq))
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func truncateCommandOutput(b []byte) ([]byte, bool) {
	if len(b) <= maxCommandOutputBytes {
		return b, false
	}
	return b[:maxCommandOutputBytes], true
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func previewCommandOutput(path string) string {
	b, err := os.ReadFile(path)
	if err != nil || len(b) == 0 {
		return ""
	}
	if len(b) <= 4096 {
		return string(b)
	}
	lines := strings.Split(string(b), "\n")
	if len(lines) <= inlineCommandPreviewLines {
		return string(b[:4096])
	}
	return strings.Join(lines[:inlineCommandPreviewLines], "\n")
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func shellCommand(command string) (string, []string) {
	if runtime.GOOS == "windows" {
		return "cmd", []string{"/C", command}
	}
	return "sh", []string{"-c", command}
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func defaultShellExec(ctx context.Context, dir, command string) ([]byte, []byte, int, error) {
	shell, args := shellCommand(command)
	return adapterkit.DefaultExec(ctx, dir, os.Environ(), nil, shell, args...)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func truncateProgressCommand(command string) string {
	command = strings.TrimSpace(command)
	runes := []rune(command)
	if len(runes) <= maxProgressCommandChars {
		return command
	}
	return string(runes[:maxProgressCommandChars-1]) + "…"
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func formatProgressElapsed(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	sec := int(d.Seconds())
	if sec < 60 {
		return fmt.Sprintf("%ds", sec)
	}
	return fmt.Sprintf("%dm %ds", sec/60, sec%60)
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func verifyCommandProgressPayload(criterionID string, commandSeq int, command string) json.RawMessage {
	b, err := json.Marshal(map[string]any{
		"criterion_id": criterionID,
		"command_seq":  commandSeq,
		"command":      command,
	})
	if err != nil {
		return []byte("{}")
	}
	return b
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (s *Service) emitCommandProgress(
	ctx context.Context,
	taskID, cycleID string,
	phaseSeq int64,
	subtype, message, criterionID string,
	commandSeq int,
	command string,
) {
	if s == nil || s.hooks.PersistProgress == nil || taskID == "" || phaseSeq <= 0 {
		return
	}
	s.hooks.PersistProgress(ctx, taskID, cycleID, phaseSeq, runner.ProgressEvent{
		Kind:    "tool_call",
		Subtype: subtype,
		Message: message,
		Tool:    ProgressToolVerifyCommand,
		Payload: verifyCommandProgressPayload(criterionID, commandSeq, command),
	})
}

// startCommandProgressHeartbeat emits "running" events until stop is closed.
//
// progressCtx is the run lifetime (parent/worker cancel or stop after Wait).
// It must NOT be the per-command kill timer: observability outlives a kill
// attempt because Wait can still block (e.g. Windows process trees).
// execCtx is the kill timer only; when it expires while Wait is still
// outstanding, one boundary event is emitted and heartbeats continue.
//
//funclogmeasure:skip category=hot-path reason="Goroutine helper; progress traces via PersistProgress."
func (s *Service) startCommandProgressHeartbeat(
	progressCtx context.Context,
	execCtx context.Context,
	taskID, cycleID string,
	phaseSeq int64,
	commandLabel, criterionID string,
	commandSeq int,
	command string,
	started time.Time,
) (stop func()) {
	interval := commandProgressHeartbeat
	if interval <= 0 {
		return func() {}
	}
	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		var execDone <-chan struct{}
		if execCtx != nil {
			execDone = execCtx.Done()
		}
		for {
			select {
			case <-done:
				return
			case <-progressCtx.Done():
				return
			case <-execDone:
				execDone = nil
				elapsed := s.clock().Sub(started)
				s.emitCommandProgress(progressCtx, taskID, cycleID, phaseSeq, "running",
					fmt.Sprintf("Timed out waiting for: %s (%s); still waiting for process exit",
						commandLabel, formatProgressElapsed(elapsed)),
					criterionID, commandSeq, command)
			case <-ticker.C:
				elapsed := s.clock().Sub(started)
				s.emitCommandProgress(progressCtx, taskID, cycleID, phaseSeq, "running",
					fmt.Sprintf("Running: %s (%s)", commandLabel, formatProgressElapsed(elapsed)),
					criterionID, commandSeq, command)
			}
		}
	}()
	return func() { close(done) }
}

// RunCriterionCommands executes checklist verify commands and records artifact paths.
func (s *Service) RunCriterionCommands(
	parentCtx context.Context,
	taskID string,
	cycleID string,
	phaseSeq int64,
	attemptSeq int64,
	snap Snapshot,
	selfReport map[string]reports.CriteriaEntry,
	execFn shellExecFunc,
) ([]CommandEvidence, error) {
	slog.Debug("trace", "cmd", calltrace.LogCmd, "operation", "agent.harness.verify.RunCriterionCommands",
		"task_id", taskID, "cycle_id", cycleID, "phase_seq", phaseSeq, "attempt_seq", attemptSeq)
	if execFn == nil {
		execFn = defaultShellExec
	}
	timeout := time.Duration(snap.VerifyCommandTimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = time.Duration(settingsdomain.DefaultVerifyCommandTimeoutSeconds) * time.Second
	}
	var out []CommandEvidence
	var persist []cyclesstore.CommandRunEntry
	for _, it := range snap.Criteria {
		entry, ok := selfReport[it.ID]
		if !ok || !entry.ClaimedDone || len(it.VerifyCommands) == 0 {
			continue
		}
		dir := commandEvidenceDir(s.reportDir, cycleID, it.ID)
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return nil, fmt.Errorf("mkdir command evidence %s: %w", dir, err)
		}
		for _, cmd := range it.VerifyCommands {
			seq := cmd.SortOrder
			base := commandArtifactBase(s.reportDir, cycleID, it.ID, seq)
			stdoutPath := base + ".stdout"
			stderrPath := base + ".stderr"
			metaPath := base + ".meta.json"

			commandLabel := truncateProgressCommand(cmd.Command)
			// progressCtx: run lifetime for live/durable updates (outlives kill attempts).
			// execCtx: per-command kill timer only — never used for progress emission.
			progressCtx := parentCtx
			s.emitCommandProgress(progressCtx, taskID, cycleID, phaseSeq, "started",
				fmt.Sprintf("Running: %s", commandLabel),
				it.ID, seq, cmd.Command)

			execCtx, cancel := context.WithTimeout(parentCtx, timeout)
			started := s.clock()
			stopHeartbeat := s.startCommandProgressHeartbeat(
				progressCtx, execCtx, taskID, cycleID, phaseSeq, commandLabel, it.ID, seq, cmd.Command, started,
			)
			stdout, stderr, exitCode, runErr := execFn(execCtx, s.workingDir, cmd.Command)
			stopHeartbeat()
			cancel()
			durationMS := s.clock().Sub(started).Milliseconds()

			stdout, truncOut := truncateCommandOutput(stdout)
			stderr, truncErr := truncateCommandOutput(stderr)
			truncated := truncOut || truncErr

			if werr := os.WriteFile(stdoutPath, stdout, 0o600); werr != nil {
				return nil, fmt.Errorf("write stdout evidence: %w", werr)
			}
			if werr := os.WriteFile(stderrPath, stderr, 0o600); werr != nil {
				return nil, fmt.Errorf("write stderr evidence: %w", werr)
			}

			meta := commandMetaFile{
				CriterionID:     it.ID,
				Seq:             seq,
				Command:         cmd.Command,
				ExpectedOutcome: cmd.ExpectedOutcome,
				ExitCode:        exitCode,
				DurationMS:      durationMS,
				StdoutBytes:     len(stdout),
				StderrBytes:     len(stderr),
				Truncated:       truncated,
			}
			if runErr != nil {
				meta.Error = runErr.Error()
				meta.ExitCode = -1
				exitCode = -1
			}
			metaBytes, _ := json.Marshal(meta)
			if werr := os.WriteFile(metaPath, metaBytes, 0o600); werr != nil {
				return nil, fmt.Errorf("write meta evidence: %w", werr)
			}

			doneSubtype := "completed"
			doneMsg := fmt.Sprintf("Finished: %s (exit %d, %s)", commandLabel, exitCode, formatProgressElapsed(time.Duration(durationMS)*time.Millisecond))
			if runErr != nil || exitCode != 0 {
				doneSubtype = "failed"
				doneMsg = fmt.Sprintf("Failed: %s (exit %d, %s)", commandLabel, exitCode, formatProgressElapsed(time.Duration(durationMS)*time.Millisecond))
			}
			s.emitCommandProgress(progressCtx, taskID, cycleID, phaseSeq, doneSubtype, doneMsg, it.ID, seq, cmd.Command)

			ev := CommandEvidence{
				CriterionID:     it.ID,
				CommandSeq:      seq,
				Command:         cmd.Command,
				ExpectedOutcome: cmd.ExpectedOutcome,
				ExitCode:        exitCode,
				DurationMS:      durationMS,
				StdoutPath:      stdoutPath,
				StderrPath:      stderrPath,
				MetaPath:        metaPath,
				Truncated:       truncated,
				StdoutPreview:   previewCommandOutput(stdoutPath),
			}
			if runErr != nil {
				ev.RunError = runErr.Error()
			}
			out = append(out, ev)
			persist = append(persist, cyclesstore.CommandRunEntry{
				CriterionID: it.ID,
				CommandSeq:  int64(seq),
				ExitCode:    exitCode,
				MetaPath:    metaPath,
			})
		}
	}
	if len(persist) > 0 {
		if err := s.store.UpsertCommandRuns(parentCtx, cycleID, attemptSeq, persist); err != nil {
			slog.Warn("agent harness UpsertCommandRuns failed",
				"cmd", calltrace.LogCmd, "operation", "agent.harness.verify.RunCriterionCommands.upsert_err",
				"cycle_id", cycleID, "attempt_seq", attemptSeq, "err", err)
		}
	}
	return out, nil
}

// FormatCommandEvidenceSection renders worker command output for verify prompts.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func FormatCommandEvidenceSection(evidence []CommandEvidence) string {
	if len(evidence) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("\n## Command evidence (worker-executed)\n\n")
	for _, ev := range evidence {
		b.WriteString(fmt.Sprintf("### [%s] command %d\n", ev.CriterionID, ev.CommandSeq))
		b.WriteString(fmt.Sprintf("Command: %s\n", ev.Command))
		if ev.ExpectedOutcome != "" {
			b.WriteString(fmt.Sprintf("Expected outcome: %s\n", ev.ExpectedOutcome))
		}
		b.WriteString(fmt.Sprintf("exit_code=%d duration_ms=%d truncated=%v\n", ev.ExitCode, ev.DurationMS, ev.Truncated))
		if ev.RunError != "" {
			b.WriteString(fmt.Sprintf("run_error: %s\n", ev.RunError))
		}
		b.WriteString(fmt.Sprintf("stdout: `%s`\n", ev.StdoutPath))
		b.WriteString(fmt.Sprintf("stderr: `%s`\n", ev.StderrPath))
		b.WriteString(fmt.Sprintf("meta: `%s`\n", ev.MetaPath))
		if ev.StdoutPreview != "" {
			b.WriteString("stdout preview:\n```\n")
			b.WriteString(ev.StdoutPreview)
			if !strings.HasSuffix(ev.StdoutPreview, "\n") {
				b.WriteString("\n")
			}
			b.WriteString("```\n")
		}
		b.WriteString("\n")
	}
	return b.String()
}
