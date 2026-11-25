// Application core: shared state, the backend bridge, preferences, the
// repository store, page registry, rendering, navigation, and shortcuts.
// Pages register themselves with registerPage() and never touch each other.

const state = {
  status: null,
  doctor: null,
  doctorRunAt: 0,
  onboarding: null,
  appInfo: null,
  // Result of ListRepositories, shared by the dashboard, lists, search,
  // identities, workspaces, and doctor.
  repos: null,
  reposError: "",
  page: "dashboard",
  history: [],
  activity: [],
  unseenActivity: 0,
};

function api() {
  return window.go?.main?.App;
}

// ---------- Preferences (kept in this browser profile, never in bgit config) ----------

const prefs = {
  get(key, fallback) {
    try {
      const raw = window.localStorage.getItem(`bgit.${key}`);
      return raw === null ? fallback : JSON.parse(raw);
    } catch (error) {
      return fallback;
    }
  },
  set(key, value) {
    try {
      window.localStorage.setItem(`bgit.${key}`, JSON.stringify(value));
    } catch (error) {
      // Preferences are a convenience; losing one is harmless.
    }
  },
};

// ---------- Theme ----------

function applyTheme() {
  const choice = prefs.get("theme", "dark");
  const resolved = choice === "system" ? (window.matchMedia("(prefers-color-scheme: light)").matches ? "light" : "dark") : choice;
  document.documentElement.dataset.theme = resolved;
}
window.matchMedia("(prefers-color-scheme: light)").addEventListener?.("change", applyTheme);
applyTheme();

// ---------- Small data helpers ----------

function identities() {
  return state.status?.identities || [];
}

function identityByAlias(alias) {
  return identities().find((identity) => identity.alias === alias) || null;
}

function activeIdentity() {
  return state.status?.activeIdentity || null;
}

function repositories() {
  return state.repos?.repositories || [];
}

function repositoryByPath(path) {
  return repositories().find((repo) => repo.path === path) || null;
}

// An identity reference used across the app: avatar and alias.
function identityChip(alias, { plain = false } = {}) {
  const identity = identityByAlias(alias);
  if (!identity) {
    return `<span class="chip is-missing ${plain ? "plain" : ""}">${escapeHtml(alias || "No identity")}</span>`;
  }
  return `<span class="chip ${plain ? "plain" : ""}" data-tip="${escapeHtml(`${identity.name} <${identity.email}>`)}">${avatar(identity, "xs")}<span class="truncate">${escapeHtml(identity.alias)}</span></span>`;
}

// Repositories the user opened in this app, most recent first.
const recentRepositories = {
  all() {
    return prefs.get("recent", {});
  },
  touch(path) {
    const recent = this.all();
    recent[path] = Date.now();
    prefs.set("recent", recent);
  },
  openedAt(path) {
    return this.all()[path] || 0;
  },
};

function syncSummary(repo) {
  if (!repo.exists) return "Folder not found";
  if (!repo.remoteUrl) return "No remote";
  const parts = [];
  if (repo.ahead > 0) parts.push(`${repo.ahead} to push`);
  if (repo.behind > 0) parts.push(`${repo.behind} to pull`);
  return parts.join(" · ") || "In sync";
}

// ---------- Activity (this session) ----------

function addActivity(title, detail = "", tone = "success") {
  state.activity.unshift({ title, detail, tone, at: Date.now() });
  state.activity = state.activity.slice(0, 20);
  state.unseenActivity += 1;
  if (typeof renderAppbar === "function") {
    renderAppbar();
  }
}

// ---------- Pages ----------

const pages = {};
const globalActions = {};

// A page: { id, nav, fill, render(), bind(root), enter(params), focus(),
//           actions: { name(el, event) }, on: { input, change, submit },
//           context(el, event), keys(event) }
function registerPage(page) {
  pages[page.id] = page;
}

function registerActions(actions) {
  Object.assign(globalActions, actions);
}

function currentPage() {
  return pages[state.page] || pages.dashboard;
}

let renderedPage = "";
const scrollMemory = {};

