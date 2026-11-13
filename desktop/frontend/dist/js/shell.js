// The application bar: brand, main navigation, global search, push protection
// status, session activity, the identity switcher, and settings.

const NAV_ITEMS = [
  { id: "dashboard", label: "Dashboard" },
  { id: "repositories", label: "Repositories" },
  { id: "identities", label: "Identities" },
  { id: "workspaces", label: "Workspaces" },
  { id: "doctor", label: "Doctor" },
];

const BRAND_MARK = `
  <svg class="brand-mark" viewBox="0 0 24 24" aria-hidden="true">
    <rect width="24" height="24" rx="7" fill="var(--accent-strong)" />
    <path d="M8.500 5.500v13" stroke="#fff" stroke-width="2.200" stroke-linecap="round" fill="none" />
    <circle cx="13" cy="14" r="4" stroke="#fff" stroke-width="2.200" fill="none" />
  </svg>`;

let appbarKey = "";

// Small markers on navigation tabs: something on that page needs attention.
function navMarker(id) {
  if (id === "repositories" && (state.repos?.needAttention || 0) > 0) {
    return `<span class="dot is-warning" data-tip="${plural(state.repos.needAttention, "repository needs", "repositories need")} attention"></span>`;
  }
  if (id === "doctor" && state.doctor && (state.doctor.errors > 0 || state.doctor.warnings > 0)) {
    return `<span class="dot ${state.doctor.errors > 0 ? "is-danger" : "is-warning"}"></span>`;
  }
  return "";
}

function renderAppbar() {
  const bar = document.getElementById("appbar");
  if (!state.status) {
    return;
  }
  const active = activeIdentity();
  const activeNav = currentPage().nav || state.page;
  const protectedNow = Boolean(state.onboarding?.pushProtection);
  const key = JSON.stringify([
    activeNav,
    active?.alias,
    active?.githubAvatarUrl,
    identities().length,
    protectedNow,
    state.unseenActivity,
    state.repos?.needAttention || 0,
    state.doctor?.errors || 0,
    state.doctor?.warnings || 0,
    state.status.configured,
  ]);
  // The bar is only re-drawn when something it shows changed, and never
  // while the search box is in use.
  if (key === appbarKey || document.activeElement?.id === "globalSearch") {
    return;
  }
  appbarKey = key;

  bar.innerHTML = `
    <div class="appbar-left">
      <button type="button" class="brand" data-action="navigate" data-page="dashboard" aria-label="bgit dashboard">${BRAND_MARK}<span>bgit</span></button>
      <nav class="nav-tabs" aria-label="Main">
        ${NAV_ITEMS.map(
          (item, index) => `
            <button type="button" class="nav-tab ${activeNav === item.id ? "active" : ""}" data-action="navigate" data-page="${item.id}" data-nav="${item.id}" ${activeNav === item.id ? 'aria-current="page"' : ""} data-tip="${MOD_KEY}+${index + 1}">
              ${escapeHtml(item.label)}${navMarker(item.id)}
            </button>`,
        ).join("")}
      </nav>
    </div>
    <div class="appbar-center">
      <label class="global-search" id="globalSearchBox">
        ${icon("search")}
        <input id="globalSearch" type="text" placeholder="Search repositories" autocomplete="off" spellcheck="false" aria-label="Search repositories" />
        <span class="kbd">${MOD_KEY}</span><span class="kbd">K</span>
      </label>
    </div>
    <div class="appbar-right">
      <button type="button" class="status-pill ${protectedNow ? "is-on" : "is-off"}" data-action="navigate" data-page="settings" data-section="protection" data-tip="${protectedNow ? "Every push is checked against the repository identity" : "Pushes are not being checked. Open Settings to turn protection on"}" data-tip-align="end">
        ${icon(protectedNow ? "shield-check" : "shield-off")}<span>${protectedNow ? "Protected" : "Unprotected"}</span>
      </button>
      <button type="button" class="icon-btn" data-action="appbar-activity" aria-label="Activity" data-tip="Activity in this session" data-tip-align="end">
        ${icon("bell")}${state.unseenActivity > 0 ? `<span class="icon-btn-dot"></span>` : ""}
      </button>
      <button type="button" class="identity-switch" id="identitySwitch" data-action="appbar-identity" aria-haspopup="menu" aria-expanded="false" data-tip="Default identity" data-tip-align="end">
        ${active ? avatar(active, "sm") : `<span class="avatar sm" aria-hidden="true">?</span>`}
        <span class="truncate">${escapeHtml(active?.alias || "No identity")}</span>
        ${icon("chevron-down", "sm")}
      </button>
      <button type="button" class="icon-btn ${state.page === "settings" ? "active" : ""}" data-action="navigate" data-page="settings" data-nav="settings" aria-label="Settings" data-tip="Settings (${MOD_KEY}+,)" data-tip-align="end">${icon("settings")}</button>
    </div>`;

  bindGlobalSearch();
}

// ---------- Identity switcher ----------

async function setDefaultIdentity(alias) {
  try {
    const result = await api().ActivateIdentity(alias);
    finishAction({ ...result, message: `'${alias}' is now the default identity` }, { activity: "Identities" });
  } catch (error) {
    notify({ title: "Could not switch identity", message: errorText(error), tone: "danger" });
  }
}

