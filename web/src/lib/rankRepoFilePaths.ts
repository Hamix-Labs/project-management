/**
 * Quick-open style path ranking: basename subsequence beats path subsequence;
 * earlier / denser matches score higher. Empty query returns paths in index order.
 */
export function rankRepoFilePaths(paths: string[], query: string): string[] {
  const q = query.trim().toLowerCase();
  if (q === "") {
    return paths.slice();
  }

  type Scored = { path: string; score: number };
  const scored: Scored[] = [];
  for (const path of paths) {
    const normalized = path.replace(/\\/g, "/");
    const base = normalized.slice(normalized.lastIndexOf("/") + 1).toLowerCase();
    const full = normalized.toLowerCase();
    const baseScore = subsequenceScore(base, q);
    const pathScore = subsequenceScore(full, q);
    const best = Math.max(
      baseScore > 0 ? baseScore + 1_000 : 0,
      pathScore,
    );
    if (best > 0) {
      scored.push({ path, score: best });
    }
  }
  scored.sort((a, b) => {
    if (b.score !== a.score) return b.score - a.score;
    return a.path.localeCompare(b.path);
  });
  return scored.map((s) => s.path);
}

/** Higher is better; 0 means no match. */
function subsequenceScore(haystack: string, needle: string): number {
  if (needle === "") return 1;
  if (haystack === needle) return 10_000;
  if (haystack.startsWith(needle)) return 5_000 + Math.max(0, 100 - haystack.length);
  if (haystack.includes(needle)) {
    return 2_000 + Math.max(0, 100 - haystack.indexOf(needle));
  }
  let hi = 0;
  let score = 100;
  let consecutive = 0;
  for (let ni = 0; ni < needle.length; ni++) {
    const ch = needle[ni]!;
    const found = haystack.indexOf(ch, hi);
    if (found < 0) return 0;
    if (found === hi) {
      consecutive += 1;
      score += 5 + consecutive;
    } else {
      consecutive = 0;
      score -= found - hi;
    }
    hi = found + 1;
  }
  return Math.max(1, score);
}
