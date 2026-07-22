/**
 * Build a Cursor deep link that opens a local folder (Mac + Windows).
 * Same shape as VS Code `vscode://file/.../` URL handlers.
 */
export function buildCursorOpenFolderUri(fsPath: string): string {
  let path = fsPath.trim().replace(/\\/g, "/");
  if (path === "") {
    throw new Error("path is required");
  }

  const winDrive = /^[A-Za-z]:\//.exec(path);
  if (winDrive) {
    path = path.replace(/\/+$/, "");
    return `cursor://file/${path}/`;
  }

  if (!path.startsWith("/")) {
    path = `/${path}`;
  }
  path = path.replace(/\/+$/, "");
  return `cursor://file${path}/`;
}
