import { maxTemplateInstantiateCountPerItem } from "@/api";

type QuantityStepperProps = {
  value: number;
  onChange: (value: number) => void;
  min?: number;
  max?: number;
  size?: "sm" | "md";
  ariaLabel?: string;
  disabled?: boolean;
  className?: string;
};

function clamp(n: number, min: number, max: number): number {
  return Math.min(max, Math.max(min, n));
}

export function QuantityStepper({
  value,
  onChange,
  min = 1,
  max = maxTemplateInstantiateCountPerItem,
  size = "md",
  ariaLabel = "Quantity",
  disabled = false,
  className,
}: QuantityStepperProps) {
  const set = (n: number) => onChange(clamp(n, min, max));

  return (
    <div
      className={[
        "quantity-stepper",
        size === "sm" ? "quantity-stepper--sm" : "quantity-stepper--md",
        className,
      ]
        .filter(Boolean)
        .join(" ")}
    >
      <button
        type="button"
        className="quantity-stepper__btn"
        aria-label="Decrease"
        disabled={disabled || value <= min}
        onClick={() => set(value - 1)}
      >
        <span aria-hidden="true">−</span>
      </button>
      <input
        type="text"
        inputMode="numeric"
        className="quantity-stepper__field"
        aria-label={ariaLabel}
        value={value}
        disabled={disabled}
        onChange={(e) => {
          const n = Number.parseInt(e.target.value.replace(/\D/g, ""), 10);
          if (Number.isNaN(n)) {
            onChange(min);
          } else {
            set(n);
          }
        }}
      />
      <button
        type="button"
        className="quantity-stepper__btn"
        aria-label="Increase"
        disabled={disabled || value >= max}
        onClick={() => set(value + 1)}
      >
        <span aria-hidden="true">+</span>
      </button>
    </div>
  );
}
