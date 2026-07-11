import { describe, expect, it } from "vitest";
import {
  parseTask,
  parseTaskCycle,
  parseTaskCycleDetail,
  parseTaskCyclePhase,
  parseTaskCycleStreamResponse,
  parseTaskCyclesListResponse,
  parseTaskEventDetail,
  parseTaskEventsResponse,
  parseTaskDraftDetail,
  parseTaskListResponse,
} from "./parseTaskApi";
import { TASK_TEST_DEFAULTS } from "@/test/taskDefaults";
import { TASK_EVENT_TYPES } from "@/types";

const validTask = {
  id: "a1",
  title: "One",
  initial_prompt: "",
  status: "ready",
  priority: "medium",
  tags: [] as string[],
  depends_on: [] as { task_id: string; satisfies?: string }[],
  ...TASK_TEST_DEFAULTS,
};

describe("parseTask", () => {
  it("accepts a well-formed task", () => {
    expect(parseTask(validTask)).toEqual(validTask);
  });

  it("defaults missing initial_prompt to empty string", () => {
    expect(
      parseTask({
        id: "a1",
        title: "One",
        status: "ready",
        priority: "medium",
      }),
    ).toEqual(validTask);
  });

  it("rejects invalid status", () => {
    expect(() =>
      parseTask({ ...validTask, status: "nope" }),
    ).toThrow(/known task status/);
  });

  it("rejects non-object", () => {
    expect(() => parseTask(null)).toThrow(/object/);
  });

  it("parses tags, milestone, depends_on, and gate", () => {
    const parsed = parseTask({
      ...validTask,
      tags: ["backend", "api"],
      milestone: "M1",
      depends_on: ["dep-1"],
      gate: {
        kind: "manual_approval",
        status: "pending_release",
        hold: false,
        criteria: [
          { id: "c1", text: "Review", done: false, sort_order: 0 },
        ],
      },
    });
    expect(parsed.tags).toEqual(["backend", "api"]);
    expect(parsed.milestone).toBe("M1");
    expect(parsed.depends_on).toEqual([{ task_id: "dep-1", satisfies: "done" }]);
    expect(parsed.gate).toEqual({
      kind: "manual_approval",
      status: "pending_release",
      hold: false,
      criteria: [
        { id: "c1", text: "Review", done: false, sort_order: 0 },
      ],
    });
  });

  it("defaults tags and depends_on to empty arrays when omitted", () => {
    const parsed = parseTask(validTask);
    expect(parsed.tags).toEqual([]);
    expect(parsed.depends_on).toEqual([]);
    expect(parsed.gate).toBeUndefined();
    expect(parsed.milestone).toBeUndefined();
  });

  it("rejects invalid gate status", () => {
    expect(() =>
      parseTask({
        ...validTask,
        gate: {
          kind: "manual_approval",
          status: "nope",
          hold: false,
        },
      }),
    ).toThrow(/gate status/);
  });
});

describe("parseTaskListResponse", () => {
  it("parses list envelope", () => {
    expect(
      parseTaskListResponse({
        tasks: [validTask],
        limit: 200,
        offset: 0,
        has_more: false,
      }),
    ).toEqual({ tasks: [validTask], limit: 200, offset: 0, has_more: false });
  });

  it("defaults has_more when omitted", () => {
    expect(
      parseTaskListResponse({
        tasks: [validTask],
        limit: 50,
        offset: 0,
      }),
    ).toEqual({ tasks: [validTask], limit: 50, offset: 0, has_more: false });
  });

  it("parses has_more true", () => {
    expect(
      parseTaskListResponse({
        tasks: [validTask],
        limit: 2,
        offset: 0,
        has_more: true,
      }),
    ).toEqual({ tasks: [validTask], limit: 2, offset: 0, has_more: true });
  });

  it("rejects invalid has_more", () => {
    expect(() =>
      parseTaskListResponse({
        tasks: [validTask],
        limit: 1,
        offset: 0,
        has_more: "yes",
      }),
    ).toThrow(/has_more/);
  });

  it("rejects non-array tasks", () => {
    expect(() =>
      parseTaskListResponse({ tasks: {}, limit: 0, offset: 0 }),
    ).toThrow(/array/);
  });

  it("treats null tasks as empty array (legacy Go nil slice JSON)", () => {
    expect(
      parseTaskListResponse({
        tasks: null,
        limit: 50,
        offset: 0,
      }),
    ).toEqual({ tasks: [], limit: 50, offset: 0, has_more: false });
  });
});

