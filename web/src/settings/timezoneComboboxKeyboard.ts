import type { KeyboardEvent, KeyboardEventHandler } from "react";
import type { TimezoneComboboxRow } from "./timezoneComboboxTypes";

export function createTimezoneListKeyDownHandler(params: {
  rowCount: number;
  rows: TimezoneComboboxRow[];
  activeIndex: number;
  setActiveIndex: (updater: (index: number) => number) => void;
  closeMenu: () => void;
  commitRow: (row: TimezoneComboboxRow) => void;
}): KeyboardEventHandler<HTMLInputElement | HTMLUListElement> {
  const { rowCount, rows, activeIndex, setActiveIndex, closeMenu, commitRow } =
    params;
  return (e) => {
    if (e.key === "Escape") {
      e.preventDefault();
      closeMenu();
      return;
    }
    if (e.key === "ArrowDown" && rowCount > 0) {
      e.preventDefault();
      setActiveIndex((i) => Math.min(i + 1, rowCount - 1));
      return;
    }
    if (e.key === "ArrowUp" && rowCount > 0) {
      e.preventDefault();
      setActiveIndex((i) => Math.max(i - 1, 0));
      return;
    }
    if (e.key === "Enter" && rowCount > 0) {
      e.preventDefault();
      const row = rows[activeIndex];
      if (row) commitRow(row);
    }
  };
}

export function onTimezoneTriggerKeyDown(
  e: KeyboardEvent<HTMLButtonElement>,
  open: boolean,
  openMenu: () => void,
  closeMenu: () => void,
) {
  if (e.key === "ArrowDown" || e.key === "Enter" || e.key === " ") {
    e.preventDefault();
    openMenu();
    return;
  }
  if (e.key === "Escape" && open) {
    e.preventDefault();
    closeMenu();
  }
}
