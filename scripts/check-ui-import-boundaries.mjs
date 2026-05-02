import { readdirSync, readFileSync, statSync } from "node:fs";
import { join } from "node:path";

const checks = [
  {
    root: "internal/site/ui/elements",
    denied: ["github.com/fastygo/ui8kit/components"],
    message: "elements must not import UI8Kit composites",
  },
  {
    root: "internal/site/views",
    denied: ["github.com/fastygo/ui8kit/components"],
    message: "views must use local blocks/elements instead of direct UI8Kit composites",
  },
];

let failed = false;

for (const check of checks) {
  for (const file of walk([check.root])) {
    if (!file.endsWith(".go") && !file.endsWith(".templ")) continue;
    const content = readFileSync(file, "utf8");
    for (const denied of check.denied) {
      if (content.includes(denied)) {
        console.error(`${file}: ${check.message}: ${denied}`);
        failed = true;
      }
    }
  }
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
