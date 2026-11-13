// Shared UI building blocks: icons, small HTML builders, popovers and menus,
// dialogs, toasts, and clipboard. No page logic lives here.

// ---------- Icons ----------
// One stroke-based set, drawn on a 24x24 grid.
const ICONS = {
  dashboard: '<rect x="3" y="3" width="7" height="9" rx="1.5"/><rect x="14" y="3" width="7" height="5" rx="1.5"/><rect x="14" y="12" width="7" height="9" rx="1.5"/><rect x="3" y="16" width="7" height="5" rx="1.5"/>',
  repo: '<path d="M4 19.500A2.500 2.500 0 0 1 6.500 17H20"/><path d="M6.500 2H20v20H6.500A2.500 2.500 0 0 1 4 19.500v-15A2.500 2.500 0 0 1 6.500 2Z"/>',
  users: '<path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M22 21v-2a4 4 0 0 0-3-3.870"/><path d="M16 3.130a4 4 0 0 1 0 7.750"/>',
  user: '<path d="M19 21v-2a4 4 0 0 0-4-4H9a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/>',
  folder: '<path d="M3 6.500A1.500 1.500 0 0 1 4.500 5H9l2 2.500h8.500A1.500 1.500 0 0 1 21 9v9.500a1.500 1.500 0 0 1-1.500 1.500h-15A1.500 1.500 0 0 1 3 18.500Z"/>',
  layers: '<path d="m12 3 9 5-9 5-9-5 9-5Z"/><path d="m3 12.500 9 5 9-5"/><path d="m3 17 9 5 9-5"/>',
  activity: '<path d="M22 12h-4l-3 9L9 3l-3 9H2"/>',
  settings: '<path d="M4 6h9"/><path d="M19 6h1"/><circle cx="16" cy="6" r="2.200"/><path d="M4 12h3"/><path d="M13 12h7"/><circle cx="10" cy="12" r="2.200"/><path d="M4 18h10"/><circle cx="17" cy="18" r="2.200"/>',
  search: '<circle cx="11" cy="11" r="7"/><path d="m20 20-3.500-3.500"/>',
  plus: '<path d="M12 5v14"/><path d="M5 12h14"/>',
  x: '<path d="M18 6 6 18"/><path d="m6 6 12 12"/>',
  check: '<path d="M20 6 9 17l-5-5"/>',
  "chevron-down": '<path d="m6 9 6 6 6-6"/>',
  "chevron-right": '<path d="m9 6 6 6-6 6"/>',
  "arrow-left": '<path d="M19 12H5"/><path d="m11 6-6 6 6 6"/>',
  more: '<circle cx="5" cy="12" r="1.200"/><circle cx="12" cy="12" r="1.200"/><circle cx="19" cy="12" r="1.200"/>',
  download: '<path d="M12 4v11"/><path d="m7 11 5 5 5-5"/><path d="M5 20h14"/>',
  upload: '<path d="M12 20V9"/><path d="m7 13 5-5 5 5"/><path d="M5 4h14"/>',
  refresh: '<path d="M21 12a9 9 0 0 1-15.500 6.200L3 16"/><path d="M3 21v-5h5"/><path d="M3 12a9 9 0 0 1 15.500-6.200L21 8"/><path d="M21 3v5h-5"/>',
  branch: '<path d="M6 3v12"/><circle cx="18" cy="6" r="3"/><circle cx="6" cy="18" r="3"/><path d="M18 9a9 9 0 0 1-9 9"/>',
  commit: '<circle cx="12" cy="12" r="3.500"/><path d="M3 12h5.500"/><path d="M15.500 12H21"/>',
  clock: '<circle cx="12" cy="12" r="9"/><path d="M12 7v5l3 2"/>',
  copy: '<rect x="9" y="9" width="11" height="11" rx="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/>',
  trash: '<path d="M3 6h18"/><path d="M8 6V4h8v2"/><path d="m19 6-1 14H6L5 6"/><path d="M10 11v5"/><path d="M14 11v5"/>',
  edit: '<path d="M12 20h9"/><path d="M16.500 3.500a2.100 2.100 0 0 1 3 3L7 19l-4 1 1-4Z"/>',
  eye: '<path d="M2 12s3.500-7 10-7 10 7 10 7-3.500 7-10 7S2 12 2 12Z"/><circle cx="12" cy="12" r="3"/>',
  external: '<path d="M15 3h6v6"/><path d="M10 14 21 3"/><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/>',
  github: '<path d="M9 19c-5 1.500-5-2.500-7-3m14 6v-3.870a3.370 3.370 0 0 0-.940-2.610c3.140-.350 6.440-1.540 6.440-7A5.440 5.440 0 0 0 20 4.770 5.070 5.070 0 0 0 19.910 1S18.730.650 16 2.480a13.380 13.380 0 0 0-7 0C6.270.650 5.090 1 5.090 1A5.070 5.070 0 0 0 5 4.770a5.440 5.440 0 0 0-1.500 3.780c0 5.420 3.300 6.610 6.440 7A3.370 3.370 0 0 0 9 18.130V22"/>',
  key: '<circle cx="7.500" cy="15.500" r="4.500"/><path d="m10.700 12.300 9.300-9.300"/><path d="m16 7 3 3"/><path d="m19 4 2 2"/>',
  wrench: '<path d="M14.700 6.300a1 1 0 0 0 0 1.400l1.600 1.600a1 1 0 0 0 1.400 0l3.770-3.770a6 6 0 0 1-7.940 7.940l-6.910 6.910a2.120 2.120 0 0 1-3-3l6.910-6.910a6 6 0 0 1 7.940-7.940l-3.760 3.760z"/>',
  alert: '<path d="M10.300 3.900 1.800 18a2 2 0 0 0 1.700 3h17a2 2 0 0 0 1.700-3L13.700 3.900a2 2 0 0 0-3.400 0Z"/><path d="M12 9v4"/><path d="M12 17h.010"/>',
  info: '<circle cx="12" cy="12" r="9"/><path d="M12 11v5"/><path d="M12 8h.010"/>',
  shield: '<path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10Z"/>',
  "shield-check": '<path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10Z"/><path d="m9 12 2 2 4-4"/>',
  "shield-off": '<path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10Z"/><path d="M12 8v4"/><path d="M12 15.500h.010"/>',
  bell: '<path d="M6 8a6 6 0 0 1 12 0c0 7 3 9 3 9H3s3-2 3-9"/><path d="M10.300 21a1.940 1.940 0 0 0 3.400 0"/>',
  sun: '<circle cx="12" cy="12" r="4"/><path d="M12 2v2"/><path d="M12 20v2"/><path d="m4.900 4.900 1.400 1.400"/><path d="m17.700 17.700 1.400 1.400"/><path d="M2 12h2"/><path d="M20 12h2"/><path d="m6.300 17.700-1.400 1.400"/><path d="m19.100 4.900-1.400 1.400"/>',
  moon: '<path d="M12 3a6 6 0 0 0 9 9 9 9 0 1 1-9-9Z"/>',
  monitor: '<rect x="2" y="3" width="20" height="14" rx="2"/><path d="M8 21h8"/><path d="M12 17v4"/>',
  file: '<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8Z"/><path d="M14 2v6h6"/>',
  link: '<path d="M10 13a5 5 0 0 0 7.100 0l2-2a5 5 0 0 0-7.100-7.100l-1.100 1.100"/><path d="M14 11a5 5 0 0 0-7.100 0l-2 2A5 5 0 0 0 12 20.100l1.100-1.100"/>',
  lock: '<rect x="4" y="11" width="16" height="10" rx="2"/><path d="M8 11V7a4 4 0 0 1 8 0v4"/>',
  globe: '<circle cx="12" cy="12" r="9"/><path d="M3 12h18"/><path d="M12 3a14 14 0 0 1 0 18"/><path d="M12 3a14 14 0 0 0 0 18"/>',
  archive: '<rect x="2" y="4" width="20" height="5" rx="1"/><path d="M4 9v10a1 1 0 0 0 1 1h14a1 1 0 0 0 1-1V9"/><path d="M10 13h4"/>',
  star: '<path d="m12 3 2.700 5.600 6.100.900-4.400 4.300 1 6.100-5.400-2.900-5.400 2.900 1-6.100L3.200 9.500l6.100-.900L12 3Z"/>',
  keyboard: '<rect x="2" y="6" width="20" height="12" rx="2"/><path d="M6 10h.010"/><path d="M10 10h.010"/><path d="M14 10h.010"/><path d="M18 10h.010"/><path d="M7 14h10"/>',
  dot: '<circle cx="12" cy="12" r="3"/>',
  sidebar: '<rect x="3" y="4" width="18" height="16" rx="2"/><path d="M9 4v16"/>',
  "sidebar-open": '<rect x="3" y="4" width="18" height="16" rx="2"/><path d="M9 4v16"/><path d="m14 10 2 2-2 2"/>',
  wrap: '<path d="M3 6h18"/><path d="M3 12h15a3 3 0 0 1 0 6h-4"/><path d="m16 16-2 2 2 2"/><path d="M3 18h7"/>',
  "chevron-up": '<path d="m18 15-6-6-6 6"/>',
  "arrow-up": '<path d="M12 19V5"/><path d="m6 11 6-6 6 6"/>',
  "arrow-down": '<path d="M12 5v14"/><path d="m18 13-6 6-6-6"/>',
};

