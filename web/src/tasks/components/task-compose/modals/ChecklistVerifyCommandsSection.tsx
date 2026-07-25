import type { ChecklistVerifyCommandInput } from "@/types";
import { MAX_VERIFY_COMMANDS_PER_ITEM } from "@/tasks/task-compose/checklistRequirement";
import type { VerifyCommandHandlers } from "./checklistCriterionModalCopy";
import { verifyCommandsHint } from "./checklistCriterionModalCopy";

type ChecklistVerifyCommandRowProps = {
  row: ChecklistVerifyCommandInput;
  index: number;
  readOnly: boolean;
  controlsDisabled: boolean;
  onUpdate: (
    index: number,
    patch: Partial<ChecklistVerifyCommandInput>,
  ) => void;
  onRemove: (index: number) => void;
};

function ChecklistVerifyCommandRow({
  row,
  index,
  readOnly,
  controlsDisabled,
  onUpdate,
  onRemove,
}: ChecklistVerifyCommandRowProps) {
  return (
    <div
      className="task-checklist-verify-commands__row"
      role="row"
    >
      <div
        className="task-checklist-verify-commands__cell task-checklist-verify-commands__cell--command"
        role="cell"
      >
        <label
          htmlFor={`checklist-verify-cmd-${index}`}
          className="visually-hidden"
        >
          Shell command {index + 1}
        </label>
        <input
          id={`checklist-verify-cmd-${index}`}
          className="task-checklist-verify-command-input"
          value={row.command}
          onChange={(ev) =>
            onUpdate(index, {
              command: ev.target.value,
            })
          }
          placeholder="go test ./pkgs/foo/..."
          disabled={controlsDisabled}
          readOnly={readOnly}
          spellCheck={false}
          autoComplete="off"
        />
      </div>
      <div
        className="task-checklist-verify-commands__cell task-checklist-verify-commands__cell--outcome"
        role="cell"
      >
        <label
          htmlFor={`checklist-verify-outcome-${index}`}
          className="visually-hidden"
        >
          Expected outcome for command {index + 1}
        </label>
        <input
          id={`checklist-verify-outcome-${index}`}
          className="task-checklist-verify-command-outcome-input"
          value={row.expected_outcome ?? ""}
          onChange={(ev) =>
            onUpdate(index, {
              expected_outcome: ev.target.value,
            })
          }
          placeholder="All tests pass"
          disabled={controlsDisabled}
          readOnly={readOnly}
        />
      </div>
      {!readOnly ? (
        <div
          className="task-checklist-verify-commands__cell task-checklist-verify-commands__cell--action"
          role="cell"
        >
          <button
            type="button"
            className="task-checklist-verify-command-card__remove"
            disabled={controlsDisabled}
            aria-label={`Remove command ${index + 1}`}
            onClick={() => onRemove(index)}
          >
            Remove
          </button>
        </div>
      ) : null}
    </div>
  );
}

type ChecklistVerifyCommandsTableProps = {
  verifyCommands: ChecklistVerifyCommandInput[];
  readOnly: boolean;
  controlsDisabled: boolean;
  onUpdate: (
    index: number,
    patch: Partial<ChecklistVerifyCommandInput>,
  ) => void;
  onRemove: (index: number) => void;
};

function ChecklistVerifyCommandsTable({
  verifyCommands,
  readOnly,
  controlsDisabled,
  onUpdate,
  onRemove,
}: ChecklistVerifyCommandsTableProps) {
  if (verifyCommands.length === 0) return null;

  return (
    <div
      className="task-checklist-verify-commands__table"
      role="table"
      aria-label="Verify commands"
    >
      <div
        className="task-checklist-verify-commands__row task-checklist-verify-commands__row--head"
        role="row"
      >
        <span
          className="task-checklist-verify-commands__cell task-checklist-verify-commands__cell--command"
          role="columnheader"
        >
          Shell command
        </span>
        <span
          className="task-checklist-verify-commands__cell task-checklist-verify-commands__cell--outcome"
          role="columnheader"
        >
          Expected outcome
        </span>
        <span
          className="task-checklist-verify-commands__cell task-checklist-verify-commands__cell--action visually-hidden"
          role="columnheader"
        >
          Remove
        </span>
      </div>
      {verifyCommands.map((row, index) => (
        <ChecklistVerifyCommandRow
          key={index}
          row={row}
          index={index}
          readOnly={readOnly}
          controlsDisabled={controlsDisabled}
          onUpdate={onUpdate}
          onRemove={onRemove}
        />
      ))}
    </div>
  );
}

type ChecklistVerifyCommandsSectionProps = {
  readOnly: boolean;
  controlsDisabled: boolean;
  verifyCommands: ChecklistVerifyCommandInput[];
  verifySectionOpen: boolean;
  handlers: VerifyCommandHandlers;
};

export function ChecklistVerifyCommandsSection({
  readOnly,
  controlsDisabled,
  verifyCommands,
  verifySectionOpen,
  handlers,
}: ChecklistVerifyCommandsSectionProps) {
  const {
    updateCommand,
    addCommandRow,
    ensureVerifySectionReady,
    removeCommandRow,
    setVerifySectionOpen,
  } = handlers;

  return (
    <details
      className="task-create-advanced task-checklist-verify-commands"
      open={verifySectionOpen}
      onToggle={(e) => {
        const open = (e.currentTarget as HTMLDetailsElement).open;
        if (readOnly) {
          setVerifySectionOpen(open);
          return;
        }
        ensureVerifySectionReady(open);
      }}
    >
      <summary
        className="task-create-advanced__summary"
        data-testid="checklist-verify-commands-toggle"
      >
        <span
          className="task-create-advanced__chevron"
          aria-hidden="true"
        />
        <span className="task-create-advanced__label">Verify commands</span>
        <span className="task-create-advanced__hint">
          {verifyCommandsHint(verifyCommands.length)}
        </span>
      </summary>
      <div className="task-checklist-verify-commands__body">
        <p className="task-checklist-verify-commands__note">
          Shell commands run in the repo during the verify phase. The same execute
          agent interprets stdout/stderr against each expected outcome — exit code alone
          does not pass the criterion.
        </p>
        <ChecklistVerifyCommandsTable
          verifyCommands={verifyCommands}
          readOnly={readOnly}
          controlsDisabled={controlsDisabled}
          onUpdate={updateCommand}
          onRemove={removeCommandRow}
        />
        {!readOnly ? (
          <button
            type="button"
            className="secondary task-checklist-verify-command-add"
            disabled={
              controlsDisabled ||
              verifyCommands.length >= MAX_VERIFY_COMMANDS_PER_ITEM
            }
            onClick={addCommandRow}
          >
            Add command
          </button>
        ) : null}
      </div>
    </details>
  );
}
