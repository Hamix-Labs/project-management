import { type UseQueryResult } from "@tanstack/react-query";
import type { ListCursorModelsResult } from "@/api/settings";
import type { SettingsFormState } from "../settingsForm";
import { SECTION_IDS } from "./sectionIds";
import {
  PhaseFieldGroup,
  PhaseModelField,
  PhasePanel,
  SectionCard,
} from "./settingsSectionLayout";
import type { HandleField } from "./settingsSectionTypes";

/**
 * Phases — execute worker and runner configuration under one section card.
 */
export function PhasesSettingsSection({
  form,
  pickupInvalid,
  parallelismInvalid,
  maxInvalid,
  streamIdleInvalid,
  cursorModelsQuery,
  modelIdsFromList,
  onField,
}: {
  form: SettingsFormState;
  pickupInvalid: boolean;
  parallelismInvalid: boolean;
  maxInvalid: boolean;
  streamIdleInvalid: boolean;
  cursorModelsQuery: UseQueryResult<ListCursorModelsResult, Error>;
  modelIdsFromList: Set<string>;
  onField: HandleField;
}) {
  return (
    <SectionCard id={SECTION_IDS.phases} title="Phases">
      <div className="settings-phases-stack">
        <PhasePanel
          id={SECTION_IDS.agentWorker}
          phase="execute"
          description="Pulls ready tasks and runs the agent to do the work."
        >
          <PhaseFieldGroup title="Worker">
            <label className="settings-field">
              <span className="settings-field-label">Pickup delay</span>
              <span className="settings-field-input-suffix">
                <input
                  type="number"
                  min={0}
                  max={604800}
                  step={1}
                  placeholder="5"
                  value={form.agentPickupDelaySeconds}
                  onChange={(e) =>
                    onField("agentPickupDelaySeconds", e.target.value)
                  }
                  aria-invalid={pickupInvalid}
                />
                <span className="settings-field-suffix" aria-hidden="true">
                  seconds
                </span>
              </span>
            </label>
            <div className="settings-field-help-block">
              <p className="settings-field-help">
                Minimum wait before the next ready task.
              </p>
              <p className="settings-field-help settings-field-help-meta">
                Default <code>5</code>s
              </p>
            </div>
            {pickupInvalid ? (
              <p role="alert" className="settings-field-error">
                Must be between 0 and 604800 (7 days).
              </p>
            ) : null}

            <label className="settings-field">
              <span className="settings-field-label">Max parallel tasks</span>
              <span className="settings-field-input-suffix">
                <input
                  type="number"
                  min={1}
                  step={1}
                  placeholder="150"
                  value={form.agentTaskParallelism}
                  onChange={(e) =>
                    onField("agentTaskParallelism", e.target.value)
                  }
                  aria-invalid={parallelismInvalid}
                />
                <span className="settings-field-suffix" aria-hidden="true">
                  tasks
                </span>
              </span>
            </label>
            <div className="settings-field-help-block">
              <p className="settings-field-help">
                How many tasks can run at once across different worktrees.
                Tasks on the same worktree still run one at a time.
              </p>
              <p className="settings-field-help settings-field-help-meta">
                Default <code>150</code>
              </p>
            </div>
            {parallelismInvalid ? (
              <p role="alert" className="settings-field-error">
                Must be an integer of at least 1.
              </p>
            ) : null}
          </PhaseFieldGroup>

          <PhaseFieldGroup title="Runner">
            <PhaseModelField
              testId="settings-cursor-model-select"
              value={form.cursorModel}
              onChange={(v) => onField("cursorModel", v)}
              query={cursorModelsQuery}
              knownIds={modelIdsFromList}
            />
            <p className="settings-field-help">
              Auto lets cursor-agent choose. Pick a model to pin it for every
              run.
            </p>

            <div id={SECTION_IDS.runTimeout} className="settings-field-block">
              <label className="settings-field">
                <span className="settings-field-label">Max execute duration</span>
                <span className="settings-field-input-suffix">
                  <input
                    type="number"
                    min={0}
                    step={1}
                    value={form.maxRunDurationSeconds}
                    onChange={(e) =>
                      onField("maxRunDurationSeconds", e.target.value)
                    }
                    aria-invalid={maxInvalid}
                  />
                  <span className="settings-field-suffix" aria-hidden="true">
                    seconds
                  </span>
                </span>
              </label>
              <div className="settings-field-help-block">
                <p className="settings-field-help">
                  Cancels the run if it takes longer than this.
                </p>
                <p className="settings-field-help settings-field-help-meta">
                  Default <code>0</code>
                </p>
              </div>
              {maxInvalid ? (
                <p role="alert" className="settings-field-error">
                  Must be a non-negative integer.
                </p>
              ) : null}
            </div>

            <div id={SECTION_IDS.streamIdle} className="settings-field-block">
              <label className="settings-field">
                <span className="settings-field-label">Stream idle timeout</span>
                <span className="settings-field-input-suffix">
                  <input
                    type="number"
                    min={0}
                    step={1}
                    value={form.streamIdleStuckSeconds}
                    onChange={(e) =>
                      onField("streamIdleStuckSeconds", e.target.value)
                    }
                    aria-invalid={streamIdleInvalid}
                  />
                  <span className="settings-field-suffix" aria-hidden="true">
                    seconds
                  </span>
                </span>
              </label>
              <div className="settings-field-help-block">
                <p className="settings-field-help">
                  Cancels the run if the agent emits no stdout for this long
                  after it has started streaming. Distinct from max execute
                  duration.
                </p>
                <p className="settings-field-help settings-field-help-meta">
                  Default <code>900</code>; <code>0</code> disables
                </p>
              </div>
              {streamIdleInvalid ? (
                <p role="alert" className="settings-field-error">
                  Must be a non-negative integer.
                </p>
              ) : null}
            </div>

          </PhaseFieldGroup>
        </PhasePanel>
      </div>
    </SectionCard>
  );
}