function icon(name, className = "") {
  return `<span class="icon ${className}" aria-hidden="true"><svg viewBox="0 0 24 24">${ICONS[name] || ICONS.dot}</svg></span>`;
}

// ---------- Small HTML builders ----------

function escapeHtml(value) {
  return String(value ?? "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#039;");
}

function identityInitials(identity) {
  const base = identity?.name || identity?.alias || identity?.githubUsername || "?";
  return (
    base
      .split(/\s+/)
      .filter(Boolean)
      .slice(0, 2)
      .map((part) => part[0].toUpperCase())
      .join("") || "?"
  );
}

// An identity's picture: the GitHub avatar when it loads, initials otherwise.
// A failed image is swapped for the initials by the error listener below.
function avatar(identity, size = "") {
  const initials = escapeHtml(identityInitials(identity));
  if (!identity?.githubAvatarUrl) {
    return `<span class="avatar ${size}" aria-hidden="true">${initials}</span>`;
  }
  return `<span class="avatar ${size}" aria-hidden="true" data-initials="${initials}"><img src="${escapeHtml(identity.githubAvatarUrl)}" alt="" loading="lazy" referrerpolicy="no-referrer" /></span>`;
}

document.addEventListener(
  "error",
  (event) => {
    const holder = event.target?.closest?.(".avatar[data-initials]");
    if (holder && event.target.tagName === "IMG") {
      holder.textContent = holder.dataset.initials;
    }
  },
  true,
);

