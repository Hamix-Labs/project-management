import {
  useCallback,
  useEffect,
  useId,
  useMemo,
  useRef,
  useState,
} from "react";
import {
  filterTimezoneSelectOptions,
  formatTimezoneMenuLabel,
} from "@/shared/time/appTimezone";
import { createTimezoneListKeyDownHandler } from "./timezoneComboboxKeyboard";
import {
  useCloseTimezoneMenuOnOutsideClick,
  useTimezoneDropdownPosition,
  useTimezoneMenuActiveIndex,
} from "./timezoneComboboxHooks";
import {
  buildTimezoneComboboxRows,
  resolveTimezoneSelectedLabel,
  timezoneRowValue,
} from "./timezoneComboboxRows";
import type { TimezoneComboboxProps } from "./timezoneComboboxTypes";

export function useTimezoneComboboxController(props: TimezoneComboboxProps) {
  const {
    value,
    onChange,
    browserTz,
    options,
    customSaved,
    testId = "settings-display-timezone-select",
  } = props;

  const baseId = useId();
  const listId = `${baseId}-list`;
  const searchId = `${baseId}-search`;
  const rootRef = useRef<HTMLDivElement>(null);
  const shellRef = useRef<HTMLDivElement>(null);
  const triggerRef = useRef<HTMLButtonElement>(null);
  const searchRef = useRef<HTMLInputElement>(null);
  const listRef = useRef<HTMLUListElement>(null);

  const [open, setOpen] = useState(false);
  const [search, setSearch] = useState("");

  const autoLabel = useMemo(
    () => `Auto-detect — ${formatTimezoneMenuLabel(browserTz)}`,
    [browserTz],
  );

  const autoHaystack = useMemo(
    () =>
      `auto auto-detect detect browser ${browserTz} ${formatTimezoneMenuLabel(browserTz)}`
        .toLowerCase(),
    [browserTz],
  );

  const selectedLabel = useMemo(
    () => resolveTimezoneSelectedLabel(value, autoLabel, options, customSaved),
    [value, autoLabel, options, customSaved],
  );

  const filteredIana = useMemo(
    () => filterTimezoneSelectOptions(options, search),
    [options, search],
  );

  const rows = useMemo(
    () => buildTimezoneComboboxRows(search, autoHaystack, filteredIana, customSaved),
    [search, autoHaystack, filteredIana, customSaved],
  );

  const rowCount = rows.length;
  const pos = useTimezoneDropdownPosition(open, shellRef);

  const closeMenu = useCallback(() => {
    setOpen(false);
    setSearch("");
    triggerRef.current?.focus();
  }, []);

  const commitRow = useCallback(
    (row: Parameters<typeof timezoneRowValue>[0]) => {
      onChange(timezoneRowValue(row));
      closeMenu();
    },
    [closeMenu, onChange],
  );

  const openMenu = useCallback(() => {
    setOpen(true);
  }, []);

  useCloseTimezoneMenuOnOutsideClick(open, baseId, rootRef, closeMenu);

  useEffect(() => {
    if (!open) return;
    searchRef.current?.focus();
  }, [open]);

  const [activeIndex, setActiveIndex] = useTimezoneMenuActiveIndex(
    open,
    search,
    rows,
    value,
    rowCount,
  );

  const listKeyDown = createTimezoneListKeyDownHandler({
    rowCount,
    rows,
    activeIndex,
    setActiveIndex,
    closeMenu,
    commitRow,
  });

  const shellClass = open
    ? "settings-dropdown-shell settings-dropdown-shell--open"
    : "settings-dropdown-shell";

  return {
    testId,
    rootRef,
    shellRef,
    triggerRef,
    listId,
    open,
    search,
    selectedLabel,
    shellClass,
    pos,
    rowCount,
    rows,
    value,
    autoLabel,
    activeIndex,
    baseId,
    searchId,
    searchRef,
    listRef,
    setSearch,
    setActiveIndex,
    openMenu,
    closeMenu,
    commitRow,
    listKeyDown,
  };
}
