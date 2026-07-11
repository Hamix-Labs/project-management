import type { ComboboxRow } from "@/components/combobox";

export type SettingsSelectOption = {
  value: string;
  label: string;
};

export type SettingsSelectRow = ComboboxRow;

export type SettingsSelectProps = {
  value: string;
  onChange: (value: string) => void;
  options: SettingsSelectOption[];
  testId: string;
  disabled?: boolean;
  ariaBusy?: boolean;
  searchable?: boolean;
  searchPlaceholder?: string;
  /** When set, renders section headers between options (e.g. model families). */
  rows?: SettingsSelectRow[];
};
