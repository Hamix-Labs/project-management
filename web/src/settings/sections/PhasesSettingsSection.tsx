import { type UseQueryResult } from "@tanstack/react-query";
import type { ListCursorModelsResult } from "@/api/settings";
import type { SettingsFormState } from "../settingsForm";
import { SECTION_IDS } from "./sectionIds";
import {
  PhaseFieldGroup,
  PhaseFlowConnector,
  PhaseModelField,
  PhasePanel,
  SectionCard,
} from "./settingsSectionLayout";
import type { HandleField } from "./settingsSectionTypes";

/**
 * Phases — execute and verify configuration under one section card.
 * Each phase is a nested panel with grouped fields (worker vs runner
 * for execute; runner model for verify).
 */
export function PhasesSettingsSection({
  form,
  pickupInvalid,
  maxInvalid,
  cursorModelsQuery,
  modelIdsFromList,
  onField,
}: {
  form: SettingsFormState;
  pickupInvalid: boolean;
  maxInvalid: boolean;
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

          </PhaseFieldGroup>
        </PhasePanel>

        <PhaseFlowConnector />

        <PhasePanel
          id={SECTION_IDS.verification}
          phase="verify"
          description="Command-only verify runs worker checks and MCP criteria claims after execute (ADR-0090)."
        >
          <PhaseFieldGroup title="Runner">
            <PhaseModelField
              testId="settings-verify-model-select"
              value={form.verifyModel}
              onChange={(v) => onField("verifyModel", v)}
              query={cursorModelsQuery}
              knownIds={modelIdsFromList}
            />
            <p className="settings-field-help">
              Auto inherits the execute model. Pick a model to pin for verify
              only.
            </p>
          </PhaseFieldGroup>
        </PhasePanel>
      </div>
    </SectionCard>
  );
}
