/**
 * Read-side limits mirrored from `pkgs/tasks/handler/readpolicy/readpolicy.go`.
 * Keep in sync with `testdata/readlimits.json` (Go + TS contract tests).
 */
export const READ_LIMITS = {
  bootstrapListLimit: 20,
  bootstrapProjectsLimit: 100,
  bootstrapDraftsLimit: 50,
  taskListDefaultLimit: 50,
  taskListMaxLimit: 200,
  taskEventsDefaultLimit: 50,
  taskEventsMaxLimit: 200,
  cycleListDefaultLimit: 50,
  cycleListMaxLimit: 200,
  cycleStreamDefaultLimit: 100,
  cycleStreamMaxLimit: 500,
} as const;

export type ReadLimitKey = keyof typeof READ_LIMITS;
