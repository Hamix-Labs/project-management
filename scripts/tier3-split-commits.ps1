# Builds Tier 3 atomic commit chain from tier3-snapshot onto current branch.
# Usage: from clean main: git checkout -b tier3/backend-blueprint main; .\scripts\tier3-split-commits.ps1
$ErrorActionPreference = "Stop"
$Snap = "tier3-snapshot"

function Take-FromSnapshot {
    param([string[]]$Paths)
    foreach ($p in $Paths) {
        if ($p -match '/$') {
            git checkout $Snap -- $p 2>$null
        } else {
            git checkout $Snap -- $p 2>$null
        }
    }
}

function Commit-Tier3 {
    param([string]$Message, [string[]]$Paths)
    Take-FromSnapshot -Paths $Paths
    git add -A
    $status = git status --porcelain
    if (-not $status) {
        Write-Host "SKIP (empty): $Message"
        return
    }
    git commit -m $Message
    Write-Host "OK: $Message"
}

# --- PR #176 projects contract ---
Commit-Tier3 "docs(adr): ADR-0055 contract colocation" @(
    "docs/adr/ADR-0055-contract-colocation.md",
    "pkgs/projects/contract/doc.go"
)
Commit-Tier3 "refactor(projects): add contract package with ProjectStore" @(
    "pkgs/projects/contract/",
    "pkgs/projects/README.md"
)
Commit-Tier3 "refactor(tasks/contract): alias ProjectStore from projects/contract" @(
    "pkgs/tasks/contract/project.go"
)
git branch -f refactor/projects-contract HEAD

# --- PR #177 settings contract ---
Commit-Tier3 "refactor(settings): add contract package with SettingsStore" @(
    "pkgs/settings/contract/",
    "pkgs/settings/store/store.go",
    "pkgs/settings/store/internal/settings/settings.go",
    "pkgs/settings/handler/handler.go",
    "pkgs/settings/handler/handler_settings.go",
    "pkgs/settings/handler/http_helpers.go",
    "pkgs/settings/handler/handler_http_settings_contract_test.go"
)
Commit-Tier3 "refactor(tasks/contract): alias SettingsStore from settings/contract" @(
    "pkgs/tasks/contract/settings.go"
)
Commit-Tier3 "docs(settings): README contract layer" @(
    "pkgs/settings/README.md"
)
git branch -f refactor/settings-contract HEAD

# --- PR #178 gitinventory contract ---
Commit-Tier3 "refactor(gitinventory): add contract package" @(
    "pkgs/gitinventory/contract/",
    "pkgs/gitinventory/handler/",
    "pkgs/gitinventory/store/",
    "pkgs/gitinventory/README.md"
)
Commit-Tier3 "refactor(tasks/contract): alias GitReadStore and GitWriteStore" @(
    "pkgs/tasks/contract/git_read.go"
)
Commit-Tier3 "chore(ci): gitinventory contract boundary + docs" @(
    "scripts/check-go.sh"
)
git branch -f refactor/gitinventory-contract HEAD

# --- PR #179 taskchecklist domain ---
Commit-Tier3 "docs(adr): ADR-0056 taskchecklist domain ownership" @(
    "docs/adr/ADR-0056-taskchecklist-domain-model.md",
    "pkgs/taskchecklist/domain/doc.go"
)
Commit-Tier3 "refactor(taskchecklist): add domain package" @(
    "pkgs/taskchecklist/domain/"
)
Commit-Tier3 "refactor(tasks/domain): alias checklist types from taskchecklist/domain" @(
    "pkgs/tasks/domain/checklist_aliases.go",
    "pkgs/tasks/domain/models.go",
    "pkgs/tasks/domain/verifier_kind.go",
    "pkgs/tasks/domain/verify_commands.go"
)
Commit-Tier3 "refactor(taskchecklist): point contract and store at local domain" @(
    "pkgs/taskchecklist/contract/",
    "pkgs/taskchecklist/handler/",
    "pkgs/taskchecklist/store/",
    "pkgs/taskchecklist/README.md"
)
git branch -f refactor/taskchecklist-domain HEAD

