# API

Minimal reference for the `taskapi` HTTP surface (REST) and `GET /events` (SSE). Endpoint behavior is documented in source: error strings, status codes, validation rules, and rate-limit specifics live in `pkgs/tasks/handler/` (godoc) and `pkgs/tasks/middleware/` (godoc).

Data model semantics: [data-model.md](./data-model.md). Configuration: [configuration.md](./configuration.md).

## Conventions

- Mux is mounted at `/` (no `/api` prefix).
- All routes return `application/json`. Error bodies are `{"error":"<message>"}`; some responses include `request_id` for correlation with `X-Request-ID` / `http.access` logs.
- Cacheable read routes (`GET /tasks`, `GET /tasks/{id}`, `GET /tasks/stats`, `GET /tasks/{id}/checklist`, `GET /tasks/{id}/dependencies`, `GET /tasks/{id}/cycles`, `GET /tasks/{id}/cycles/{cycleId}`, `GET /projects`, `GET /projects/{id}`, `GET /projects/{id}/context`, `GET /settings`) emit a strong `ETag` header and `Cache-Control: private, no-cache, must-revalidate`; the server returns `304 Not Modified` with no body when `If-None-Match` matches the current ETag. All other endpoints (mutations, SSE, `/metrics`, `/health*`, `/system/health`, `/repo/*`, `/tasks/cycle-failures`, `/tasks/activity`, drafts, runners) return `Cache-Control: no-store` and do not participate in revalidation.
- `X-Actor` header: `user` (default) or `agent`. The handler ignores any body `triggered_by` and uses this header.
- `Idempotency-Key` (≤ 128 bytes) caches successful (2xx) `POST`/`PATCH`/`DELETE` responses for `HAMIX_IDEMPOTENCY_TTL` (default 24h, in-process only). Replays are byte-identical.
- Rate limit: `HAMIX_RATE_LIMIT_PER_MIN` per `RemoteAddr` (default 120; `0` disables). `429` returns `Retry-After: 60`.
- Request body cap: `HAMIX_MAX_REQUEST_BODY_BYTES` (default 1 MiB; `0` disables).
- `HAMIX_API_TOKEN`, when set, requires `Authorization: Bearer <token>` on all routes except `/health*` and `/metrics`.

## Health and metrics

| Method | Path | Notes |
|---|---|---|
| GET | `/health` | Liveness; returns `version` from `runtime/debug.ReadBuildInfo`. No DB probe. |
| GET | `/health/live` | Same shape as `/health`. |
| GET | `/health/ready` | Readiness; DB ping + `SELECT 1`; `checks.schema` compares code vs DB `SchemaRevision` (`503` when `pending` or `downgrade`, with `schema.code_revision`, `schema.db_revision`, `schema.remediation`); workspace/repo checks as before. |
| GET | `/metrics` | Prometheus text. Standard Go / process collectors + `taskapi_build_info` + `taskapi_db_pool_*` + `taskapi_http_*` + `hamix_agent_runs_*` + `taskapi_sse_*` + `taskapi_agent_queue_*`. |
| GET | `/system/health` | Aggregated JSON for the SPA observability page: build, DB pool gauges, HTTP totals, SSE totals, agent queue + runs + paused. |
| POST | `/v1/rum` | Browser RUM ingest; one batched line per call, capped fields. |
| GET | `/v1/bootstrap` | Cold-start aggregate. Returns `{ settings, tasks: {tasks, limit, offset, has_more}, stats, projects: {projects, limit}, drafts: {drafts} }` in a single round trip; each field mirrors the corresponding per-endpoint wire shape. Default limits match [`policy`](../../pkgs/tasks/handler/policy/read_limits.go) (`BootstrapListLimit` 20, `BootstrapProjectsLimit` 100, `BootstrapDraftsLimit` 50). Honors `ETag` / `If-None-Match` (`304` on match). 5xx on any sub-call failure; clients must tolerate absence and fall back to per-endpoint fan-out. |

## Projects

| Method | Path | Notes |
|---|---|---|
| POST | `/projects` | Create. Body `{ id?, name, description?, repository_id }`. Publishes `project_created`. |
| GET | `/projects` | List. `?limit` (0–100, default 50), `?include_archived=true`. |
| GET | `/projects/{id}` | Single project. |
| PATCH | `/projects/{id}` | Partial. At least one of `name`, `description`, `status`. Default project (`00000000-0000-4000-8000-000000000001`) cannot be renamed / archived (409). Publishes `project_updated`. |
| DELETE | `/projects/{id}` | `204`. Blocked while tasks reference it (409). Default project cannot be deleted. Publishes `project_deleted`. |

Project memory (`/projects/{id}/context*`, `context_summary`, task `project_context_item_ids`) was removed — see [ADR-0087](./adr/ADR-0087-remove-project-context.md).

