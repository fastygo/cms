const GOCMS_PLAYGROUND_DB_NAME = "gocms-playground";
const GOCMS_PLAYGROUND_DB_VERSION = 1;
const GOCMS_PLAYGROUND_SNAPSHOTS_STORE = "snapshots";
const GOCMS_PLAYGROUND_MEDIA_METADATA_STORE = "media_metadata";
const GOCMS_PLAYGROUND_MEDIA_BLOB_STORE = "media_blobs";
const GOCMS_PLAYGROUND_SETTINGS_STORE = "settings";
const GOCMS_PLAYGROUND_SNAPSHOT_KEY = "snapshot";
const GOCMS_PLAYGROUND_SETTINGS_KEY_SOURCE = "settings-source";
const GOCMS_PLAYGROUND_SETTINGS_KEY_IMPORTED_AT = "settings-imported-at";
const GOCMS_PLAYGROUND_SETTINGS_KEY_LAUNCH = "settings-launch";
const GOCMS_PLAYGROUND_POST_LIMIT = 10;
const GOCMS_PLAYGROUND_BLUEPRINT_VERSION = "gocms.playground.blueprint.v1";
const GOCMS_PLAYGROUND_LAUNCH_VERSION = "gocms.playground.launch.v1";

const GOCMS_PLAYGROUND_SCHEMA = {
  snapshotVersion: "gocms.playground.v1",
  routes: {
    "/wp-json/wp/v2/posts": [],
    "/wp-json/wp/v2/pages": [],
    "/wp-json/wp/v2/categories": [],
    "/wp-json/wp/v2/tags": [],
    "/wp-json/wp/v2/media": [],
  },
};

const GOCMS_PLAYGROUND_LAUNCH_SCHEMA = {
  blueprintVersion: GOCMS_PLAYGROUND_BLUEPRINT_VERSION,
  launchVersion: GOCMS_PLAYGROUND_LAUNCH_VERSION,
  query: {
    source: "gocms",
    snapshot: "gocms_snapshot",
    blueprint: "gocms_blueprint",
    route: "gocms_route",
    theme: "gocms_theme",
    preset: "gocms_preset",
    demo: "gocms_demo",
    embed: "gocms_embed",
  },
};

function playgroundStoreConfig() {
  return {
    [GOCMS_PLAYGROUND_DB_NAME]: [
      GOCMS_PLAYGROUND_SNAPSHOTS_STORE,
      GOCMS_PLAYGROUND_MEDIA_METADATA_STORE,
      GOCMS_PLAYGROUND_MEDIA_BLOB_STORE,
      GOCMS_PLAYGROUND_SETTINGS_STORE,
    ],
  };
}

function isIndexedDbAvailable() {
  return typeof window !== "undefined" && "indexedDB" in window;
}

function createRequest() {
  if (!isIndexedDbAvailable()) {
    throw new Error("IndexedDB is not available");
  }
}

function normalizeSnapshot(value) {
  if (!value || typeof value !== "object") {
    return null;
  }
  const routes = {};
  const knownRoutes = value.routes || {};
  for (const route of Object.keys(GOCMS_PLAYGROUND_SCHEMA.routes)) {
    routes[route] = knownRoutes[route] || GOCMS_PLAYGROUND_SCHEMA.routes[route];
  }
  const source = value.source || { kind: "wp-json", base_url: "" };
  return {
    snapshot_version: GOCMS_PLAYGROUND_SCHEMA.snapshotVersion,
    source,
    routes,
    settings: Array.isArray(value.settings) ? value.settings : [],
    local: value.local || { media_blobs: "excluded" },
  };
}

async function openPlaygroundDB() {
  createRequest();
  return await new Promise((resolve, reject) => {
    const request = indexedDB.open(
      GOCMS_PLAYGROUND_DB_NAME,
      GOCMS_PLAYGROUND_DB_VERSION,
    );
    request.onerror = () => reject(request.error || new Error("IndexedDB open failed"));
    request.onsuccess = () => resolve(request.result);
    request.onupgradeneeded = () => {
      const database = request.result;
      if (!database.objectStoreNames.contains(GOCMS_PLAYGROUND_SNAPSHOTS_STORE)) {
        database.createObjectStore(GOCMS_PLAYGROUND_SNAPSHOTS_STORE, {
          keyPath: "id",
        });
      }
      if (!database.objectStoreNames.contains(GOCMS_PLAYGROUND_MEDIA_METADATA_STORE)) {
        database.createObjectStore(GOCMS_PLAYGROUND_MEDIA_METADATA_STORE, {
          keyPath: "id",
        });
      }
      if (!database.objectStoreNames.contains(GOCMS_PLAYGROUND_MEDIA_BLOB_STORE)) {
        database.createObjectStore(GOCMS_PLAYGROUND_MEDIA_BLOB_STORE, {
          keyPath: "id",
        });
      }
      if (!database.objectStoreNames.contains(GOCMS_PLAYGROUND_SETTINGS_STORE)) {
        database.createObjectStore(GOCMS_PLAYGROUND_SETTINGS_STORE, {
          keyPath: "key",
        });
      }
    };
  });
}

