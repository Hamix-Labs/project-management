export const DEMO_PRIMARY_PROJECT_ID = "00000000-0000-4000-8000-000000000001";
export const DEMO_SECOND_PROJECT_ID = "22222222-2222-4222-8222-222222222222";
export const DEMO_THIRD_PROJECT_ID = "33333333-3333-4333-8333-333333333333";

const ISO = "2026-03-10T12:00:00Z";

/** Fixed anchors so sample tasks span relative-time buckets (newest → oldest). */
const CREATED = {
  justNow: "2026-06-20T17:59:00Z",
  min12: "2026-06-20T17:48:00Z",
  min45: "2026-06-20T17:15:00Z",
  h3: "2026-06-20T15:00:00Z",
  h9: "2026-06-20T09:00:00Z",
  d1: "2026-06-19T10:00:00Z",
  d3: "2026-06-17T14:00:00Z",
  d5: "2026-06-15T11:00:00Z",
  w1: "2026-06-13T16:00:00Z",
  w2: "2026-06-06T12:00:00Z",
  w3: "2026-05-30T08:00:00Z",
  mo1: "2026-05-20T12:00:00Z",
  mo2: "2026-04-18T12:00:00Z",
  mo3: "2026-03-10T12:00:00Z",
  mo5: "2026-01-15T12:00:00Z",
} as const;

const DEMO_PROJECT_IDS = new Set([
  DEMO_PRIMARY_PROJECT_ID,
  DEMO_SECOND_PROJECT_ID,
  DEMO_THIRD_PROJECT_ID,
]);

export function isDemoProjectId(id: string): boolean {
  return DEMO_PROJECT_IDS.has(id);
}

function task(
  id: string,
  title: string,
  status: string,
  priority: string,
  initialPrompt: string,
  createdAt: string,
  opt: Record<string, unknown> = {},
): Record<string, unknown> {
  return {
    id,
    title,
    initial_prompt: initialPrompt,
    status,
    priority,
    runner: "cursor",
    cursor_model: "",
    checklist_inherit: true,
    tags: [],
    depends_on: [],
    created_at: createdAt,
    ...opt,
  };
}

const M_JWT = "JWT rollout";
const M_SESS = "Session hardening";
const M_DISC = "Discovery";
const M_TEST = "Load testing";
const M_REL = "Release";

const C1 = "c1111111-1111-4111-8111-111111111111";
const C2 = "c2222222-2222-4222-8222-222222222222";
const C3 = "c3333333-3333-4333-8333-333333333333";
const C4 = "c4444444-4444-4444-8444-444444444444";
const C5 = "c5555555-5555-4555-8555-555555555555";
const C6 = "c6666666-6666-4666-8666-666666666666";

const ALL_TASK_IDS: string[] = [];

function reg(id: string) {
  ALL_TASK_IDS.push(id);
  return id;
}