### Git repositories, worktrees, and branches

Git context follows [ADR-0037](./adr/ADR-0037-global-repos-project-tree.md) (global repositories, optional project overlay) and [ADR-0039](./adr/ADR-0039-fixed-worktree-branch.md) (one fixed `branch_id` per worktree; tasks bind `worktree_id`). Worktree JSON includes `branch_id`. Error responses use `{ "error", "code", "request_id?" }`.

**Global routes (preferred):**

| Method | Path | Notes |
|---|---|---|
| GET | `/git/repositories` | `{ repositories: [...] }`. Each repository includes `main_branch_name` (branch checked out on the main worktree, when bound) and `linked_worktree_count` (non-main worktrees with a bound `branch_id`, matching the Repositories list UI). |
| POST | `/git/repositories` | Register checkout. Body `{ path, host_path? }`. Resolves main worktree path and `git_common_dir`. **201**. Seeds the main worktree row. Does **not** create a project (the system Default is global — see [ADR-0094](./adr/ADR-0094-global-default-project.md)). Does not auto-create additional worktrees/branches. **409** `not_a_git_repository`, `duplicate` (same git object database). |
| DELETE | `/git/repositories/{repoId}` | Deletes the repository inventory row, its worktrees/branches, and **user projects for that repo**. The global system Default is not deleted. **204**. **409** `has_running_task`. |
| GET | `/git/repositories/{repoId}` | Single repository. **404** `repository_not_found`. |
| GET | `/git/repositories/{repoId}/worktrees` | `{ worktrees: [...] }` including optional `stale` (idle terminal tasks older than 24h). |
| GET | `/git/worktrees/{worktreeId}` | Single worktree with resolved `repository_path`, `repository_host_path`, and `branch_name` (for task-detail binding). **404** `worktree_not_found`. No `stale` enrichment. |
| GET | `/git/repositories/{repoId}/worktrees/checkout-status` | Live git checkout state for **branch-bound** registered worktrees only: `{ worktrees: [{ worktree_id, available, reason?, dirty?, detached?, head_commit_at?, has_upstream?, ahead?, behind?, upstream? }] }`. |
| POST | `/git/repositories/{repoId}/sync` | Fetch `origin` and refresh registered metadata without discovering operator worktrees. **202** `{ status, report }` (same shape as reconcile). **400** when fetch fails. |
| POST | `/git/repositories/{repoId}/reconcile` | Repair registered repository/worktree paths against `git worktree list`. Body `{ bootstrap_path?, repair?, dry_run? }` (all optional). Does **not** insert unregistered worktrees. When the stored main path is missing, pass `bootstrap_path` or use **Relocate**. **202** `{ status, report }`. |
| POST | `/git/repositories/{repoId}/relocate` | Operator alias: body `{ path }` runs reconcile with `bootstrap_path=path`, `repair=true`. **202** same shape as reconcile. |
| POST | `/git/worktrees/{worktreeId}/relocate` | Manual path fix for one registered worktree. Body `{ path }`. **200** worktree JSON after probe + UPDATE. **409** `bootstrap_mismatch` when path belongs to a different repo. |
| DELETE | `/git/worktrees/{worktreeId}` | Default: unregister from Hamix (**204**); checkout stays on disk. Query `?remove_from_disk=true` runs `git worktree remove` then deletes the row; optional `&force=true` for dirty trees. **409** `has_running_task`. **400** when target is the main worktree and `remove_from_disk=true`. |
| GET | `/git/repositories/{repoId}/branches` | Registered branches `{ branches: [...] }`. |
| GET | `/git/repositories/{repoId}/projects` | Projects tied to this repo `{ projects, limit }`. |

**Projects:** `POST /projects` accepts optional `repository_id` (repo must exist). Task create requires `repository_id` (+ `project_id`); the server returns the task immediately and **eagerly allocates** a managed worktree asynchronously (sets `worktree_id` when ready). Agents refuse main/`is_main` and the repository default branch, and do not pick up until `worktree_id` is set.

Stable error codes: `not_a_git_repository`, `path_exists`, `branch_exists`, `branch_checked_out`, `branch_bound_to_worktree`, `project_repo_mismatch`, `has_running_task`, `bootstrap_mismatch`, `repository_not_found`, `worktree_not_found`, `branch_not_found`, `duplicate`.

## Tasks

Model semantics (tags, milestone, `depends_on`, gate, worker readiness): [data-model.md](./data-model.md).

