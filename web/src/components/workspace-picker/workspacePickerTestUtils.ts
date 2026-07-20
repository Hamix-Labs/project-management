export function jsonResponse(body: unknown): Response {
  return new Response(JSON.stringify(body), {
    status: 200,
    headers: { "content-type": "application/json" },
  });
}

export type BrowseFixture = {
  path: string;
  parent_path: string;
  is_git_repo?: boolean;
  entries: Array<{
    name: string;
    path: string;
    has_children: boolean;
    is_git_repo: boolean;
  }>;
};

export function browseRouter(fixtures: Record<string, BrowseFixture>) {
  return (rawUrl: string): Response => {
    const u = new URL(rawUrl, "http://test.local");
    const p = u.searchParams.get("path") ?? "";
    const fx = fixtures[p];
    if (!fx) {
      return new Response("not found", { status: 404 });
    }
    return jsonResponse(fx);
  };
}
