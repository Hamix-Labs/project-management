import {
  useCallback,
  useEffect,
  useLayoutEffect,
  useState,
  type RefObject,
} from "react";
import type { TimezoneComboboxRow, TimezoneDropdownPosition } from "./timezoneComboboxTypes";
import { isTimezoneRowSelected } from "./timezoneComboboxRows";

export function useTimezoneDropdownPosition(
  open: boolean,
  shellRef: RefObject<HTMLDivElement | null>,
) {
  const [pos, setPos] = useState<TimezoneDropdownPosition | null>(null);

  const updatePosition = useCallback(() => {
    const el = shellRef.current;
    if (!el) return;
    const r = el.getBoundingClientRect();
    setPos({ top: r.bottom + 6, left: r.left, width: r.width });
  }, [shellRef]);

  useLayoutEffect(() => {
    if (!open) {
      setPos(null);
      return;
    }
    updatePosition();
    const onMove = () => updatePosition();
    window.addEventListener("scroll", onMove, true);
    window.addEventListener("resize", onMove);
    return () => {
      window.removeEventListener("scroll", onMove, true);
      window.removeEventListener("resize", onMove);
    };
  }, [open, updatePosition]);

  return pos;
}

export function useCloseTimezoneMenuOnOutsideClick(
  open: boolean,
  baseId: string,
  rootRef: RefObject<HTMLDivElement | null>,
  closeMenu: () => void,
) {
  useEffect(() => {
    if (!open) return;
    const onDoc = (e: MouseEvent) => {
      if (rootRef.current?.contains(e.target as Node)) return;
      const panel = document.getElementById(`${baseId}-panel`);
      if (panel?.contains(e.target as Node)) return;
      closeMenu();
    };
    document.addEventListener("mousedown", onDoc);
    return () => document.removeEventListener("mousedown", onDoc);
  }, [open, baseId, rootRef, closeMenu]);
}

export function useTimezoneMenuActiveIndex(
  open: boolean,
  search: string,
  rows: TimezoneComboboxRow[],
  value: string,
  rowCount: number,
) {
  const [activeIndex, setActiveIndex] = useState(0);

  useEffect(() => {
    if (!open) return;
    if (!search.trim()) {
      const idx = rows.findIndex((row) => isTimezoneRowSelected(row, value));
      setActiveIndex(idx >= 0 ? idx : 0);
      return;
    }
    setActiveIndex(0);
  }, [open, search, rows, value]);

  useEffect(() => {
    if (activeIndex >= rowCount) setActiveIndex(Math.max(0, rowCount - 1));
  }, [activeIndex, rowCount]);

  return [activeIndex, setActiveIndex] as const;
}