async function withStore(storeName, mode, callback) {
  const database = await openPlaygroundDB();
  return await new Promise((resolve, reject) => {
    try {
      const tx = database.transaction(storeName, mode);
      const store = tx.objectStore(storeName);
      let pending = true;
      let resultValue;
      const finish = (value, err) => {
        pending = false;
        database.close();
        if (err) {
          reject(err);
        } else {
          resolve(value);
        }
      };
      const wrapped = {
        get: (key) =>
          new Promise((ok, fail) => {
            const req = store.get(key);
            req.onsuccess = () => ok(req.result);
            req.onerror = () => fail(req.error);
          }),
        put: (value) =>
          new Promise((ok, fail) => {
            const req = store.put(value);
            req.onsuccess = () => ok(req.result);
            req.onerror = () => fail(req.error);
          }),
        clear: () =>
          new Promise((ok, fail) => {
            const req = store.clear();
            req.onsuccess = () => ok(true);
            req.onerror = () => fail(req.error);
          }),
        delete: (key) =>
          new Promise((ok, fail) => {
            const req = store.delete(key);
            req.onsuccess = () => ok(true);
            req.onerror = () => fail(req.error);
          }),
        all: () =>
          new Promise((ok, fail) => {
            const req = store.getAll();
            req.onsuccess = () => ok(req.result || []);
            req.onerror = () => fail(req.error);
          }),
      };
      tx.oncomplete = () => {
        if (pending) {
          finish(resultValue, null);
        }
      };
      tx.onerror = () => {
        finish(null, tx.error || new Error("IndexedDB transaction failed"));
      };
      Promise.resolve(callback(wrapped))
        .then((value) => {
          resultValue = value;
        })
        .catch((err) => {
          tx.abort();
          finish(null, err);
        });
    } catch (err) {
      database.close();
      reject(err);
    }
  });
}

async function snapshotFromDB() {
  return await withStore(GOCMS_PLAYGROUND_SNAPSHOTS_STORE, "readonly", async (store) => {
    return await store.get(GOCMS_PLAYGROUND_SNAPSHOT_KEY);
  });
}

async function hasSnapshot() {
  const snapshot = await snapshotFromDB();
  return Boolean(snapshot && snapshot.value);
}

async function saveSnapshot(snapshot) {
  return await withStore(GOCMS_PLAYGROUND_SNAPSHOTS_STORE, "readwrite", async (store) => {
    return await store.put({
      id: GOCMS_PLAYGROUND_SNAPSHOT_KEY,
      value: snapshot,
      updatedAt: new Date().toISOString(),
    });
  });
}

async function readSetting(key) {
  return await withStore(GOCMS_PLAYGROUND_SETTINGS_STORE, "readonly", async (store) => {
    const value = await store.get(key);
    return value ? value.value : null;
  });
}

async function writeSetting(key, value) {
  return await withStore(GOCMS_PLAYGROUND_SETTINGS_STORE, "readwrite", async (store) => {
    return await store.put({
      key,
      value,
      updatedAt: new Date().toISOString(),
    });
  });
}

async function mediaMetadataFromDB() {
  return await withStore(GOCMS_PLAYGROUND_MEDIA_METADATA_STORE, "readonly", async (store) => {
    return await store.all();
  });
}

