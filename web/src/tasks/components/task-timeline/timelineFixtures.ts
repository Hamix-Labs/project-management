import type { TimelineEvent } from "./timelineTypes";

const DAY_MS = 24 * 60 * 60 * 1000;

/** Build ISO `at` for a local calendar day offset and clock time. */
function atLocal(
  now: Date,
  daysAgo: number,
  hour: number,
  minute: number,
): string {
  const d = new Date(now.getTime());
  d.setHours(hour, minute, 0, 0);
  d.setTime(d.getTime() - daysAgo * DAY_MS);
  return d.toISOString();
}

/**
 * Fixture activity feed shaped for the home Timeline view.
 * Pass `now` in tests for stable relative grouping.
 */
export function createTimelineFixtures(now: Date = new Date()): TimelineEvent[] {
  return [
    {
      id: "ev-1",
      kind: "verification-passed",
      category: "verification",
      at: atLocal(now, 0, 10, 42),
      title: "Verification passed",
      highlight: "checkout-flow.spec",
      detail: "All 14 assertions succeeded on the release candidate build.",
      taskId: "f0000131-0000-4000-8000-000000000131",
      taskRef: "f0000131",
      meta: ["14 / 14 assertions", "1.8s"],
    },
    {
      id: "ev-2",
      kind: "agent-started",
      category: "agents",
      at: atLocal(now, 0, 10, 31),
      title: "Agent run started",
      highlight: "refactor-agent",
      detail: "Triggered by task status change to Ready for agent.",
      taskId: "f0000131-0000-4000-8000-000000000131",
      taskRef: "f0000131",
      meta: ["refactor-agent", "queued"],
    },
    {
      id: "ev-3",
      kind: "status-changed",
      category: "tasks",
      at: atLocal(now, 0, 9, 7),
      title: "Status changed",
      highlight: "on f0000151",
      detail:
        "In Progress moved forward as the Redis migration entered review.",
      taskId: "f0000151-0000-4000-8000-000000000151",
      taskRef: "f0000151",
      meta: ["In Progress → Verification"],
    },
    {
      id: "ev-4",
      kind: "status-changed",
      category: "tasks",
      at: atLocal(now, 1, 16, 18),
      title: "Status changed",
      highlight: "on f0000131",
      detail:
        "In review → Ready for agent after the reviewer approved the split.",
      taskId: "f0000131-0000-4000-8000-000000000131",
      taskRef: "f0000131",
      meta: ["In review → Ready for agent"],
    },
    {
      id: "ev-5",
      kind: "verification-failed",
      category: "verification",
      at: atLocal(now, 1, 14, 5),
      title: "Verification failed",
      highlight: "payment-api.spec",
      detail: "2 of 9 assertions failed — timeout on retry logic under load.",
      taskId: "f0000142-0000-4000-8000-000000000142",
      taskRef: "f0000142",
      meta: ["2 / 9 failed", "timeout"],
    },
    {
      id: "ev-6",
      kind: "comment",
      category: "tasks",
      at: atLocal(now, 1, 11, 24),
      title: "New comment",
      highlight: "on f0000127",
      detail:
        '"The flake only reproduces when the seed data runs in parallel."',
      taskId: "f0000127-0000-4000-8000-000000000127",
      taskRef: "f0000127",
      meta: ["from SO"],
    },
    {
      id: "ev-7",
      kind: "agent-finished",
      category: "agents",
      at: atLocal(now, 2, 15, 40),
      title: "Agent run finished",
      highlight: "docs-agent",
      detail: "Generated a first draft of the webhook payload reference.",
      taskId: "f0000138-0000-4000-8000-000000000138",
      taskRef: "f0000138",
      meta: ["docs-agent", "42s", "success"],
    },
    {
      id: "ev-8",
      kind: "task-created",
      category: "tasks",
      at: atLocal(now, 2, 9, 0),
      title: "Task created",
      highlight: "f0000142",
      detail:
        '"Add retry with backoff to the sync worker" was added to Backlog.',
      taskId: "f0000142-0000-4000-8000-000000000142",
      taskRef: "f0000142",
      meta: ["Backlog", "High"],
    },
  ];
}

/** Default fixtures relative to module load time (SPA home view). */
export const TIMELINE_FIXTURES = createTimelineFixtures();
