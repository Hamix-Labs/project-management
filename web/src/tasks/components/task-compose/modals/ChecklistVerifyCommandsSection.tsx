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

function timeoutMode(row: ChecklistVerifyCommandInput): "none" | "seconds" {
  return typeof row.timeout_seconds === "number" && row.timeout_seconds > 0
    ? "seconds"
    : "none";
}

function ChecklistVerifyCommandRow({
  row,
  index,
  readOnly,
  controlsDisabled,
  onUpdate,
  onRemove,
}: ChecklistVerifyCommandRowProps) {
  const mode = timeoutMode(row);
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
      <div
        className="task-checklist-verify-commands__cell task-checklist-verify-commands__cell--timeout"
        role="cell"
      >
        <div className="task-checklist-verify-command-timeout">
          <label
            htmlFor={`checklist-verify-timeout-mode-${index}`}
            className="visually-hidden"
          >
            Timeout mode for command {index + 1}
          </label>
          <select
            id={`checklist-verify-timeout-mode-${index}`}
            className="task-checklist-verify-command-timeout-mode"
            value={mode}
            disabled={controlsDisabled || readOnly}
            onChange={(ev) => {
              if (ev.target.value === "none") {
                onUpdate(index, { timeout_seconds: null });
                return;
              }
              const current =
                typeof row.timeout_seconds === "number" &&
                row.timeout_seconds > 0
                  ? row.timeout_seconds
                  : 120;
              onUpdate(index, { timeout_seconds: current });
            }}
          >
            <option value="none">No timeout</option>
            <option value="seconds">Timeout</option>
          </select>
          {mode === "seconds" ? (
            <>
              <label
                htmlFor={`checklist-verify-timeout-sec-${index}`}
                className="visually-hidden"
              >
                Timeout seconds for command {index + 1}
              </label>
              <input
                id={`checklist-verify-timeout-sec-${index}`}
                className="task-checklist-verify-command-timeout-input"
                type="number"
                min={1}
                step={1}
                value={row.timeout_seconds ?? ""}
                disabled={controlsDisabled || readOnly}
                readOnly={readOnly}
                onChange={(ev) => {
                  const raw = ev.target.value;
                  if (raw === "") {
                    onUpdate(index, { timeout_seconds: null });
                    return;
                  }
                  const n = Number(raw);
                  if (!Number.isFinite(n)) return;
                  onUpdate(index, {
                    timeout_seconds: Math.max(1, Math.floor(n)),
                  });
                }}
                aria-label={`Timeout seconds for command ${index + 1}`}
              />
              <span
                className="task-checklist-verify-command-timeout-unit"
                aria-hidden="true"
              >
                s
              </span>
            </>
          ) : null}
        </div>
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
          className="task-checklist-verify-commands__cell task-checklist-verify-commands__cell--timeout"
          role="columnheader"
        >
          Timeout
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
          Checks the execute agent must run in the repo before claiming done.
          Default is no timeout — unbounded commands run until the cycle is
          cancelled. Set an optional per-command timeout when you want a
          wall-clock kill. Exit code alone does not pass — the agent reports
          claimed_done with evidence via MCP.
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