async function saveMediaBlob(metadata, blob) {
  if (!metadata || !metadata.id) {
    throw new Error("Media metadata requires an id");
  }
  const normalized = {
    id: metadata.id,
    filename: metadata.filename || "",
    mime_type: metadata.mime_type || (blob ? blob.type : ""),
    width: Number(metadata.width || 0),
    height: Number(metadata.height || 0),
    size: Number(metadata.size || (blob ? blob.size : 0)),
    alt: metadata.alt || "",
    caption: metadata.caption || "",
    created_at: metadata.created_at || new Date().toISOString(),
    attached_to: metadata.attached_to || "",
    blob_status: blob ? "local-only" : "missing-local-blob",
  };
  await withStore(GOCMS_PLAYGROUND_MEDIA_METADATA_STORE, "readwrite", async (store) => {
    return await store.put(normalized);
  });
  if (blob) {
    await withStore(GOCMS_PLAYGROUND_MEDIA_BLOB_STORE, "readwrite", async (store) => {
      return await store.put({ id: metadata.id, blob });
    });
  }
  return normalized;
}

function mediaPlaceholder(metadata) {
  const width = Number(metadata && metadata.width ? metadata.width : 160);
  const height = Number(metadata && metadata.height ? metadata.height : 96);
  return {
    filename: metadata && metadata.filename ? metadata.filename : "missing media",
    width,
    height,
    aspectRatio: width > 0 && height > 0 ? width / height : 160 / 96,
    blob_status: "missing-local-blob",
  };
}

function routeForAdminScreen(screen) {
  if (screen === "pages" || screen === "pages-edit") {
    return "/wp-json/wp/v2/pages";
  }
  if (screen === "posts" || screen === "posts-edit") {
    return "/wp-json/wp/v2/posts";
  }
  return "";
}

function routeForCurrentLocation() {
  if (window.location.pathname.indexOf("/go-admin/pages") === 0) {
    return "/wp-json/wp/v2/pages";
  }
  if (window.location.pathname.indexOf("/go-admin/posts") === 0) {
    return "/wp-json/wp/v2/posts";
  }
  return "";
}

function textValue(value) {
  if (typeof value === "string") {
    return value;
  }
  if (value && typeof value.rendered === "string") {
    return value.rendered.replace(/<[^>]*>/g, "").trim();
  }
  if (value == null) {
    return "";
  }
  return String(value);
}

function contentID(item) {
  return item && item.id != null ? String(item.id) : "";
}

function rowFromItem(item, route) {
  const row = document.createElement("tr");
  const slug = item.slug || "";
  const status = item.status || "draft";
  const base = route.indexOf("pages") >= 0 ? "/go-admin/pages/new" : "/go-admin/posts/new";
  const editURL = `${base}?playground_id=${encodeURIComponent(contentID(item))}`;
  for (const value of [textValue(item.title), slug, status, item.author != null ? String(item.author) : ""]) {
    const cell = document.createElement("td");
    cell.textContent = value;
    row.appendChild(cell);
  }
  const actionCell = document.createElement("td");
  const action = document.createElement("a");
  action.href = editURL;
  action.textContent = "Edit";
  action.className = "ui-button ui-button--outline";
  actionCell.appendChild(action);
  row.appendChild(actionCell);
  return row;
}

async function hydrateContentTable(snapshot) {
  const marker = document.querySelector("[data-gocms-screen]");
  const screen = marker ? marker.getAttribute("data-gocms-screen") : "";
  const route = routeForAdminScreen(screen);
  const tableBody = document.querySelector('table[data-gocms-resource="content"] tbody');
  if (!route || !tableBody || !snapshot.routes || !Array.isArray(snapshot.routes[route])) {
    return;
  }
  tableBody.replaceChildren(...snapshot.routes[route].map((item) => rowFromItem(item, route)));
}

function fillEditorFromItem(item) {
  const values = {
    title: textValue(item.title),
    slug: item.slug || "",
    content: textValue(item.content),
    excerpt: textValue(item.excerpt),
    author_id: item.author != null ? String(item.author) : "",
    featured_media_id: item.featured_media != null ? String(item.featured_media) : "",
    status: item.status || "draft",
  };
  for (const [name, value] of Object.entries(values)) {
    const field = document.querySelector(`[name="${name}"]`);
    if (field) {
      field.value = value;
    }
  }
}

async function hydrateContentEditor(snapshot) {
  const route = routeForCurrentLocation();
  const id = new URLSearchParams(window.location.search).get("playground_id");
  if (!route || !id || !snapshot.routes || !Array.isArray(snapshot.routes[route])) {
    return;
  }
  const item = snapshot.routes[route].find((candidate) => contentID(candidate) === id);
  if (item) {
    fillEditorFromItem(item);
  }
}