# --- PR #180 taskchecklist model ---
Commit-Tier3 "refactor(taskchecklist): add store/model package" @(
    "pkgs/taskchecklist/store/model/"
)
Commit-Tier3 "refactor(tasks/store/model): import checklist models from taskchecklist" @(
    "pkgs/tasks/store/model/migrate_models.go",
    "pkgs/tasks/store/model/parity.go"
)
Commit-Tier3 "refactor(taskchecklist/store): use local model mappers" @(
    "pkgs/taskchecklist/store/internal/checklist/"
)
Commit-Tier3 "refactor(tasks): remove duplicated checklist model files" @(
    "pkgs/tasks/store/model/task_checklist_item.go",
    "pkgs/tasks/store/model/task_checklist_item_command.go",
    "pkgs/tasks/store/model/task_checklist_completion.go",
    "pkgs/tasks/store/model/map_checklist.go"
)
Commit-Tier3 "chore(ci): taskchecklist domain boundary + README" @(
    "scripts/check-go.sh",
    "docs/agent-map.md"
)
git branch -f refactor/taskchecklist-model HEAD

# --- PR #181 taskevents domain ---
Commit-Tier3 "docs(adr): ADR-0057 taskevents domain ownership" @(
    "docs/adr/ADR-0057-taskevents-domain-model.md",
    "pkgs/taskevents/domain/"
)
Commit-Tier3 "refactor(tasks/domain): alias TaskEvent and EventType from taskevents/domain" @(
    "pkgs/tasks/domain/event_aliases.go",
    "pkgs/tasks/domain/enums.go",
    "pkgs/tasks/domain/response_thread.go",
    "pkgs/tasks/domain/event_user_response.go",
    "pkgs/tasks/domain/event_user_response_test.go",
    "pkgs/tasks/domain/event_types_manifest_test.go"
)
Commit-Tier3 "refactor(taskevents): point contract, store, handler at local domain" @(
    "pkgs/taskevents/contract/",
    "pkgs/taskevents/handler/",
    "pkgs/taskevents/store/",
    "pkgs/taskevents/README.md"
)
git branch -f refactor/taskevents-domain HEAD

# --- PR #182 taskevents model ---
Commit-Tier3 "refactor(taskevents): add store/model for task_events table" @(
    "pkgs/taskevents/store/model/"
)
Commit-Tier3 "refactor(tasks/store/model): wire migrate to taskevents models" @(
    "pkgs/tasks/store/model/migrate_models.go",
    "pkgs/tasks/store/model/parity.go"
)
Commit-Tier3 "refactor(taskevents/store): use local model mappers" @(
    "pkgs/taskevents/store/internal/events/"
)
Commit-Tier3 "refactor(tasks): remove central task_event model files" @(
    "pkgs/tasks/store/model/task_event.go",
    "pkgs/tasks/store/model/map_task_event.go"
)
Commit-Tier3 "chore(ci): taskevents domain boundary + README" @(
    "scripts/check-go.sh",
    "docs/agent-map.md"
)
git branch -f refactor/taskevents-model HEAD

# --- PR #183 taskcycles domain ---
Commit-Tier3 "docs(adr): ADR-0058 taskcycles domain ownership" @(
    "docs/adr/ADR-0058-taskcycles-domain-model.md"
)
Commit-Tier3 "refactor(taskcycles): add domain package — entity types" @(
    "pkgs/taskcycles/domain/cycle.go",
    "pkgs/taskcycles/domain/enums.go",
    "pkgs/taskcycles/domain/reports.go",
    "pkgs/taskcycles/domain/doc.go"
)
Commit-Tier3 "refactor(taskcycles): add domain package — state machine" @(
    "pkgs/taskcycles/domain/cycle_state.go",
    "pkgs/taskcycles/domain/cycle_state_test.go",
    "pkgs/taskcycles/domain/cycle_correlation.go",
    "pkgs/taskcycles/domain/cycle_correlation_test.go"
)
Commit-Tier3 "refactor(tasks/domain): alias cycle types from taskcycles/domain" @(
    "pkgs/tasks/domain/cycle_aliases.go",
    "pkgs/tasks/domain/cycle_state.go",
    "pkgs/tasks/domain/cycle_state_test.go",
    "pkgs/tasks/domain/cycle_correlation.go",
    "pkgs/tasks/domain/cycle_correlation_test.go",
    "pkgs/tasks/domain/models.go"
)
Commit-Tier3 "refactor(taskcycles): repoint contract, store, handler imports" @(
    "pkgs/taskcycles/contract/",
    "pkgs/taskcycles/handler/",
    "pkgs/taskcycles/store/",
    "pkgs/taskcycles/README.md"
)
git branch -f refactor/taskcycles-domain HEAD

