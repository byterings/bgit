// Repositories: every repository bgit manages, with its identity, branch,
// pending work, and health. Also owns the clone and add dialogs, which other
// screens open through global actions.
(() => {
  const view = {
    query: "",
    filter: "all",
    identity: "",
    sort: prefs.get("repoSort", "name"),
  };

  const FILTERS = [
    { id: "all", label: "All", test: () => true },
    { id: "attention", label: "Needs attention", test: (repo) => repo.health !== "ok" },
    { id: "changes", label: "Changes", test: (repo) => repo.changes > 0 },
    { id: "sync", label: "To sync", test: (repo) => repo.ahead > 0 || repo.behind > 0 },
  ];
  const HEALTH = {
    attention: { label: "Needs fix", tone: "warning" },
    blocked: { label: "Blocked", tone: "danger" },
    missing: { label: "Missing", tone: "danger" },
  };

  function identityOptions(selected) {
    return identities()
      .map(
        (identity) =>
          `<option value="${escapeHtml(identity.alias)}" ${identity.alias === selected ? "selected" : ""}>${escapeHtml(identity.alias)} · ${escapeHtml(identity.email)}</option>`,
      )
      .join("");
  }

  // ---------- List ----------

  function visibleRepositories() {
    const needle = view.query.trim().toLowerCase();
    const filter = FILTERS.find((item) => item.id === view.filter) || FILTERS[0];
    const list = repositories().filter(
      (repo) =>
        filter.test(repo) &&
        (!view.identity || repo.alias === view.identity) &&
        (!needle || repo.name.toLowerCase().includes(needle) || repo.path.toLowerCase().includes(needle)),
    );
    const byName = (a, b) => a.name.localeCompare(b.name);
    if (view.sort === "recent") {
      list.sort((a, b) => recentRepositories.openedAt(b.path) - recentRepositories.openedAt(a.path) || byName(a, b));
    } else if (view.sort === "changes") {
      list.sort((a, b) => b.changes - a.changes || byName(a, b));
    } else {
      list.sort(byName);
    }
    return list;
  }

  function toolbar() {
    const repos = repositories();
    return `
      <div class="list-toolbar">
        <label class="search grow">
          ${icon("search")}
          <input class="input" id="repoSearch" type="text" placeholder="Filter by name or folder" value="${escapeHtml(view.query)}" autocomplete="off" spellcheck="false" aria-label="Filter repositories" />
        </label>
        <div class="segmented" role="tablist" aria-label="Show">
          ${FILTERS.map((filter) => {
            const count = repos.filter(filter.test).length;
            return `<button type="button" role="tab" aria-selected="${view.filter === filter.id}" class="${view.filter === filter.id ? "active" : ""}" data-action="set-filter" data-filter="${filter.id}">${filter.label}<span class="n">${count}</span></button>`;
          }).join("")}
        </div>
        <select class="select sm" id="repoIdentity" aria-label="Identity" style="width: 150px">
          <option value="">All identities</option>
          ${identities()
            .map((identity) => `<option value="${escapeHtml(identity.alias)}" ${view.identity === identity.alias ? "selected" : ""}>${escapeHtml(identity.alias)}</option>`)
            .join("")}
        </select>
        <select class="select sm" id="repoSort" aria-label="Sort" style="width: 150px">
          <option value="name" ${view.sort === "name" ? "selected" : ""}>Sort: Name</option>
          <option value="recent" ${view.sort === "recent" ? "selected" : ""}>Sort: Recently opened</option>
          <option value="changes" ${view.sort === "changes" ? "selected" : ""}>Sort: Most changes</option>
        </select>
      </div>`;
  }

  function row(repo) {
    const health = HEALTH[repo.health];
    const opened = recentRepositories.openedAt(repo.path);
    const problems = repo.problems || [];
    return `
      <div class="repo-row ${repo.health !== "ok" ? `is-${health.tone}` : ""}" data-context="repo" data-path="${escapeHtml(repo.path)}">
        <button type="button" class="repo-cell repo-main" data-action="open-repo" data-path="${escapeHtml(repo.path)}" ${repo.exists ? "" : "disabled"}>
          <span class="repo-title"><strong class="truncate">${escapeHtml(repo.name)}</strong>${health ? badge(health.label, health.tone) : ""}</span>
          <span class="mono faint truncate-start"><span>${escapeHtml(repo.path)}</span></span>
        </button>
        <span class="repo-cell truncate">${repo.exists ? `${icon("branch", "sm")}<span class="truncate">${escapeHtml(repo.branch || "No branch")}</span>` : `<span class="faint">Folder not found</span>`}</span>
        <span class="repo-cell">${identityChip(repo.alias)}${repo.source === "workspace" ? `<span class="faint" data-tip="Identity comes from the workspace folder ${escapeHtml(repo.sourcePath || "")}">${icon("layers", "sm")}</span>` : ""}</span>
        <span class="repo-cell ${repo.changes > 0 ? "" : "faint"}">${repo.changes > 0 ? plural(repo.changes, "change", "changes") : "Clean"}</span>
        <span class="repo-cell sync-cell">
          ${repo.ahead > 0 ? `<span data-tip="${plural(repo.ahead, "commit", "commits")} to push">${icon("upload", "sm")}${repo.ahead}</span>` : ""}
          ${repo.behind > 0 ? `<span data-tip="${plural(repo.behind, "commit", "commits")} to pull">${icon("download", "sm")}${repo.behind}</span>` : ""}
          ${repo.ahead === 0 && repo.behind === 0 ? `<span class="faint">${escapeHtml(syncSummary(repo))}</span>` : ""}
        </span>
        <span class="repo-cell faint">${opened ? timeAgo(opened) : "Not opened"}</span>
        <span class="repo-cell repo-actions">
          ${repo.canFix ? `<button type="button" class="btn btn-sm" data-action="fix-repo" data-path="${escapeHtml(repo.path)}">${icon("wrench", "sm")}Fix</button>` : ""}
          <button type="button" class="icon-btn sm" data-action="repo-menu" data-path="${escapeHtml(repo.path)}" aria-label="Actions for ${escapeHtml(repo.name)}" aria-haspopup="menu">${icon("more")}</button>
        </span>
        ${
          problems.length > 0
            ? `<ul class="repo-problems">${problems.map((problem) => `<li>${icon("alert", "sm")}<span>${escapeHtml(problem)}</span></li>`).join("")}</ul>`
            : ""
        }
      </div>`;
  }

  function list() {
    if (state.reposError) {
      return banner({ tone: "danger", title: "Repositories could not be loaded", text: state.reposError, actions: `<button type="button" class="btn btn-sm" data-action="retry-repos">${icon("refresh", "sm")}Try again</button>` });
    }
    if (!state.repos) {
      return `<section class="card">${skeletonRows(5)}</section>`;
    }
    if (repositories().length === 0) {
      return `<section class="card">${emptyState({
        iconName: "repo",
        title: "No repositories yet",
        text: "Clone one from GitHub, or add one that is already on this computer. bgit keeps each on the identity you choose.",
        actions: `<button type="button" class="btn" data-action="add-repo">${icon("plus")}Add existing</button><button type="button" class="btn btn-primary" data-action="clone-repo">${icon("download")}Clone repository</button>`,
      })}</section>`;
    }
    const visible = visibleRepositories();
    const fixable = repositories().filter((repo) => repo.canFix).length;
    return `
      ${
        state.repos.needAttention > 0
          ? banner({
              tone: "warning",
              title: `${plural(state.repos.needAttention, "repository needs", "repositories need")} attention`,
              text: "Pushes are blocked until a repository matches its identity.",
              actions: fixable > 0 ? `<button type="button" class="btn btn-sm btn-primary" data-action="fix-all-repos">${icon("wrench", "sm")}Fix ${fixable === 1 ? "it" : `all ${fixable}`}</button>` : "",
            })
          : ""
      }
      ${toolbar()}
      <section class="card repo-table">
        <div class="repo-row repo-head">
          <span class="eyebrow">Repository</span><span class="eyebrow">Branch</span><span class="eyebrow">Identity</span><span class="eyebrow">Changes</span><span class="eyebrow">Sync</span><span class="eyebrow">Opened</span><span></span>
        </div>
        <div id="repoRows">
          ${visible.length > 0 ? visible.map(row).join("") : emptyState({ iconName: "search", title: "Nothing matches", text: "Try another name, or clear the filters.", actions: `<button type="button" class="btn btn-sm" data-action="clear-filters">Clear filters</button>`, compact: true })}
        </div>
      </section>`;
  }

  function page() {
    return `
      <div class="page">
        <header class="page-head">
          <div class="page-title">
            <h1>Repositories</h1>
            <p>Each repository commits and pushes with the identity shown next to it.</p>
          </div>
          <div class="cluster">
            <button type="button" class="btn" data-action="add-repo" ${identities().length ? "" : "disabled"}>${icon("plus")}Add existing</button>
            <button type="button" class="btn btn-primary" data-action="clone-repo" ${identities().length ? "" : "disabled"}>${icon("download")}Clone repository</button>
          </div>
        </header>
        ${identities().length === 0 ? emptyState({ iconName: "users", title: "Add an identity first", text: "Repositories are linked to an identity.", actions: `<button type="button" class="btn btn-primary" data-action="add-identity">${icon("plus")}Add identity</button>` }) : list()}
      </div>`;
  }

  // Re-draws only the rows while typing, so the filter box keeps focus.
  function redrawRows() {
    const rows = document.getElementById("repoRows");
    if (!rows) {
      render();
      return;
    }
    const visible = visibleRepositories();
    rows.innerHTML = visible.length > 0 ? visible.map(row).join("") : emptyState({ iconName: "search", title: "Nothing matches", text: "Try another name, or clear the filters.", actions: `<button type="button" class="btn btn-sm" data-action="clear-filters">Clear filters</button>`, compact: true });
  }

  function repoMenuItems(repo) {
    return [
      { label: "Open", icon: "repo", disabled: !repo.exists, onSelect: () => openRepository(repo.path) },
      repo.canFix ? { label: "Fix repository", icon: "wrench", onSelect: () => fixRepository(repo.path) } : null,
      { label: "Change identity", icon: "users", disabled: !repo.exists, onSelect: () => openAddRepositoryDialog(repo) },
      { separator: true },
      { label: "Show in file manager", icon: "folder", disabled: !repo.exists, onSelect: () => revealRepository(repo.path) },
      { label: "Copy path", icon: "copy", onSelect: () => copyText(repo.path, "Path copied") },
      repo.source === "binding" ? { separator: true } : null,
      repo.source === "binding" ? { label: "Remove from bgit", icon: "trash", danger: true, onSelect: () => removeRepository(repo) } : null,
    ];
  }

  registerPage({
    id: "repositories",
    render: page,
    enter(params) {
      if (params.filter) {
        view.filter = params.filter;
        view.query = "";
        view.identity = "";
      }
      loadRepositories();
    },
    focus() {
      loadRepositories();
    },
    actions: {
      "set-filter"(element) {
        view.filter = element.dataset.filter;
        render();
      },
      "clear-filters"() {
        Object.assign(view, { query: "", filter: "all", identity: "" });
        render();
      },
      "repo-menu"(element) {
        const repo = repositoryByPath(element.dataset.path);
        if (repo) {
          openMenu(element, repoMenuItems(repo), { align: "end" });
        }
      },
    },
    on: {
      input(target) {
        if (target.id === "repoSearch") {
          view.query = target.value;
          redrawRows();
        }
      },
      change(target) {
        if (target.id === "repoIdentity") {
          view.identity = target.value;
          render();
        }
        if (target.id === "repoSort") {
          view.sort = target.value;
          prefs.set("repoSort", view.sort);
          render();
        }
      },
    },
    context(element, event) {
      const repo = repositoryByPath(element.dataset.path);
      if (repo) {
        openContextMenu(event, repoMenuItems(repo));
      }
    },
    keys(event) {
      if (event.key === "/" && !["INPUT", "TEXTAREA", "SELECT"].includes(document.activeElement?.tagName)) {
        event.preventDefault();
        document.getElementById("repoSearch")?.focus();
      }
    },
  });

  // ---------- Actions shared with other screens ----------

  async function fixRepository(path, button) {
    try {
      const result = await withBusy(button, "Fixing", () => api().FixRepository(path));
      finishAction(result, { activity: "Repositories" });
      return true;
    } catch (error) {
      notify({ title: "Could not fix the repository", message: errorText(error), tone: "danger" });
      return false;
    }
  }
  window.fixRepository = fixRepository;

  async function revealRepository(path) {
    try {
      await api().RevealRepository(path);
    } catch (error) {
      notify({ title: "Could not open the folder", message: errorText(error), tone: "danger" });
    }
  }
  window.revealRepository = revealRepository;

  async function removeRepository(repo) {
    const confirmed = await confirmAction({
      title: `Remove ${repo.name} from bgit?`,
      message: `bgit stops managing this repository with identity '${repo.alias}'.`,
      warning: "The folder and its files are not deleted. The repository goes back to your default identity.",
      acceptLabel: "Remove",
      danger: true,
    });
    if (!confirmed) {
      return false;
    }
    try {
      finishAction(await api().RemoveRepository(repo.path), { activity: "Repositories" });
      return true;
    } catch (error) {
      notify({ title: "Could not remove the repository", message: errorText(error), tone: "danger" });
      return false;
    }
  }
  window.removeRepository = removeRepository;

  registerActions({
    "open-repo"(element) {
      openRepository(element.dataset.path);
    },
    "clone-repo"(element) {
      openCloneDialog({ alias: element.dataset.alias || "" });
    },
    "add-repo"() {
      openAddRepositoryDialog();
    },
    "fix-repo"(element) {
      fixRepository(element.dataset.path, element);
    },
    async "fix-all-repos"(element) {
      try {
        const result = await withBusy(element, "Fixing", () => api().FixAllRepositories());
        state.repos = result;
        addActivity(result.message || "Repositories fixed", "Repositories");
        notify({ title: result.message || "Repositories fixed" });
      } catch (error) {
        notify({ title: "Could not fix repositories", message: errorText(error), tone: "danger" });
      }
      render();
    },
  });

  // ---------- Clone dialog ----------

  function futureCommitsLine(alias) {
    const identity = identityByAlias(alias);
    return identity ? `Commits here will be recorded as ${identity.name} <${identity.email}>.` : "";
  }

  // preset.alias picks the identity up front (for example from a workspace).
  window.openCloneDialog = async (preset = {}) => {
    if (identities().length === 0) {
      notify({ title: "Add an identity first", message: "A clone is linked to an identity.", tone: "warning" });
      return;
    }
    let folderEdited = false;
    const result = await openDialog({
      id: "cloneDialog",
      title: "Clone a repository",
      description: "The clone is linked to an identity and uses that identity's SSH key.",
      submitLabel: "Clone",
      busyLabel: "Cloning",
      body: `
        <div class="field">
          <label for="cloneUrl">Repository address</label>
          <input class="input mono" id="cloneUrl" name="url" placeholder="https://github.com/owner/repository" autocomplete="off" spellcheck="false" autofocus />
          <span class="field-hint">Paste the link from GitHub, or type owner/repository.</span>
        </div>
        <div class="field">
          <label for="cloneAlias">Identity</label>
          <select class="select" id="cloneAlias" name="alias">${identityOptions(preset.alias || state.status?.activeAlias)}</select>
          <span class="field-hint" data-future></span>
        </div>
        <div class="field">
          <label for="cloneFolder">Clone into</label>
          <div class="input-group">
            <input class="input mono" id="cloneFolder" name="parentDir" placeholder="Choose a folder" autocomplete="off" spellcheck="false" />
            <button type="button" class="btn" data-browse>${icon("folder")}Browse</button>
          </div>
          <span class="field-hint" data-target></span>
        </div>`,
      onOpen({ form, setError }) {
        const preview = () => {
          const url = form.elements.url.value.trim().replace(/\/+$/, "");
          const parent = form.elements.parentDir.value.trim();
          const name = (url.split(/[/:]/).pop() || "").replace(/\.git$/, "");
          const separator = parent.includes("\\") && !parent.includes("/") ? "\\" : "/";
          form.querySelector("[data-target]").textContent =
            parent && name ? `Creates ${parent.replace(/[\\/]+$/, "")}${separator}${name}` : "The repository gets its own folder inside this one.";
          form.querySelector("[data-future]").textContent = futureCommitsLine(form.elements.alias.value);
        };
        // Suggest the identity's workspace folder (or the last folder used),
        // unless the user already picked one.
        const suggest = async () => {
          try {
            const defaults = await api().GetCloneDefaults(form.elements.alias.value);
            if (!folderEdited) {
              form.elements.parentDir.value = (defaults.fromWorkspace ? "" : prefs.get("cloneFolder", "")) || defaults.parentDir || "";
              preview();
            }
          } catch (error) {
            // Without a suggestion the user picks the folder.
          }
        };
        form.elements.url.addEventListener("input", preview);
        form.elements.parentDir.addEventListener("input", () => {
          folderEdited = true;
          preview();
        });
        form.elements.alias.addEventListener("change", () => {
          preview();
          suggest();
        });
        form.querySelector("[data-browse]").addEventListener("click", async () => {
          try {
            const folder = await api().ChooseCloneFolder();
            if (folder) {
              form.elements.parentDir.value = folder;
              folderEdited = true;
              preview();
            }
          } catch (error) {
            setError(errorText(error));
          }
        });
        preview();
        suggest();
      },
      async onSubmit(form) {
        const request = {
          url: form.elements.url.value.trim(),
          alias: form.elements.alias.value,
          parentDir: form.elements.parentDir.value.trim(),
          folderName: "",
        };
        if (!request.url) throw new Error("Enter the repository address.");
        if (!request.parentDir) throw new Error("Choose the folder to clone into.");
        const cloned = await api().CloneRepository(request);
        prefs.set("cloneFolder", request.parentDir);
        return cloned;
      },
    });
    if (result?.repository) {
      finishAction(result, { activity: "Repositories" });
      openRepository(result.repository.path, result.repository);
    }
  };

  // ---------- Add existing / change identity dialog ----------

  // Without a repository this adds one; with one it changes its identity.
  window.openAddRepositoryDialog = async (repo = null) => {
    if (identities().length === 0) {
      notify({ title: "Add an identity first", message: "A repository is linked to an identity.", tone: "warning" });
      return null;
    }
    const changing = Boolean(repo);
    const result = await openDialog({
      id: "addRepoDialog",
      title: changing ? `Change the identity of ${repo.name}` : "Add an existing repository",
      description: changing
        ? "The repository's commit name, email, and GitHub remote address change to the identity you choose."
        : "bgit sets the repository's commit name and email to the identity, and routes its GitHub remote through the identity's SSH key.",
      submitLabel: changing ? "Change identity" : "Add repository",
      busyLabel: "Saving",
      body: `
        <div class="field">
          <label for="addRepoPath">Repository folder</label>
          <div class="input-group">
            <input class="input mono" id="addRepoPath" name="path" value="${escapeHtml(repo?.path || "")}" placeholder="Choose the repository folder" autocomplete="off" spellcheck="false" ${changing ? "readonly" : "autofocus"} />
            <button type="button" class="btn" data-browse ${changing ? "disabled" : ""}>${icon("folder")}Browse</button>
          </div>
        </div>
        <div class="field">
          <label for="addRepoAlias">Identity</label>
          <select class="select" id="addRepoAlias" name="alias" ${changing ? "autofocus" : ""}>${identityOptions(repo?.alias || state.status?.activeAlias)}</select>
          <span class="field-hint" data-future></span>
        </div>`,
      onOpen({ form, setError }) {
        const future = () => {
          form.querySelector("[data-future]").textContent = futureCommitsLine(form.elements.alias.value);
        };
        form.elements.alias.addEventListener("change", future);
        form.querySelector("[data-browse]").addEventListener("click", async () => {
          try {
            const folder = await api().ChooseRepositoryPath();
            if (folder) {
              form.elements.path.value = folder;
            }
          } catch (error) {
            setError(errorText(error));
          }
        });
        future();
      },
      async onSubmit(form) {
        const request = { path: form.elements.path.value.trim(), alias: form.elements.alias.value };
        if (!request.path) throw new Error("Choose the repository folder.");
        return api().AddRepository(request);
      },
    });
    if (result?.repository) {
      finishAction(result, { activity: "Repositories" });
    }
    return result?.repository || null;
  };
})();