function itemFromEditor(route, existing) {
  const form = document.querySelector('form[data-gocms-action="save-content"]');
  const data = new FormData(form);
  const now = new Date().toISOString();
  const id = contentID(existing) || data.get("id") || `playground-${Date.now()}`;
  return {
    ...(existing || {}),
    id,
    slug: String(data.get("slug") || ""),
    status: String(data.get("status") || "draft"),
    author: Number(data.get("author_id") || 0),
    featured_media: data.get("featured_media_id") || null,
    date: existing && existing.date ? existing.date : now,
    modified: now,
    type: route.indexOf("pages") >= 0 ? "page" : "post",
    title: { rendered: String(data.get("title") || "") },
    content: { rendered: String(data.get("content") || "") },
    excerpt: { rendered: String(data.get("excerpt") || "") },
  };
}

async function saveEditorToSnapshot(event) {
  const route = routeForCurrentLocation();
  if (!route) {
    return;
  }
  event.preventDefault();
  const snapshotRecord = await snapshotFromDB();
  const snapshot = snapshotRecord && snapshotRecord.value ? snapshotRecord.value : normalizeSnapshot({});
  snapshot.routes = snapshot.routes || {};
  const rows = Array.isArray(snapshot.routes[route]) ? snapshot.routes[route] : [];
  const id = new URLSearchParams(window.location.search).get("playground_id");
  const index = rows.findIndex((item) => contentID(item) === id);
  const nextItem = itemFromEditor(route, index >= 0 ? rows[index] : null);
  const nextRows = rows.slice();
  if (index >= 0) {
    nextRows[index] = nextItem;
  } else {
    nextRows.unshift(nextItem);
  }
  snapshot.routes[route] = nextRows.slice(0, GOCMS_PLAYGROUND_POST_LIMIT);
  await saveSnapshot(snapshot);
  window.alert("Saved to this browser's playground storage.");
  window.location.href = route.indexOf("pages") >= 0 ? "/go-admin/pages" : "/go-admin/posts";
}

async function hydratePlaygroundAdmin() {
  const snapshotRecord = await snapshotFromDB();
  if (!snapshotRecord || !snapshotRecord.value) {
    return;
  }
  await hydrateContentTable(snapshotRecord.value);
  await hydrateContentEditor(snapshotRecord.value);
  const form = document.querySelector('form[data-gocms-action="save-content"]');
  if (form) {
    form.addEventListener("submit", saveEditorToSnapshot);
  }
}

function buildSnapshotFromData(payload) {
  const normalized = normalizeSnapshot(payload);
  normalized.source = normalized.source || {};
  normalized.source.imported_at = normalized.source.imported_at || new Date().toISOString();
  if (!normalized.source.base_url) {
    normalized.source.base_url = "";
  }
  return normalized;
}

async function importFromJson(file) {
  const text = typeof file === "string" ? file : await file.text();
  const parsed = JSON.parse(text);
  const snapshot = buildSnapshotFromData(parsed);
  const database = await openPlaygroundDB();
  database.close();
  await Promise.all([
    saveSnapshot(snapshot),
    writeSetting(GOCMS_PLAYGROUND_SETTINGS_KEY_SOURCE, snapshot.source.base_url || ""),
    writeSetting(GOCMS_PLAYGROUND_SETTINGS_KEY_IMPORTED_AT, snapshot.source.imported_at || new Date().toISOString()),
  ]);
  return snapshot;
}

async function fetchRouteJSON(url) {
  const response = await fetch(url, { headers: { "Accept": "application/json" } });
  if (!response.ok) {
    throw new Error("Unable to load " + url);
  }
  const data = await response.json();
  return data;
}

async function importFromSource(source, options = {}) {
  if (!options.force && await hasSnapshot()) {
    const existing = await snapshotFromDB();
    return existing.value;
  }
  const base = normalizeSource(source);
  const snapshotRoutes = {};
  for (const route of Object.keys(GOCMS_PLAYGROUND_SCHEMA.routes)) {
    const params = new URLSearchParams({ "per_page": String(GOCMS_PLAYGROUND_POST_LIMIT) });
    if (route.indexOf("posts") !== -1) {
      params.set("orderby", "modified");
      params.set("order", "desc");
    }
    const responseURL = base + route + "?" + params.toString();
    const rows = await fetchRouteJSON(responseURL);
    snapshotRoutes[route] = rows;
  }
  const snapshot = {
    snapshot_version: GOCMS_PLAYGROUND_SCHEMA.snapshotVersion,
    source: { kind: "wp-json", base_url: base, imported_at: new Date().toISOString() },
    routes: snapshotRoutes,
    settings: [],
    local: { media_blobs: "excluded" },
  };
  await saveSnapshot(snapshot);
  await writeSetting(GOCMS_PLAYGROUND_SETTINGS_KEY_SOURCE, base);
  await writeSetting(GOCMS_PLAYGROUND_SETTINGS_KEY_IMPORTED_AT, snapshot.source.imported_at);
  return snapshot;
}

