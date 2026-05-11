function gocmsPluginAction() {
  const params = new URLSearchParams(window.location.search);
  return params.get("plugin-action") || "";
}

function clearPluginAction() {
  const url = new URL(window.location.href);
  url.searchParams.delete("plugin-action");
  window.history.replaceState({}, "", url.pathname + url.search);
}

function isAdminChrome() {
  return window.location.pathname.includes("/go-admin");
}

function readPluginActionFromAnchor(anchor) {
  try {
    const href = anchor.getAttribute("href");
    if (!href) {
      return "";
    }
    return new URL(href, window.location.href).searchParams.get("plugin-action") || "";
  } catch {
    return "";
  }
}

async function postPlugin(url, body) {
  const response = await fetch(url, {
    method: "POST",
    body,
    credentials: "same-origin",
    headers: body instanceof FormData ? undefined : { Accept: "application/json" },
  });
  if (!response.ok) {
    const message = await response.text();
    throw new Error(message || "Plugin action failed");
  }
}

async function importSnapshotFile() {
  return await new Promise((resolve, reject) => {
    const input = document.createElement("input");
    input.type = "file";
    input.accept = "application/json";
    input.addEventListener(
      "change",
      async () => {
        try {
          const file = input.files && input.files[0];
          if (!file) {
            resolve(false);
            return;
          }
          const form = new FormData();
          form.append("snapshot", file);
          await postPlugin("/go-admin/plugins/json-import-export/import", form);
          resolve(true);
        } catch (error) {
          reject(error);
        }
      },
      { once: true },
    );
    input.click();
  });
}

async function runPluginAction(action) {
  if (action === "json-import-export.import") {
    const imported = await importSnapshotFile();
    if (imported) {
      window.alert("JSON snapshot imported.");
    }
    return;
  }
  if (action === "json-import-export.export-site-package") {
    await postPlugin("/go-admin/plugins/json-import-export/export-site-package");
    window.alert("Site package exported.");
    return;
  }
  if (action === "json-import-export.import-site-package") {
    await postPlugin("/go-admin/plugins/json-import-export/import-site-package");
    window.alert("Site package imported.");
    return;
  }
}

async function runSnapshotActionFromURL() {
  const action = gocmsPluginAction();
  if (!action) {
    return;
  }
  try {
    await runPluginAction(action);
  } catch (error) {
    window.alert(error instanceof Error ? error.message : "Snapshot action failed.");
  } finally {
    clearPluginAction();
  }
}

function installPluginActionClickHandler() {
  document.addEventListener(
    "click",
    (event) => {
      if (!isAdminChrome()) {
        return;
      }
      const anchor = event.target && event.target.closest && event.target.closest("a[href*='plugin-action=']");
      if (!anchor) {
        return;
      }
      const action = readPluginActionFromAnchor(anchor);
      if (!action) {
        return;
      }
      event.preventDefault();
      runPluginAction(action).catch((error) => {
        window.alert(error instanceof Error ? error.message : "Snapshot action failed.");
      });
    },
    true,
  );
}

if (typeof window !== "undefined" && isAdminChrome()) {
  installPluginActionClickHandler();
  if (document.readyState === "loading") {
    window.addEventListener(
      "DOMContentLoaded",
      () => {
        runSnapshotActionFromURL().catch(() => {});
      },
      { once: true },
    );
  } else {
    runSnapshotActionFromURL().catch(() => {});
  }
}