| Method | Path | Notes |
|---|---|---|
| POST | `/tasks` | Create. Title required; `priority` required; `checklist_items` required. **`project_id` and `repository_id` required**. Optional **`worktree_id`**: when set to an existing non-main managed worktree in that repository, the task binds immediately and **skips** allocate (enqueue onto a shared worktree). When omitted, returns flat `domain.Task` with `worktree_id` often still null; server then eagerly allocates a managed worktree (`hamix/task-<8 hex>`), fetches `origin`, and patches `worktree_id` (publishes `task_updated`). Allocate failure marks the task `failed`. Agents refuse main/`is_main` and the repo default branch and do not dequeue until `worktree_id` is set. Optional `id`, `draft_id`, `pickup_not_before`, `cursor_model`, `tags`, `milestone`, `depends_on`. Publishes `task_created` on insert. Read responses may include computed `worktree_root_task_id` (task whose id named the managed branch) when `worktree_id` is set. |
| GET | `/tasks` | List all tasks (flat). Pagination: `?limit` (0–200, default **50** when omitted) + `?offset` (≥ 0) **or** `?after_id` (keyset, mutually exclusive with offset). Optional **`?worktree_id=`** (offset pagination only) restricts to tasks bound to that worktree. Envelope `{ tasks, limit, offset, has_more }`. Each element is a flat `domain.Task` (no nested `children`); may include computed `worktree_root_task_id`. Rows are ordered **newest first** by `created_at` (from the `task_created` audit event). **SPA note:** the Hamix web app always sends `limit=20` (`BootstrapListLimit` / `TASK_LIST_PAGE_SIZE`) explicitly; it does not rely on the server default. See [`policy`](../../pkgs/tasks/handler/policy/read_limits.go). |
| GET | `/tasks/stats` | Counters: `total`, `ready`, `critical`, `scheduled`, `by_status`, `by_priority`, `cycles`, `phases`, `runner`, `recent_failures`. |
| GET | `/tasks/cycle-failures` | Paginated terminal cycle failures. `?limit`, `?offset`, `?sort ∈ at_desc | at_asc | reason_asc | reason_desc`. |
| GET | `/tasks/activity` | Paginated cross-task activity feed. Fixed type filter: `status_changed`, `phase_failed`, `approval_granted`. Query params: `limit` (1–200, default 50), `offset` (≥ 0), `since` (optional RFC3339 lower bound on `at`). Ordered newest-first (`at DESC, seq DESC`). Envelope: `{ total, limit, offset, events: [{ task_id, seq, at, type, by, data, task_title?, task_number?, task_priority?, task_project_id?, task_tags? }] }`. Joined task fields (when the owning task still exists) power client-side Timeline filters. `Cache-Control: no-store`. |
| GET | `/tasks/{id}` | Single flat `domain.Task`. |
| PATCH | `/tasks/{id}` | At least one of: `title`, `initial_prompt`, `status`, `priority`, `project_id`, `worktree_id`, `pickup_not_before`, `cursor_model`, `tags`, `milestone`, `gate`, `depends_on`. Publishes `task_updated` (+ `task_gate_changed` / `task_dependency_changed` when those fields change). Writable `status` values for `X-Actor: user`: `ready`, `running`, `blocked`, `review`, `failed`, `on_hold` (`done` via approve only; `pr_ready` via open-pr finalize only; `closed` via close only). See [data-model.md](./data-model.md). |
| POST | `/tasks/{id}/close` | Marks the task `closed` (idempotent). Cancels an in-flight agent run for this task only, drops it from the ready queue, cancels pickup wake. Does **not** delete the row or remove the worktree. Publishes `task_updated`. |
| POST | `/tasks/{id}/reopen` | `closed` → `ready`. `409` if not closed. Publishes `task_updated`. |
| GET | `/tasks/{id}/events` | Audit log. Default: ascending all rows. With `limit` / `before_seq` / `after_seq`: keyset-paged newest-first slice with `range_*`, `has_more_*`, `approval_pending`. Deep dive: [domain/task-events.md](./domain/task-events.md). |
| GET | `/tasks/{id}/events/{seq}` | Single event row. |
| PATCH | `/tasks/{id}/events/{seq}` | Append a user-response message (max 10 000 bytes after trim, thread cap 200). Only for `approval_requested` and `task_failed`. Publishes `task_event_changed`. |
| GET | `/tasks/{id}/dependencies` | `{ depends_on: [{ task_id, satisfies }] }`. |
| POST | `/tasks/{id}/dependencies` | Body `{ depends_on_task_id, satisfies? }` (default `done`). Cycles / self-deps rejected. Publishes `task_dependency_changed`. |
| DELETE | `/tasks/{id}/dependencies/{depId}` | `204`. Publishes `task_dependency_changed`. |
| PATCH | `/tasks/{id}/gate` | Body `{ action: release | hold | clear_hold }`. Publishes `task_gate_changed` and `task_updated`. |
| POST | `/tasks/{id}/retry` | Operator retry after task `failed`. Body `{ mode: fresh|resume, parent_cycle_id? }`. Requires `X-Actor: user`. Resolves `parent_cycle_id` to the latest terminal cycle (`failed` or `aborted`, max `attempt_seq`) when omitted. Sets ephemeral `pending_retry` on the task row and `status=ready`. `409` when already `ready` with a different pending intent; idempotent `200` when the same mode+parent is re-posted. Appends `task_retry_requested` audit event. Publishes `task_updated`. Bare `PATCH failed→ready` without this route leaves `pending_retry` null (legacy run). |
| POST | `/tasks/{id}/open-pr` | Human approve-and-open-PR from agent-complete. Requires `X-Actor: user` and `status=review`. Optional body `{ parent_cycle_id? }`. Queues `pending_retry` kind `open_pr` (mode `resume`), transitions `review→ready`, appends `approval_granted` and `open_pr_requested`. Publishes `task_updated` and notifies the worker. See [ADR-0082](./adr/ADR-0082-human-approval-gate.md). |
| POST | `/tasks/{id}/approve` | Human completion after a PR is open. Requires `X-Actor: user` and `status=pr_ready`. Checklist must be complete. Transitions `pr_ready→done`, appends `status_changed` and `on_task_done` (commits from latest succeeded cycle). Publishes `task_updated` and unblocks dependents. Free-form `PATCH { status: done }` is rejected (`400`). See [ADR-0082](./adr/ADR-0082-human-approval-gate.md). |
| POST | `/tasks/{id}/polish` | Human rework from `review`. Body `{ instructions, parent_cycle_id?, flagged_criterion_ids?, new_criteria? }`. Requires `X-Actor: user`, non-empty `instructions`. Optional `flagged_criterion_ids` reopen those checklist completions; optional `new_criteria` (objects `{ text, verify_commands? }` or legacy text strings via structured items) append definition rows with optional verify commands in the same transaction. Resolves `parent_cycle_id` to the latest **succeeded** cycle when omitted. Sets `pending_retry` with `kind=polish`, `mode=resume`, instructions, flagged/new IDs, and `skip_verify` when neither flags nor new criteria are present; transitions `review→ready`. Idempotent when the same polish intent is already queued; `409` on conflicting pending intent. Appends `task_polish_requested`. Worker resumes the parent execute Cursor session with a **human polish directive** (not failure/process-restart recovery) and always runs execute. When `skip_verify` is true, prior done criteria stay locked, verify is skipped, and the task returns to `review` after execute. Otherwise verify runs only for non-locked criteria (flagged + newly added), then returns to `review`. |