const ROOT_TASKS: Record<string, unknown>[] = [
  task(
    reg("f0000001-0000-4000-8000-000000000001"),
    "Auth refactor rollout",
    "running",
    "high",
    "Replace session cookies with short-lived JWT access tokens and refresh rotation across internal services. Acceptance: zero-downtime deploy with a 72-hour backward-compatibility window.",
    CREATED.h3,
    {
      project_id: DEMO_PRIMARY_PROJECT_ID,
      milestone: M_JWT,
      tags: ["auth"],
    },
  ),
  task(
    reg("f0000002-0000-4000-8000-000000000002"),
    "Session invalidation sweep",
    "ready",
    "medium",
    "Audit and revoke stale sessions older than 90 days. Cross-reference recent password resets and security alerts before bulk invalidation.",
    CREATED.d3,
    {
      project_id: DEMO_PRIMARY_PROJECT_ID,
      milestone: M_SESS,
      tags: ["auth"],
    },
  ),
  task(
    reg("f0000003-0000-4000-8000-000000000003"),
    "OAuth consent copy review",
    "blocked",
    "low",
    "Legal must approve updated OAuth consent screen copy before the partner portal ships. Blocked on compliance review ticket #4421.",
    CREATED.w2,
    {
      project_id: DEMO_PRIMARY_PROJECT_ID,
      milestone: M_DISC,
    },
  ),
  task(
    reg("f0000004-0000-4000-8000-000000000004"),
    "Load-test harness for login",
    "review",
    "critical",
    "Build k6 scripts simulating 500 concurrent logins against staging. Agent finished the scripts; operator review needed for thresholds and alerting config.",
    CREATED.h9,
    {
      project_id: DEMO_PRIMARY_PROJECT_ID,
      milestone: M_TEST,
    },
  ),
  task(
    reg("f0000005-0000-4000-8000-000000000005"),
    "Release checklist: AuthV2",
    "done",
    "medium",
    "Final pre-launch checklist: feature flags, rollback plan, monitoring dashboards, and on-call runbook. All items verified and signed off.",
    CREATED.mo3,
    {
      project_id: DEMO_PRIMARY_PROJECT_ID,
      milestone: M_REL,
    },
  ),
  task(
    reg("f0000006-0000-4000-8000-000000000006"),
    "Backfill audit logs",
    "ready",
    "medium",
    "Backfill missing audit events from the Jan–Feb migration window into the central log store. Estimated volume: 2.4M records.",
    CREATED.w1,
    {
      project_id: DEMO_PRIMARY_PROJECT_ID,
    },
  ),
  task(
    reg("f0000007-0000-4000-8000-000000000007"),
    "Customer migration dry run",
    "failed",
    "high",
    "Dry-run migration of 50 pilot customers to AuthV2. Failed at step 3 when session mapping hit an edge case with federated accounts.",
    CREATED.d1,
    {
      project_id: DEMO_PRIMARY_PROJECT_ID,
      milestone: M_TEST,
    },
  ),
  task(
    reg("f0000008-0000-4000-8000-000000000008"),
    "Billing webhook resilience",
    "running",
    "critical",
    "Add exponential backoff, a dead-letter queue, and idempotency keys to Stripe webhook processing. Must dedupe events within a 24-hour window.",
    CREATED.min12,
    {
      project_id: DEMO_SECOND_PROJECT_ID,
    },
  ),
  task(
    reg("f0000009-0000-4000-8000-000000000009"),
    "Usage dashboard tiles",
    "ready",
    "medium",
    "Implement three summary tiles on the usage dashboard: current-period spend, forecast, and anomaly alerts. Match existing billing design tokens.",
    CREATED.d5,
    {
      project_id: DEMO_SECOND_PROJECT_ID,
    },
  ),
  task(
    reg("f000000a-0000-4000-8000-00000000000a"),
    "Unassigned triage: docs site",
    "ready",
    "low",
    "Several broken links reported on the developer docs site. Triage scope, estimate effort, and assign an owner.",
    CREATED.justNow,
    {},
  ),
  task(
    reg("f000000b-0000-4000-8000-00000000000b"),
    "Parent: onboarding epic",
    "running",
    "medium",
    "Coordinate new-user onboarding improvements across empty states, analytics beacons, and help-center links.",
    CREATED.mo2,
    {
      project_id: DEMO_PRIMARY_PROJECT_ID,
      milestone: M_DISC,
      children: [
        task(
          reg("f000000c-0000-4000-8000-00000000000c"),
          "Child: empty state illustrations",
          "done",
          "low",
          "Replace placeholder illustrations on the welcome screen and first-run checklist with final brand assets from design.",
          CREATED.mo3,
          {
            parent_id: "f000000b-0000-4000-8000-00000000000b",
            project_id: DEMO_PRIMARY_PROJECT_ID,
          },
        ),
        task(
          reg("f000000d-0000-4000-8000-00000000000d"),
          "Child: analytics beacon",
          "ready",
          "medium",
          "Wire onboarding step completion events to the product analytics pipeline. Include step name, duration, and drop-off reason.",
          CREATED.d3,
          {
            parent_id: "f000000b-0000-4000-8000-00000000000b",
            project_id: DEMO_PRIMARY_PROJECT_ID,
          },
        ),
      ],
    },
  ),
  task(
    reg("fafaf001-fafa-4afa-bafa-000000000001"),
    "Rotate JWT signing keys for staging",
    "done",
    "medium",
    "Rotate staging signing keys ahead of production rollout. Verify all services pick up the new JWKS endpoint within the 72-hour overlap window.",
    CREATED.mo5,
    {
      project_id: DEMO_PRIMARY_PROJECT_ID,
      milestone: M_JWT,
    },
  ),
  task(
    reg("fafaf002-fafa-4afa-bafa-000000000002"),
    "Add retry logic to Stripe webhook handler",
    "ready",
    "high",
    "Retry failed webhook deliveries with exponential backoff and surface permanent failures to the ops dashboard.",
    CREATED.d1,
    {
      project_id: DEMO_SECOND_PROJECT_ID,
    },
  ),
  task(
    reg("fafaf003-fafa-4afa-bafa-000000000003"),
    "Document API rate limits in developer portal",
    "ready",
    "low",
    "Add a rate-limits page to the developer portal covering per-endpoint quotas, burst behavior, and 429 response headers.",
    CREATED.w1,
    {},
  ),
  task(
    reg("fafaf004-fafa-4afa-bafa-000000000004"),
    "Migrate legacy sessions to new token format",
    "blocked",
    "high",
    "Convert remaining cookie-based sessions to JWT pairs on next login. Blocked until the mobile app ships token refresh support.",
    CREATED.mo1,
    {
      project_id: DEMO_PRIMARY_PROJECT_ID,
      milestone: M_JWT,
    },
  ),
  task(
    reg("fafaf005-fafa-4afa-bafa-000000000005"),
    "Fix timezone display on usage export CSV",
    "ready",
    "medium",
    "Usage export CSVs show UTC timestamps without a label. Format timestamps in the account timezone and add a column header note.",
    CREATED.min45,
    {
      project_id: DEMO_SECOND_PROJECT_ID,
    },
  ),
  task(
    reg("fafaf006-fafa-4afa-bafa-000000000006"),
    "Add CSRF protection to login form",
    "done",
    "medium",
    "Issue a CSRF token on the login page and validate it server-side on POST. Regression-test OAuth redirect flows.",
    CREATED.mo2,
    {
      project_id: DEMO_PRIMARY_PROJECT_ID,
      milestone: M_JWT,
    },
  ),
  task(
    reg("fafaf007-fafa-4afa-bafa-000000000007"),
    "Implement invoice PDF download endpoint",
    "ready",
    "medium",
    "Expose GET /invoices/{id}/pdf returning a cached PDF with correct Content-Disposition headers and audit logging.",
    CREATED.d5,
    {
      project_id: DEMO_SECOND_PROJECT_ID,
    },
  ),
  task(
    reg("fafaf008-fafa-4afa-bafa-000000000008"),
    "Update dependency audit in CI pipeline",
    "ready",
    "low",
    "Switch CI from npm audit to osv-scanner and fail the build on critical CVEs in production dependencies.",
    CREATED.w3,
    {},
  ),
  task(
    reg("fafaf009-fafa-4afa-bafa-000000000009"),
    "Resolve OAuth redirect loop on mobile Safari",
    "blocked",
    "high",
    "Users on iOS 17 Safari hit an infinite redirect during OAuth login. Reproduce on device farm and patch redirect URI validation.",
    CREATED.d3,
    {
      project_id: DEMO_PRIMARY_PROJECT_ID,
    },
  ),
  task(
    reg("fafaf00a-fafa-4afa-bafa-00000000000a"),
    "Normalize currency formatting in billing dashboard",
    "done",
    "low",
    "Some dashboard tiles mix USD and locale-formatted amounts. Standardize on Intl.NumberFormat with the account currency setting.",
    CREATED.mo1,
    {
      project_id: DEMO_SECOND_PROJECT_ID,
    },
  ),
];

