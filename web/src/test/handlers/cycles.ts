import { http, HttpResponse, type JsonBodyType } from "msw";

/** Placeholder for cycle endpoints referenced by full-page flows. */
export function taskCyclesEmpty(taskId: string) {
  return http.get(`/tasks/${taskId}/cycles`, () =>
    HttpResponse.json({
      task_id: taskId,
      cycles: [],
      limit: 50,
      has_more: false,
    }),
  );
}

export function taskCommitsEmpty(taskId: string) {
  return http.get(`/tasks/${taskId}/commits`, () =>
    HttpResponse.json({ task_id: taskId, commits: [] }),
  );
}

export function cycleDetailGet(
  taskId: string,
  cycleId: string,
  body: JsonBodyType,
) {
  return http.get(`/tasks/${taskId}/cycles/${cycleId}`, () =>
    HttpResponse.json(body),
  );
}

export function cycleStreamGet(
  taskId: string,
  cycleId: string,
  events: unknown[],
) {
  return http.get(`/tasks/${taskId}/cycles/${cycleId}/stream`, () =>
    HttpResponse.json({
      task_id: taskId,
      cycle_id: cycleId,
      events,
      limit: 500,
      has_more: false,
    }),
  );
}

export function cyclePageAuditEvents(taskId: string, events: unknown[]) {
  return http.get(`/tasks/${taskId}/events`, ({ request }) => {
    const url = new URL(request.url);
    if (url.searchParams.get("limit") !== "200") {
      return new HttpResponse(null, { status: 404 });
    }
    return HttpResponse.json({
      task_id: taskId,
      events,
      approval_pending: false,
    });
  });
}

export function cycleVerdictsGet(
  taskId: string,
  cycleId: string,
  body: JsonBodyType,
) {
  return http.get(`/tasks/${taskId}/cycles/${cycleId}/verdicts`, () =>
    HttpResponse.json(body),
  );
}

/** Handlers for TaskCycleDetailPage routing tests. */
export function cyclePageHandlers(options: {
  taskId: string;
  cycleId: string;
  cycle: JsonBodyType;
  streamEvents: JsonBodyType[];
  auditEvents: JsonBodyType[];
  verdicts?: JsonBodyType;
}) {
  const { taskId, cycleId, cycle, streamEvents, auditEvents, verdicts } =
    options;
  return [
    cycleDetailGet(taskId, cycleId, cycle),
    cycleStreamGet(taskId, cycleId, streamEvents),
    cyclePageAuditEvents(taskId, auditEvents),
    cycleVerdictsGet(
      taskId,
      cycleId,
      verdicts ?? {
        task_id: taskId,
        cycle_id: cycleId,
        criteria_reports: [],
        verify_reports: [],
        command_runs: [],
        commits: [],
      },
    ),
  ];
}
