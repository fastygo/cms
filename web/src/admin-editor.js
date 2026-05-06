import { Editor } from "@tiptap/core";
import StarterKit from "@tiptap/starter-kit";
import Link from "@tiptap/extension-link";

const readyHosts = new WeakSet();
const providers = new Map();
const instances = new WeakMap();
const selections = new WeakMap();
const TOOLBAR_BUTTON_CLASS = "inline-flex h-8 shrink-0 items-center justify-center gap-2 whitespace-nowrap rounded border border-border bg-background px-3 text-sm font-medium text-foreground transition-colors hover:bg-accent hover:text-accent-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:opacity-50";
const TOOLBAR_BUTTON_ACTIVE_CLASSES = ["bg-accent", "text-accent-foreground"];
const EDITOR_SURFACE_CLASS = "gocms-richtext-surface prose max-w-none";

function editorHosts(root) {
  return Array.from(root.querySelectorAll("[data-gocms-editor-provider]"));
}

function bootProviders(root) {
  for (const host of editorHosts(root)) {
    const providerId = host.dataset.gocmsEditorProvider || "";
    const activate = providers.get(providerId);
    if (!activate || readyHosts.has(host)) {
      continue;
    }
    activate(host);
    readyHosts.add(host);
  }
}

function registerProvider(id, activate) {
  if (!id || typeof activate !== "function") {
    return;
  }
  providers.set(id, activate);
  if (typeof document !== "undefined" && document.readyState !== "loading") {
    bootProviders(document);
  }
}

function fieldParts(host) {
  return {
    input: host.querySelector("[data-gocms-editor-input]"),
    surface: host.querySelector("[data-gocms-editor-surface]"),
    toolbar: host.querySelector("[data-gocms-editor-toolbar]"),
  };
}

function setHidden(element, hidden) {
  if (!(element instanceof HTMLElement)) {
    return;
  }
  if (hidden) {
    element.setAttribute("hidden", "hidden");
  } else {
    element.removeAttribute("hidden");
  }
}

function syncInput(editor, input) {
  input.value = editor.isEmpty ? "" : editor.getHTML();
}

function rememberSelection(editor) {
  selections.set(editor, {
    from: editor.state.selection.from,
    to: editor.state.selection.to,
  });
}

function runEditorCommand(editor, apply) {
  const selection = selections.get(editor);
  let chain = editor.chain().focus();
  if (selection && Number.isInteger(selection.from) && Number.isInteger(selection.to)) {
    chain = chain.setTextSelection(selection);
  }
  const completed = apply(chain).run();
  rememberSelection(editor);
  return completed;
}

function button(editor, label, onClick) {
  const element = document.createElement("button");
  element.type = "button";
  element.className = TOOLBAR_BUTTON_CLASS;
  element.textContent = label;
  element.addEventListener("mousedown", (event) => {
    event.preventDefault();
    onClick();
  });
  return element;
}

function promptForLink(editor) {
  const current = editor.getAttributes("link").href || "https://";
  const next = window.prompt("Enter URL", current);
  if (next === null) {
    return;
  }
  const value = next.trim();
  if (!value) {
    runEditorCommand(editor, (chain) => chain.extendMarkRange("link").unsetLink());
    return;
  }
  runEditorCommand(editor, (chain) => chain.extendMarkRange("link").setLink({ href: value }));
}