const DEMO_TASK_BY_ID = new Map<string, Record<string, unknown>>();
for (const row of ROOT_TASKS) {
  DEMO_TASK_BY_ID.set(row.id as string, row);
  const ch = row.children as Record<string, unknown>[] | undefined;
  if (ch) {
    for (const c of ch) {
      DEMO_TASK_BY_ID.set(c.id as string, c);
    }
  }
}

export function demoProjectsListWire(): unknown {
  return {
    projects: [
      {
        id: DEMO_PRIMARY_PROJECT_ID,
        name: "AuthV2",
        description: "JWT + session hardening across services.",
        status: "active",
        context_summary: "Primary operator sandbox project.",
        is_default: false,
        created_at: ISO,
        updated_at: ISO,
      },
      {
        id: DEMO_SECOND_PROJECT_ID,
        name: "Billing insights",
        description: "Usage metering, exports, and anomaly detection.",
        status: "active",
        context_summary: "Cross-team billing context.",
        is_default: false,
        created_at: ISO,
        updated_at: ISO,
      },
      {
        id: DEMO_THIRD_PROJECT_ID,
        name: "Archived pilot",
        description: "Superseded experiment — kept for layout regression.",
        status: "archived",
        context_summary: "",
        is_default: false,
        created_at: ISO,
        updated_at: ISO,
      },
    ],
    limit: 100,
  };
}

export function demoProjectWire(id: string): unknown | null {
  if (!isDemoProjectId(id)) return null;
  const row = (demoProjectsListWire() as { projects: { id: string }[] }).projects.find((p) => p.id === id);
  return row ?? null;
}