# --- PR #184 taskcycles model A ---
Commit-Tier3 "refactor(taskcycles): add store/model for cycles and phases" @(
    "pkgs/taskcycles/store/model/task_cycle.go",
    "pkgs/taskcycles/store/model/task_cycle_phase.go",
    "pkgs/taskcycles/store/model/task_cycle_stream_event.go",
    "pkgs/taskcycles/store/model/map_cycles.go",
    "pkgs/taskcycles/store/model/doc.go",
    "pkgs/taskcycles/store/model/migrate_models.go",
    "pkgs/taskcycles/store/model/task_row.go"
)
Commit-Tier3 "refactor(taskcycles/store): internal/cycles uses local models" @(
    "pkgs/taskcycles/store/internal/cycles/"
)
Commit-Tier3 "refactor(tasks/store/model): migrate imports cycle core models" @(
    "pkgs/tasks/store/model/migrate_models.go",
    "pkgs/tasks/store/model/parity.go"
)
Commit-Tier3 "refactor(tasks): delete central cycle core model files" @(
    "pkgs/tasks/store/model/task_cycle.go",
    "pkgs/tasks/store/model/task_cycle_phase.go",
    "pkgs/tasks/store/model/task_cycle_stream_event.go",
    "pkgs/tasks/store/model/map_cycles.go"
)
git branch -f refactor/taskcycles-model-cycles HEAD

# --- PR #185 taskcycles model B ---
Commit-Tier3 "refactor(taskcycles): add store/model for commits and reports" @(
    "pkgs/taskcycles/store/model/task_cycle_commit.go",
    "pkgs/taskcycles/store/model/task_cycle_criteria_report.go",
    "pkgs/taskcycles/store/model/task_cycle_verify_report.go",
    "pkgs/taskcycles/store/model/task_cycle_command_run.go",
    "pkgs/taskcycles/store/model/map_commits.go",
    "pkgs/taskcycles/store/model/map_reports.go",
    "pkgs/taskcycles/store/model/parity.go",
    "pkgs/taskcycles/store/model/map_json.go"
)
Commit-Tier3 "refactor(taskcycles/store): internal/commits and internal/reports use local models" @(
    "pkgs/taskcycles/store/internal/commits/",
    "pkgs/taskcycles/store/internal/reports/"
)
Commit-Tier3 "refactor(tasks): remove remaining cycle model files" @(
    "pkgs/tasks/store/model/task_cycle_commit.go",
    "pkgs/tasks/store/model/task_cycle_criteria_report.go",
    "pkgs/tasks/store/model/task_cycle_verify_report.go",
    "pkgs/tasks/store/model/task_cycle_command_run.go",
    "pkgs/tasks/store/model/map_commits.go",
    "pkgs/tasks/store/model/map_reports.go"
)
Commit-Tier3 "chore(ci): taskcycles domain boundary + README" @(
    "scripts/check-go.sh",
    "docs/agent-map.md"
)
git branch -f refactor/taskcycles-model-artifacts HEAD

