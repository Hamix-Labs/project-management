import {
  resolveComboboxActiveIndexForOpen,
} from "@/components/combobox";
import {
  useCallback,
  useEffect,
  useLayoutEffect,
  useState,
  type RefObject,
} from "react";
import type { SettingsSelectRow } from "./settingsSelectTypes";

export function useDropdownPanelPosition(
  shellRef: RefObject<HTMLDivElement>,
  open: boolean,
) {
  const [pos, setPos] = useState<{
    top: number;
    left: number;
    width: number;
  } | null>(null);

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

export function useCloseOnOutsideClick(
  open: boolean,
  baseId: string,
  rootRef: RefObject<HTMLDivElement>,
  onClose: () => void,
) {
  useEffect(() => {
    if (!open) return;
    const onDoc = (e: MouseEvent) => {
      if (rootRef.current?.contains(e.target as Node)) return;
      const panel = document.getElementById(`${baseId}-panel`);
      if (panel?.contains(e.target as Node)) return;
      onClose();
    };
    document.addEventListener("mousedown", onDoc);
    return () => document.removeEventListener("mousedown", onDoc);
  }, [open, baseId, rootRef, onClose]);
}

export function useFocusMenuOnOpen(
  open: boolean,
  searchable: boolean,
  searchRef: RefObject<HTMLInputElement>,
  listRef: RefObject<HTMLUListElement>,
) {
  useEffect(() => {
    if (!open) return;
    if (searchable) {
      searchRef.current?.focus();
    } else {
      listRef.current?.focus();
    }
  }, [open, searchable, searchRef, listRef]);
}

export function useSyncActiveIndexOnOpen(
  open: boolean,
  search: string,
  filteredRows: SettingsSelectRow[],
  value: string,
  setActiveIndex: (index: number) => void,
) {
  useEffect(() => {
    if (!open) return;
    setActiveIndex(resolveComboboxActiveIndexForOpen(filteredRows, value, search));
  }, [open, search, filteredRows, value, setActiveIndex]);
}
