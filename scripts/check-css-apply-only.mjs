import { readdirSync, readFileSync, statSync } from "node:fs";
import { join } from "node:path";

const roots = ["web/static/css"];
const bannedDeclaration = /^\s*[a-z-]+\s*:/i;
let failed = false;

const rawCSSExceptions = new Set(["editor.css"]);

for (const file of walk(roots)) {
  if (!file.endsWith(".css")) continue;
  if (file.endsWith("tokens.css") || file.endsWith("input.css")) continue;
  if (rawCSSExceptions.has(file.split(sep()).pop())) continue;
  if (file.includes(`${sep()}ui8kit${sep()}`) || file.endsWith("fonts.css")) continue;
  const lines = readFileSync(file, "utf8").split(/\r?\n/);
  lines.forEach((line, index) => {
    const trimmed = line.trim();
    if (trimmed === "" || trimmed.startsWith("/*") || trimmed.startsWith("*") || trimmed.startsWith("*/")) {
      return;
    }
    if (bannedDeclaration.test(line) && !trimmed.startsWith("@apply")) {
      console.error(`${file}:${index + 1}: app CSS selectors must use @apply only`);
      failed = true;
    }
  });
}

function sep() {
  return process.platform === "win32" ? "\\" : "/";
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