# --- PR #186 taskcore domain ---
Commit-Tier3 "docs(adr): ADR-0059 taskcore extraction" @(
    "docs/adr/ADR-0059-taskcore-extraction.md",
    "pkgs/taskcore/domain/doc.go",
    "pkgs/taskcore/README.md"
)
Commit-Tier3 "refactor(taskcore): add domain package" @(
    "pkgs/taskcore/domain/"
)
Commit-Tier3 "refactor(tasks/domain): alias Task and shared enums from taskcore/domain" @(
    "pkgs/tasks/domain/actor_aliases.go",
    "pkgs/tasks/domain/enums.go",
    "pkgs/tasks/domain/errors.go",
    "pkgs/tasks/domain/models.go",
    "pkgs/tasks/domain/task_gate.go",
    "pkgs/tasks/domain/gate_criterion.go",
    "pkgs/tasks/domain/retry.go",
    "pkgs/tasks/domain/dependency_satisfies.go",
    "pkgs/tasks/domain/task_model_validate.go",
    "pkgs/tasks/domain/sqltypes.go"
)
Commit-Tier3 "refactor(pkgs): repoint BC domains to taskcore/domain for Actor" @(
    "pkgs/taskchecklist/domain/",
    "pkgs/taskevents/domain/",
    "pkgs/taskcycles/domain/",
    "pkgs/storekernel/",
    "pkgs/taskcompose/store/internal/"
)
git branch -f refactor/taskcore-domain HEAD

# --- PR #187 taskcore store ---
Commit-Tier3 "refactor(taskcore): add store skeleton and model package" @(
    "pkgs/taskcore/store/model/",
    "pkgs/taskcore/store/store.go",
    "pkgs/taskcore/store/pickup_wake.go"
)
Commit-Tier3 "refactor(taskcore): move internal/tasks CRUD" @(
    "pkgs/taskcore/store/internal/tasks/"
)
Commit-Tier3 "refactor(taskcore): move ready, stats, devmirror, health internals" @(
    "pkgs/taskcore/store/internal/ready/",
    "pkgs/taskcore/store/internal/stats/",
    "pkgs/taskcore/store/internal/devmirror/",
    "pkgs/taskcore/store/internal/health/"
)
Commit-Tier3 "refactor(tasks/store): facade_tasks delegates to taskcore/store" @(
    "pkgs/tasks/store/facade_tasks.go",
    "pkgs/tasks/store/facade_ready.go",
    "pkgs/tasks/store/facade_stats.go",
    "pkgs/tasks/store/facade_devmirror.go",
    "pkgs/tasks/store/facade_health.go",
    "pkgs/tasks/store/store.go",
    "pkgs/tasks/store/README.md"
)
Commit-Tier3 "test(taskcore): store integration tests" @(
    "pkgs/tasks/store/facade_tasks_test.go",
    "pkgs/tasks/store/scheduling_parity_test.go",
    "pkgs/tasks/postgres/migrate_repo_default_projects.go",
    "pkgs/tasks/postgres/migrate_repo_default_projects_test.go",
    "pkgs/tasks/postgres/migrate_fixed_worktree_branch_test.go"
)
git branch -f refactor/taskcore-store HEAD

# --- PR #188 taskcore handler ---
Commit-Tier3 "refactor(taskcore): add contract package" @(
    "pkgs/taskcore/contract/",
    "pkgs/tasks/contract/task_crud.go",
    "pkgs/tasks/contract/stats_types.go",
    "pkgs/tasks/contract/health.go",
    "pkgs/tasks/contract/task_types.go"
)
Commit-Tier3 "refactor(taskcore): add handler package with Register" @(
    "pkgs/taskcore/handler/"
)
Commit-Tier3 "refactor(tasks/handler): call taskcorehandler.Register from handler_routes" @(
    "pkgs/tasks/handler/handler_routes.go",
    "pkgs/tasks/handler/handler.go",
    "pkgs/tasks/handler/handler_taskcore_wire.go",
    "pkgs/tasks/handler/handler_taskcore_compat.go",
    "pkgs/tasks/handler/handler_taskcore_compose.go",
    "pkgs/tasks/handler/handler_task_git_binding.go",
    "pkgs/tasks/handler/README.md",
    "pkgs/tasks/handler/sse_notify.go",
    "pkgs/tasks/handler/events_test_helpers_test.go",
    "pkgs/tasks/handler/handler_helpers_test.go",
    "pkgs/tasks/handler/handler_http_cycles_test.go",
    "pkgs/tasks/handler/handler_http_cycles_wire_test.go",
    "pkgs/tasks/handler/handler_http_runners_contract_test.go"
)
Commit-Tier3 "test(tasks/handler): contract tests still on full mux" @(
    "pkgs/tasks/handler/"
)
git branch -f refactor/taskcore-handler HEAD

