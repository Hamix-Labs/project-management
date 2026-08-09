import uFuzzy from "@leeoniya/ufuzzy";

/**
 * Splitting the needle on non-alphanumerics lets "web/src main" and
 * "websrcmain" both reach `web/src/main.tsx`, which is how people type a path
 * fragment. `intraIns: 1` forgives a single typo inside a term.
 */
const fuzzy = new uFuzzy({ intraIns: 1 });

function basenameStart(path: string): number {
  return path.lastIndexOf("/") + 1;
}

/** A path is "hidden" when any segment starts with a dot. */
function isDotPath(path: string): boolean {
  return path.startsWith(".") || path.includes("/.");
}

function depth(path: string): number {
  let count = 0;
  for (let i = 0; i < path.length; i += 1) {
    if (path[i] === "/") count += 1;
  }
  return count;
}

/**
 * Order for an empty query, where there is nothing to score against.
 *
 * Walk order used to put `.codegraph/` and `.cursor/` at the top and never
 * reach `web/src` inside the result cap, so dot-paths go last, then shallower
 * paths, then alphabetical. Stable for a given list, so callers memoize it.
 */
export function browseOrderPaths(paths: readonly string[]): string[] {
  return [...paths].sort((a, b) => {
    const dotA = isDotPath(a);
    const dotB = isDotPath(b);
    if (dotA !== dotB) return dotA ? 1 : -1;
    const depthDelta = depth(a) - depth(b);
    if (depthDelta !== 0) return depthDelta;
    return a.localeCompare(b);
  });
}

/**
 * Ranks paths against a query, best first, returning every match.
 *
 * uFuzzy scores match quality; on top of that, a match landing inside the file
 * name beats one that only matched directory segments, and dot-paths sink —
 * typing "test" should surface `taskTest.ts` before `.cursor/rules/test.mdc`.
 */
export function rankMentionPaths(
  paths: readonly string[],
  query: string,
): string[] {
  const needle = query.trim();
  if (needle === "") return browseOrderPaths(paths);

  const haystack = paths as string[];
  const idxs = fuzzy.filter(haystack, needle);
  if (!idxs || idxs.length === 0) return [];

  const info = fuzzy.info(idxs, haystack, needle);
  const order = fuzzy.sort(info, haystack, needle);

  const inName: string[] = [];
  const inPath: string[] = [];
  const hidden: string[] = [];
  for (const position of order) {
    const path = haystack[info.idx[position]];
    if (path === undefined) continue;
    if (isDotPath(path)) {
      hidden.push(path);
      continue;
    }
    const firstMatch = info.ranges[position]?.[0] ?? 0;
    if (firstMatch >= basenameStart(path)) {
      inName.push(path);
    } else {
      inPath.push(path);
    }
  }
  return [...inName, ...inPath, ...hidden];
}
