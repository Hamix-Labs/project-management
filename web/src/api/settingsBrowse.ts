import { fetchWithTimeout, apiErrorFromResponse } from "./shared";
import { isRecord, parseNonEmptyString, parseString } from "./parseTaskApiCore";

export type WorkspaceBrowseCategory =
  | "registered"
  | "install"
  | "home"
  | "documents"
  | "desktop"
  | "downloads"
  | "pictures"
  | "music"
  | "videos"
  | "custom";

export type WorkspaceBrowseRoot = {
  id: string;
  path: string;
  label: string;
  category?: WorkspaceBrowseCategory;
  available: boolean;
  unavailable_reason?: string;
};

export type WorkspaceRootsResponse = {
  roots: WorkspaceBrowseRoot[];
  environment: "native";
};

export type BrowseDirEntry = {
  name: string;
  path: string;
  has_children: boolean;
  is_git_repo: boolean;
};

export type BrowseDirsResponse = {
  path?: string;
  parent_path?: string;
  /** True when the listed path itself is a git checkout root. */
  is_git_repo?: boolean;
  entries: BrowseDirEntry[];
};

const BROWSE_CATEGORIES: readonly WorkspaceBrowseCategory[] = [
  "registered",
  "install",
  "home",
  "documents",
  "desktop",
  "downloads",
  "pictures",
  "music",
  "videos",
  "custom",
];

function parseBrowseCategory(
  raw: unknown,
  field: string,
): WorkspaceBrowseCategory | undefined {
  if (raw === undefined || raw === null || raw === "") {
    return undefined;
  }
  if (typeof raw !== "string") {
    throw new Error(`Invalid API response: ${field} must be a string`);
  }
  if (!(BROWSE_CATEGORIES as readonly string[]).includes(raw)) {
    throw new Error(
      `Invalid API response: ${field} must be a known browse category`,
    );
  }
  return raw as WorkspaceBrowseCategory;
}

function parseBrowseRoot(raw: unknown, path: string): WorkspaceBrowseRoot {
  if (!isRecord(raw)) {
    throw new Error(`Invalid API response: ${path} must be an object`);
  }
  const out: WorkspaceBrowseRoot = {
    id: parseNonEmptyString(raw.id, `${path}.id`),
    path: parseNonEmptyString(raw.path, `${path}.path`),
    label: parseNonEmptyString(raw.label, `${path}.label`),
    available: raw.available === true,
  };
  const category = parseBrowseCategory(raw.category, `${path}.category`);
  if (category !== undefined) {
    out.category = category;
  }
  if (raw.unavailable_reason !== undefined && raw.unavailable_reason !== null) {
    out.unavailable_reason = parseString(
      raw.unavailable_reason,
      `${path}.unavailable_reason`,
    );
  }
  return out;
}

function parseBrowseDirEntry(raw: unknown, path: string): BrowseDirEntry {
  if (!isRecord(raw)) {
    throw new Error(`Invalid API response: ${path} must be an object`);
  }
  return {
    name: parseNonEmptyString(raw.name, `${path}.name`),
    path: parseNonEmptyString(raw.path, `${path}.path`),
    has_children: raw.has_children === true,
    is_git_repo: raw.is_git_repo === true,
  };
}

export function parseWorkspaceRootsResponse(raw: unknown): WorkspaceRootsResponse {
  if (!isRecord(raw)) {
    throw new Error("Invalid API response: workspace roots must be an object");
  }
  const rootsRaw = raw.roots;
  if (!Array.isArray(rootsRaw)) {
    throw new Error("Invalid API response: workspace roots missing roots array");
  }
  if (raw.environment !== "native") {
    throw new Error(
      "Invalid API response: workspace roots environment must be native",
    );
  }
  return {
    roots: rootsRaw.map((item, i) => parseBrowseRoot(item, `roots[${i}]`)),
    environment: "native",
  };
}

export function parseBrowseDirsResponse(raw: unknown): BrowseDirsResponse {
  if (!isRecord(raw)) {
    throw new Error("Invalid API response: browse dirs must be an object");
  }
  const entriesRaw = raw.entries;
  if (!Array.isArray(entriesRaw)) {
    throw new Error("Invalid API response: browse dirs missing entries array");
  }
  const out: BrowseDirsResponse = {
    is_git_repo: raw.is_git_repo === true,
    entries: entriesRaw.map((item, i) =>
      parseBrowseDirEntry(item, `entries[${i}]`),
    ),
  };
  if (raw.path !== undefined && raw.path !== null) {
    out.path = parseString(raw.path, "path");
  }
  if (raw.parent_path !== undefined && raw.parent_path !== null) {
    out.parent_path = parseString(raw.parent_path, "parent_path");
  }
  return out;
}

export type WorkspaceRootsScope = "default" | "expanded";

export type FetchWorkspaceRootsOptions = {
  scope?: WorkspaceRootsScope;
  signal?: AbortSignal;
};

export async function fetchWorkspaceRoots(
  options?: FetchWorkspaceRootsOptions,
): Promise<WorkspaceRootsResponse> {
  const params = new URLSearchParams();
  if (options?.scope === "expanded") {
    params.set("scope", "expanded");
  }
  const qs = params.toString();
  const url = qs ? `/settings/workspace-roots?${qs}` : "/settings/workspace-roots";
  const res = await fetchWithTimeout(url, {
    signal: options?.signal,
    headers: { Accept: "application/json" },
  });
  if (!res.ok) {
    throw await apiErrorFromResponse(res);
  }
  return parseWorkspaceRootsResponse(await res.json());
}

export async function browseWorkspaceDirs(
  path?: string,
  init?: RequestInit,
): Promise<BrowseDirsResponse> {
  const params = new URLSearchParams();
  if (path && path.trim() !== "") {
    params.set("path", path);
  }
  const qs = params.toString();
  const url = qs ? `/settings/browse-dirs?${qs}` : "/settings/browse-dirs";
  const res = await fetchWithTimeout(url, {
    ...init,
    headers: { Accept: "application/json", ...(init?.headers ?? {}) },
  });
  if (!res.ok) {
    throw await apiErrorFromResponse(res);
  }
  return parseBrowseDirsResponse(await res.json());
}