### Checklist

| Method | Path | Notes |
|---|---|---|
| GET | `/tasks/{id}/checklist` | `{ items: [...] }` ordered by `sort_order`. Each item includes optional `verify_commands: [{ sort_order, command, expected_outcome, timeout_seconds? }]`. `timeout_seconds` is omitted when unlimited. |
| POST | `/tasks/{id}/checklist/items` | Body `{ text, verify_commands? }` where each command is `{ command, expected_outcome?, timeout_seconds? }`. `timeout_seconds` omitted or null = no wall-clock timeout; when set must be `> 0`. Rejected `409` when the task is `running` or a cycle is running. Allowed on `done` tasks for post-completion edits. Publishes `task_updated`. |
| PATCH | `/tasks/{id}/checklist/items/{itemId}` | Body: exactly one of `{ text }`, `{ verify_commands }`, or `{ done: true|false }`. Rejected `409` when the task is `running` or a cycle is running. Satisfied criteria remain locked until the task is `done`. `done:true` requires `X-Actor: agent` plus `evidence` + optional `verified_by`. Publishes `task_updated`. |
| DELETE | `/tasks/{id}/checklist/items/{itemId}` | `204`. Rejected `409` when the task is `running` or a cycle is running. Publishes `task_updated`. |

### Task drafts

| Method | Path | Notes |
|---|---|---|
| POST | `/task-drafts` | Upsert. Body `{ id?, name, payload }`. Never publishes on SSE. |
| GET | `/task-drafts` | List summaries (without `payload`). `?limit` (0–100). |
| GET | `/task-drafts/{id}` | Full draft with `payload` defaulted to `{}`. |
| DELETE | `/task-drafts/{id}` | `204`. |

### Draft assist (compose AI)

In-memory prompt-assist sessions for the compose page (ADR-0101). Not part of the task SSE hub. Heartbeats are SSE comments every 3s while a run is active.

