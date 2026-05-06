function gocmsPluginAction() {
  const params = new URLSearchParams(window.location.search);
  return params.get("plugin-action") || "";
}

function clearPluginAction() {
  const url = new URL(window.location.href);
  url.searchParams.delete("plugin-action");
  window.history.replaceState({}, "", url.pathname + url.search);
}

async function postPlugin(url, body) {
  const response = await fetch(url, {
    method: "POST",
    body,
    headers: body ? undefined : { "Accept": "application/json" },
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
    input.addEventListener("change", async () => {
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
    }, { once: true });
    input.click();
  });
}

async function runSnapshotAction() {
  const action = gocmsPluginAction();
  if (!action) {
    return;
  }
  try {
    if (action === "json-import-export.import") {
      const imported = await importSnapshotFile();
      if (imported) {
        window.alert("JSON snapshot imported.");
      }
    } else if (action === "json-import-export.export-site-package") {
      await postPlugin("/go-admin/plugins/json-import-export/export-site-package");
      window.alert("Site package exported.");
    } else if (action === "json-import-export.import-site-package") {
      await postPlugin("/go-admin/plugins/json-import-export/import-site-package");
      window.alert("Site package imported.");
    } else {
      return;
    }
  } catch (error) {
    window.alert(error instanceof Error ? error.message : "Snapshot action failed.");
  } finally {
    clearPluginAction();
  }
}

if (typeof window !== "undefined" && window.location.pathname.startsWith("/go-admin")) {
  if (document.readyState === "loading") {
    window.addEventListener("DOMContentLoaded", () => {
      runSnapshotAction().catch(() => {});
    }, { once: true });
  } else {
    runSnapshotAction().catch(() => {});
  }
}