describe("parseTaskEventsResponse", () => {
  it("parses events envelope", () => {
    const at = "2026-01-01T12:00:00Z";
    expect(
      parseTaskEventsResponse({
        task_id: "tid",
        events: [
          {
            seq: 1,
            at,
            type: "task_created",
            by: "user",
            data: {},
          },
        ],
      }),
    ).toEqual({
      task_id: "tid",
      events: [
        {
          seq: 1,
          at,
          type: "task_created",
          by: "user",
          data: {},
        },
      ],
      approval_pending: false,
      has_more_newer: false,
      has_more_older: false,
    });
  });

  it("parses optional user_response on events", () => {
    const at = "2026-01-01T12:00:00Z";
    expect(
      parseTaskEventsResponse({
        task_id: "tid",
        events: [
          {
            seq: 2,
            at,
            type: "approval_requested",
            by: "agent",
            data: {},
            user_response: "Approved",
          },
        ],
        approval_pending: false,
      }),
    ).toEqual({
      task_id: "tid",
      events: [
        {
          seq: 2,
          at,
          type: "approval_requested",
          by: "agent",
          data: {},
          user_response: "Approved",
          response_thread: [{ at, by: "user", body: "Approved" }],
        },
      ],
      approval_pending: false,
      has_more_newer: false,
      has_more_older: false,
    });
  });

  it("accepts every server-declared EventType (regression: cycle/phase mirrors)", () => {
    // The backend emits cycle_started / cycle_completed / cycle_failed /
    // phase_started / phase_completed / phase_failed / phase_skipped audit
    // mirrors as soon as a real agent run dispatches (see
    // pkgs/tasks/domain/enums.go). When TASK_EVENT_TYPES drifted from the
    // server enum, parseTaskEventsResponse rejected the entire /events
    // payload with "event type must be a known value" the moment any of
    // those rows landed, collapsing the whole Updates section into an
    // error banner. Walk every declared TaskEventType through the parser
    // so future server-side additions either get mirrored here or fail
    // this test loudly instead of breaking the timeline silently in prod.
    const at = "2026-01-01T12:00:00Z";
    const events = TASK_EVENT_TYPES.map((type, idx) => ({
      seq: idx + 1,
      at,
      type,
      by: "agent" as const,
      data: {},
    }));
    const out = parseTaskEventsResponse({ task_id: "tid", events });
    expect(out.events.map((e) => e.type)).toEqual(
      TASK_EVENT_TYPES as readonly string[],
    );
  });

  it("parses keyset-paged envelope", () => {
    const at = "2026-01-01T12:00:00Z";
    expect(
      parseTaskEventsResponse({
        task_id: "tid",
        limit: 20,
        total: 45,
        range_start: 21,
        range_end: 40,
        has_more_newer: true,
        has_more_older: true,
        approval_pending: true,
        events: [
          {
            seq: 3,
            at,
            type: "sync_ping",
            by: "user",
            data: {},
          },
        ],
      }),
    ).toEqual({
      task_id: "tid",
      limit: 20,
      total: 45,
      range_start: 21,
      range_end: 40,
      has_more_newer: true,
      has_more_older: true,
      approval_pending: true,
      events: [
        {
          seq: 3,
          at,
          type: "sync_ping",
          by: "user",
          data: {},
        },
      ],
    });
  });
});