function toolbarItems(editor) {
  return [
    {
      label: "P",
      action: () => runEditorCommand(editor, (chain) => chain.setParagraph()),
      active: () => editor.isActive("paragraph"),
      enabled: () => editor.can().chain().focus().setParagraph().run(),
    },
    {
      label: "H2",
      action: () => runEditorCommand(editor, (chain) => chain.toggleHeading({ level: 2 })),
      active: () => editor.isActive("heading", { level: 2 }),
      enabled: () => editor.can().chain().focus().toggleHeading({ level: 2 }).run(),
    },
    {
      label: "H3",
      action: () => runEditorCommand(editor, (chain) => chain.toggleHeading({ level: 3 })),
      active: () => editor.isActive("heading", { level: 3 }),
      enabled: () => editor.can().chain().focus().toggleHeading({ level: 3 }).run(),
    },
    {
      label: "B",
      action: () => runEditorCommand(editor, (chain) => chain.toggleBold()),
      active: () => editor.isActive("bold"),
      enabled: () => editor.can().chain().focus().toggleBold().run(),
    },
    {
      label: "I",
      action: () => runEditorCommand(editor, (chain) => chain.toggleItalic()),
      active: () => editor.isActive("italic"),
      enabled: () => editor.can().chain().focus().toggleItalic().run(),
    },
    {
      label: "Quote",
      action: () => runEditorCommand(editor, (chain) => chain.toggleBlockquote()),
      active: () => editor.isActive("blockquote"),
      enabled: () => editor.can().chain().focus().toggleBlockquote().run(),
    },
    {
      label: "UL",
      action: () => runEditorCommand(editor, (chain) => chain.toggleBulletList()),
      active: () => editor.isActive("bulletList"),
      enabled: () => editor.can().chain().focus().toggleBulletList().run(),
    },
    {
      label: "OL",
      action: () => runEditorCommand(editor, (chain) => chain.toggleOrderedList()),
      active: () => editor.isActive("orderedList"),
      enabled: () => editor.can().chain().focus().toggleOrderedList().run(),
    },
    {
      label: "Link",
      action: () => promptForLink(editor),
      active: () => editor.isActive("link"),
      enabled: () => true,
    },
    {
      label: "Clear",
      action: () => runEditorCommand(editor, (chain) => chain.clearNodes().unsetAllMarks()),
      active: () => false,
      enabled: () => true,
    },
    {
      label: "Undo",
      action: () => runEditorCommand(editor, (chain) => chain.undo()),
      active: () => false,
      enabled: () => editor.can().chain().focus().undo().run(),
    },
    {
      label: "Redo",
      action: () => runEditorCommand(editor, (chain) => chain.redo()),
      active: () => false,
      enabled: () => editor.can().chain().focus().redo().run(),
    },
  ];
}

function renderToolbar(toolbar, editor) {
  const controls = [];
  toolbar.replaceChildren();
  for (const item of toolbarItems(editor)) {
    const control = button(editor, item.label, item.action);
    controls.push({ element: control, item });
    toolbar.appendChild(control);
  }

  const refresh = () => {
    for (const control of controls) {
      const active = control.item.active();
      const enabled = control.item.enabled();
      control.element.disabled = !enabled;
      control.element.setAttribute("aria-pressed", active ? "true" : "false");
      for (const className of TOOLBAR_BUTTON_ACTIVE_CLASSES) {
        control.element.classList.toggle(className, active);
      }
    }
  };

  editor.on("selectionUpdate", refresh);
  editor.on("transaction", refresh);
  refresh();
}

function enhanceTipTapBasic(host) {
  const existing = instances.get(host);
  if (existing) {
    return existing;
  }

  const { input, surface, toolbar } = fieldParts(host);
  if (!(input instanceof HTMLTextAreaElement) || !(surface instanceof HTMLElement) || !(toolbar instanceof HTMLElement)) {
    return null;
  }

  const editor = new Editor({
    element: surface,
    injectCSS: false,
    extensions: [
      StarterKit.configure({
        link: false,
      }),
      Link.configure({
        openOnClick: false,
        autolink: true,
        linkOnPaste: true,
        HTMLAttributes: {
          rel: "noopener noreferrer nofollow",
        },
      }),
    ],
    content: input.value || "",
    editorProps: {
      attributes: {
        class: EDITOR_SURFACE_CLASS,
        role: "textbox",
        "aria-multiline": "true",
        spellcheck: "true",
      },
    },
    onCreate: ({ editor: instance }) => {
      syncInput(instance, input);
      rememberSelection(instance);
    },
    onUpdate: ({ editor: instance }) => {
      syncInput(instance, input);
      rememberSelection(instance);
    },
    onSelectionUpdate: ({ editor: instance }) => {
      rememberSelection(instance);
    },
  });

  instances.set(host, editor);
  renderToolbar(toolbar, editor);
  setHidden(toolbar, false);
  setHidden(surface, false);
  input.hidden = true;
  input.setAttribute("aria-hidden", "true");

  const form = input.form;
  if (form) {
    form.addEventListener("submit", () => {
      syncInput(editor, input);
    });
  }
  return editor;
}

if (typeof window !== "undefined") {
  window.gocmsRegisterEditorProvider = registerProvider;
  registerProvider("tiptap-basic", enhanceTipTapBasic);
  registerProvider("classic-html", enhanceTipTapBasic);

  if (document.readyState === "loading") {
    window.addEventListener("DOMContentLoaded", () => {
      bootProviders(document);
    }, { once: true });
  } else {
    bootProviders(document);
  }
}
