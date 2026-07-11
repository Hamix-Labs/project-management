#!/usr/bin/env python3
"""One-shot migration: pkgs/tasks/domain -> BC domain imports (Tier 4)."""

from __future__ import annotations

import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
OLD_IMPORT = '"github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"'
SKIP_DIRS = {".git", "node_modules", "vendor", ".codegraph"}

TASKCORE = {
    "Actor", "ActorUser", "ActorAgent", "Task", "Status", "Priority", "GateStatus",
    "TaskGate", "GateKind", "GateCriterion", "PendingRetry", "RetryMode", "RetryFresh",
    "RetryResume", "TaskContextSnapshot", "TaskDependency", "DependencyEdge",
    "DependencySatisfies", "DependencySatisfiesDone", "ErrNotFound", "ErrInvalidInput",
    "ErrConflict", "ValidateTaskTag", "ValidateTaskTags", "ValidateTaskMilestone",
    "NormalizeTaskTags", "GateKindManualApproval", "GateCriteriaAllDone",
    "ValidDependencySatisfies", "NormalizeDependencySatisfies", "StatusReady",
    "StatusRunning", "StatusBlocked", "StatusReview", "StatusDone", "StatusFailed",
    "StatusOnHold", "PriorityLow", "PriorityMedium", "PriorityHigh", "PriorityCritical",
    "GateStatusLocked", "GateStatusActive", "GateStatusPendingRelease", "GateStatusReleased",
}

TASKCYCLES = {
    "TaskCycle", "TaskCyclePhase", "TaskCycleStreamEvent", "TaskCycleCriteriaReport",
    "TaskCycleVerifyReport", "TaskCycleCommandRun", "TaskCycleCommit", "Phase",
    "CycleStatus", "PhaseStatus", "PhaseExecute", "PhaseVerify", "CycleStatusRunning",
    "CycleStatusSucceeded", "CycleStatusFailed", "CycleStatusAborted", "PhaseStatusRunning",
    "PhaseStatusSucceeded", "PhaseStatusFailed", "PhaseStatusSkipped",
    "ExecuteCriteriaReportAttemptSeq", "PhaseInterruptReason", "PhaseDetailsRunCorrelationID",
    "PhaseDetailsSessionID", "ValidPhaseTransition", "ValidInterruptResumeTransition",
    "ValidVerifyOnlyRetryTransition", "TerminalCycleStatus", "TerminalPhaseStatus",
    "RunCorrelationIDFromDetailsJSON", "SessionIDFromDetailsJSON",
}

TASKCHECKLIST = {
    "TaskChecklistItem", "TaskChecklistItemCommand", "TaskChecklistCompletion", "VerifierKind",
    "MaxVerifyCommandsPerItem", "MaxVerifyCommandLen", "MaxVerifyExpectedOutcomeLen",
    "VerifierAgentSelf", "VerifierVerifyAgent", "VerifierDeterministicCheck",
    "VerifierHumanOverride", "VerifierLegacy", "ValidVerifierKind",
}

TASKEVENTS = {
    "TaskEvent", "EventType", "ResponseThreadEntry", "EventTypeAcceptsUserResponse",
    "EventTaskCreated", "EventStatusChanged", "EventPriorityChanged", "EventPromptAppended",
    "EventContextAdded", "EventConstraintAdded", "EventSuccessCriterionAdded",
    "EventNonGoalAdded", "EventPlanAdded", "EventChecklistItemAdded", "EventChecklistItemToggled",
    "EventChecklistItemUpdated", "EventChecklistItemRemoved", "EventMessageAdded",
    "EventArtifactAdded", "EventApprovalRequested", "EventApprovalGranted", "EventTaskCompleted",
    "EventOnTaskDone", "EventTaskFailed", "EventTaskRetryRequested", "EventTaskPickupFailed",
    "EventCycleStarted", "EventCycleCompleted", "EventCycleFailed", "EventPhaseStarted",
    "EventPhaseCompleted", "EventPhaseFailed", "EventPhaseSkipped", "EventSyncPing",
}

PACKAGES = [
    ("taskcoredomain", TASKCORE, "github.com/AlexsanderHamir/Hamix/pkgs/taskcore/domain"),
    ("cyclesdomain", TASKCYCLES, "github.com/AlexsanderHamir/Hamix/pkgs/taskcycles/domain"),
    ("checklistdomain", TASKCHECKLIST, "github.com/AlexsanderHamir/Hamix/pkgs/taskchecklist/domain"),
    ("taskeventsdomain", TASKEVENTS, "github.com/AlexsanderHamir/Hamix/pkgs/taskevents/domain"),
]