export function demoContextWire(projectId: string): unknown {
  if (projectId !== DEMO_PRIMARY_PROJECT_ID) {
    return { items: [], edges: [], limit: 100 };
  }
  return {
    items: [
      {
        id: C1,
        project_id: DEMO_PRIMARY_PROJECT_ID,
        tag: "decision",
        title: "JWT-first for partner APIs",
        description: "Partner auth choice",
        body: "Partners accept bearer tokens only; cookies reserved for first-party.",
        created_by: "user",
        pinned: true,
        created_at: ISO,
        updated_at: ISO,
      },
      {
        id: C2,
        project_id: DEMO_PRIMARY_PROJECT_ID,
        tag: "constraint",
        title: "No PII in logs",
        description: "Logging redaction rule",
        body: "Structured logs must redact email and phone by default.",
        created_by: "user",
        pinned: false,
        created_at: ISO,
        updated_at: ISO,
      },
      {
        id: C3,
        project_id: DEMO_PRIMARY_PROJECT_ID,
        tag: "note",
        title: "Rotation cadence",
        description: "Key rotation schedule",
        body: "Signing keys rotate every 30 days; overlap window 72h.",
        created_by: "agent",
        pinned: false,
        created_at: ISO,
        updated_at: ISO,
      },
      {
        id: C4,
        project_id: DEMO_PRIMARY_PROJECT_ID,
        tag: "decision",
        title: "Session fixation mitigation",
        description: "Session hardening",
        body: "Regenerate session id post-auth; SameSite=Lax default.",
        created_by: "user",
        pinned: false,
        created_at: ISO,
        updated_at: ISO,
      },
      {
        id: C5,
        project_id: DEMO_PRIMARY_PROJECT_ID,
        tag: "constraint",
        title: "EU residency",
        description: "Data residency constraint",
        body: "Auth metadata stores primary region EU.",
        created_by: "user",
        pinned: false,
        created_at: ISO,
        updated_at: ISO,
      },
      {
        id: C6,
        project_id: DEMO_PRIMARY_PROJECT_ID,
        tag: "note",
        title: "Load test window",
        description: "Allowed load-test hours",
        body: "Saturdays 02:00–06:00 UTC only.",
        created_by: "user",
        pinned: false,
        created_at: ISO,
        updated_at: ISO,
      },
    ],
    edges: [],
    limit: 100,
  };
}

const cyclesPhasesEmpty = {
  cycles: { by_status: {}, by_triggered_by: {} },
  phases: {
    by_phase_status: {
      execute: {},
      verify: {},
    },
  },
  runner: {
    by_runner: {},
    by_model: {},
    by_runner_model: {},
    by_runner_model_resolved: {},
  },
  recent_failures: [] as unknown[],
};

export function demoTaskStatsWire(): unknown {
  const byStatus: Record<string, number> = {};
  for (const row of ROOT_TASKS) {
    const s = row.status as string;
    byStatus[s] = (byStatus[s] ?? 0) + 1;
  }
  const total = ROOT_TASKS.length;
  return {
    total,
    ready: byStatus.ready ?? 0,
    critical: byStatus.critical ?? 0,
    scheduled: 2,
    by_status: byStatus,
    by_priority: { low: 4, medium: total - 8, high: 4, critical: 4 },
    ...cyclesPhasesEmpty,
  };
}

export function demoTasksListWire(
  limit: number,
  offset: number,
  afterId: string | null | undefined,
): unknown {
  if (afterId) {
    return { tasks: [], limit, offset: 0, has_more: false };
  }
  const slice = ROOT_TASKS.slice(offset, offset + limit);
  return {
    tasks: slice,
    limit,
    offset,
    has_more: offset + slice.length < ROOT_TASKS.length,
  };
}

export function demoTaskWire(id: string): unknown | null {
  const row = DEMO_TASK_BY_ID.get(id);
  return row ? { ...row } : null;
}

export function demoTaskDraftsWire(): unknown {
  return {
    drafts: [
      {
        id: "d1111111-1111-4111-8111-111111111111",
        name: "Draft: incident retro",
        created_at: ISO,
        updated_at: ISO,
      },
      {
        id: "d2222222-2222-4222-8222-222222222222",
        name: "Draft: Q2 planning",
        created_at: ISO,
        updated_at: ISO,
      },
    ],
  };
}

export function demoTaskTemplatesWire(): unknown {
  return {
    templates: [
      {
        id: "t1111111-1111-4111-8111-111111111111",
        name: "Template: incident runbook",
        created_at: ISO,
        updated_at: ISO,
      },
    ],
  };
}

export function demoTaskEventsWire(taskId: string): unknown {
  return {
    task_id: taskId,
    events: [],
    approval_pending: false,
    has_more_newer: false,
    has_more_older: false,
    limit: 200,
  };
}

export function demoTaskCyclesListWire(taskId: string): unknown {
  return {
    task_id: taskId,
    cycles: [],
    limit: 50,
    has_more: false,
  };
}

export function demoTaskChecklistWire(): unknown {
  return { items: [] };
}

export function demoCycleFailuresWire(): unknown {
  return {
    total: 0,
    limit: 50,
    offset: 0,
    sort: "at_desc",
    reason_sort_truncated: false,
    failures: [],
  };
}

export function isDemoTaskId(id: string): boolean {
  return DEMO_TASK_BY_ID.has(id);
}

export function allRegisteredDemoTaskIds(): readonly string[] {
  return ALL_TASK_IDS;
}
