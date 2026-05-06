import { readdirSync, readFileSync, statSync } from "node:fs";
import { join } from "node:path";

const roots = ["internal/site"];
const tagPattern = /(^|[^\w])<\/?[a-z][\w:-]*(\s|>|\/)/;
// Local UI8Kit-style "tag suppliers" owned by GoCMS. They are allowed to emit
// raw HTML so the rest of the application can stay raw-tag-free.
const allowedTagSuppliers = new Set([
  "internal/site/ui/elements/markers.templ",
  "internal/site/ui/elements/head.templ",
  "internal/site/ui/blocks/auth_document.templ",
]);
let failed = false;

for (const file of walk(roots)) {
  if (!file.endsWith(".templ")) continue;
  const normalized = file.replaceAll("\\", "/");
  if (allowedTagSuppliers.has(normalized)) continue;
  const lines = readFileSync(file, "utf8").split(/\r?\n/);
  lines.forEach((line, index) => {
    const trimmed = line.trim();
    if (tagPattern.test(trimmed)) {
      console.error(`${file}:${index + 1}: raw HTML tag is not allowed in GoCMS app templates`);
      failed = true;
    }
  });
}

process.exit(failed ? 1 : 0);

function* walk(paths) {
  for (const path of paths) {
    if (!exists(path)) continue;
    const stats = statSync(path);
    if (stats.isDirectory()) {
      for (const name of readdirSync(path)) {
        yield* walk([join(path, name)]);
      }
    } else {
      yield path;
    }
  }
}

function exists(path) {
  try {
    statSync(path);
    return true;
  } catch {
    return false;
  }
}