function normalizeSource(rawSource) {
  const trimmed = String(rawSource || "").trim();
  if (!trimmed) {
    return "";
  }
  if (trimmed.startsWith("http://") || trimmed.startsWith("https://")) {
    return trimmed.replace(/\/+$/, "");
  }
  return "https://" + trimmed.replace(/\/+$/, "");
}

async function exportToJson() {
  const snapshot = await snapshotFromDB();
  const mediaMetadata = await mediaMetadataFromDB();
  if (!snapshot || !snapshot.value) {
    return {
      snapshot_version: GOCMS_PLAYGROUND_SCHEMA.snapshotVersion,
      source: { kind: "wp-json", base_url: "", imported_at: "" },
      routes: GOCMS_PLAYGROUND_SCHEMA.routes,
      settings: [],
      local: { media_blobs: "excluded", media_metadata: mediaMetadata },
    };
  }
  const current = snapshot.value;
  return {
    snapshot_version: GOCMS_PLAYGROUND_SCHEMA.snapshotVersion,
    source: current.source || { kind: "wp-json", base_url: "", imported_at: "" },
    routes: current.routes || {},
    settings: Array.isArray(current.settings) ? current.settings : [],
    local: { media_blobs: "excluded", media_metadata: mediaMetadata },
  };
}

async function resetPlaygroundStorage() {
  const deleted = await new Promise((resolve, reject) => {
    if (!isIndexedDbAvailable()) {
      reject(new Error("IndexedDB is not available"));
      return;
    }
    const request = indexedDB.deleteDatabase(GOCMS_PLAYGROUND_DB_NAME);
    request.onsuccess = () => resolve(true);
    request.onerror = () => reject(request.error || new Error("IndexedDB delete failed"));
  });
  return deleted;
}

function ensurePlaygroundControls() {
  if (typeof document === "undefined") {
    return;
  }
  document.addEventListener("click", async function (event) {
    const link = event.target.closest("a[href*='playground=']");
    if (!link) {
      return;
    }
    const url = new URL(link.getAttribute("href"), window.location.origin);
    const action = url.searchParams.get("playground");
    if (!action || !url.pathname.startsWith("/go-admin")) {
      return;
    }
    event.preventDefault();
    try {
      if (action === "import-json") {
        await triggerPlaygroundFileImport();
      } else if (action === "export-json") {
        await triggerPlaygroundExport();
      } else if (action === "reset-storage") {
        await triggerPlaygroundReset();
      } else if (action === "import-source" || action === "refresh-source") {
        await triggerPlaygroundSourceImport(url);
      }
    } catch (err) {
      window.alert(err.message || "Playground action failed");
    }
  });
}

async function triggerPlaygroundSourceImport(url) {
  const source = url.searchParams.get("gocms") || new URLSearchParams(window.location.search).get("gocms");
  const sourceValue = source || (await readSetting(GOCMS_PLAYGROUND_SETTINGS_KEY_SOURCE));
  if (!sourceValue) {
    window.alert("No source configured. Use ?gocms=host to set one.");
    return;
  }
  const action = url.searchParams.get("playground");
  const force = action === "refresh-source";
  if (force && !window.confirm("This will replace local playground content from the source.")) {
    return;
  }
  await importFromSource(sourceValue, { force });
  window.alert("Imported content from source.");
}

async function triggerPlaygroundFileImport() {
  const input = document.createElement("input");
  input.type = "file";
  input.accept = "application/json";
  input.addEventListener("change", async () => {
    if (!input.files || input.files.length === 0) {
      return;
    }
    const [file] = input.files;
    await importFromJson(file);
    window.alert("Imported JSON snapshot.");
  }, { once: true });
  input.click();
}