function render() {
  if (!state.status) {
    return;
  }
  renderAppbar();
  const view = document.getElementById("view");
  const page = currentPage();
  const changed = renderedPage !== state.page;
  if (renderedPage && changed) {
    scrollMemory[renderedPage] = view.scrollTop;
  }
  const scrollTop = changed ? scrollMemory[state.page] || 0 : view.scrollTop;

  // Pages with their own scrolling panes fill the window instead of scrolling.
  view.classList.toggle("fill", Boolean(page.fill));
  // Only a change of page fades in; re-drawing the same page must not animate.
  view.innerHTML = `<div class="page-root${changed ? " page-enter" : ""}">${state.status.configured || page.allowUnconfigured ? page.render() : unconfiguredView()}</div>`;
  renderedPage = state.page;
  page.bind?.(view);
  view.scrollTop = scrollTop;
}

function unconfiguredView() {
  return `
    <div class="page">
      ${emptyState({
        iconName: "shield",
        title: "bgit is not set up on this computer",
        text: "Setup takes about a minute: it turns on push protection and adds your first identity.",
        actions: `<button type="button" class="btn btn-primary" data-action="open-setup">Start setup</button>`,
      })}
    </div>`;
}

function navigate(pageId, params = {}, { remember = true } = {}) {
  if (!pages[pageId]) {
    return;
  }
  closePopover();
  if (state.page !== pageId && remember) {
    state.history.push(state.page);
  }
  state.page = pageId;
  pages[pageId].enter?.(params);
  render();
}

function goBack() {
  navigate(state.history.pop() || "dashboard", {}, { remember: false });
}

// ---------- Event delegation ----------
// Elements declare what they do with data-action; the current page's actions
// win over global ones. Nothing needs re-binding after a re-draw.

document.addEventListener("click", (event) => {
  const element = event.target.closest("[data-action]");
  if (!element || element.disabled || element.getAttribute("aria-disabled") === "true") {
    return;
  }
  const handler = currentPage().actions?.[element.dataset.action] || globalActions[element.dataset.action];
  if (handler) {
    handler(element, event);
  }
});

for (const type of ["input", "change", "submit"]) {
  document.addEventListener(type, (event) => {
    if (!event.target.closest?.("#view")) {
      return;
    }
    currentPage().on?.[type]?.(event.target, event);
  });
}

document.addEventListener("contextmenu", (event) => {
  const element = event.target.closest?.("[data-context]");
  if (element && event.target.closest("#view")) {
    currentPage().context?.(element, event);
  }
});

registerActions({
  navigate(element) {
    navigate(element.dataset.page, { ...element.dataset });
  },
  back() {
    goBack();
  },
  copy(element) {
    copyText(element.dataset.value, element.dataset.copied || "Copied to clipboard");
  },
  "retry-repos"(element) {
    withBusy(element, "Loading", () => loadRepositories());
  },
  "open-setup"() {
    openOnboarding("welcome");
  },
  async "open-link"(element) {
    try {
      await api().OpenExternalURL(element.dataset.url);
    } catch (error) {
      notify({ title: "Could not open link", message: errorText(error), tone: "danger" });
    }
  },
});

// ---------- Loading and refreshing ----------

function applyStatus(status) {
  if (status) {
    state.status = status;
  }
}

async function load() {
  const view = document.getElementById("view");
  try {
    if (!api()) {
      throw new Error("The bgit backend is not available. Open this screen in the bgit app window.");
    }
    if (!state.status) {
      view.innerHTML = loadingLine("Starting bgit");
    }
    const [status, doctor, onboarding] = await Promise.all([api().GetStatus(), api().GetDoctorStatus(), api().GetOnboardingState()]);
    state.status = status;
    state.doctor = doctor;
    state.doctorRunAt = Date.now();
    state.onboarding = onboarding;
    render();
    if (status.configured) {
      loadRepositories();
    }
    if (!state.appInfo) {
      api()
        .GetAppInfo()
        .then((info) => {
          state.appInfo = info;
        })
        .catch(() => {});
    }
    // First run: the setup guide covers the app until setup is finished.
    if (typeof syncOnboarding === "function" && !window.settingsShowsRemoval?.()) {
      syncOnboarding(onboarding);
    }
  } catch (error) {
    view.innerHTML = `<div class="page">${emptyState({ iconName: "alert", title: "bgit could not start", text: errorText(error) })}</div>`;
  }
}