function badge(label, tone = "", withDot = false) {
  return `<span class="badge ${tone ? `is-${tone}` : ""}">${withDot ? `<span class="dot ${tone ? `is-${tone}` : ""}"></span>` : ""}${escapeHtml(label)}</span>`;
}

function banner({ tone = "info", title, text = "", actions = "", iconName = "" }) {
  const defaultIcon = { warning: "alert", danger: "alert", success: "check", info: "info" }[tone] || "info";
  return `
    <div class="banner is-${tone}" role="${tone === "danger" || tone === "warning" ? "alert" : "status"}">
      ${icon(iconName || defaultIcon)}
      <div class="banner-text"><strong>${escapeHtml(title)}</strong>${text ? `<span>${escapeHtml(text)}</span>` : ""}</div>
      ${actions}
    </div>`;
}

function emptyState({ iconName = "info", title, text = "", actions = "", compact = false }) {
  return `
    <div class="empty ${compact ? "compact" : ""}">
      <div class="empty-icon">${icon(iconName, "lg")}</div>
      <h2>${escapeHtml(title)}</h2>
      ${text ? `<p>${escapeHtml(text)}</p>` : ""}
      ${actions ? `<div class="cluster">${actions}</div>` : ""}
    </div>`;
}

// Placeholder rows for a list that is still loading.
function skeletonRows(count) {
  return `<div aria-busy="true" aria-label="Loading">${'<div class="skeleton-row"><span class="skeleton"></span><span class="skeleton"></span><span class="skeleton"></span><span class="skeleton"></span></div>'.repeat(count)}</div>`;
}