registerActions({
  "appbar-identity"(button) {
    const items = [{ heading: "Default identity" }];
    for (const identity of identities()) {
      items.push({
        html: `${avatar(identity, "sm")}<span class="menu-label"><strong class="truncate">${escapeHtml(identity.alias)}</strong><span class="muted truncate">${escapeHtml(identity.email)}</span></span>`,
        checked: identity.active,
        onSelect: () => {
          if (!identity.active) {
            setDefaultIdentity(identity.alias);
          }
        },
      });
    }
    if (identities().length === 0) {
      items.push({ label: "No identities yet", disabled: true });
    }
    items.push({ separator: true });
    items.push({ label: "Add identity", icon: "plus", onSelect: () => openIdentityDialog() });
    items.push({ label: "Manage identities", icon: "users", onSelect: () => navigate("identities") });
    openMenu(button, items, {
      align: "end",
      width: 300,
      className: "identity-menu",
      // Switching here does not touch repositories that have their own identity.
      note: "The default identity applies to repositories without an identity of their own. Bound repositories and workspaces keep theirs.",
    });
  },

  "appbar-activity"(button) {
    state.unseenActivity = 0;
    const items = state.activity;
    const html = `
      <div class="activity-pop">
        <div class="menu-heading">This session</div>
        ${
          items.length === 0
            ? `<p class="activity-empty">Nothing yet. Actions you take in bgit are listed here.</p>`
            : items
                .map(
                  (item) => `
                    <div class="activity-item">
                      <span class="dot is-${item.tone === "error" || item.tone === "danger" ? "danger" : item.tone === "warn" || item.tone === "warning" ? "warning" : "success"}"></span>
                      <div><strong>${escapeHtml(item.title)}</strong><span class="muted">${escapeHtml([item.detail, timeAgo(item.at)].filter(Boolean).join(" · "))}</span></div>
                    </div>`,
                )
                .join("")
        }
      </div>`;
    openPopover(button, html, { align: "end", width: 320, onClose: renderAppbar });
  },
});

// ---------- Global search ----------

let searchIndex = 0;

function searchResults(query) {
  const needle = query.trim().toLowerCase();
  const repos = repositories().filter((repo) => !needle || repo.name.toLowerCase().includes(needle) || repo.path.toLowerCase().includes(needle));
  if (!needle) {
    // Nothing typed yet: offer the repositories opened most recently.
    repos.sort((a, b) => recentRepositories.openedAt(b.path) - recentRepositories.openedAt(a.path) || a.name.localeCompare(b.name));
  }
  return repos.slice(0, 8);
}

function drawSearchResults(input) {
  const results = searchResults(input.value);
  searchIndex = Math.min(searchIndex, Math.max(0, results.length - 1));
  const html =
    results.length === 0
      ? `<p class="activity-empty">${repositories().length === 0 ? "No repositories yet." : "No repository matches."}</p>`
      : `<div class="menu">
          <div class="menu-heading">${input.value.trim() ? "Repositories" : "Recent repositories"}</div>
          ${results
            .map(
              (repo, index) => `
                <button type="button" class="menu-item ${index === searchIndex ? "active" : ""}" data-search-path="${escapeHtml(repo.path)}" ${repo.exists ? "" : "disabled"}>
                  ${icon("repo")}
                  <span class="menu-label"><strong class="truncate">${escapeHtml(repo.name)}</strong><span class="muted mono truncate-start"><span>${escapeHtml(repo.path)}</span></span></span>
                  ${identityChip(repo.alias, { plain: true })}
                </button>`,
            )
            .join("")}
        </div>`;
  const box = document.getElementById("globalSearchBox");
  openPopover(box, html, {
    replace: true,
    width: Math.max(box.offsetWidth, 420),
    className: "search-pop",
    onOpen(element) {
      element.addEventListener("mousedown", (event) => {
        const button = event.target.closest("[data-search-path]");
        if (button && !button.disabled) {
          event.preventDefault();
          openSearchResult(button.dataset.searchPath);
        }
      });
    },
  });
  return results;
}

function openSearchResult(path) {
  const input = document.getElementById("globalSearch");
  input.value = "";
  input.blur();
  closePopover();
  openRepository(path);
}

function bindGlobalSearch() {
  const input = document.getElementById("globalSearch");
  input.addEventListener("focus", () => {
    searchIndex = 0;
    drawSearchResults(input);
  });
  input.addEventListener("input", () => {
    searchIndex = 0;
    drawSearchResults(input);
  });
  input.addEventListener("blur", () => {
    // Let a click on a result land before the list goes away.
    window.setTimeout(() => {
      if (document.activeElement !== input) {
        closePopover();
        renderAppbar();
      }
    }, 120);
  });
  input.addEventListener("keydown", (event) => {
    if (event.key === "Escape") {
      input.value = "";
      input.blur();
      return;
    }
    const results = searchResults(input.value);
    if (event.key === "ArrowDown" || event.key === "ArrowUp") {
      event.preventDefault();
      if (results.length > 0) {
        searchIndex = (searchIndex + (event.key === "ArrowDown" ? 1 : -1) + results.length) % results.length;
        drawSearchResults(input);
      }
    }
    if (event.key === "Enter" && results[searchIndex]?.exists) {
      event.preventDefault();
      openSearchResult(results[searchIndex].path);
    }
  });
}

function focusGlobalSearch() {
  document.getElementById("globalSearch")?.focus();
}
