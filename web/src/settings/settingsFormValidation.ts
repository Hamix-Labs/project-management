import {
  detectBrowserTimezone,
  formatInAppTimezone,
  getTimezoneSelectOptions,
} from "@/shared/time/appTimezone";
import type { SettingsFormState } from "./settingsForm";

export type SettingsNumericValidation = {
  maxInvalid: boolean;
  parallelismInvalid: boolean;
  pickupInvalid: boolean;
};

export type TimezoneDisplayContext = {
  browserTz: string;
  effectiveDisplayTimezone: string;
  lastUpdatedFormatted: string;
  showCustomTz: boolean;
};

function parseNonNegativeIntField(raw: string): boolean {
  const parsed = Number.parseInt(raw.trim() || "0", 10);
  return !Number.isFinite(parsed) || parsed < 0;
}

export function parseSettingsNumericValidation(
  form: SettingsFormState | null,
): SettingsNumericValidation {
  if (!form) {
    return {
      maxInvalid: false,
      parallelismInvalid: false,
      pickupInvalid: false,
    };
  }
  const parallelismParsed = Number.parseInt(form.agentTaskParallelism.trim() || "0", 10);
  const parallelismInvalid = !Number.isFinite(parallelismParsed) || parallelismParsed < 1;
  const pickupParsed = Number.parseInt(form.agentPickupDelaySeconds.trim() || "0", 10);
  const pickupInvalid =
    !Number.isFinite(pickupParsed) || pickupParsed < 0 || pickupParsed > 604800;
  return {
    maxInvalid: parseNonNegativeIntField(form.maxRunDurationSeconds),
    parallelismInvalid,
    pickupInvalid,
  };
}

export function computeTimezoneDisplayContext(
  form: SettingsFormState,
  lastUpdated: string,
  tzValueSet: Set<string>,
): TimezoneDisplayContext {
  const showCustomTz =
    form.displayTimezone.trim() !== "" && !tzValueSet.has(form.displayTimezone.trim());
  const browserTz = detectBrowserTimezone();
  const effectiveDisplayTimezone = form.displayTimezone.trim() || browserTz;
  const lastUpdatedFormatted = lastUpdated
    ? formatInAppTimezone(lastUpdated, effectiveDisplayTimezone, {
        timeZoneName: "longOffset",
      })
    : "";
  return {
    browserTz,
    effectiveDisplayTimezone,
    lastUpdatedFormatted,
    showCustomTz,
  };
}

export function buildTimezoneSelectValueSet(): Set<string> {
  return new Set(getTimezoneSelectOptions().map((o) => o.value));
}
