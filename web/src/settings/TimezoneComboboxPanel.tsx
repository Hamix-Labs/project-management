import { createPortal } from "react-dom";
import type { KeyboardEventHandler, LegacyRef, RefObject } from "react";
import {
  SettingsSelectCheckIcon,
} from "./settingsSelectIcons";
import type { TimezoneComboboxRow, TimezoneDropdownPosition } from "./timezoneComboboxTypes";
import {
  isTimezoneRowSelected,
  timezoneRowKey,
  timezoneRowLabel,
} from "./timezoneComboboxRows";

type TimezoneComboboxPanelProps = {
  baseId: string;
  listId: string;
  searchId: string;
  pos: TimezoneDropdownPosition;
  search: string;
  rowCount: number;
  rows: TimezoneComboboxRow[];
  value: string;
  autoLabel: string;
  activeIndex: number;
  searchRef: RefObject<HTMLInputElement | null>;
  listRef: RefObject<HTMLUListElement | null>;
  onSearchChange: (value: string) => void;
  onSearchKeyDown: KeyboardEventHandler<HTMLInputElement>;
  onListKeyDown: KeyboardEventHandler<HTMLUListElement>;
  onActiveIndexChange: (index: number) => void;
  onCommitRow: (row: TimezoneComboboxRow) => void;
};

export function TimezoneComboboxPanel({
  baseId,
  listId,
  searchId,
  pos,
  search,
  rowCount,
  rows,
  value,
  autoLabel,
  activeIndex,
  searchRef,
  listRef,
  onSearchChange,
  onSearchKeyDown,
  onListKeyDown,
  onActiveIndexChange,
  onCommitRow,
}: TimezoneComboboxPanelProps) {
  return createPortal(
    <div
      id={`${baseId}-panel`}
      className="settings-dropdown-panel"
      style={{
        position: "fixed",
        top: pos.top,
        left: pos.left,
        width: pos.width,
        zIndex: "var(--z-portal-popover, 13000)",
      }}
    >
      <div className="settings-dropdown-panel-search">
        <input
          ref={searchRef as LegacyRef<HTMLInputElement>}
          id={searchId}
          type="search"
          className="settings-dropdown-panel-search-input"
          placeholder="Search by city, region, or GMT offset…"
          value={search}
          autoComplete="off"
          spellCheck={false}
          aria-controls={listId}
          aria-autocomplete="list"
          onChange={(e) => onSearchChange(e.target.value)}
          onKeyDown={onSearchKeyDown}
        />
      </div>
      {rowCount > 0 ? (
        <ul
          ref={listRef as LegacyRef<HTMLUListElement>}
          id={listId}
          role="listbox"
          tabIndex={-1}
          className="settings-dropdown-list settings-dropdown-list--portal"
          aria-activedescendant={
            rowCount > 0 ? `${baseId}-opt-${activeIndex}` : undefined
          }
          onKeyDown={onListKeyDown}
        >
          {rows.map((row, idx) => {
            const id = `${baseId}-opt-${idx}`;
            const isActive = idx === activeIndex;
            const isSelected = isTimezoneRowSelected(row, value);
            const text = timezoneRowLabel(row, autoLabel);
            return (
              <li
                key={timezoneRowKey(row)}
                id={id}
                role="option"
                aria-selected={isSelected}
                className={[
                  "settings-dropdown-option",
                  isActive ? "settings-dropdown-option--active" : "",
                  isSelected ? "settings-dropdown-option--selected" : "",
                ]
                  .filter(Boolean)
                  .join(" ")}
                onMouseEnter={() => onActiveIndexChange(idx)}
                onMouseDown={(e) => e.preventDefault()}
                onClick={() => onCommitRow(row)}
              >
                <span className="settings-dropdown-option-check-slot">
                  {isSelected ? <SettingsSelectCheckIcon /> : null}
                </span>
                <span className="settings-dropdown-option-label">{text}</span>
              </li>
            );
          })}
        </ul>
      ) : (
        <div
          className="settings-dropdown-empty settings-dropdown-empty--portal"
          role="status"
        >
          No matching timezones
        </div>
      )}
    </div>,
    document.body,
  );
}
