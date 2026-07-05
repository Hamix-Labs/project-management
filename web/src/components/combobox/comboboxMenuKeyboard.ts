import type { KeyboardEvent } from "react";

export type ComboboxSelectableRow = {
  type: "option";
  value: string;
  label: string;
};

export type ComboboxHeaderRow = {
  type: "header";
  label: string;
};

export type ComboboxRow = ComboboxHeaderRow | ComboboxSelectableRow;

export function isComboboxSelectableRow(
  row: ComboboxRow,
): row is ComboboxSelectableRow {
  return row.type === "option";
}

export function comboboxSelectableRows(rows: ComboboxRow[]): ComboboxSelectableRow[] {
  return rows.filter(isComboboxSelectableRow);
}

export function firstComboboxSelectableIndex(rows: ComboboxRow[]): number {
  return rows.findIndex(isComboboxSelectableRow);
}

export function nextComboboxSelectableIndex(rows: ComboboxRow[], from: number): number {
  for (let i = from + 1; i < rows.length; i++) {
    if (isComboboxSelectableRow(rows[i])) return i;
  }
  for (let i = 0; i < rows.length; i++) {
    if (isComboboxSelectableRow(rows[i])) return i;
  }
  return from;
}

export function prevComboboxSelectableIndex(rows: ComboboxRow[], from: number): number {
  for (let i = from - 1; i >= 0; i--) {
    if (isComboboxSelectableRow(rows[i])) return i;
  }
  for (let i = rows.length - 1; i >= 0; i--) {
    if (isComboboxSelectableRow(rows[i])) return i;
  }
  return from;
}

export function lastComboboxSelectableIndex(rows: ComboboxRow[]): number {
  for (let i = rows.length - 1; i >= 0; i--) {
    if (isComboboxSelectableRow(rows[i])) return i;
  }
  return 0;
}

export type ComboboxMenuKeyboardContext = {
  filteredRows: ComboboxRow[];
  activeIndex: number;
  setActiveIndex: (index: number | ((current: number) => number)) => void;
  closeMenu: () => void;
  commitRow: (row: ComboboxSelectableRow) => void;
};

function handleComboboxMenuArrowNavigation(
  e: KeyboardEvent,
  ctx: ComboboxMenuKeyboardContext,
  direction: "down" | "up",
) {
  e.preventDefault();
  ctx.setActiveIndex((i) =>
    direction === "down"
      ? nextComboboxSelectableIndex(ctx.filteredRows, i)
      : prevComboboxSelectableIndex(ctx.filteredRows, i),
  );
}

function handleComboboxMenuEnterSelection(
  e: KeyboardEvent,
  ctx: ComboboxMenuKeyboardContext,
) {
  if (comboboxSelectableRows(ctx.filteredRows).length === 0) return;
  e.preventDefault();
  const row = ctx.filteredRows[ctx.activeIndex];
  if (row && isComboboxSelectableRow(row)) ctx.commitRow(row);
}

/** Shared listbox keyboard handler for searchable and non-searchable combobox menus. */
export function createComboboxMenuKeyDownHandler(ctx: ComboboxMenuKeyboardContext) {
  return (e: KeyboardEvent<HTMLInputElement | HTMLUListElement>) => {
    if (e.key === "Escape") {
      e.preventDefault();
      ctx.closeMenu();
      return;
    }
    if (e.key === "ArrowDown") {
      handleComboboxMenuArrowNavigation(e, ctx, "down");
      return;
    }
    if (e.key === "ArrowUp") {
      handleComboboxMenuArrowNavigation(e, ctx, "up");
      return;
    }
    if (e.key === "Enter") {
      handleComboboxMenuEnterSelection(e, ctx);
    }
  };
}

export function createComboboxTriggerKeyDownHandler(
  disabled: boolean,
  open: boolean,
  openMenu: () => void,
  closeMenu: () => void,
) {
  return (e: KeyboardEvent<HTMLButtonElement>) => {
    if (disabled) return;
    if (e.key === "ArrowDown" || e.key === "Enter" || e.key === " ") {
      e.preventDefault();
      openMenu();
      return;
    }
    if (e.key === "Escape" && open) {
      e.preventDefault();
      closeMenu();
    }
  };
}

export function filterComboboxRowsWithHeaders(
  baseRows: ComboboxRow[],
  search: string,
): ComboboxRow[] {
  const q = search.trim().toLowerCase();
  if (!q) return baseRows;
  const out: ComboboxRow[] = [];
  let pendingHeader: ComboboxHeaderRow | null = null;
  for (const row of baseRows) {
    if (row.type === "header") {
      pendingHeader = row;
      continue;
    }
    const haystack = `${row.label} ${row.value}`.toLowerCase();
    if (!haystack.includes(q)) continue;
    if (pendingHeader) {
      out.push(pendingHeader);
      pendingHeader = null;
    }
    out.push(row);
  }
  return out;
}

export function resolveComboboxActiveIndexForOpen(
  filteredRows: ComboboxRow[],
  value: string,
  search: string,
): number {
  if (!search.trim()) {
    const idx = filteredRows.findIndex(
      (row) => isComboboxSelectableRow(row) && row.value === value,
    );
    return idx >= 0 ? idx : firstComboboxSelectableIndex(filteredRows);
  }
  return firstComboboxSelectableIndex(filteredRows);
}
