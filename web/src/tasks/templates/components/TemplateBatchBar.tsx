import { QuantityStepper } from "../QuantityStepper";

type TemplateBatchBarProps = {
  selectedCount: number;
  totalTaskCount: number;
  batchDefaultCount: number;
  instantiatePending: boolean;
  onClear: () => void;
  onBatchDefaultCountChange: (count: number) => void;
  onApplyToAll: () => void;
  onCreate: () => void;
};

export function TemplateBatchBar({
  selectedCount,
  totalTaskCount,
  batchDefaultCount,
  instantiatePending,
  onClear,
  onBatchDefaultCountChange,
  onApplyToAll,
  onCreate,
}: TemplateBatchBarProps) {
  if (selectedCount === 0) return null;

  return (
    <div className="template-batch-bar" role="region" aria-label="Batch actions">
      <div className="template-batch-bar__left">
        <span className="template-batch-bar__count-badge" aria-hidden="true">
          {selectedCount}
        </span>
        <span className="template-batch-bar__count-label">selected</span>
        <button type="button" className="template-batch-bar__clear" onClick={onClear}>
          Clear
        </button>
      </div>

      <div className="template-batch-bar__right">
        <span className="template-batch-bar__instances-label">Instances each</span>
        <QuantityStepper
          value={batchDefaultCount}
          ariaLabel="Instances per selected template"
          disabled={instantiatePending}
          onChange={onBatchDefaultCountChange}
        />
        <button
          type="button"
          className="secondary template-batch-bar__apply"
          disabled={instantiatePending}
          onClick={onApplyToAll}
        >
          Apply to all
        </button>
        <span className="template-batch-bar__divider" aria-hidden="true" />
        <button
          type="button"
          className="template-batch-bar__create"
          disabled={instantiatePending}
          onClick={onCreate}
        >
          Create tasks
          <span className="template-batch-bar__create-badge">{totalTaskCount}</span>
        </button>
      </div>
    </div>
  );
}
