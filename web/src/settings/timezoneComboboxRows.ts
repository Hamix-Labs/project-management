import {
  formatTimezoneMenuLabel,
  getTimezoneSearchHaystack,
  matchesTimezoneSearchQuery,
  type TimezoneSelectOption,
} from "@/shared/time/appTimezone";
import type {
  TimezoneComboboxProps,
  TimezoneComboboxRow,
} from "./timezoneComboboxTypes";

export function timezoneRowKey(row: TimezoneComboboxRow): string {
  if (row.kind === "auto") return "auto";
  if (row.kind === "iana") return row.opt.value;
  return `custom-${row.value}`;
}

export function timezoneRowLabel(row: TimezoneComboboxRow, autoLabel: string): string {
  if (row.kind === "auto") return autoLabel;
  if (row.kind === "iana") return row.opt.label;
  return row.label;
}

export function timezoneRowValue(row: TimezoneComboboxRow): string {
  if (row.kind === "auto") return "";
  if (row.kind === "iana") return row.opt.value;
  return row.value;
}

export function isTimezoneRowSelected(row: TimezoneComboboxRow, value: string): boolean {
  return timezoneRowValue(row) === value;
}

export function resolveTimezoneSelectedLabel(
  value: string,
  autoLabel: string,
  options: TimezoneSelectOption[],
  customSaved: TimezoneComboboxProps["customSaved"],
): string {
  if (value === "") return autoLabel;
  const hit = options.find((o) => o.value === value);
  if (hit) return hit.label;
  if (customSaved && value === customSaved.value) return customSaved.label;
  return formatTimezoneMenuLabel(value);
}

export function buildTimezoneComboboxRows(
  search: string,
  autoHaystack: string,
  filteredIana: TimezoneSelectOption[],
  customSaved: TimezoneComboboxProps["customSaved"],
): TimezoneComboboxRow[] {
  const out: TimezoneComboboxRow[] = [];
  const q = search.trim();
  if (!q || matchesTimezoneSearchQuery(autoHaystack, q)) {
    out.push({ kind: "auto" });
  }
  for (const opt of filteredIana) {
    out.push({ kind: "iana", opt });
  }
  if (customSaved) {
    const ch = getTimezoneSearchHaystack({
      value: customSaved.value,
      label: customSaved.label,
    });
    if (!q || matchesTimezoneSearchQuery(ch, q)) {
      out.push({
        kind: "custom",
        value: customSaved.value,
        label: customSaved.label,
      });
    }
  }
  return out;
}
