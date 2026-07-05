import { apiErrorFromResponse, fetchWithTimeout } from "./shared";

export async function fetchNamedEntityJson<T>(
  path: string,
  init: RequestInit | undefined,
  parse: (raw: unknown) => T,
): Promise<T> {
  const res = await fetchWithTimeout(path, init);
  if (!res.ok) throw await apiErrorFromResponse(res);
  const raw: unknown = await res.json();
  return parse(raw);
}

export async function fetchNamedEntityVoid(path: string, init?: RequestInit): Promise<void> {
  const res = await fetchWithTimeout(path, init);
  if (!res.ok) throw await apiErrorFromResponse(res);
}
