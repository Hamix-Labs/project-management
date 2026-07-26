import { useEffect, useId, useState } from "react";
import {
  searchRepoEntries,
  searchRepoSymbols,
  type RepoSearchEntry,
  type RepoSymbolHit,
} from "@/api/repo";
import { useDebouncedTrimmedValue } from "@/hooks/useDebouncedTrimmedValue";
import type {
  TemplateFunctionInputKind,
  TemplateFunctionRef,
} from "@/types";

type Props = {
  kind: TemplateFunctionInputKind;
  worktreeId: string | null;
  multiple: boolean;
  paths: string[];
  functions: TemplateFunctionRef[];
  disabled?: boolean;
  onPathsChange: (paths: string[]) => void;
  onFunctionsChange: (fns: TemplateFunctionRef[]) => void;
};

export function RepoScopePicker({
  kind,
  worktreeId,
  multiple,
  paths,
  functions,
  disabled,
  onPathsChange,
  onFunctionsChange,
}: Props) {
  const listboxId = useId();
  const [query, setQuery] = useState("");
  const debounced = useDebouncedTrimmedValue(query, 250);
  const [entries, setEntries] = useState<RepoSearchEntry[]>([]);
  const [symbols, setSymbols] = useState<RepoSymbolHit[]>([]);
  const [busy, setBusy] = useState(false);
  const [unavailable, setUnavailable] = useState(false);
  const [open, setOpen] = useState(false);

  useEffect(() => {
    if (!worktreeId || disabled) {
      setEntries([]);
      setSymbols([]);
      setUnavailable(!worktreeId);
      return;
    }
    let cancelled = false;
    const run = async () => {
      setBusy(true);
      try {
        if (kind === "function") {
          if (!debounced) {
            if (!cancelled) setSymbols([]);
            return;
          }
          const hits = await searchRepoSymbols(debounced, { worktreeId });
          if (cancelled) return;
          if (hits === null) {
            setUnavailable(true);
            setSymbols([]);
            return;
          }
          setUnavailable(false);
          setSymbols(hits);
          return;
        }
        const kinds =
          kind === "dir" ? { file: false, dir: true } : { file: true, dir: false };
        const found = await searchRepoEntries(debounced, { worktreeId, kinds });
        if (cancelled) return;
        if (found === null) {
          setUnavailable(true);
          setEntries([]);
          return;
        }
        setUnavailable(false);
        setEntries(found.filter((e) => e.kind === kind));
      } finally {
        if (!cancelled) setBusy(false);
      }
    };
    void run();
    return () => {
      cancelled = true;
    };
  }, [debounced, disabled, kind, worktreeId]);

  const pickPath = (path: string) => {
    if (multiple) {
      if (paths.includes(path)) return;
      onPathsChange([...paths, path]);
    } else {
      onPathsChange([path]);
    }
    setQuery("");
    setOpen(false);
  };

  const pickSymbol = (hit: RepoSymbolHit) => {
    const ref: TemplateFunctionRef = {
      path: hit.path,
      name: hit.name,
      line: hit.line,
    };
    if (multiple) {
      onFunctionsChange([...functions, ref]);
    } else {
      onFunctionsChange([ref]);
    }
    setQuery("");
    setOpen(false);
  };

  return (
    <div className="repo-scope-picker">
      {unavailable ? (
        <p className="hint" role="status">
          Repository search is unavailable for this template (no worktree).
        </p>
      ) : null}
      <input
        type="search"
        className="repo-scope-picker__input"
        value={query}
        disabled={disabled || !worktreeId}
        placeholder={
          kind === "function"
            ? "Search functions…"
            : kind === "dir"
              ? "Search directories…"
              : "Search files…"
        }
        aria-controls={listboxId}
        aria-expanded={open}
        onChange={(e) => {
          setQuery(e.target.value);
          setOpen(true);
        }}
        onFocus={() => setOpen(true)}
      />
      {busy ? <p className="hint">Searching…</p> : null}
      {open && kind !== "function" && entries.length > 0 ? (
        <ul id={listboxId} className="repo-scope-picker__list" role="listbox">
          {entries.map((e) => (
            <li key={`${e.kind}:${e.path}`}>
              <button type="button" onMouseDown={(ev) => ev.preventDefault()} onClick={() => pickPath(e.path)}>
                {e.path}
              </button>
            </li>
          ))}
        </ul>
      ) : null}
      {open && kind === "function" && symbols.length > 0 ? (
        <ul id={listboxId} className="repo-scope-picker__list" role="listbox">
          {symbols.map((s) => (
            <li key={`${s.path}:${s.name}:${s.line}`}>
              <button
                type="button"
                onMouseDown={(ev) => ev.preventDefault()}
                onClick={() => pickSymbol(s)}
              >
                {s.name} — {s.path}:{s.line}
              </button>
            </li>
          ))}
        </ul>
      ) : null}
      <div className="repo-scope-picker__chips">
        {kind === "function"
          ? functions.map((f) => (
              <button
                key={`${f.path}:${f.name}:${f.line}`}
                type="button"
                className="repo-scope-picker__chip"
                disabled={disabled}
                onClick={() =>
                  onFunctionsChange(
                    functions.filter(
                      (x) =>
                        !(x.path === f.path && x.name === f.name && x.line === f.line),
                    ),
                  )
                }
              >
                {f.name}@{f.path}:{f.line} ×
              </button>
            ))
          : paths.map((p) => (
              <button
                key={p}
                type="button"
                className="repo-scope-picker__chip"
                disabled={disabled}
                onClick={() => onPathsChange(paths.filter((x) => x !== p))}
              >
                {p} ×
              </button>
            ))}
      </div>
    </div>
  );
}
