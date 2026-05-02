import { copyFile, mkdir, readdir, readFile, rm, writeFile } from "node:fs/promises";
import { createHash } from "node:crypto";
import { basename, dirname, extname, join, relative } from "node:path";

const root = process.cwd();
const staticRoot = join(root, "web", "static");
const manifestPath = join(staticRoot, "asset-manifest.json");
const cssAssets = ["web/static/css/app.css"];
const ui8kitManifestPath = join(staticRoot, "js", "manifest.json");

function versionedName(filePath, hash) {
  const ext = extname(filePath);
  const name = basename(filePath, ext);
  return `${name}.${hash}${ext}`;
}

async function cleanOldVersionedFiles(filePath, hashLength = 12) {
  const dir = dirname(filePath);
  const ext = extname(filePath);
  const name = basename(filePath, ext);
  const pattern = new RegExp(`^${escapeRegExp(name)}\\.[a-f0-9]{${hashLength}}${escapeRegExp(ext)}$`);

  for (const entry of await readdir(dir)) {
    if (pattern.test(entry)) {
      await rm(join(dir, entry), { force: true });
    }
  }
}

function staticPath(filePath) {
  const relativePath = relative(staticRoot, filePath).replaceAll("\\", "/");
  return `/static/${relativePath}`;
}

function escapeRegExp(value) {
  return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}

const manifest = {};

for (const relativePath of cssAssets) {
  const source = join(root, relativePath);
  const contents = await readFile(source);
  const hash = createHash("sha256").update(contents).digest("hex").slice(0, 12);
  const target = join(dirname(source), versionedName(source, hash));

  await cleanOldVersionedFiles(source);
  await copyFile(source, target);
  manifest[staticPath(source)] = staticPath(target);
}

await cleanOldVersionedFiles(join(staticRoot, "js", "theme.js"));
await cleanOldVersionedFiles(join(staticRoot, "js", "ui8kit.js"));

const ui8kitManifest = JSON.parse(await readFile(ui8kitManifestPath, "utf8"));
if (ui8kitManifest.theme?.file) {
  manifest["/static/js/theme.js"] = `/static/js/${ui8kitManifest.theme.file}`;
}
if (ui8kitManifest.ui8kit?.file) {
  manifest["/static/js/ui8kit.js"] = `/static/js/${ui8kitManifest.ui8kit.file}`;
}

await mkdir(dirname(manifestPath), { recursive: true });
await writeFile(manifestPath, `${JSON.stringify(manifest, null, 2)}\n`);

console.log(`Versioned ${Object.keys(manifest).length} static assets and removed stale versions.`);