| Method | Path | Notes |
|---|---|---|
| GET | `/draft-assist/ready` | `{ ready, runner: "sdk"\|"fake"\|"missing", reason? }`. Reasons when not ready: `no_runner`, `missing_key`, `sidecar_down`. |
| POST | `/draft-assist/sessions` | Create session. Body `{ worktree_id?, snapshot }`. `201` `{ id, nonce, … }` |
| GET | `/draft-assist/sessions/{id}` | Session + snapshot |
| PUT | `/draft-assist/sessions/{id}/snapshot` | Replace form snapshot |
| GET | `/draft-assist/sessions/{id}/events` | SSE: `session` (includes `schema_version: 1`), `status`, `token`, `tool`, `patch`, `error`, `done`. Heartbeats are `: heartbeat` comments every 3s while a run is active. Ring size 256; `Last-Event-ID` replay. |
| POST | `/draft-assist/sessions/{id}/runs` | **202** `{ run_id }` immediately. Body `{ user_message, snapshot? }`. Concurrent run → **409**. |
| POST | `/draft-assist/sessions/{id}/runs/{runId}/cancel` | **202** `{ status: "cancelling" }`; SSE emits `status=cancelling` then `done{cancelled}` |
| DELETE | `/draft-assist/sessions/{id}` | `204` |

See [domain/draft-assist.md](./domain/draft-assist.md).

### Task templates

Named, durable task compose blueprints. Payload shape matches task create fields (title, prompt, status, priority, checklist, runner, project, schedule, tags, milestone, `depends_on`) plus optional template-only `function_inputs` (see [ADR-0088](./adr/ADR-0088-template-functions.md)). Never publishes on SSE for CRUD; instantiate accepts with **202** and publishes `task_created` per created task as each commit finishes (same enrichment as `POST /tasks`).

| Method | Path | Notes |
|---|---|---|
| POST | `/task-templates` | Upsert. Body `{ id?, name?, payload }`. `name` defaults to trimmed `payload.title`. Validates like compose drafts (title, priority, checklist, runner/model, git binding). Optional `payload.function_inputs`: `[{ id, kind: dir\|file\|function, label, required?, multiple? }]` (`id` matches `[a-z][a-z0-9_]*`, unique). **201** summary `{ id, name, created_at, updated_at, primary_tag?, project_id?, repository_id?, instantiate_count, is_function?, input_kinds? }`. `primary_tag` is the first entry in `payload.tags` when present. `project_id` / `repository_id` are peeked from the payload when set. `is_function` / `input_kinds` are derived when `function_inputs` is non-empty. `instantiate_count` defaults **0**. |
| GET | `/task-templates` | List summaries (without `payload`). `?limit` (0–100, default 50). `?q=` ILIKE search on `name`. `?sort=` one of `updated_at` (default), `name`, `instantiate_count`. `?order=` `asc` or `desc` (default `desc`). Invalid `sort` or `order` → **400**. `?tag=` case-insensitive match on the first `payload.tags` entry. Summary fields match POST **201**. |
| GET | `/task-templates/{id}` | Full template with `payload` (may include `function_inputs`). |
| PATCH | `/task-templates/{id}` | Partial `{ name?, payload? }`. **200** full detail. |
| DELETE | `/task-templates/{id}` | `204`. |
| POST | `/task-templates/instantiate` | Body `{ template_ids: string[], count?: number }` **or** `{ items: { template_id, count?, function_bindings? }[] }`. When `items` is non-empty it takes precedence. `function_bindings`: `[{ input_id, paths?, functions?: [{ path, name, line }] }]`. Required for templates with `function_inputs`; omitted/empty for ordinary templates. Sync gate loads templates and applies bindings; missing templates / bind failures appear in `errors` without blocking other items. Omitted `count` defaults to **1**. Per-item/top-level `count` **1..25**; total creates ≤ **100**. Duplicate `template_id` in `items` → **400**. **Accepts asynchronously:** **202** `{ accepted, total, errors: [{ template_id, error }] }` when at least one item is queued (`accepted: true`, `total` = queued creates). Creates run on an owned worker; each successful create publishes enriched **`task_created` SSE immediately** (same as `POST /tasks`) so the task list can insert rows progressively — the HTTP response does **not** wait for creates and does not return `tasks[]`. Queue full → **503**. Strips `depends_on`; omits past `pickup_not_before`. |

### Execution cycles

See [data-model.md](./data-model.md) for state machine and substrate semantics.

