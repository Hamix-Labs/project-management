import type { ComponentPropsWithoutRef } from "react";

type HamixWordmarkProps = ComponentPropsWithoutRef<"span">;

/** Scalable product wordmark — split “Ha” / “mix” for redesigned shell chrome. */
export function HamixWordmark({ className, ...props }: HamixWordmarkProps) {
  return (
    <span className={className} {...props}>
      <span className="hamix-wordmark__ha" aria-hidden="true">
        Ha
      </span>
      <span className="hamix-wordmark__mix" aria-hidden="true">
        mix
      </span>
      <span className="visually-hidden">Hamix</span>
    </span>
  );
}