# --- PR #189 taskcore wire ---
Commit-Tier3 "refactor(tasks/store): embed taskcore/store in Store struct" @(
    "pkgs/tasks/store/store.go",
    "pkgs/tasks/store/facade_checklist.go",
    "pkgs/tasks/store/facade_cycles.go",
    "pkgs/tasks/store/facade_events.go",
    "pkgs/tasks/store/facade_commits.go",
    "pkgs/tasks/store/facade_reports.go",
    "pkgs/tasks/store/facade_git_test.go",
    "pkgs/tasks/store/facade_checklist_test.go",
    "pkgs/tasks/store/facade_cycles_test.go",
    "pkgs/tasks/store/facade_events_test.go"
)
Commit-Tier3 "refactor(tasks): delete moved store internals" @(
    "pkgs/tasks/store/internal/tasks/",
    "pkgs/tasks/store/internal/ready/",
    "pkgs/tasks/store/internal/stats/",
    "pkgs/tasks/store/internal/devmirror/",
    "pkgs/tasks/store/internal/health/",
    "pkgs/tasks/store/model/task.go",
    "pkgs/tasks/store/model/task_dependency.go",
    "pkgs/tasks/store/model/task_context_snapshot.go",
    "pkgs/tasks/store/model/map_task.go",
    "pkgs/tasks/store/model/map_task_dependency.go",
    "pkgs/tasks/store/model/map_task_context_snapshot.go",
    "pkgs/tasks/store/model/map_roundtrip_test.go"
)
Commit-Tier3 "refactor(tasks): delete moved handler files" @(
    "pkgs/tasks/handler/handler_task_crud.go",
    "pkgs/tasks/handler/handler_task_dependencies.go",
    "pkgs/tasks/handler/handler_depends_on_json.go",
    "pkgs/tasks/handler/handler_task_gate.go",
    "pkgs/tasks/handler/handler_tasks_retry.go",
    "pkgs/tasks/handler/handler_task_create_compose.go",
    "pkgs/tasks/handler/handler_create_checklist.go",
    "pkgs/tasks/handler/handler_task_json.go",
    "pkgs/tasks/handler/handler_task_pickup.go",
    "pkgs/tasks/handler/handler_task_runner.go",
    "pkgs/tasks/handler/handler_compose_wire.go",
    "pkgs/tasks/handler/patch_fields.go",
    "pkgs/tasks/handler/domain_metrics.go"
)
Commit-Tier3 "chore(ci): taskcore boundary + docs + schema revision" @(
    "scripts/check-go.sh",
    "scripts/check-schema-revision.sh",
    "pkgs/tasks/postgres/schema_revision.go",
    "pkgs/tasks/contract/",
    "pkgs/tasks/domain/",
    "pkgs/tasks/store/model/",
    "pkgs/projects/handler/",
    "pkgs/runners/handler/",
    "docs/agent-map.md"
)
git branch -f refactor/taskcore-wire HEAD

# Remaining snapshot files (web lockfile, etc.)
Take-FromSnapshot @("web/package-lock.json")
$left = git status --porcelain
if ($left) {
    git add -A
    git commit -m "chore: remaining tier3 collateral (lockfile, gofmt)"
}

Write-Host "`nDone. Branch tips:"
git log --oneline refactor/projects-contract..refactor/taskcore-wire 2>$null
git branch -v | Select-String "refactor/"