describe("parseTaskEventDetail", () => {
  it("parses GET /tasks/{id}/events/{seq} envelope", () => {
    const at = "2026-01-02T15:30:00.000Z";
    expect(
      parseTaskEventDetail({
        task_id: "tid",
        seq: 4,
        at,
        type: "approval_requested",
        by: "agent",
        data: { reason: "review" },
      }),
    ).toEqual({
      task_id: "tid",
      seq: 4,
      at,
      type: "approval_requested",
      by: "agent",
      data: { reason: "review" },
    });
  });

  it("parses user_response on event detail", () => {
    const at = "2026-01-02T15:30:00.000Z";
    const user_response_at = "2026-01-02T16:00:00.000Z";
    expect(
      parseTaskEventDetail({
        task_id: "tid",
        seq: 4,
        at,
        type: "task_failed",
        by: "agent",
        data: {},
        user_response: "Retry scheduled",
        user_response_at,
      }),
    ).toEqual({
      task_id: "tid",
      seq: 4,
      at,
      type: "task_failed",
      by: "agent",
      data: {},
      user_response: "Retry scheduled",
      user_response_at,
      response_thread: [
        { at: user_response_at, by: "user", body: "Retry scheduled" },
      ],
    });
  });
});

const validCycle = {
  id: "cyc-1",
  task_id: "task-1",
  attempt_seq: 1,
  status: "running",
  started_at: "2026-04-18T10:00:00.000Z",
  triggered_by: "user",
  meta: { source: "manual" },
};

const emptyCycleMeta = {
  runner: "",
  runner_version: "",
  cursor_model: "",
  cursor_model_effective: "",
  prompt_hash: "",
};

describe("parseTaskCycle", () => {
  it("accepts a well-formed running cycle and defaults meta when missing", () => {
    expect(parseTaskCycle(validCycle)).toEqual({
      ...validCycle,
      cycle_meta: emptyCycleMeta,
    });
    const noMeta = { ...validCycle };
    delete (noMeta as Partial<typeof validCycle>).meta;
    expect(parseTaskCycle(noMeta)).toEqual({
      ...validCycle,
      meta: {},
      cycle_meta: emptyCycleMeta,
    });
  });

  it("includes optional ended_at and parent_cycle_id when present", () => {
    const out = parseTaskCycle({
      ...validCycle,
      status: "succeeded",
      ended_at: "2026-04-18T10:05:00.000Z",
      parent_cycle_id: "cyc-0",
    });
    expect(out.ended_at).toBe("2026-04-18T10:05:00.000Z");
    expect(out.parent_cycle_id).toBe("cyc-0");
  });

  it("rejects unknown status, bad actor, and unparseable started_at", () => {
    expect(() => parseTaskCycle({ ...validCycle, status: "weird" })).toThrow(
      /known cycle status/,
    );
    expect(() =>
      parseTaskCycle({ ...validCycle, triggered_by: "robot" }),
    ).toThrow(/user or agent/);
    expect(() =>
      parseTaskCycle({ ...validCycle, started_at: "not-a-date" }),
    ).toThrow(/started_at/);
  });

  it("extracts the typed cycle_meta projection when the server provides it", () => {
    const out = parseTaskCycle({
      ...validCycle,
      meta: {
        runner: "cursor-cli",
        runner_version: "0.42.0",
        cursor_model: "",
        cursor_model_effective: "opus",
        prompt_hash: "deadbeef",
      },
      cycle_meta: {
        runner: "cursor-cli",
        runner_version: "0.42.0",
        cursor_model: "",
        cursor_model_effective: "opus",
        prompt_hash: "deadbeef",
      },
    });
    expect(out.cycle_meta).toEqual({
      runner: "cursor-cli",
      runner_version: "0.42.0",
      cursor_model: "",
      cursor_model_effective: "opus",
      prompt_hash: "deadbeef",
    });
  });

  it("falls back to meta when cycle_meta is absent (forward/back compat)", () => {
    const out = parseTaskCycle({
      ...validCycle,
      meta: {
        runner: "cursor-cli",
        runner_version: "0.42.0",
        cursor_model: "opus",
        cursor_model_effective: "opus",
        prompt_hash: "deadbeef",
      },
    });
    // Same shape as the cycle_meta object the server would have sent.
    expect(out.cycle_meta).toEqual({
      runner: "cursor-cli",
      runner_version: "0.42.0",
      cursor_model: "opus",
      cursor_model_effective: "opus",
      prompt_hash: "deadbeef",
    });
  });

  it("preserves empty strings as semantic values, not coerced to undefined", () => {
    const out = parseTaskCycle({
      ...validCycle,
      cycle_meta: {
        runner: "cursor-cli",
        runner_version: "0.42.0",
        cursor_model: "",
        cursor_model_effective: "",
        prompt_hash: "",
      },
    });
    // "" is the truth: no model anywhere — must NOT be coerced to undefined.
    expect(out.cycle_meta.cursor_model).toBe("");
    expect(out.cycle_meta.cursor_model_effective).toBe("");
    expect(out.cycle_meta.prompt_hash).toBe("");
  });
});

