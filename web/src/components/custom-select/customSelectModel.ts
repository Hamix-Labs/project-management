import {
  firstComboboxSelectableIndex,
  lastComboboxSelectableIndex,
  nextComboboxSelectableIndex,
  prevComboboxSelectableIndex,
  type ComboboxRow,
} from "@/components/combobox";

export type CustomSelectOption =
  | { type: "header"; label: string }
  | {
      value: string;
      label: string;
      pillClass?: string;
      /** Visual indent steps for hierarchical lists (e.g. parent task picker). */
      depth?: number;
      /** Short leading label (e.g. Top level / Subtask in parent picker). */
      rowTag?: string;
    };

export function isCustomSelectHeader(
  o: CustomSelectOption,
): o is { type: "header"; label: string } {
  return "type" in o && o.type === "header";
}

function customSelectOptionsToComboboxRows(
  opts: CustomSelectOption[],
): ComboboxRow[] {
  return opts.map((o) =>
    isCustomSelectHeader(o)
      ? { type: "header", label: o.label }
      : { type: "option", value: o.value, label: o.label },
  );
}

export function firstSelectableIndex(opts: CustomSelectOption[]): number {
  if (opts.length === 0) return 0;
  const idx = firstComboboxSelectableIndex(customSelectOptionsToComboboxRows(opts));
  return idx >= 0 ? idx : 0;
}

export function lastSelectableIndex(opts: CustomSelectOption[]): number {
  return lastComboboxSelectableIndex(customSelectOptionsToComboboxRows(opts));
}

export function nextSelectable(
  opts: CustomSelectOption[],
  from: number,
): number {
  return nextComboboxSelectableIndex(customSelectOptionsToComboboxRows(opts), from);
}

export function prevSelectable(
  opts: CustomSelectOption[],
  from: number,
): number {
  return prevComboboxSelectableIndex(customSelectOptionsToComboboxRows(opts), from);
}
