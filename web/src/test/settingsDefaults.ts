import type { AppSettings } from "@/api/settings";

/**
 * Full AppSettings fixture with all required fields. Tests override
 * individual fields via the spread operator.
 */
export const APP_SETTINGS_DEFAULTS: AppSettings = {
  agent_paused: false,
  runner: "cursor",
  cursor_bin: "",
  cursor_model: "",
  verify_model: "",
  max_run_duration_seconds: 0,
  agent_task_parallelism: 8,
  agent_pickup_delay_seconds: 5,
  display_timezone: "UTC",
  optimistic_mutations_enabled: true,
  sse_replay_enabled: false,
};