function loadingLine(text) {
  return `<div class="loading-line"><span class="spinner" aria-hidden="true"></span>${escapeHtml(text)}</div>`;
}

function plural(count, one, many) {
  return `${count} ${count === 1 ? one : many}`;
}

function baseName(path) {
  return String(path || "").split(/[\\/]/).filter(Boolean).pop() || String(path || "");
}

function timeAgo(milliseconds) {
  if (!milliseconds) {
    return "";
  }
  const seconds = Math.max(0, (Date.now() - milliseconds) / 1000);
  if (seconds < 60) return "just now";
  if (seconds < 3600) return `${Math.floor(seconds / 60)} min ago`;
  if (seconds < 86400) return `${Math.floor(seconds / 3600)} h ago`;
  if (seconds < 86400 * 30) return `${Math.floor(seconds / 86400)} d ago`;
  return new Date(milliseconds).toLocaleDateString();
}

// ---------- Tooltips ----------
// Tooltips are drawn by CSS from data-tip. Once an element is clicked its
// tooltip would only be in the way, so it stays hidden until the pointer
// leaves the element.
document.addEventListener("mousedown", (event) => {
  const element = event.target.closest?.("[data-tip]");
  if (element) {
    element.classList.add("tip-off");
    element.addEventListener("mouseleave", () => element.classList.remove("tip-off"), { once: true });
  }
});

// ---------- Popovers and menus ----------

let activePopover = null;

function closePopover() {
  if (!activePopover) {
    return;
  }
  const { element, anchor, onClose } = activePopover;
  activePopover = null;
  element.remove();
  anchor?.setAttribute?.("aria-expanded", "false");
  onClose?.();
}

// Places a floating panel under its anchor, flipping above or shifting left
// when it would leave the window.
function placePopover(element, rect, align) {
  const margin = 8;
  const width = element.offsetWidth;
  const height = element.offsetHeight;
  let left = align === "end" ? rect.right - width : rect.left;
  left = Math.max(margin, Math.min(left, window.innerWidth - width - margin));
  let top = rect.bottom + 6;
  if (top + height > window.innerHeight - margin && rect.top - height - 6 > margin) {
    top = rect.top - height - 6;
  }
  top = Math.max(margin, Math.min(top, window.innerHeight - height - margin));
  element.style.left = `${Math.round(left)}px`;
  element.style.top = `${Math.round(top)}px`;
}

// Opens a floating panel anchored to an element (or to a point, for context
// menus). Opening it again from the same anchor closes it.
function openPopover(anchor, html, { align = "start", width = 0, className = "", onOpen, onClose, point = null, replace = false } = {}) {
  const sameAnchor = activePopover && anchor && activePopover.anchor === anchor;
  closePopover();
  // replace keeps a panel that re-draws itself (search results) open.
  if (sameAnchor && !replace) {
    return null;
  }
  const element = document.createElement("div");
  element.className = `popover ${className}`;
  if (width) {
    element.style.width = `${width}px`;
  }
  element.innerHTML = html;
  document.body.appendChild(element);
  const rect = point ? { left: point.x, right: point.x, top: point.y, bottom: point.y } : anchor.getBoundingClientRect();
  placePopover(element, rect, align);
  anchor?.setAttribute?.("aria-expanded", "true");
  activePopover = { element, anchor, onClose };
  onOpen?.(element);
  return element;
}

document.addEventListener("mousedown", (event) => {
  if (!activePopover) {
    return;
  }
  const inside = activePopover.element.contains(event.target);
  const onAnchor = activePopover.anchor?.contains?.(event.target);
  if (!inside && !onAnchor) {
    closePopover();
  }
});
window.addEventListener("blur", closePopover);
window.addEventListener("resize", closePopover);

