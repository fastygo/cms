import { copyFile, cp, mkdir, readFile, rm, writeFile } from "node:fs/promises";
import { spawnSync } from "node:child_process";
import { dirname, join } from "node:path";

const root = process.cwd();
const distDir = join(root, "dist");
const staticRoot = join(root, "web", "static");
const profileRoot = join(root, "web", "static-profiles");
const manifestPath = join(staticRoot, "asset-manifest.json");
const presets = [
  "offline-json-sql",
  "ssh-fixtures",
  "full",
  "headless",
  "playground",
];

const sourceManifest = JSON.parse(await readFile(manifestPath, "utf8"));

await mkdir(distDir, { recursive: true });

for (const preset of presets) {
  const profile = await loadProfile(preset);
  const bundleDir = join(distDir, preset);

  await rm(bundleDir, { recursive: true, force: true });
  await mkdir(bundleDir, { recursive: true });
  buildBinary(preset, join(bundleDir, binaryName(preset)));

  if (profile.static === false) {
    console.log(`Packaged ${preset} without static assets.`);
    continue;
  }

  await packageStatic(profile, join(bundleDir, "web", "static"));
  console.log(`Packaged ${preset} with ${profile.assets.length} logical static assets.`);
}

async function loadProfile(preset) {
  const payload = await readFile(join(profileRoot, `${preset}.json`), "utf8");
  const profile = JSON.parse(payload);
  if (profile.id !== preset) {
    throw new Error(`static profile ${preset} has mismatched id ${profile.id}`);
  }
  return {
    static: true,
    assets: [],
    directories: [],
    ...profile,
  };
}

function buildBinary(preset, output) {
  const ldflags = `-X github.com/fastygo/cms/internal/platform/preset.DefaultPreset=${preset}`;
  const result = spawnSync("go", ["build", "-ldflags", ldflags, "-o", output, "./cmd/server"], {
    cwd: root,
    stdio: "inherit",
  });
  if (result.status !== 0) {
    process.exit(result.status ?? 1);
  }
}

async function packageStatic(profile, targetStaticRoot) {
  const bundleManifest = {};
  await mkdir(targetStaticRoot, { recursive: true });

  for (const logicalPath of profile.assets) {
    const resolvedPath = sourceManifest[logicalPath] ?? logicalPath;
    bundleManifest[logicalPath] = resolvedPath;
    await copyPublicPath(resolvedPath, targetStaticRoot);
  }

  for (const publicDir of profile.directories) {
    await copyPublicDir(publicDir, targetStaticRoot);
  }

  await writeFile(join(targetStaticRoot, "asset-manifest.json"), `${JSON.stringify(bundleManifest, null, 2)}\n`);
}

async function copyPublicPath(publicPath, targetStaticRoot) {
  const relativePath = staticRelativePath(publicPath);
  const source = join(staticRoot, relativePath);
  const target = join(targetStaticRoot, relativePath);
  await mkdir(dirname(target), { recursive: true });
  await copyFile(source, target);
}

async function copyPublicDir(publicDir, targetStaticRoot) {
  const relativePath = staticRelativePath(publicDir);
  await cp(join(staticRoot, relativePath), join(targetStaticRoot, relativePath), {
    recursive: true,
    force: true,
  });
}

function staticRelativePath(publicPath) {
  if (!publicPath.startsWith("/static/")) {
    throw new Error(`static asset must start with /static/: ${publicPath}`);
  }
  return publicPath.slice("/static/".length);
}

function binaryName(preset) {
  const suffix = process.platform === "win32" ? ".exe" : "";
  return `gocms-${preset}${suffix}`;
}
