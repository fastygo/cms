import { spawn } from "node:child_process";

const preset = process.argv[2];

if (!preset) {
  console.error("Usage: node ./scripts/run-preset.mjs <preset>");
  process.exit(1);
}

const child = spawn("go", ["run", "./cmd/server"], {
  cwd: process.cwd(),
  stdio: "inherit",
  env: {
    ...process.env,
    GOCMS_PRESET: preset,
  },
});

child.on("exit", (code) => {
  process.exit(code ?? 0);
});
