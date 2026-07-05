import type { CustomSelectOption } from "@/components/custom-select/customSelectModel";
import type { ComboboxRow } from "./comboboxMenuKeyboard";

/** Maps settings-style grouped rows to CustomSelect options (headers preserved). */
export function comboboxRowsToCustomSelectOptions(
  rows: ComboboxRow[],
): CustomSelectOption[] {
  return rows.map((row) =>
    row.type === "header"
      ? { type: "header" as const, label: row.label }
      : {
          value: row.value,
          label: row.label,
        },
  );
}

export type { ComboboxRow, ComboboxSelectableRow } from "./comboboxMenuKeyboard";
export {
  comboboxSelectableRows,
  createComboboxMenuKeyDownHandler,
  createComboboxTriggerKeyDownHandler,
  filterComboboxRowsWithHeaders,
  firstComboboxSelectableIndex,
  isComboboxSelectableRow,
  lastComboboxSelectableIndex,
  nextComboboxSelectableIndex,
  prevComboboxSelectableIndex,
  resolveComboboxActiveIndexForOpen,
} from "./comboboxMenuKeyboard";
