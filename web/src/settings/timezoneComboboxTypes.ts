import type {
  TimezoneSelectOption,
} from "@/shared/time/appTimezone";

export type TimezoneComboboxRow =
  | { kind: "auto" }
  | { kind: "iana"; opt: TimezoneSelectOption }
  | { kind: "custom"; value: string; label: string };

export type TimezoneComboboxProps = {
  value: string;
  onChange: (value: string) => void;
  browserTz: string;
  options: TimezoneSelectOption[];
  /** Saved IANA not present in `options` — show a third list row. */
  customSaved?: { value: string; label: string } | null;
  testId?: string;
};

export type TimezoneDropdownPosition = {
  top: number;
  left: number;
  width: number;
};
