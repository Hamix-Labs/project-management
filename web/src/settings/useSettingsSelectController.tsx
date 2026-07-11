import {
  useCallback,
  useId,
  useMemo,
  useRef,
  useState,
  type KeyboardEvent,
  type ReactNode,
  type RefObject,
} from "react";
import {
  comboboxSelectableRows,
  createComboboxMenuKeyDownHandler,
  createComboboxTriggerKeyDownHandler,
  filterComboboxRowsWithHeaders,
} from "@/components/combobox";
import { SettingsSelectPanel } from "./SettingsSelectPanel";
import {
  useCloseOnOutsideClick,
  useDropdownPanelPosition,
  useFocusMenuOnOpen,
  useSyncActiveIndexOnOpen,
} from "./settingsSelectHooks";
import type {
  SettingsSelectOption,
  SettingsSelectProps,
  SettingsSelectRow,
} from "./settingsSelectTypes";

export type SettingsSelectController = {
  rootRef: RefObject<HTMLDivElement>;
  shellRef: RefObject<HTMLDivElement>;
  triggerRef: RefObject<HTMLButtonElement>;
  baseId: string;
  listId: string;
  open: boolean;
  selectedLabel: string;
  shellClass: string;
  testId: string;
  disabled: boolean;
  ariaBusy: boolean;
  onTriggerClick: () => void;
  onTriggerKeyDown: (e: KeyboardEvent<HTMLButtonElement>) => void;
  panel: ReactNode;
};

export function useSettingsSelectController({
  value,
  onChange,
  options,
  testId,
  disabled = false,
  ariaBusy = false,
  searchable: searchableProp,
  searchPlaceholder = "Search…",
  rows: rowsProp,
}: SettingsSelectProps): SettingsSelectController {
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
  const [activeIndex, setActiveIndex] = useState(0);

  const searchable = searchableProp ?? options.length > 10;

  const baseRows = useMemo(
    (): SettingsSelectRow[] =>
      rowsProp ??
      options.map((o) => ({ type: "option" as const, value: o.value, label: o.label })),
    [options, rowsProp],
  );

  const selectedLabel = useMemo(() => {
    const hit = options.find((o) => o.value === value);
    return hit?.label ?? value;
  }, [options, value]);

  const filteredRows = useMemo(
    () => filterComboboxRowsWithHeaders(baseRows, search),
    [baseRows, search],
  );

  const selectable = useMemo(
    () => comboboxSelectableRows(filteredRows),
    [filteredRows],
  );

  const closeMenu = useCallback(() => {
    setOpen(false);
    setSearch("");
    triggerRef.current?.focus();
  }, []);

  const pos = useDropdownPanelPosition(shellRef, open);

  useCloseOnOutsideClick(open, baseId, rootRef, closeMenu);
  useFocusMenuOnOpen(open, searchable, searchRef, listRef);
  useSyncActiveIndexOnOpen(open, search, filteredRows, value, setActiveIndex);

  const commitOption = useCallback(
    (opt: SettingsSelectOption) => {
      onChange(opt.value);
      closeMenu();
    },
    [closeMenu, onChange],
  );

  const openMenu = useCallback(() => {
    if (disabled) return;
    setOpen(true);
  }, [disabled]);

  const keyboardCtx = useMemo(
    () => ({
      filteredRows,
      activeIndex,
      setActiveIndex,
      closeMenu,
      commitRow: commitOption,
    }),
    [filteredRows, activeIndex, closeMenu, commitOption],
  );

  const onTriggerKeyDown = useMemo(
    () => createComboboxTriggerKeyDownHandler(disabled, open, openMenu, closeMenu),
    [disabled, open, openMenu, closeMenu],
  );
  const onMenuKeyDown = createComboboxMenuKeyDownHandler(keyboardCtx);

  const shellClass = open
    ? "settings-dropdown-shell settings-dropdown-shell--open"
    : "settings-dropdown-shell";

  const panel =
    open && pos ? (
      <SettingsSelectPanel
        baseId={baseId}
        listId={listId}
        searchId={searchId}
        pos={pos}
        searchable={searchable}
        searchPlaceholder={searchPlaceholder}
        search={search}
        setSearch={setSearch}
        searchRef={searchRef}
        listRef={listRef}
        filteredRows={filteredRows}
        selectable={selectable}
        activeIndex={activeIndex}
        value={value}
        setActiveIndex={setActiveIndex}
        commitOption={commitOption}
        onMenuKeyDown={onMenuKeyDown}
      />
    ) : null;

  return {
    rootRef,
    shellRef,
    triggerRef,
    baseId,
    listId,
    open,
    selectedLabel,
    shellClass,
    testId,
    disabled,
    ariaBusy,
    onTriggerClick: () => (open ? closeMenu() : openMenu()),
    onTriggerKeyDown,
    panel,
  };
}
