import { createPortal } from "react-dom";
import type { KeyboardEvent, RefObject } from "react";
import { isComboboxSelectableRow } from "@/components/combobox";
import type { SettingsSelectOption, SettingsSelectRow } from "./settingsSelectTypes";
import {
  SettingsSelectCheckIcon,
} from "./settingsSelectIcons";

type SettingsSelectPanelProps = {
  baseId: string;
  listId: string;
  searchId: string;
  pos: { top: number; left: number; width: number };
  searchable: boolean;
  searchPlaceholder: string;
  search: string;
  setSearch: (value: string) => void;
  searchRef: RefObject<HTMLInputElement>;
  listRef: RefObject<HTMLUListElement>;
  filteredRows: SettingsSelectRow[];
  selectable: SettingsSelectOption[];
  activeIndex: number;
  value: string;
  setActiveIndex: (index: number) => void;
  commitOption: (opt: SettingsSelectOption) => void;
  onMenuKeyDown: (e: KeyboardEvent<HTMLInputElement | HTMLUListElement>) => void;
};

export function SettingsSelectPanel({
  baseId,
  listId,
  searchId,
  pos,
  searchable,
  searchPlaceholder,
  search,
  setSearch,
  searchRef,
  listRef,
  filteredRows,
  selectable,
  activeIndex,
  value,
  setActiveIndex,
  commitOption,
  onMenuKeyDown,
}: SettingsSelectPanelProps) {
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
      {searchable ? (
        <div className="settings-dropdown-panel-search">
          <input
            ref={searchRef}
            id={searchId}
            type="search"
            className="settings-dropdown-panel-search-input"
            placeholder={searchPlaceholder}
            value={search}
            autoComplete="off"
            spellCheck={false}
            aria-controls={listId}
            aria-autocomplete="list"
            onChange={(e) => setSearch(e.target.value)}
            onKeyDown={onMenuKeyDown}
          />
        </div>
      ) : null}
      {selectable.length > 0 ? (
        <ul
          ref={listRef}
          id={listId}
          role="listbox"
          tabIndex={searchable ? -1 : 0}
          className="settings-dropdown-list settings-dropdown-list--portal"
          aria-activedescendant={
            filteredRows[activeIndex] && isComboboxSelectableRow(filteredRows[activeIndex])
              ? `${baseId}-opt-${activeIndex}`
              : undefined
          }
          onKeyDown={onMenuKeyDown}
        >
          {filteredRows.map((row, idx) => {
            if (row.type === "header") {
              return (
                <li
                  key={`header-${row.label}-${idx}`}
                  role="presentation"
                  className="settings-dropdown-option-header"
                >
                  {row.label}
                </li>
              );
            }
            const id = `${baseId}-opt-${idx}`;
            const isActive = idx === activeIndex;
            const isSelected = row.value === value;
            return (
              <li
                key={`${row.value}-${row.label}`}
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
                onMouseEnter={() => setActiveIndex(idx)}
                onMouseDown={(e) => e.preventDefault()}
                onClick={() => commitOption(row)}
              >
                <span className="settings-dropdown-option-check-slot">
                  {isSelected ? <SettingsSelectCheckIcon /> : null}
                </span>
                <span className="settings-dropdown-option-label">{row.label}</span>
              </li>
            );
          })}
        </ul>
      ) : (
        <div
          className="settings-dropdown-empty settings-dropdown-empty--portal"
          role="status"
        >
          No matches
        </div>
      )}
    </div>,
    document.body,
  );
}