SYMBOL_TO_ALIAS: dict[str, str] = {}
for alias, symbols, _ in PACKAGES:
    for sym in symbols:
        SYMBOL_TO_ALIAS[sym] = alias


def find_go_files() -> list[Path]:
    files: list[Path] = []
    for path in ROOT.rglob("*.go"):
        if any(part in SKIP_DIRS for part in path.parts):
            continue
        if "pkgs/tasks/domain" in str(path).replace("\\", "/"):
            continue
        text = path.read_text(encoding="utf-8")
        if OLD_IMPORT in text or "pkgs/tasks/domain" in text:
            files.append(path)
    return sorted(files)


def uses_tasks_domain(text: str) -> bool:
    return OLD_IMPORT in text or re.search(
        r'github\.com/AlexsanderHamir/Hamix/pkgs/tasks/domain', text
    )


def collect_domain_refs(text: str) -> set[str]:
    refs = set(re.findall(r"\bdomain\.([A-Z][A-Za-z0-9_]*)", text))
    refs |= set(re.findall(r"\bdomain\.([A-Z][A-Za-z0-9_]*)\b", text))
    # bare domain import without alias
    if re.search(r'^\s*"github.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"', text, re.M):
        refs |= set(re.findall(r"\bdomain\.([A-Z][A-Za-z0-9_]*)", text))
    return refs


def rewrite_file(path: Path) -> bool:
    text = path.read_text(encoding="utf-8")
    if not uses_tasks_domain(text):
        return False

    refs = collect_domain_refs(text)
    needed_aliases: dict[str, str] = {}
    unknown: set[str] = set()
    for ref in refs:
        alias = SYMBOL_TO_ALIAS.get(ref)
        if alias:
            needed_aliases[alias] = next(p for a, _, p in PACKAGES if a == alias)
        elif ref:
            unknown.add(ref)

    if unknown:
        print(f"SKIP {path}: unknown domain.{unknown}", file=sys.stderr)
        return False

    # Replace domain.Symbol with alias.Symbol
    new_text = text
    for alias, _, _ in PACKAGES:
        symbols = [s for s, a in SYMBOL_TO_ALIAS.items() if a == alias]
        for sym in sorted(symbols, key=len, reverse=True):
            new_text = re.sub(rf"\bdomain\.{sym}\b", f"{alias}.{sym}", new_text)

    # Remove old import lines (plain and alias)
    new_text = re.sub(
        r'^\s*"?github\.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"?\s*\n',
        "",
        new_text,
        flags=re.M,
    )
    new_text = re.sub(
        r'^\s*\w+\s+"github\.com/AlexsanderHamir/Hamix/pkgs/tasks/domain"\s*\n',
        "",
        new_text,
        flags=re.M,
    )

    if not needed_aliases:
        # import only, no refs — drop import
        if OLD_IMPORT not in new_text and "pkgs/tasks/domain" not in new_text:
            path.write_text(new_text, encoding="utf-8", newline="\n")
            return True
        return False

    # Build import block additions
    import_lines = []
    for alias in sorted(needed_aliases.keys()):
        import_lines.append(f'\t{alias} "{needed_aliases[alias]}"')

    # Insert into existing import block or create one
    if re.search(r"^import \(", new_text, re.M):
        def inject(match: re.Match[str]) -> str:
            block = match.group(0)
            for line in import_lines:
                if line.strip().split()[1].strip('"') in block:
                    continue
                if line.split()[0] not in block:
                    block = block.rstrip(")\n") + "\n" + line + "\n)\n"
            return block

        new_text = re.sub(r"import \([^)]*\)", inject, new_text, count=1, flags=re.DOTALL)
    else:
        block = "import (\n" + "\n".join(import_lines) + "\n)\n"
        new_text = re.sub(r"(package \w+\n\n)", r"\1" + block, new_text, count=1)

    if new_text == text:
        return False
    path.write_text(new_text, encoding="utf-8", newline="\n")
    return True


def main() -> int:
    changed = 0
    for path in find_go_files():
        if rewrite_file(path):
            changed += 1
            print(path.relative_to(ROOT))
    print(f"migrated {changed} files", file=sys.stderr)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