const validPhase = {
  id: "ph-1",
  cycle_id: "cyc-1",
  phase: "execute",
  phase_seq: 1,
  status: "running",
  started_at: "2026-04-18T10:00:01.000Z",
  details: {},
};

describe("parseTaskCyclePhase", () => {
  it("accepts a well-formed running phase and defaults details when missing", () => {
    expect(parseTaskCyclePhase(validPhase)).toEqual(validPhase);
    const noDetails = { ...validPhase };
    delete (noDetails as Partial<typeof validPhase>).details;
    expect(parseTaskCyclePhase(noDetails)).toEqual({
      ...validPhase,
      details: {},
    });
  });

  it("includes optional summary, ended_at, event_seq when present", () => {
    const out = parseTaskCyclePhase({
      ...validPhase,
      status: "succeeded",
      ended_at: "2026-04-18T10:01:00.000Z",
      summary: "diagnosed root cause",
      event_seq: 7,
      details: { hint: "x" },
    });
    expect(out.summary).toBe("diagnosed root cause");
    expect(out.ended_at).toBe("2026-04-18T10:01:00.000Z");
    expect(out.event_seq).toBe(7);
    expect(out.details).toEqual({ hint: "x" });
  });

  it("rejects unknown phase or status", () => {
    expect(() => parseTaskCyclePhase({ ...validPhase, phase: "ship" })).toThrow(
      /known phase/,
    );
    expect(() => parseTaskCyclePhase({ ...validPhase, status: "weird" })).toThrow(
      /known phase status/,
    );
  });

  // Historical cycle rows that predate the diagnose/persist trim must
  // still parse so the SPA can render an honest audit trail for old
  // attempts instead of throwing on the detail page. Pinned because the
  // write-side PHASES enum no longer includes these values — accidentally
  // tightening parsePhase to PHASES alone would silently break old data.
  it("accepts legacy diagnose / persist phase values on read", () => {
    expect(parseTaskCyclePhase({ ...validPhase, phase: "diagnose" }).phase).toBe(
      "diagnose",
    );
    expect(parseTaskCyclePhase({ ...validPhase, phase: "persist" }).phase).toBe(
      "persist",
    );
  });
});

describe("parseTaskCyclesListResponse", () => {
  it("parses an empty list with limit and has_more", () => {
    expect(
      parseTaskCyclesListResponse({
        task_id: "task-1",
        cycles: [],
        limit: 50,
        has_more: false,
      }),
    ).toEqual({ task_id: "task-1", cycles: [], limit: 50, has_more: false });
  });

  it("parses cycles array element-by-element with index in error", () => {
    expect(() =>
      parseTaskCyclesListResponse({
        task_id: "task-1",
        cycles: [validCycle, { ...validCycle, status: "weird" }],
        limit: 10,
        has_more: false,
      }),
    ).toThrow(/cycles\[1\]/);
  });

  it("rejects when cycles is missing or not an array", () => {
    expect(() =>
      parseTaskCyclesListResponse({
        task_id: "task-1",
        limit: 10,
        has_more: false,
      }),
    ).toThrow(/cycles must be an array/);
  });
});

describe("parseTaskCycleDetail", () => {
  it("parses cycle + ordered phases envelope", () => {
    const out = parseTaskCycleDetail({
      ...validCycle,
      phases: [
        validPhase,
        {
          ...validPhase,
          id: "ph-2",
          phase: "execute",
          phase_seq: 2,
          status: "running",
        },
      ],
    });
    expect(out.id).toBe("cyc-1");
    expect(out.phases).toHaveLength(2);
    expect(out.phases[1].phase).toBe("execute");
  });

  it("rejects when phases is missing", () => {
    expect(() => parseTaskCycleDetail(validCycle)).toThrow(
      /phases must be an array/,
    );
  });
});

