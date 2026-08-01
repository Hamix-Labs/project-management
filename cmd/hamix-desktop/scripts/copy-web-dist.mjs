import { cpSync, mkdirSync, rmSync, existsSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const here = dirname(fileURLToPath(import.meta.url));
const desktopRoot = join(here, "..");
const webDist = join(desktopRoot, "..", "..", "web", "dist");
const outDist = join(desktopRoot, "frontend", "dist");

if (!existsSync(webDist)) {
  console.error("web/dist missing — run npm run build in web/ first");
  process.exit(1);
}

rmSync(outDist, { recursive: true, force: true });
mkdirSync(outDist, { recursive: true });
cpSync(webDist, outDist, { recursive: true });
console.log("copied web/dist -> cmd/hamix-desktop/frontend/dist");
