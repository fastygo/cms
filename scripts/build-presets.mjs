import { mkdir } from "node:fs/promises";
import { spawnSync } from "node:child_process";
import { join } from "node:path";

const root = process.cwd();
const distDir = join(root, "dist");
const presets = [
  "offline-json-sql",
  "ssh-fixtures",
  "full",
  "headless",
  "playground",
];

await mkdir(distDir, { recursive: true });

for (const preset of presets) {
  const output = join(distDir, binaryName(preset));
  const ldflags = `-X github.com/fastygo/cms/internal/platform/preset.DefaultPreset=${preset}`;
  const result = spawnSync("go", ["build", "-ldflags", ldflags, "-o", output, "./cmd/server"], {
    cwd: root,
    stdio: "inherit",
  });
  if (result.status !== 0) {
    process.exit(result.status ?? 1);
  }
  console.log(`Built ${output}`);
}

function binaryName(preset) {
  const suffix = process.platform === "win32" ? ".exe" : "";
  return `gocms-${preset}${suffix}`;
}