| Method | Path | Notes |
|---|---|---|
| POST | `/tasks/{id}/cycles` | Start a cycle. Body `{ parent_cycle_id?, meta? }`. Returns `taskCycleResponse` (with typed `cycle_meta` projection). Optional `token_usage` when execute/verify phases have parseable Cursor `usage` in `details_json` (see [ADR-0083](./adr/ADR-0083-task-token-usage.md)). Publishes `task_cycle_changed`. |
| GET | `/tasks/{id}/cycles` | List. `?limit` (1–200), `?before_attempt_seq` keyset cursor. Newest first. Each cycle may include optional `token_usage` when that attempt has parseable phase usage. |
| GET | `/tasks/{id}/cycles/{cycleId}` | One cycle with `phases[]` ordered ascending. Optional `token_usage` aggregates execute/verify usage for that cycle when present. |
| GET | `/tasks/{id}/token-usage` | Task-wide token accounting. Returns `{ task_id, token_usage, attempts: [{ cycle_id, attempt_seq, token_usage, share_of_task_pct }] }`. `token_usage` and each attempt's `token_usage` use the same projection as cycle responses (`consumed_tokens`, `execute_consumed_tokens`, `verify_consumed_tokens`, component fields, `known`). `attempts` includes only cycles with at least one parseable usage row, ordered by `attempt_seq ASC`. `share_of_task_pct` is a percentage of task `consumed_tokens` when both task and attempt usage are known and task consumed &gt; 0; otherwise `null`. **404** when the task does not exist. Semantics: [ADR-0083](./adr/ADR-0083-task-token-usage.md). |
| PATCH | `/tasks/{id}/cycles/{cycleId}` | Terminate. Body `{ status: succeeded|failed|aborted, reason? }`. Publishes `task_cycle_changed`. The agent worker emits `verification_failed:<id>,<id>,…` on terminal verify failure (sorted, deduped failing criterion IDs); the `verification_failed` prefix is contract-stable — clients MUST use prefix matching. Bare `verification_failed` (older cycles) remains a valid value. The reason column is 256 chars; long lists are truncated with `…` while the prefix stays intact. |
| GET | `/tasks/{id}/cycles/{cycleId}/stream` | Normalized Cursor live-run history. `?limit` (1–500), `?after_seq` keyset. |
| GET | `/tasks/{id}/commits` | Task-wide git commits (deduped by SHA, earliest `committed_at` wins for the kept row). Returns `{ task_id, commits: [{ cycle_id, attempt_seq, seq, repo, worktree, branch, sha, committed_at, message }] }` ordered newest-first by `committed_at`. Refetch on `task_cycle_changed` after execute ingest. |
| GET | `/tasks/{id}/cycles/{cycleId}/verdicts` | Per-criterion verdict evidence for one cycle. Returns `{ task_id, cycle_id, git_context?, commits: [...], criteria_reports: [...], verify_reports: [...], command_runs: [...] }`. `git_context` (`{ repo, worktree, branch }`) is omitted when no commits were indexed; when present, `repo`/`worktree` come from the first commit and `branch` from the last commit with a non-empty branch (fallback: first). `commits[]` is always non-null (empty when no rows); entries are `{ seq, repo, worktree, branch, sha, committed_at, message }` ordered `seq ASC`. Criteria, verify, and command arrays are non-null (empty when no rows mirrored); those rows are ordered `(attempt_seq ASC, criterion_id ASC)` (command runs also by `command_seq ASC`). Each criteria row carries `claimed_done` + `evidence` from the execute agent's self-report; each verify row carries `verified` + `verifier_kind` + `reasoning`; each command run carries `exit_code` + `meta_path` for worker-executed verify shell checks. `verifier_kind` is the same enum as `task_checklist_completions.verified_by`. Pre-ADR-0014 cycles return empty `commits` and omit `git_context`. |
| POST | `/tasks/{id}/cycles/{cycleId}/phases` | Start a phase. Body `{ phase: execute|verify }`. Transitions follow `domain.ValidPhaseTransition`. Publishes `task_cycle_changed`. |
| PATCH | `/tasks/{id}/cycles/{cycleId}/phases/{phaseSeq}` | Terminate a phase. Body `{ status: succeeded|failed|skipped, summary?, details? }`. Publishes `task_cycle_changed`. |

## Runners

Runner adapters register at compile time via `pkgs/agents/runner/registry`. The SPA Settings page discovers available runners through these routes. Full plug-in model: [domain/runner-adapters.md](./domain/runner-adapters.md).

| Method | Path | Notes |
|---|---|---|
| GET | `/runners` | Array of `{ id, label, default_binary_hint, config_schema? }`. `config_schema` present when the adapter implements `ConfigSchemaProvider`. |
| GET | `/runners/{id}/config-schema` | Returns the adapter config schema. **404** unknown runner. **501** runner does not expose a schema. |
| POST | `/runners/{id}/validate-config` | Body: opaque JSON config blob. **200** `{ valid: true }` or **422** `{ valid: false, error }`. **404** unknown runner. **501** no validator. |
| POST | `/runners/{id}/probe` | Body `{ binary_path? }`. When `binary_path` is omitted, falls back to `app_settings.cursor_bin` for `cursor` only. Probe/CLI failures return **200** `{ ok: false, error, runner, binary_path? }` so the SPA renders inline. **404** unknown runner. **501** adapter does not implement `Prober`. Success: **200** `{ ok: true, version, binary_path, runner }`. |
| POST | `/runners/{id}/list-models` | Same body and soft-failure semantics as probe. Success: **200** `{ ok: true, models: [{ id, label }], binary_path, runner }`. **501** when `ModelLister` is not implemented. Timeout: 30s. |

