import { SettingsSelectChevronIcon } from "./settingsSelectIcons";
import { groupModelSelectRows } from "./settingsSelectGroupModel";
import type {
  SettingsSelectOption,
  SettingsSelectRow,
  SettingsSelectProps,
} from "./settingsSelectTypes";
import { useSettingsSelectController } from "./useSettingsSelectController";

export type { SettingsSelectOption, SettingsSelectRow };
export { groupModelSelectRows };

export function SettingsSelect(props: SettingsSelectProps) {
  const {
    rootRef,
    shellRef,
    triggerRef,
    listId,
    open,
    selectedLabel,
    shellClass,
    testId,
    disabled,
    ariaBusy,
    onTriggerClick,
    onTriggerKeyDown,
    panel,
  } = useSettingsSelectController(props);

  return (
    <div ref={rootRef} className="settings-dropdown">
      <div ref={shellRef} className={shellClass}>
        <button
          ref={triggerRef}
          type="button"
          data-testid={testId}
          role="combobox"
          aria-expanded={open}
          aria-controls={open ? listId : undefined}
          aria-busy={ariaBusy || undefined}
          disabled={disabled}
          className="settings-dropdown-trigger"
          onClick={onTriggerClick}
          onKeyDown={onTriggerKeyDown}
        >
          <span className="settings-dropdown-trigger-label">{selectedLabel}</span>
        </button>
        <SettingsSelectChevronIcon open={open} />
      </div>
      {panel}
    </div>
  );
}
