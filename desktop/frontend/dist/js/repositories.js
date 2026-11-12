// Repositories page: every repository bgit manages, with its identity and
// health, plus clone, add existing, change identity, fix, and remove.
//
// Uses the shared helpers defined in index.html.
(() => {
  const state = {
    list: null,
    loading: false,
    error: "",
    // Reload on the next render (set when the page is opened or data changed).
    stale: true,
  };

  const healthLabels = {
    ok: { text: "Ready", tone: "good" },
    attention: { text: "Needs fix", tone: "warn" },
    blocked: { text: "Blocked", tone: "error" },
    missing: { text: "Missing", tone: "error" },
  };

  function identities() {
    return currentStatus?.identities || [];
  }

  function identityFor(alias) {
    return identities().find((identity) => identity.alias === alias);
  }

  function identityOptions(selected) {
    return identities()
      .map(
        (identity) =>
          `<option value="${escapeHtml(identity.alias)}" ${identity.alias === selected ? "selected" : ""}>${escapeHtml(identity.alias)} (${escapeHtml(identity.email)})</option>`,
      )
      .join("");
  }

  // ---------- Loading ----------

  async function loadList() {
    if (state.loading) {
      return;
    }
    state.loading = true;
    state.stale = false;
    try {
      state.list = await api().ListRepositories();
      state.error = "";
    } catch (error) {
      state.error = String(error);
    } finally {
      state.loading = false;
      if (currentPage === "repositories") {
        render(currentStatus, currentDoctor);
      }
    }
  }

  // Other modules call this after they change repositories.
  window.invalidateRepositories = () => {
    state.stale = true;
  };

  // ---------- Rendering ----------

  function page() {
    const actions = `
      <button type="button" data-repos="add"><span class="icon" aria-hidden="true">${iconSvg("plus")}</span>Add existing</button>
      <button type="button" class="primary" data-repos="clone"><span class="icon" aria-hidden="true">${iconSvg("download")}</span>Clone repository</button>
    `;
    return `
      ${pageHeading("Repositories", "Each repository commits and pushes with the identity shown here.", actions)}
      ${body()}
    `;
  }

  function body() {
    if (identities().length === 0) {
      return `<section class="panel empty"><p>Add an identity first. Repositories are linked to an identity.</p></section>`;
    }
    if (state.error) {
      return `<section class="panel"><div class="note error">${escapeHtml(state.error)}</div></section>`;
    }
    if (!state.list) {
      return `<section class="panel"><p class="muted"><span class="spinner" aria-hidden="true"></span> Loading repositories</p></section>`;
    }

    const repositories = state.list.repositories || [];
    if (repositories.length === 0) {
      return `
        <section class="panel empty">
          <div class="empty-stack">
            <p>No repositories yet. Clone one from GitHub, or add one that is already on this computer.</p>
            <div class="actions">
              <button type="button" data-repos="add">Add existing</button>
              <button type="button" class="primary" data-repos="clone">Clone repository</button>
            </div>
          </div>
        </section>
      `;
    }

    const fixable = repositories.filter((repo) => repo.canFix).length;
    return `
      ${
        state.list.needAttention > 0
          ? `<div class="note warn repo-banner">
              <span><strong>${state.list.needAttention === 1 ? "1 repository needs attention" : `${state.list.needAttention} repositories need attention`}</strong>Pushes are blocked until a repository matches its identity.</span>
              ${fixable > 0 ? `<button type="button" data-repos="fix-all"><span class="icon" aria-hidden="true">${iconSvg("wrench")}</span>Fix ${fixable === 1 ? "it" : `all ${fixable}`}</button>` : ""}
            </div>`
          : ""
      }
      <section class="repo-list">
        ${repositories.map(row).join("")}
      </section>
    `;
  }

  function changeSummary(repo) {
    const parts = [];
    if (repo.changes > 0) {
      parts.push(`${repo.changes} change${repo.changes === 1 ? "" : "s"}`);
    }
    if (repo.ahead > 0) {
      parts.push(`${repo.ahead} to push`);
    }
    if (repo.behind > 0) {
      parts.push(`${repo.behind} to pull`);
    }
    return parts.join(" · ") || "Up to date";
  }

  function row(repo) {
    const health = healthLabels[repo.health] || healthLabels.blocked;
    const identity = identityFor(repo.alias);
    const problems = [...(repo.problems || [])];
    const notes = repo.health === "ok" ? repo.notes || [] : [];
    return `
      <article class="repo-record ${escapeHtml(repo.health)}">
        <div class="repo-main">
          <div class="repo-title">
            <strong class="truncate" title="${escapeHtml(repo.name)}">${escapeHtml(repo.name)}</strong>
            <span class="badge ${health.tone}">${health.text}</span>
          </div>
          <span class="muted truncate repo-location" title="${escapeHtml(repo.path)}">${escapeHtml(repo.path)}</span>
        </div>
        <div class="repo-identity">
          ${identity ? avatarBlock(identity) : `<span class="avatar" aria-hidden="true">?</span>`}
          <div class="identity-details">
            <strong class="truncate">${escapeHtml(repo.alias || "No identity")}</strong>
            <span class="muted truncate">${repo.source === "workspace" ? "From workspace" : "Bound"}</span>
          </div>
        </div>
        <div class="repo-branch">
          ${
            repo.exists
              ? `<span class="truncate" title="${escapeHtml(repo.branch || "")}"><span class="icon" aria-hidden="true">${iconSvg("branch")}</span>${escapeHtml(repo.branch || "No branch")}</span>
                 <span class="muted">${escapeHtml(changeSummary(repo))}</span>`
              : `<span class="muted">Folder not found</span>`
          }
        </div>
        <div class="actions repo-actions">
          ${repo.canFix ? iconButton({ action: "repos-fix", icon: "wrench", label: `Fix ${repo.name}`, path: repo.path }) : ""}
          ${repo.exists ? iconButton({ action: "repos-identity", icon: "users", label: `Change identity of ${repo.name}`, path: repo.path, alias: repo.alias }) : ""}
          ${repo.exists ? iconButton({ action: "repos-reveal", icon: "folder", label: `Show ${repo.name} in file manager`, path: repo.path }) : ""}
          ${repo.source === "binding" ? iconButton({ action: "repos-remove", icon: "trash", label: `Remove ${repo.name} from bgit`, path: repo.path, alias: repo.alias, danger: true }) : ""}
          ${repo.exists ? `<button type="button" class="primary" data-action="repos-open" data-path="${escapeHtml(repo.path)}">Open</button>` : ""}
        </div>
        ${
          problems.length > 0 || notes.length > 0
            ? `<ul class="repo-problems">
                ${problems.map((problem) => `<li>${escapeHtml(problem)}</li>`).join("")}
                ${notes.map((note) => `<li class="muted">${escapeHtml(note)}</li>`).join("")}
              </ul>`
            : ""
        }
      </article>
    `;
  }

  // ---------- Dialogs ----------

  function ensureDialogs() {
    if (document.getElementById("cloneDialog")) {
      return;
    }
    document.body.insertAdjacentHTML(
      "beforeend",
      `
      <dialog id="cloneDialog">
        <form id="cloneForm" class="dialog-body" method="dialog" novalidate>
          <h2>Clone a repository</h2>
          <div class="form-stack">
            <label>
              Repository address
              <input name="url" placeholder="https://github.com/owner/repository" autocomplete="off" spellcheck="false" required />
              <span class="field-hint">Paste the link from GitHub, or type owner/repository.</span>
            </label>
            <label>
              Identity
              <select name="alias" required></select>
              <span class="field-hint">The clone is linked to this identity and uses its SSH key.</span>
            </label>
            <div class="field-group">
              <span>Clone into</span>
              <div class="field-control">
                <input name="parentDir" placeholder="Choose a folder" required />
                <button type="button" data-clone="choose" title="Choose folder" aria-label="Choose folder">
                  <span class="icon" aria-hidden="true">${iconSvg("folder")}</span>
                </button>
              </div>
              <span class="field-hint" id="cloneTarget"></span>
            </div>
            <div id="cloneError" class="note error" role="alert" hidden></div>
          </div>
          <div class="dialog-actions">
            <button type="button" data-clone="cancel">Cancel</button>
            <button type="submit" class="primary" id="cloneSubmit">Clone</button>
          </div>
        </form>
      </dialog>

      <dialog id="addRepoDialog">
        <form id="addRepoForm" class="dialog-body" method="dialog" novalidate>
          <h2 id="addRepoTitle">Add an existing repository</h2>
          <div class="form-stack">
            <div class="field-group">
              <span>Repository folder</span>
              <div class="field-control">
                <input name="path" placeholder="Choose the repository folder" required />
                <button type="button" data-addrepo="choose" title="Choose repository folder" aria-label="Choose repository folder">
                  <span class="icon" aria-hidden="true">${iconSvg("folder")}</span>
                </button>
              </div>
            </div>
            <label>
              Identity
              <select name="alias" required></select>
            </label>
            <div class="note">bgit sets this repository's commit name and email to the identity, and changes its GitHub remote address to use the identity's SSH key.</div>
            <div id="addRepoError" class="note error" role="alert" hidden></div>
          </div>
          <div class="dialog-actions">
            <button type="button" data-addrepo="cancel">Cancel</button>
            <button type="submit" class="primary" id="addRepoSubmit">Add repository</button>
          </div>
        </form>
      </dialog>
    `,
    );
    bindCloneDialog();
    bindAddDialog();
  }

  function showError(box, message) {
    box.textContent = message;
    box.hidden = !message;
  }

  function setBusy(button, busyText) {
    const previous = button.innerHTML;
    button.disabled = true;
    button.innerHTML = `<span class="spinner" aria-hidden="true"></span>${escapeHtml(busyText)}`;
    return () => {
      button.disabled = false;
      button.innerHTML = previous;
    };
  }

  // The folder a clone will create, shown while the user types.
  function clonePreview(form) {
    const url = form.elements.url.value.trim().replace(/\/+$/, "");
    const parent = form.elements.parentDir.value.trim();
    const name = (url.split(/[/:]/).pop() || "").replace(/\.git$/, "");
    const target = document.getElementById("cloneTarget");
    if (parent && name) {
      const separator = parent.includes("\\") && !parent.includes("/") ? "\\" : "/";
      target.textContent = `Creates ${parent.replace(/[\\/]+$/, "")}${separator}${name}`;
    } else {
      target.textContent = "The repository gets its own folder inside this one.";
    }
  }

  // Suggests the clone folder for the chosen identity, unless the user
  // already picked one themselves.
  async function suggestCloneFolder(form) {
    if (form.dataset.folderEdited === "true") {
      return;
    }
    try {
      const defaults = await api().GetCloneDefaults(form.elements.alias.value);
      const remembered = defaults.fromWorkspace ? "" : rememberedCloneFolder();
      if (form.dataset.folderEdited !== "true") {
        form.elements.parentDir.value = remembered || defaults.parentDir || "";
        clonePreview(form);
      }
    } catch (error) {
      // A missing suggestion only means the user picks the folder.
    }
  }

  function rememberedCloneFolder() {
    try {
      return window.localStorage.getItem("bgit.cloneFolder") || "";
    } catch (error) {
      return "";
    }
  }

  function rememberCloneFolder(folder) {
    try {
      window.localStorage.setItem("bgit.cloneFolder", folder);
    } catch (error) {
      // Not being able to remember the folder is harmless.
    }
  }

  function openCloneDialog() {
    ensureDialogs();
    const form = document.getElementById("cloneForm");
    form.reset();
    form.dataset.folderEdited = "false";
    form.elements.alias.innerHTML = identityOptions(currentStatus?.activeAlias);
    showError(document.getElementById("cloneError"), "");
    clonePreview(form);
    suggestCloneFolder(form);
    document.getElementById("cloneDialog").showModal();
    form.elements.url.focus();
  }

  function bindCloneDialog() {
    const dialog = document.getElementById("cloneDialog");
    const form = document.getElementById("cloneForm");
    const errorBox = document.getElementById("cloneError");

    form.elements.url.addEventListener("input", () => {
      showError(errorBox, "");
      clonePreview(form);
    });
    form.elements.parentDir.addEventListener("input", () => {
      form.dataset.folderEdited = "true";
      clonePreview(form);
    });
    form.elements.alias.addEventListener("change", () => suggestCloneFolder(form));
    dialog.querySelector("[data-clone='cancel']").addEventListener("click", () => dialog.close());
    dialog.querySelector("[data-clone='choose']").addEventListener("click", async () => {
      try {
        const folder = await api().ChooseCloneFolder();
        if (folder) {
          form.elements.parentDir.value = folder;
          form.dataset.folderEdited = "true";
          clonePreview(form);
        }
      } catch (error) {
        showError(errorBox, String(error));
      }
    });

    form.addEventListener("submit", async (event) => {
      event.preventDefault();
      const request = {
        url: form.elements.url.value.trim(),
        alias: form.elements.alias.value,
        parentDir: form.elements.parentDir.value.trim(),
        folderName: "",
      };
      if (!request.url) {
        showError(errorBox, "Enter the repository address.");
        return;
      }
      if (!request.parentDir) {
        showError(errorBox, "Choose the folder to clone into.");
        return;
      }
      showError(errorBox, "");
      const done = setBusy(document.getElementById("cloneSubmit"), "Cloning");
      try {
        const result = await api().CloneRepository(request);
        rememberCloneFolder(request.parentDir);
        dialog.close();
        finishChange(result);
        if (result.repository?.path && typeof openRepository === "function") {
          openRepository(result.repository.path);
        }
      } catch (error) {
        showError(errorBox, String(error));
      } finally {
        done();
      }
    });
  }

  // Opens the add dialog. With a path it becomes "change identity" for that
  // repository.
  function openAddDialog(path = "", alias = "") {
    ensureDialogs();
    const form = document.getElementById("addRepoForm");
    form.reset();
    const changing = Boolean(path);
    document.getElementById("addRepoTitle").textContent = changing ? "Change repository identity" : "Add an existing repository";
    document.getElementById("addRepoSubmit").textContent = changing ? "Change identity" : "Add repository";
    form.elements.path.value = path;
    form.elements.path.readOnly = changing;
    form.querySelector("[data-addrepo='choose']").disabled = changing;
    form.elements.alias.innerHTML = identityOptions(alias || currentStatus?.activeAlias);
    showError(document.getElementById("addRepoError"), "");
    document.getElementById("addRepoDialog").showModal();
  }

  function bindAddDialog() {
    const dialog = document.getElementById("addRepoDialog");
    const form = document.getElementById("addRepoForm");
    const errorBox = document.getElementById("addRepoError");

    dialog.querySelector("[data-addrepo='cancel']").addEventListener("click", () => dialog.close());
    dialog.querySelector("[data-addrepo='choose']").addEventListener("click", async () => {
      try {
        const folder = await api().ChooseRepositoryPath();
        if (folder) {
          form.elements.path.value = folder;
        }
      } catch (error) {
        showError(errorBox, String(error));
      }
    });

    form.addEventListener("submit", async (event) => {
      event.preventDefault();
      const request = { path: form.elements.path.value.trim(), alias: form.elements.alias.value };
      if (!request.path) {
        showError(errorBox, "Choose the repository folder.");
        return;
      }
      showError(errorBox, "");
      const done = setBusy(document.getElementById("addRepoSubmit"), "Saving");
      try {
        const result = await api().AddRepository(request);
        dialog.close();
        finishChange(result);
      } catch (error) {
        showError(errorBox, String(error));
      } finally {
        done();
      }
    });
  }

  // ---------- Actions ----------

  // Applies the result of a change: fresh status, reloaded list, feedback.
  function finishChange(result) {
    state.stale = true;
    if (result?.message) {
      addActivity(result.message, "Repositories", "success");
      notify({ title: "Done", message: result.message, tone: "success" });
    }
    render(result?.status || currentStatus, currentDoctor);
  }

  async function runRowAction(button, busyTitle, action) {
    button.disabled = true;
    setActionText(busyTitle);
    try {
      finishChange(await action());
    } catch (error) {
      notify({ title: "Action failed", message: String(error), tone: "error" });
    } finally {
      button.disabled = false;
      setActionText("");
    }
  }

  async function fixAll(button) {
    const done = setBusy(button, "Fixing");
    try {
      state.list = await api().FixAllRepositories();
      notify({ title: "Done", message: state.list.message || "Repositories fixed", tone: "success" });
      addActivity(state.list.message || "Repositories fixed", "Repositories", "success");
    } catch (error) {
      notify({ title: "Fix failed", message: String(error), tone: "error" });
    } finally {
      done();
      render(currentStatus, currentDoctor);
    }
  }

  async function removeRepository(button) {
    const path = button.dataset.path;
    const confirmed = await confirmAction({
      title: "Remove repository from bgit",
      message: `Stop managing this repository with identity '${button.dataset.alias}'?`,
      warning: "The folder and its files are not deleted. The repository goes back to your default identity.",
      acceptLabel: "Remove",
      danger: true,
    });
    if (confirmed) {
      await runRowAction(button, "Removing repository", () => api().RemoveRepository(path));
    }
  }

  function bind(target) {
    ensureDialogs();
    if (state.stale && identities().length > 0) {
      loadList();
    }

    target.querySelectorAll("[data-repos]").forEach((button) => {
      button.addEventListener("click", () => {
        switch (button.dataset.repos) {
          case "clone":
            openCloneDialog();
            break;
          case "add":
            openAddDialog();
            break;
          case "fix-all":
            fixAll(button);
            break;
          default:
            break;
        }
      });
    });

    target.querySelectorAll("[data-action^='repos-']").forEach((button) => {
      button.addEventListener("click", async () => {
        const path = button.dataset.path;
        switch (button.dataset.action) {
          case "repos-open":
            if (typeof openRepository === "function") {
              openRepository(path);
            }
            break;
          case "repos-fix":
            await runRowAction(button, "Fixing repository", () => api().FixRepository(path));
            break;
          case "repos-identity":
            openAddDialog(path, button.dataset.alias);
            break;
          case "repos-reveal":
            try {
              await api().RevealRepository(path);
            } catch (error) {
              notify({ title: "Could not open folder", message: String(error), tone: "error" });
            }
            break;
          case "repos-remove":
            await removeRepository(button);
            break;
          default:
            break;
        }
      });
    });

    // Opened from the wizard or the dashboard with a specific goal.
    const intent = window.pendingRepositoryIntent;
    if (intent && identities().length > 0) {
      window.pendingRepositoryIntent = "";
      if (intent === "clone") {
        openCloneDialog();
      } else {
        openAddDialog();
      }
    }
  }

  registerPage({
    id: "repositories",
    render: page,
    bind,
    enter() {
      state.stale = true;
    },
  });
})();