Legacy cursor-named routes under `/settings` (`probe-cursor`, `list-cursor-models`) remain; prefer `/runners/*` for new UI work.

## App settings

Singleton row (`id=1`) seeded on first read with `domain.DefaultAppSettings`. Full field reference: [configuration.md](./configuration.md).

| Method | Path | Notes |
|---|---|---|
| GET | `/settings` | Returns the full `AppSettings` row. Always available. |
| GET | `/settings/workspace-roots` | `{ roots: [{ id, path, label, category?, available, unavailable_reason? }], environment: "native" }`. Browse roots for the workspace folder picker. Optional query `scope=expanded` merges OS bootstrap places (Home, Documents, …) with registered repository paths — used by **Create worktree** parent-folder browse. `category` is one of `registered`, `install`, `home`, `documents`, `desktop`, `downloads`, `pictures`, `music`, `videos`, or `custom`. Does not require `repo_root`. `Cache-Control: no-store`. |
| GET | `/settings/browse-dirs?path=` | `{ path?, parent_path?, is_git_repo?, entries: [{ name, path, has_children, is_git_repo }] }`. Lists immediate subdirectories under allowed browse roots. When `path` is set, `is_git_repo` reflects whether that directory is a git checkout. Empty `path` lists available roots. Does not require `repo_root`. **400** when path escapes roots. `Cache-Control: no-store`. |
| GET | `/settings/git-probe?path=` | `{ path, main_path?, is_main?, is_git_repository, current_branch?, branches: [{ name, head_sha }] }`. Opens the path with git and lists local branches without registering a repository. When the path is a checkout, `path` is the opened toplevel, `main_path` is the canonical main repository root (linked worktrees resolve here), and `is_main` is true when they match. `is_git_repository: false` and empty `branches` when the path is not a checkout. Does not require `repo_root`. **400** when `path` is missing. `Cache-Control: no-store`. |
| PATCH | `/settings` | Partial; pointer fields distinguish "not provided" from explicit zero. On success, supervisor reloads in-process and SSE publishes `settings_changed`. |
| POST | `/settings/probe-cursor` | Body `{ runner?, binary_path? }`. Probe failures return `200 { ok: false, error }` so the SPA renders inline. |
| POST | `/settings/list-cursor-models` | Same fallback semantics as probe. CLI failures return `200 { ok: false }`. |
| POST | `/settings/cancel-current-run` | Cancels any in-flight `runner.Run`. Cycle terminated with `cancelled_by_operator`. Publishes `agent_run_cancelled` when something was running. |

## Workspace repo

Deep dive: [domain/workspace-repo.md](./domain/workspace-repo.md). Wired only when `app_settings.repo_root` is set. When unset, every `/repo/*` route returns `409 { error: "repo root is not configured", reason: "repo_root_not_configured" }`. When `OpenRoot` rejects the path (missing, not a directory, symlink loop), routes return `500 { reason: "repo_root_open_failed", error }`.

| Method | Path | Notes |
|---|---|---|
| GET | `/repo/search?q=&worktree_id=` | `{ paths: string[], entries?: [{ path, kind }] }`. `paths` is always the file hits (compat for `@`-mentions). Optional `kinds=file,dir` (comma-separated); when set, `entries` includes matching files and/or directories. Default kinds = files only. Caps: empty `q` ≤ 250 browse hits; filtered ≤ 100. `q` ≤ 512 bytes. **400** on invalid `kinds`. Requires `worktree_id`. |
| GET | `/repo/files?worktree_id=&limit=&after=&q=` | Cursor page of referenceable files for `@` index warm. Response: `{ paths, next_after?, has_more, source, truncated? }`. Paths are sorted, slash-separated. Default `limit` is 500 (max 2000). `after` is the last path from the previous page. Optional `q` filters with a case-insensitive substring (warm-incomplete fallback; the SPA ranks locally once the client index is ready). `source: "git"` means `git ls-files --cached --others --exclude-standard` (gitignore-aware; tracked-but-ignored files stay listed). `source: "walk"` is the non-git fallback with a fixed skip list. The server keeps a short TTL in-process cache of the full sorted list so page requests do not re-run git on every warm batch. Requires `worktree_id`. |
| GET | `/repo/symbols?q=&worktree_id=` | `{ symbols: [{ path, name, line, kind }] }`. Best-effort declaration search (Go/TS/JS/Python/Rust regex). Empty `q` → empty `symbols`. Cap 50. `q` ≤ 512 bytes. Requires `worktree_id`. |
| GET | `/repo/file?path=` | `{ path, content, binary, truncated, size_bytes, line_count, warning? }`. Binary or invalid UTF-8 returns `binary: true` with empty `content`. Files over 32 MiB are truncated. |
| GET | `/repo/validate-range?path=&start=&end=` | `{ ok, line_count?, warning? }`. Used by the SPA to warn about invalid `@`-mentions inline. |
| GET | `/repo/diff?worktree_id=&sha=` | `{ sha, patch, truncated, size_bytes, author?, author_email?, parent_sha?, files_changed?, insertions?, deletions? }`. Unified diff for one commit via `git show` in the worktree opened by `worktree_id` (required); `sha` is 7–40 hex chars (≤ 64 bytes query). Patch capped at 512 KiB (`truncated: true` when clipped). Author and shortstat come from `git show --format` / `--shortstat`. **400** when `worktree_id` is missing; **404** when the worktree or SHA is absent. |

