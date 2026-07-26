import type { TemplateFunctionInputDef, TemplateFunctionInputKind } from "@/types";

type Props = {
  idsPrefix: string;
  inputs: TemplateFunctionInputDef[];
  disabled: boolean;
  onChange: (next: TemplateFunctionInputDef[]) => void;
};

const KINDS: TemplateFunctionInputKind[] = ["dir", "file", "function"];

function slugFromLabel(label: string): string {
  const base = label
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "_")
    .replace(/^_+|_+$/g, "");
  if (!base) return "input";
  if (/^[a-z]/.test(base)) return base.slice(0, 32);
  return `i_${base}`.slice(0, 32);
}

function uniqueId(preferred: string, existing: string[], skipIndex: number): string {
  let id = preferred || "input";
  let n = 2;
  while (existing.some((other, i) => i !== skipIndex && other === id)) {
    id = `${preferred}_${n}`.slice(0, 32);
    n += 1;
  }
  return id;
}

export function TaskCreateModalFunctionInputsField({
  idsPrefix,
  inputs,
  disabled,
  onChange,
}: Props) {
  const updateRow = (index: number, patch: Partial<TemplateFunctionInputDef>) => {
    onChange(
      inputs.map((row, i) => {
        if (i !== index) return row;
        const next = { ...row, ...patch };
        if (patch.label !== undefined && (!row.id || row.id === slugFromLabel(row.label))) {
          next.id = uniqueId(
            slugFromLabel(patch.label),
            inputs.map((r) => r.id),
            index,
          );
        }
        if (next.kind === "function") {
          next.multiple = false;
        }
        return next;
      }),
    );
  };

  return (
    <div className="task-create-function-inputs">
      <p className="hint">
        Optional inputs collected when creating tasks from this template (dirs, files, or
        functions). The agent is prompted to stay within the chosen scope.
      </p>
      {inputs.map((row, index) => (
        <div key={`${idsPrefix}-fn-${index}`} className="task-create-function-inputs__row">
          <label className="field">
            <span className="field-label">Label</span>
            <input
              type="text"
              value={row.label}
              disabled={disabled}
              onChange={(e) => updateRow(index, { label: e.target.value })}
              placeholder="Target directory"
            />
          </label>
          <label className="field">
            <span className="field-label">Kind</span>
            <select
              value={row.kind}
              disabled={disabled}
              onChange={(e) =>
                updateRow(index, { kind: e.target.value as TemplateFunctionInputKind })
              }
            >
              {KINDS.map((k) => (
                <option key={k} value={k}>
                  {k}
                </option>
              ))}
            </select>
          </label>
          <label className="field field--inline">
            <input
              type="checkbox"
              checked={row.required !== false}
              disabled={disabled}
              onChange={(e) => updateRow(index, { required: e.target.checked })}
            />
            <span>Required</span>
          </label>
          {row.kind !== "function" ? (
            <label className="field field--inline">
              <input
                type="checkbox"
                checked={row.multiple === true}
                disabled={disabled}
                onChange={(e) => updateRow(index, { multiple: e.target.checked })}
              />
              <span>Multiple</span>
            </label>
          ) : null}
          <button
            type="button"
            className="secondary"
            disabled={disabled}
            onClick={() => onChange(inputs.filter((_, i) => i !== index))}
          >
            Remove
          </button>
        </div>
      ))}
      <button
        type="button"
        className="secondary"
        disabled={disabled || inputs.length >= 20}
        onClick={() =>
          onChange([
            ...inputs,
            {
              id: uniqueId("scope", inputs.map((r) => r.id), -1),
              kind: "dir",
              label: "",
              required: true,
            },
          ])
        }
      >
        Add input
      </button>
    </div>
  );
}
