import { useEffect, useState } from "react";

export function useDebouncedTrimmedValue(value: string, delayMs: number): string {
  const [debounced, setDebounced] = useState(value.trim());

  useEffect(() => {
    const timer = window.setTimeout(() => setDebounced(value.trim()), delayMs);
    return () => window.clearTimeout(timer);
  }, [value, delayMs]);

  return debounced;
}
