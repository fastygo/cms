/**
 * After ui8kit sync-assets, inject public locale menu sync (data-gocms-menu-location)
 * and refresh ui8kit hashed filename + js/manifest.json so Docker and CI match local builds.
 */
import { createHash } from "node:crypto";
import { readFile, readdir, writeFile, rm } from "node:fs/promises";
import { join } from "node:path";

const root = process.cwd();
const injectPath = join(root, "web", "src", "gocms-locale-public-inject.js");
const staticJs = join(root, "web", "static", "js");
const manifestPath = join(staticJs, "manifest.json");

const marker = "syncLocaleMenuRegions";

function shortHash(data) {
  return createHash("sha256").update(data).digest("hex").slice(0, 8);
}

function sriSHA384(data) {
  const hash = createHash("sha384").update(data).digest("base64");
  return "sha384-" + hash;
}

async function main() {
  const inject = await readFile(injectPath, "utf8");
  const ui8Path = join(staticJs, "ui8kit.js");
  let bundle = await readFile(ui8Path, "utf8");
  if (bundle.includes(marker)) {
    console.log("append-gocms-locale-sync: already applied, skipping");
    return;
  }

  const needle = `    if (currentTarget && nextTarget) {
      currentTarget.innerHTML = nextTarget.innerHTML;
    }

    var parsedTitle`;
  if (!bundle.includes(needle)) {
    throw new Error(
      "append-gocms-locale-sync: expected ui8kit locale snippet not found; update needle or sync-assets output",
    );
  }

  const replacement = `    if (currentTarget && nextTarget) {
      currentTarget.innerHTML = nextTarget.innerHTML;
    }

    syncLocaleMenuRegions(parsed);

    var parsedTitle`;

  const anchor = "  function replaceMainContent(button, html) {";
  if (!bundle.includes(anchor)) {
    throw new Error("append-gocms-locale-sync: replaceMainContent anchor not found");
  }

  bundle = bundle.replace(
    anchor,
    inject.trimEnd() + "\n\n  " + anchor.trimStart(),
  );
  bundle = bundle.replace(needle, replacement);

  const hash = shortHash(Buffer.from(bundle, "utf8"));
  const hashedName = `ui8kit.${hash}.js`;
  const hashedPath = join(staticJs, hashedName);

  for (const name of await readdir(staticJs)) {
    if (name === "manifest.json") continue;
    if (/^ui8kit\.[0-9a-f]+\.js$/i.test(name)) {
      await rm(join(staticJs, name), { force: true });
    }
  }

  await writeFile(ui8Path, bundle, "utf8");
  await writeFile(hashedPath, bundle, "utf8");

  const manifest = JSON.parse(await readFile(manifestPath, "utf8"));
  if (!manifest.ui8kit) {
    throw new Error("append-gocms-locale-sync: manifest.json missing ui8kit entry");
  }
  manifest.ui8kit.file = hashedName;
  manifest.ui8kit.sri = sriSHA384(Buffer.from(bundle, "utf8"));
  await writeFile(manifestPath, JSON.stringify(manifest, null, 2) + "\n", "utf8");

  console.log(`append-gocms-locale-sync: wrote ${hashedName} and updated manifest.json`);
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