function menuHtml(items) {
  return `<div class="menu" role="menu">${items
    .map((item, index) => {
      if (!item) return "";
      if (item.separator) return `<div class="menu-sep" role="separator"></div>`;
      if (item.heading) return `<div class="menu-heading">${escapeHtml(item.heading)}</div>`;
      return `
        <button type="button" class="menu-item ${item.danger ? "danger" : ""} ${item.active ? "active" : ""}" role="menuitem" data-menu-index="${index}" ${item.disabled ? "disabled" : ""}>
          ${item.html || `${item.icon ? icon(item.icon) : ""}<span class="menu-label truncate">${escapeHtml(item.label)}</span>`}
          ${item.hint ? `<span class="menu-hint">${escapeHtml(item.hint)}</span>` : ""}
          ${item.checked ? icon("check", "sm") : ""}
        </button>`;
    })
    .join("")}</div>`;
}

// Opens a menu. Items: { label, icon, hint, danger, disabled, checked, onSelect }
// | { separator: true } | { heading: "..." }. Arrow keys move, Enter selects.
function openMenu(anchor, items, options = {}) {
  const footer = options.note ? `<div class="menu-note">${escapeHtml(options.note)}</div>` : "";
  return openPopover(anchor, menuHtml(items) + footer, {
    ...options,
    onOpen(element) {
      element.addEventListener("click", (event) => {
        const button = event.target.closest("[data-menu-index]");
        if (!button || button.disabled) {
          return;
        }
        const item = items[Number(button.dataset.menuIndex)];
        closePopover();
        item.onSelect?.();
      });
      element.addEventListener("keydown", (event) => {
        const buttons = [...element.querySelectorAll(".menu-item:not(:disabled)")];
        const current = buttons.indexOf(document.activeElement);
        if (event.key === "ArrowDown" || event.key === "ArrowUp") {
          event.preventDefault();
          const step = event.key === "ArrowDown" ? 1 : -1;
          buttons[(current + step + buttons.length) % buttons.length]?.focus();
        }
        if (event.key === "Home" || event.key === "End") {
          event.preventDefault();
          buttons[event.key === "Home" ? 0 : buttons.length - 1]?.focus();
        }
      });
      element.querySelector(".menu-item:not(:disabled)")?.focus({ preventScroll: true });
      options.onOpen?.(element);
    },
  });
}

function openContextMenu(event, items) {
  event.preventDefault();
  return openMenu(null, items, { point: { x: event.clientX, y: event.clientY } });
}

// ---------- Dialogs ----------