async function triggerPlaygroundExport() {
  const payload = await exportToJson();
  const blob = new Blob([JSON.stringify(payload, null, 2)], { type: "application/json" });
  const href = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = href;
  link.download = "gocms-playground.json";
  link.click();
  URL.revokeObjectURL(href);
}

async function triggerPlaygroundReset() {
  const approved = window.confirm("This will clear local playground content.");
  if (!approved) {
    return;
  }
  await resetPlaygroundStorage();
  window.alert("Playground storage cleared.");
}

async function bootstrapFromQuery() {
  const params = new URLSearchParams(window.location.search);
  const launch = launchOptionsFromQuery(params);
  if (!launch.source_url && !launch.snapshot_url) {
    return;
  }
  const existingSnapshot = await hasSnapshot();
  if (!existingSnapshot) {
    if (launch.snapshot_url) {
      await importFromSnapshotURL(launch.snapshot_url);
    } else if (launch.source_url) {
      const normalized = normalizeSource(launch.source_url);
      if (!normalized) {
        return;
      }
      await importFromSource(normalized);
    }
  }
  await writeSetting(GOCMS_PLAYGROUND_SETTINGS_KEY_LAUNCH, JSON.stringify(launch));
  if (launch.initial_path && window.location.pathname === "/go-admin") {
    const next = new URL(launch.initial_path, window.location.origin);
    if (launch.theme) {
      next.searchParams.set("preview_theme", launch.theme);
    }
    if (launch.preset) {
      next.searchParams.set("preview_preset", launch.preset);
    }
    if (String(next) !== String(window.location)) {
      window.location.href = next.toString();
    }
  }
}

async function importFromSnapshotURL(url) {
  const response = await fetch(url, { headers: { "Accept": "application/json" } });
  if (!response.ok) {
    throw new Error("Unable to load " + url);
  }
  const payload = await response.json();
  const snapshot = normalizeSnapshot(payload);
  if (!snapshot) {
    throw new Error("Invalid playground snapshot payload");
  }
  await saveSnapshot(snapshot);
  return snapshot;
}

function launchOptionsFromQuery(params) {
  return {
    launch_version: GOCMS_PLAYGROUND_LAUNCH_SCHEMA.launchVersion,
    source_url: params.get(GOCMS_PLAYGROUND_LAUNCH_SCHEMA.query.source) || "",
    snapshot_url: params.get(GOCMS_PLAYGROUND_LAUNCH_SCHEMA.query.snapshot) || params.get(GOCMS_PLAYGROUND_LAUNCH_SCHEMA.query.blueprint) || "",
    initial_path: params.get(GOCMS_PLAYGROUND_LAUNCH_SCHEMA.query.route) || "",
    theme: params.get(GOCMS_PLAYGROUND_LAUNCH_SCHEMA.query.theme) || "",
    preset: params.get(GOCMS_PLAYGROUND_LAUNCH_SCHEMA.query.preset) || "",
    demo_mode: asBoolean(params.get(GOCMS_PLAYGROUND_LAUNCH_SCHEMA.query.demo)),
    embedded: asBoolean(params.get(GOCMS_PLAYGROUND_LAUNCH_SCHEMA.query.embed)),
  };
}

function asBoolean(value) {
  const normalized = String(value || "").trim().toLowerCase();
  return normalized === "1" || normalized === "true" || normalized === "yes";
}

if (typeof window !== "undefined") {
  window.GoCMSPlayground = {
    storeConfig: playgroundStoreConfig,
    openPlaygroundDB,
    importFromJson,
    exportToJson,
    importFromSource,
    resetPlaygroundStorage,
    bootstrapFromQuery,
    importFromSnapshotURL,
    launchOptionsFromQuery,
    hasSnapshot,
    saveMediaBlob,
    mediaPlaceholder,
    mediaMetadataFromDB,
    hydratePlaygroundAdmin,
    snapshotFromDB,
    readSetting,
    writeSetting,
    schema: GOCMS_PLAYGROUND_SCHEMA,
    launchSchema: GOCMS_PLAYGROUND_LAUNCH_SCHEMA,
  };
  if (window.location.pathname.startsWith("/go-admin")) {
    const start = async () => {
      ensurePlaygroundControls();
      try {
        await bootstrapFromQuery();
        await hydratePlaygroundAdmin();
      } catch (_) {
        // Ignore bootstrap errors by default; keep UI functional.
      }
    };
    if (document.readyState === "loading") {
      window.addEventListener("DOMContentLoaded", start);
    } else {
      start();
    }
  }
}

