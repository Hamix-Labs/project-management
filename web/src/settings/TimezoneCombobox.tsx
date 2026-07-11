import { SettingsSelectChevronIcon } from "./settingsSelectIcons";
import { onTimezoneTriggerKeyDown } from "./timezoneComboboxKeyboard";
import { TimezoneComboboxPanel } from "./TimezoneComboboxPanel";
import type { TimezoneComboboxProps } from "./timezoneComboboxTypes";
import { useTimezoneComboboxController } from "./useTimezoneComboboxController";

export function TimezoneCombobox(props: TimezoneComboboxProps) {
  const controller = useTimezoneComboboxController(props);

  const panel =
    controller.open && controller.pos ? (
      <TimezoneComboboxPanel
        baseId={controller.baseId}
        listId={controller.listId}
        searchId={controller.searchId}
        pos={controller.pos}
        search={controller.search}
        rowCount={controller.rowCount}
        rows={controller.rows}
        value={controller.value}
        autoLabel={controller.autoLabel}
        activeIndex={controller.activeIndex}
        searchRef={controller.searchRef}
        listRef={controller.listRef}
        onSearchChange={controller.setSearch}
        onSearchKeyDown={controller.listKeyDown}
        onListKeyDown={controller.listKeyDown}
        onActiveIndexChange={controller.setActiveIndex}
        onCommitRow={controller.commitRow}
      />
    ) : null;

  return (
    <div ref={controller.rootRef} className="settings-dropdown">
      <div ref={controller.shellRef} className={controller.shellClass}>
        <button
          ref={controller.triggerRef}
          type="button"
          data-testid={controller.testId}
          role="combobox"
          aria-expanded={controller.open}
          aria-controls={controller.open ? controller.listId : undefined}
          className="settings-dropdown-trigger"
          onClick={() =>
            controller.open ? controller.closeMenu() : controller.openMenu()
          }
          onKeyDown={(e) =>
            onTimezoneTriggerKeyDown(
              e,
              controller.open,
              controller.openMenu,
              controller.closeMenu,
            )
          }
        >
          <span className="settings-dropdown-trigger-label">
            {controller.selectedLabel}
          </span>
        </button>
        <SettingsSelectChevronIcon open={controller.open} />
      </div>
      {panel}
    </div>
  );
}