// Opens a modal dialog built from a form. onSubmit receives the form and may
// throw (the message is shown inside the dialog) or return a value, which
// closes the dialog and resolves the promise. Cancelling resolves undefined.
function openDialog({
  id,
  title,
  description = "",
  body = "",
  submitLabel = "Save",
  cancelLabel = "Cancel",
  busyLabel = "",
  danger = false,
  wide = false,
  footNote = "",
  hideSubmit = false,
  onOpen,
  onSubmit,
}) {
  document.getElementById(id)?.remove();
  const dialog = document.createElement("dialog");
  dialog.id = id;
  dialog.className = `dialog ${wide ? "wide" : ""}`;
  dialog.innerHTML = `
    <form class="dialog-form" method="dialog" novalidate>
      <div class="dialog-head">
        <h2>${escapeHtml(title)}</h2>
        ${description ? `<p>${escapeHtml(description)}</p>` : ""}
      </div>
      <div class="dialog-body">
        ${body}
        <div class="dialog-error" data-dialog-error role="alert" hidden></div>
      </div>
      <div class="dialog-foot">
        ${footNote ? `<span class="dialog-foot-note">${escapeHtml(footNote)}</span>` : ""}
        <button type="button" class="btn" data-dialog-cancel>${escapeHtml(cancelLabel)}</button>
        ${hideSubmit ? "" : `<button type="submit" class="btn ${danger ? "btn-danger" : "btn-primary"}" data-dialog-submit>${escapeHtml(submitLabel)}</button>`}
      </div>
    </form>`;
  document.body.appendChild(dialog);

  const form = dialog.querySelector("form");
  const errorBox = dialog.querySelector("[data-dialog-error]");
  const submitButton = dialog.querySelector("[data-dialog-submit]");
  const control = {
    element: dialog,
    form,
    setError(message) {
      errorBox.textContent = message || "";
      errorBox.hidden = !message;
      if (message) {
        errorBox.scrollIntoView({ block: "nearest" });
      }
    },
    close(value) {
      dialog.returnValue = "";
      dialog._result = value;
      dialog.close();
    },
  };

  return new Promise((resolve) => {
    dialog.addEventListener("close", () => {
      const result = dialog._result;
      dialog.remove();
      resolve(result);
    });
    const cancelButton = dialog.querySelector("[data-dialog-cancel]");
    cancelButton.addEventListener("click", () => control.close(undefined));
    // While the action is running the dialog stays: closing it would hide
    // the result of something that is still happening.
    dialog.addEventListener("cancel", (event) => {
      if (submitButton?.disabled) {
        event.preventDefault();
      }
    });
    form.addEventListener("input", () => control.setError(""));
    form.addEventListener("submit", async (event) => {
      event.preventDefault();
      if (!onSubmit) {
        control.close(true);
        return;
      }
      control.setError("");
      const idleHtml = submitButton?.innerHTML;
      if (submitButton) {
        submitButton.disabled = true;
        submitButton.innerHTML = `<span class="spinner" aria-hidden="true"></span>${escapeHtml(busyLabel || submitLabel)}`;
      }
      cancelButton.disabled = true;
      try {
        const result = await onSubmit(form, control);
        if (dialog.open) {
          control.close(result === undefined ? true : result);
        }
      } catch (error) {
        control.setError(errorText(error));
      } finally {
        cancelButton.disabled = false;
        if (submitButton) {
          submitButton.disabled = false;
          submitButton.innerHTML = idleHtml;
        }
      }
    });
    closePopover();
    dialog.showModal();
    onOpen?.(control);
    (dialog.querySelector("[autofocus]") || submitButton)?.focus();
  });
}

function confirmAction({ title, message, warning = "", acceptLabel = "Continue", cancelLabel = "Cancel", danger = false }) {
  return openDialog({
    id: "confirmDialog",
    title,
    description: message,
    body: warning ? banner({ tone: danger ? "danger" : "warning", title: warning }) : "",
    submitLabel: acceptLabel,
    cancelLabel,
    danger,
  }).then((result) => result === true);
}

// Backend errors arrive as plain strings; keep whatever is readable.
function errorText(error) {
  if (!error) return "Something went wrong.";
  if (typeof error === "string") return error;
  return error.message || String(error);
}

// ---------- Toasts ----------

function notify({ title, message = "", tone = "success", timeout = 4800 }) {
  const region = document.getElementById("toasts");
  const toast = document.createElement("div");
  const toneClass = tone === "error" ? "danger" : tone === "warn" ? "warning" : tone;
  toast.className = `toast is-${toneClass}`;
  toast.setAttribute("role", toneClass === "danger" ? "alert" : "status");
  toast.innerHTML = `
    ${icon(toneClass === "success" ? "check" : toneClass === "info" ? "info" : "alert")}
    <div class="toast-text"><strong>${escapeHtml(title)}</strong>${message ? `<span>${escapeHtml(message)}</span>` : ""}</div>
    <button type="button" class="icon-btn sm" aria-label="Dismiss">${icon("x", "sm")}</button>`;
  region.appendChild(toast);
  // Keep the stack short so notifications never cover the page.
  while (region.children.length > 3) {
    region.firstElementChild.remove();
  }
  const remove = () => toast.remove();
  toast.querySelector("button").addEventListener("click", remove);
  if (timeout > 0) {
    window.setTimeout(remove, timeout);
  }
}

// ---------- Clipboard ----------

async function copyText(value, message = "Copied to clipboard") {
  if (!value) {
    return false;
  }
  try {
    if (window.runtime?.ClipboardSetText) {
      await window.runtime.ClipboardSetText(value);
    } else {
      await navigator.clipboard.writeText(value);
    }
    notify({ title: "Copied", message, timeout: 2200 });
    return true;
  } catch (error) {
    notify({ title: "Could not copy", message: errorText(error), tone: "danger" });
    return false;
  }
}

