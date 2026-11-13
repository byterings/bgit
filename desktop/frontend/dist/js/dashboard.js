// Dashboard: where things stand and what to do next.
(() => {
  const RECENT_LIMIT = 6;

  function quickActions() {
    return `
      <div class="cluster">
        <button type="button" class="btn" data-action="navigate" data-page="identities">${icon("users")}Manage identities</button>
        <button type="button" class="btn" data-action="add-repo">${icon("plus")}Add existing</button>
        <button type="button" class="btn btn-primary" data-action="clone-repo">${icon("download")}Clone repository</button>
      </div>`;
  }

  function heading() {
    const active = activeIdentity();
    const firstName = (active?.name || "").split(/\s+/)[0];
    const count = repositories().length;
    const sub = active
      ? `Default identity: ${active.alias} (${active.email}). ${count > 0 ? `${plural(count, "repository is", "repositories are")} managed by bgit.` : ""}`
      : "Choose a default identity to start.";
    return `
      <header class="page-head">
        <div class="page-title">
          <h1>${firstName ? `Welcome back, ${escapeHtml(firstName)}` : "Welcome to bgit"}</h1>
          <p>${escapeHtml(sub)}</p>
        </div>
        ${quickActions()}
      </header>`;
  }

  // Things that need a decision, most important first.
  function attention() {
    const items = [];
    if (state.onboarding && !state.onboarding.pushProtection) {
      items.push(
        banner({
          tone: "warning",
          iconName: "shield-off",
          title: "Push protection is off",
          text: "Pushes are not checked against the repository identity right now.",
          actions: `<button type="button" class="btn btn-sm" data-action="navigate" data-page="settings" data-section="protection">Open Settings</button>`,
        }),
      );
    }
    const needing = state.repos?.needAttention || 0;
    if (needing > 0) {
      const fixable = repositories().filter((repo) => repo.canFix).length;
      items.push(
        banner({
          tone: "warning",
          title: `${plural(needing, "repository needs", "repositories need")} attention`,
          text: "Pushes are blocked until a repository matches its identity.",
          actions: `
            <button type="button" class="btn btn-sm" data-action="navigate" data-page="repositories" data-filter="attention">Review</button>
            ${fixable > 0 ? `<button type="button" class="btn btn-sm btn-primary" data-action="fix-all-repos">${icon("wrench", "sm")}Fix ${fixable === 1 ? "it" : `all ${fixable}`}</button>` : ""}`,
        }),
      );
    }
    return items.join("");
  }

  function stat(label, value, filter, tone = "") {
    return `
      <button type="button" class="stat ${tone}" data-action="navigate" data-page="repositories" data-filter="${filter}">
        <span class="stat-value">${value}</span>
        <span class="stat-label">${escapeHtml(label)}</span>
      </button>`;
  }

  function stats() {
    const repos = repositories();
    const changed = repos.filter((repo) => repo.changes > 0).length;
    const toSync = repos.filter((repo) => repo.ahead > 0 || repo.behind > 0).length;
    const attentionCount = state.repos?.needAttention || 0;
    return `
      <section class="stat-row" aria-label="Repository summary">
        ${stat("Repositories", repos.length, "all")}
        ${stat("With uncommitted changes", changed, "changes", changed > 0 ? "is-accent" : "")}
        ${stat("With commits to push or pull", toSync, "sync", toSync > 0 ? "is-info" : "")}
        ${stat("Needing attention", attentionCount, "attention", attentionCount > 0 ? "is-warning" : "")}
      </section>`;
  }

  function recentRow(repo) {
    const opened = recentRepositories.openedAt(repo.path);
    const health = repo.health === "ok" ? "" : `<span class="dot is-${repo.health === "attention" ? "warning" : "danger"}" data-tip="${escapeHtml((repo.problems || [])[0] || "Needs attention")}" data-tip-align="start"></span>`;
    return `
      <button type="button" class="recent-row" data-action="open-repo" data-path="${escapeHtml(repo.path)}" ${repo.exists ? "" : "disabled"}>
        <span class="recent-name">
          <span class="recent-title">${health}<strong class="truncate">${escapeHtml(repo.name)}</strong></span>
          <span class="mono faint truncate-start"><span>${escapeHtml(repo.path)}</span></span>
        </span>
        <span class="recent-branch truncate">${repo.exists ? `${icon("branch", "sm")}<span class="truncate">${escapeHtml(repo.branch || "No branch")}</span>` : ""}</span>
        <span>${identityChip(repo.alias)}</span>
        <span class="recent-sync ${repo.changes > 0 ? "" : "muted"}">${repo.changes > 0 ? plural(repo.changes, "change", "changes") : escapeHtml(syncSummary(repo))}</span>
        <span class="faint recent-opened">${opened ? `Opened ${timeAgo(opened)}` : ""}</span>
      </button>`;
  }

  function recent() {
    const repos = [...repositories()].sort(
      (a, b) => recentRepositories.openedAt(b.path) - recentRepositories.openedAt(a.path) || a.name.localeCompare(b.name),
    );
    return `
      <section class="card">
        <div class="card-head">
          <h2>${repos.some((repo) => recentRepositories.openedAt(repo.path)) ? "Recent repositories" : "Repositories"}</h2>
          <button type="button" class="btn btn-sm btn-ghost" data-action="navigate" data-page="repositories">View all${icon("chevron-right", "sm")}</button>
        </div>
        <div class="recent-list">${repos.slice(0, RECENT_LIMIT).map(recentRow).join("")}</div>
      </section>`;
  }

  function identitiesCard() {
    const counts = {};
    for (const repo of repositories()) {
      counts[repo.alias] = (counts[repo.alias] || 0) + 1;
    }
    return `
      <section class="card">
        <div class="card-head">
          <h2>Identities</h2>
          <button type="button" class="btn btn-sm btn-ghost" data-action="navigate" data-page="identities">Manage${icon("chevron-right", "sm")}</button>
        </div>
        <div class="identity-mini-list">
          ${identities()
            .map(
              (identity) => `
                <button type="button" class="identity-mini" data-action="navigate" data-page="identities" data-alias="${escapeHtml(identity.alias)}">
                  ${avatar(identity)}
                  <span class="identity-mini-text">
                    <span class="identity-mini-title"><strong class="truncate">${escapeHtml(identity.alias)}</strong>${identity.active ? badge("Default", "accent") : ""}</span>
                    <span class="muted truncate">${escapeHtml(identity.email)}</span>
                  </span>
                  <span class="faint">${plural(counts[identity.alias] || 0, "repo", "repos")}</span>
                </button>`,
            )
            .join("")}
        </div>
      </section>`;
  }

  // Health at a glance, from the last Doctor run and push protection. Every
  // number here comes from a real check.
  function healthCard() {
    const doctor = state.doctor;
    const protectedNow = Boolean(state.onboarding?.pushProtection);
    const checks = (doctor?.sections || []).flatMap((section) => section.checks || []);
    const failed = doctor?.errors || 0;
    const warned = doctor?.warnings || 0;
    const passed = Math.max(0, checks.length - failed - warned);
    const tone = failed > 0 ? "danger" : warned > 0 ? "warning" : "success";
    const summary = !doctor ? "Not checked yet" : failed > 0 ? plural(failed, "problem", "problems") : warned > 0 ? plural(warned, "warning", "warnings") : "Everything looks good";
    return `
      <section class="card">
        <div class="card-head">
          <h2>Health</h2>
          <button type="button" class="btn btn-sm btn-ghost" data-action="navigate" data-page="doctor">Doctor${icon("chevron-right", "sm")}</button>
        </div>
        <div class="health-rows">
          <button type="button" class="health-row" data-action="navigate" data-page="settings" data-section="protection">
            <span class="check-icon is-${protectedNow ? "success" : "warning"}">${icon(protectedNow ? "shield-check" : "shield-off")}</span>
            <span class="health-text"><strong>Push protection</strong><span class="muted">${protectedNow ? "Every push is checked" : "Pushes are not checked"}</span></span>
          </button>
          <button type="button" class="health-row" data-action="navigate" data-page="doctor">
            <span class="check-icon is-${doctor ? tone : ""}">${icon(tone === "success" ? "check" : "alert")}</span>
            <span class="health-text"><strong>${escapeHtml(summary)}</strong><span class="muted">${doctor ? `${plural(passed, "check", "checks")} passed${state.doctorRunAt ? ` · ${timeAgo(state.doctorRunAt)}` : ""}` : "Open Doctor to run the checks"}</span></span>
          </button>
        </div>
      </section>`;
  }

  // First use: no repositories yet. Shows the way in instead of empty numbers.
  function getStarted() {
    const active = activeIdentity();
    return `
      <section class="card start-card">
        <div class="start-copy">
          <span class="eyebrow">Get started</span>
          <h2>Bring in your first repository</h2>
          <p>bgit links each repository to an identity, so its commits and pushes always use the right name, email, and SSH key.</p>
          <div class="cluster">
            <button type="button" class="btn btn-primary btn-lg" data-action="clone-repo">${icon("download")}Clone from GitHub</button>
            <button type="button" class="btn btn-lg" data-action="add-repo">${icon("folder")}Add a repository on this computer</button>
          </div>
        </div>
        <ol class="start-steps">
          <li class="done">${icon("check")}<div><strong>Identity ready</strong><span>${active ? `${escapeHtml(active.alias)} · ${escapeHtml(active.email)}` : "Add an identity first"}</span></div></li>
          <li>${icon("repo")}<div><strong>Add a repository</strong><span>Clone it, or pick a folder you already have.</span></div></li>
          <li>${icon("upload")}<div><strong>Commit and push</strong><span>bgit checks every push before it leaves this computer.</span></div></li>
        </ol>
      </section>`;
  }

  function page() {
    if (identities().length === 0) {
      return `<div class="page">${heading()}${emptyState({
        iconName: "users",
        title: "Add your first identity",
        text: "An identity is a Git name, email, and SSH key. Repositories are linked to one.",
        actions: `<button type="button" class="btn btn-primary" data-action="add-identity">${icon("plus")}Add identity</button>`,
      })}</div>`;
    }
    if (!state.repos) {
      return `<div class="page">${heading()}${
        state.reposError
          ? banner({ tone: "danger", title: "Repositories could not be loaded", text: state.reposError, actions: `<button type="button" class="btn btn-sm" data-action="retry-repos">${icon("refresh", "sm")}Try again</button>` })
          : `<section class="card">${skeletonRows(4)}</section>`
      }</div>`;
    }
    if (repositories().length === 0) {
      return `<div class="page">${heading()}${attention()}<div class="dash-grid">${getStarted()}<aside class="stack">${identitiesCard()}${healthCard()}</aside></div></div>`;
    }
    return `
      <div class="page">
        ${heading()}
        ${attention()}
        ${stats()}
        <div class="dash-grid">
          ${recent()}
          <aside class="stack">${identitiesCard()}${healthCard()}</aside>
        </div>
      </div>`;
  }

  registerPage({ id: "dashboard", render: page });
})();