`POST /tasks` and `PATCH /tasks/{id}` validate `@`-mentions in `initial_prompt` against the configured repo. Failures return `400` with the offending mention in the error message (`@<path>` or `@<path> (<start>-<end>)`). Validation is skipped when `repo_root` is unset, `initial_prompt` is omitted, or `initial_prompt = ""`.

## SSE — `GET /events`

Deep dive: [domain/sse-hub.md](./domain/sse-hub.md). `text/event-stream`. First frame: `retry: 3000`. Frames are id + JSON:

```text
id: 42
data: {"type":"task_updated","id":"<task-uuid>"}

```

Lossless reconnects via `Last-Event-ID`: a ring buffer (default 1024 entries) replays unseen frames on reconnect. Out-of-window reconnects emit one `resync` frame and the client drops caches. Slow consumers (full per-connection buffer) are evicted with a `resync` frame. Heartbeat `: heartbeat` comment every 15s. Identical `{type,id}` frames within 50ms are coalesced (except `task_cycle_changed` and `agent_run_progress`).

### Event types

| Type | When | Payload |
|---|---|---|
| `task_created` | `POST /tasks` succeeds. | `{ type, id, data: <task> }` |
| `task_updated` | Task-row mutations (PATCH, checklist, gate, retry; agent status including pickup `ready→running` and terminals). `data` carries the full flat task when the publisher enriches post-commit; hint-only frames omit `data`. | `{ type, id, data?: <task> }` |
| `task_event_changed` | Audit event thread append (`PATCH /tasks/{id}/events/{seq}`). Does not mutate the `tasks` row. | `{ type, id, event_seq }` |
| `task_dependency_changed` | Dependency add/remove/replace. | `{ type, id }` |
| `task_gate_changed` | Gate create/patch/action. | `{ type, id }` |
| `task_cycle_changed` | Cycle/phase mutation. | `{ type, id, cycle_id, data?: <cycle detail> }` |
| `agent_run_progress` | Live Cursor activity hint while a phase runs. Not persisted in `task_events`; durable history via `GET /tasks/{id}/cycles/{cycleId}/stream`. Throttled to one frame per 750ms per running phase. | `{ type, id, cycle_id, phase_seq, progress: { kind, subtype, message, tool } }` |
| `project_created` / `project_updated` / `project_deleted` | Project CRUD. | `{ type, id }` |
| `settings_changed` | `PATCH /settings` after supervisor reload. | `{ type }` (no id) |
| `agent_run_cancelled` | `POST /settings/cancel-current-run` actually cancelled something. | `{ type }` (no id) |
| `resync` | Hub-emitted. Out-of-window reconnect or slow-consumer eviction. No `id:` line on wire (preserves `Last-Event-ID` cursor). | `{ type }` |

Read-only GETs never publish. Failed writes never publish. Drafts (`/task-drafts/*`), task templates CRUD (`/task-templates` except instantiate), and `POST /settings/probe-cursor` are not part of the SSE surface.

### Dev synthetic SSE (`HAMIX_SSE_TEST=1`)

For local UI work, `taskapi` can start a background ticker (no extra routes). Set `HAMIX_SSE_TEST=1`; interval via `HAMIX_SSE_TEST_INTERVAL` (default `3s`; `0` disables). Tunables: `HAMIX_SSE_TEST_EVENTS_PER_TICK`, `HAMIX_SSE_TEST_SYNC_ROW`, `HAMIX_SSE_TEST_USER_RESPONSE`, `HAMIX_SSE_TEST_LIFECYCLE`, `HAMIX_SSE_TEST_LIFECYCLE_EVERY`. Never enable in production without intent. Source: `pkgs/tasks/devsim/`.