// Shows a spinner inside a button while an async action runs.
async function withBusy(button, busyLabel, action) {
  if (!button) {
    return action();
  }
  const idleHtml = button.innerHTML;
  const wasDisabled = button.disabled;
  button.disabled = true;
  const iconOnly = button.classList.contains("icon-btn");
  button.innerHTML = iconOnly ? `<span class="spinner" aria-hidden="true"></span>` : `<span class="spinner" aria-hidden="true"></span>${escapeHtml(busyLabel)}`;
  try {
    return await action();
  } finally {
    if (button.isConnected) {
      button.disabled = wasDisabled;
      button.innerHTML = idleHtml;
    }
  }
}

// ---------- Resizable panes ----------
// A splitter is a thin handle between two panes. Dragging it (or using the
// arrow keys when it has focus) changes one CSS variable on the container,
// e.g. --side-w. CSS clamps the variable, so the layout stays valid when the
// window shrinks. The size is remembered per key; double-click resets it.
// reverse: the resized pane sits after the handle (right or below).
function bindSplitter(handle, { container, variable, key, fallback, min, max, axis = "x", reverse = false }) {
  if (!handle || !container) {
    return;
  }
  const limit = (value) => Math.round(Math.max(min, Math.min(max(), value)));
  const apply = (value, remember) => {
    const size = limit(value);
    container.style.setProperty(variable, `${size}px`);
    handle.setAttribute("aria-valuenow", String(size));
    if (remember) {
      prefs.set(`split.${key}`, size);
    }
    return size;
  };
  // Measure what is really shown: CSS may have clamped the stored size.
  const current = () => {
    const pane = reverse ? handle.nextElementSibling : handle.previousElementSibling;
    const rect = pane?.getBoundingClientRect();
    return rect ? (axis === "x" ? rect.width : rect.height) : prefs.get(`split.${key}`, fallback);
  };

  handle.setAttribute("role", "separator");
  handle.setAttribute("tabindex", "0");
  handle.setAttribute("aria-orientation", axis === "x" ? "vertical" : "horizontal");
  handle.setAttribute("aria-valuemin", String(min));
  apply(prefs.get(`split.${key}`, fallback), false);

  handle.addEventListener("pointerdown", (event) => {
    if (event.button !== 0) {
      return;
    }
    event.preventDefault();
    const start = axis === "x" ? event.clientX : event.clientY;
    const startSize = current();
    handle.setPointerCapture(event.pointerId);
    handle.classList.add("dragging");
    document.body.classList.add(axis === "x" ? "resizing-x" : "resizing-y");
    const move = (moveEvent) => {
      const delta = (axis === "x" ? moveEvent.clientX : moveEvent.clientY) - start;
      apply(startSize + (reverse ? -delta : delta), false);
    };
    const stop = () => {
      handle.removeEventListener("pointermove", move);
      handle.classList.remove("dragging");
      document.body.classList.remove("resizing-x", "resizing-y");
      apply(current(), true);
    };
    handle.addEventListener("pointermove", move);
    handle.addEventListener("pointerup", stop, { once: true });
    handle.addEventListener("pointercancel", stop, { once: true });
  });
  handle.addEventListener("dblclick", () => apply(fallback, true));
  handle.addEventListener("keydown", (event) => {
    const grow = axis === "x" ? (reverse ? "ArrowLeft" : "ArrowRight") : reverse ? "ArrowUp" : "ArrowDown";
    const shrink = axis === "x" ? (reverse ? "ArrowRight" : "ArrowLeft") : reverse ? "ArrowDown" : "ArrowUp";
    const step = event.shiftKey ? 64 : 16;
    let next = null;
    if (event.key === grow) next = current() + step;
    if (event.key === shrink) next = current() - step;
    if (event.key === "Home") next = min;
    if (event.key === "End") next = max();
    if (next !== null) {
      event.preventDefault();
      apply(next, true);
    }
  });
}