let repositoriesRequest = null;

// Loads the repository list. Callers get the same request while one is in
// flight, and the page is re-drawn only when the data actually changed.
function loadRepositories() {
  if (!state.status?.configured || identities().length === 0) {
    state.repos = { repositories: [], needAttention: 0 };
    return Promise.resolve(state.repos);
  }
  if (repositoriesRequest) {
    return repositoriesRequest;
  }
  const before = JSON.stringify([state.repos, state.reposError]);
  repositoriesRequest = api()
    .ListRepositories()
    .then((result) => {
      state.repos = result;
      state.reposError = "";
    })
    .catch((error) => {
      state.reposError = errorText(error);
    })
    .then(() => {
      repositoriesRequest = null;
      if (JSON.stringify([state.repos, state.reposError]) !== before && !currentPage().fill) {
        render();
      } else if (currentPage().fill) {
        renderAppbar();
      }
      return state.repos;
    });
  return repositoriesRequest;
}

// Applies the result of an action that changed something: fresh status when
// the backend sent one, a reloaded repository list, and feedback.
function finishAction(result, { activity = "" } = {}) {
  applyStatus(result?.status);
  if (result?.message) {
    addActivity(result.message, activity);
    notify({ title: result.message });
  }
  render();
  loadRepositories();
}

// The bgit CLI and Git share this app's data. When the window comes back into
// focus, pick up whatever changed outside the app, without re-drawing anything
// when nothing changed.
let lastFocusRefresh = 0;

async function refreshOnFocus() {
  if (!api() || !state.status) {
    return;
  }
  if (document.querySelector("dialog[open]") || !document.getElementById("onboarding").hidden) {
    return;
  }
  const now = Date.now();
  if (now - lastFocusRefresh < 4000) {
    return;
  }
  lastFocusRefresh = now;
  try {
    const [status, onboarding] = await Promise.all([api().GetStatus(), api().GetOnboardingState()]);
    const changed = JSON.stringify([status, onboarding]) !== JSON.stringify([state.status, state.onboarding]);
    if (changed) {
      state.status = status;
      state.onboarding = onboarding;
      if (onboarding.needed) {
        // For example bgit was removed with the CLI: start over.
        await load();
        return;
      }
      const typing = ["INPUT", "TEXTAREA", "SELECT"].includes(document.activeElement?.tagName);
      if (currentPage().fill || typing) {
        renderAppbar();
      } else {
        render();
      }
    }
    if (status.configured) {
      loadRepositories();
    }
    currentPage().focus?.();
  } catch (error) {
    // A failed background refresh is not worth interrupting the user.
  }
}
window.addEventListener("focus", refreshOnFocus);

// ---------- Keyboard shortcuts ----------

const NAV_ORDER = ["dashboard", "repositories", "identities", "workspaces", "doctor"];
const MOD_KEY = /Mac|iPhone|iPad/.test(navigator.platform) ? "⌘" : "Ctrl";

document.addEventListener("keydown", (event) => {
  if (event.key === "Escape" && activePopover) {
    closePopover();
    return;
  }
  if (document.querySelector("dialog[open]") || !document.getElementById("onboarding").hidden) {
    return;
  }
  const mod = event.ctrlKey || event.metaKey;
  if (mod && !event.shiftKey && !event.altKey) {
    if (event.key === "k") {
      event.preventDefault();
      focusGlobalSearch();
      return;
    }
    if (event.key === ",") {
      event.preventDefault();
      navigate("settings");
      return;
    }
    const index = Number(event.key) - 1;
    if (index >= 0 && index < NAV_ORDER.length) {
      event.preventDefault();
      navigate(NAV_ORDER[index]);
      return;
    }
  }
  currentPage().keys?.(event);
});
