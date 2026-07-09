import type { Task } from "@/types";

/** Flat row for list display (depth is always 0 for flat tasks). */
export type TaskWithDepth = Task & { depth: number };

/** Maps root tasks to flat list rows for the home list and pickers. */
export function flattenTaskTreeRoots(nodes: Task[]): TaskWithDepth[] {
  return nodes.map((n) => ({ ...n, depth: 0 }));
}