describe("parseTaskCycleStreamResponse", () => {
  it("parses persisted stream events", () => {
    expect(
      parseTaskCycleStreamResponse({
        task_id: "task-1",
        cycle_id: "cyc-1",
        events: [
          {
            id: "stream-1",
            task_id: "task-1",
            cycle_id: "cyc-1",
            phase_seq: 2,
            stream_seq: 1,
            at: "2026-01-01T00:00:00Z",
            source: "cursor",
            kind: "tool_call",
            subtype: "started",
            message: "Read file",
            tool: "ReadFile",
            payload: { kind: "tool_call" },
          },
        ],
        limit: 100,
        has_more: true,
        next_after_seq: 1,
      }),
    ).toEqual({
      task_id: "task-1",
      cycle_id: "cyc-1",
      events: [
        {
          id: "stream-1",
          task_id: "task-1",
          cycle_id: "cyc-1",
          phase_seq: 2,
          stream_seq: 1,
          at: "2026-01-01T00:00:00Z",
          source: "cursor",
          kind: "tool_call",
          subtype: "started",
          message: "Read file",
          tool: "ReadFile",
          payload: { kind: "tool_call" },
        },
      ],
      limit: 100,
      has_more: true,
      next_after_seq: 1,
    });
  });
});

describe("parseTaskDraftDetail (payload.priority validation)", () => {
  // Regression: parseDraftPayload used to do
  //   priority: (value.priority as TaskDraftPayload["priority"]) ?? "",
  // which let arbitrary server values through unvalidated, even though
  // parsePriority(). The downstream UI (PrioritySelect) silently drops
  // invalid values; the parser is the chokepoint that should reject them.
  const baseDraft = {
    id: "d1",
    name: "draft",
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
    payload: {
      title: "t",
      initial_prompt: "p",
      priority: "medium",
      checklist_items: [],
    },
  };

  it("accepts a valid PriorityChoice value on the parent draft", () => {
    const out = parseTaskDraftDetail(baseDraft);
    expect(out.payload.priority).toBe("medium");
  });

  it("accepts an empty-string priority (user has not selected one yet)", () => {
    const out = parseTaskDraftDetail({
      ...baseDraft,
      payload: { ...baseDraft.payload, priority: "" },
    });
    expect(out.payload.priority).toBe("");
  });

  it("defaults a missing priority to empty string", () => {
    const { priority: _omit, ...rest } = baseDraft.payload;
    const out = parseTaskDraftDetail({ ...baseDraft, payload: rest });
    expect(out.payload.priority).toBe("");
  });

  it("rejects an unknown priority string on the parent draft", () => {
    expect(() =>
      parseTaskDraftDetail({
        ...baseDraft,
        payload: { ...baseDraft.payload, priority: "Critical" },
      }),
    ).toThrow(/known task priority/);
  });

  it("rejects a non-string priority on the parent draft", () => {
    expect(() =>
      parseTaskDraftDetail({
        ...baseDraft,
        payload: { ...baseDraft.payload, priority: 42 },
      }),
    ).toThrow(/known task priority/);
  });

  it("parses checklist_items objects with verify_commands", () => {
    const out = parseTaskDraftDetail({
      ...baseDraft,
      payload: {
        ...baseDraft.payload,
        checklist_items: [
          {
            text: "Ship with tests",
            verify_commands: [
              { command: "go test ./...", expected_outcome: "exit 0" },
            ],
          },
        ],
      },
    });
    expect(out.payload.checklist_items).toEqual([
      {
        text: "Ship with tests",
        verify_commands: [
          { command: "go test ./...", expected_outcome: "exit 0" },
        ],
      },
    ]);
  });

  it("accepts legacy string checklist_items entries", () => {
    const out = parseTaskDraftDetail({
      ...baseDraft,
      payload: {
        ...baseDraft.payload,
        checklist_items: ["Legacy criterion"],
      },
    });
    expect(out.payload.checklist_items).toEqual([{ text: "Legacy criterion" }]);
  });
});
