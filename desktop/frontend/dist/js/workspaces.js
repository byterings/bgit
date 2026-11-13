// Workspaces: folders where every repository inside uses one identity.
(() => {
  // Repositories that live inside a workspace folder, bound or not.
  function reposIn(workspace) {
    const prefix = workspace.path.replace(/[\\/]+$/, "");
    return repositories().filter((repo) => repo.path.startsWith(`${prefix}/`) || repo.path.startsWith(`${prefix}\\`));
  }

  function card(workspace) {
    const repos = reposIn(workspace);
    const problem = !workspace.exists ? "This folder no longer exists." : workspace.missingIdentity ? `Identity '${workspace.alias}' no longer exists.` : "";
    return `
      <section class="card workspace-card">
        <div class="card-head">
          <div class="workspace-title">
            <span class="workspace-icon">${icon("layers")}</span>
            <div class="grow">
              <h2 class="truncate">${escapeHtml(baseName(workspace.path))}</h2>
              <span class="mono faint truncate-start"><span>${escapeHtml(workspace.path)}</span></span>
            </div>
          </div>
          <div class="cluster">
            ${identityChip(workspace.alias)}
            <button type="button" class="icon-btn sm" data-action="workspace-menu" data-path="${escapeHtml(workspace.path)}" data-alias="${escapeHtml(workspace.alias)}" aria-label="Workspace actions" aria-haspopup="menu">${icon("more")}</button>
          </div>
        </div>
        ${problem ? `<div class="card-body">${banner({ tone: "danger", title: problem, text: "Remove this workspace, or restore what is missing." })}</div>` : ""}
        ${
          workspace.exists
            ? repos.length > 0
              ? `<div class="mini-list padded">
                  ${repos
                    .map(
                      (repo) => `
                        <button type="button" class="mini-row" data-action="open-repo" data-path="${escapeHtml(repo.path)}">
                          ${icon("repo", "sm")}<strong class="truncate">${escapeHtml(repo.name)}</strong>
                          <span class="faint truncate">${icon("branch", "sm")}${escapeHtml(repo.branch || "No branch")}</span>
                          ${repo.alias !== workspace.alias ? `<span class="faint">bound to ${escapeHtml(repo.alias)}</span>` : ""}
                          ${repo.health === "ok" ? "" : badge(repo.health === "attention" ? "Needs fix" : "Blocked", repo.health === "attention" ? "warning" : "danger")}
                        </button>`,
                    )
                    .join("")}
                </div>`
              : emptyState({
                  iconName: "repo",
                  title: "No repositories in this folder yet",
                  text: `Anything cloned into this folder uses '${workspace.alias}'.`,
                  actions: `<button type="button" class="btn btn-sm" data-action="clone-repo" data-alias="${escapeHtml(workspace.alias)}">${icon("download", "sm")}Clone into this workspace</button>`,
                  compact: true,
                })
            : ""
        }
      </section>`;
  }

  // The order bgit uses to choose an identity for a repository (see
  // core/check ResolveExpected). Shown so workspaces are understood in
  // relation to bound repositories and the default identity.
  function precedenceCard() {
    const fallback = activeIdentity();
    return `
      <section class="card">
        <div class="card-head"><h2>How bgit picks an identity</h2></div>
        <ol class="precedence">
          <li><span class="precedence-n">1</span><div><strong>Bound repository</strong><span class="muted">A repository you added or cloned with an identity keeps it.</span></div></li>
          <li><span class="precedence-n">2</span><div><strong>Workspace folder</strong><span class="muted">Otherwise, the identity of the workspace the repository is in.</span></div></li>
          <li><span class="precedence-n">3</span><div><strong>Remote address</strong><span class="muted">Otherwise, an identity named in the remote address.</span></div></li>
          <li><span class="precedence-n">4</span><div><strong>Default identity</strong><span class="muted">Everything else${fallback ? `: currently ${escapeHtml(fallback.alias)}` : ""}.</span></div></li>
        </ol>
      </section>`;
  }

  function page() {
    const workspaces = state.status?.workspaces || [];
    const canAdd = identities().length > 0;
    return `
      <div class="page">
        <header class="page-head">
          <div class="page-title">
            <h1>Workspaces</h1>
            <p>A workspace is a folder. Every repository inside it uses the workspace's identity.</p>
          </div>
          <div class="cluster">
            <button type="button" class="btn btn-primary" data-action="add-workspace" ${canAdd ? "" : "disabled"}>${icon("plus")}Add workspace</button>
          </div>
        </header>
        <div class="side-layout">
        <div class="stack">
        ${
          workspaces.length === 0
            ? `<section class="card">${emptyState({
                iconName: "layers",
                title: "No workspaces yet",
                text: canAdd
                  ? "Point bgit at a folder such as work or clients. Repositories you put there get that identity without binding each one."
                  : "Add an identity first, then choose a folder for it.",
                actions: canAdd ? `<button type="button" class="btn btn-primary" data-action="add-workspace">${icon("plus")}Add workspace</button>` : `<button type="button" class="btn btn-primary" data-action="add-identity">${icon("plus")}Add identity</button>`,
              })}</section>`
            : workspaces.map(card).join("")
        }
        </div>
        <aside class="stack">${precedenceCard()}</aside>
        </div>
      </div>`;
  }

  async function addWorkspace() {
    const result = await openDialog({
      id: "workspaceDialog",
      title: "Add a workspace",
      description: "Repositories in this folder and its subfolders use the identity you choose.",
      submitLabel: "Add workspace",
      busyLabel: "Adding",
      body: `
        <div class="field">
          <label for="workspacePath">Folder</label>
          <div class="input-group">
            <input class="input mono" id="workspacePath" name="path" placeholder="Choose a folder" autocomplete="off" spellcheck="false" autofocus />
            <button type="button" class="btn" data-browse>${icon("folder")}Browse</button>
          </div>
        </div>
        <div class="field">
          <label for="workspaceAlias">Identity</label>
          <select class="select" id="workspaceAlias" name="alias">
            ${identities()
              .map((identity) => `<option value="${escapeHtml(identity.alias)}" ${identity.active ? "selected" : ""}>${escapeHtml(identity.alias)} · ${escapeHtml(identity.email)}</option>`)
              .join("")}
          </select>
        </div>`,
      onOpen({ form, setError }) {
        form.querySelector("[data-browse]").addEventListener("click", async () => {
          try {
            const folder = await api().ChooseWorkspaceFolder();
            if (folder) {
              form.elements.path.value = folder;
            }
          } catch (error) {
            setError(errorText(error));
          }
        });
      },
      async onSubmit(form) {
        const request = { path: form.elements.path.value.trim(), alias: form.elements.alias.value };
        if (!request.path) throw new Error("Choose the workspace folder.");
        const added = await api().AddWorkspace(request);
        return { ...added, path: request.path };
      },
    });
    if (!result?.status) {
      return;
    }
    finishAction(result, { activity: "Workspaces" });
    const added = (result.status.workspaces || []).find((workspace) => workspace.path === result.path);
    if (added?.repositoryCount > 0) {
      notify({
        title: `${plural(added.repositoryCount, "repository", "repositories")} found in this workspace`,
        message: "Open Repositories to check that they match the identity.",
        tone: "warning",
        timeout: 9000,
      });
    }
  }

  async function removeWorkspace(path) {
    const confirmed = await confirmAction({
      title: `Remove workspace ${baseName(path)}?`,
      message: "bgit stops applying an identity to this folder.",
      warning: "The folder and its repositories are not deleted. Repositories that are not bound go back to your default identity.",
      acceptLabel: "Remove",
      danger: true,
    });
    if (!confirmed) {
      return;
    }
    try {
      finishAction(await api().RemoveWorkspace(path), { activity: "Workspaces" });
    } catch (error) {
      notify({ title: "Could not remove the workspace", message: errorText(error), tone: "danger" });
    }
  }

  registerPage({
    id: "workspaces",
    render: page,
    enter() {
      loadRepositories();
    },
    actions: {
      "add-workspace": () => addWorkspace(),
      "workspace-menu"(element) {
        openMenu(
          element,
          [
            { label: "Clone into this workspace", icon: "download", onSelect: () => openCloneDialog({ alias: element.dataset.alias }) },
            { label: "Copy path", icon: "copy", onSelect: () => copyText(element.dataset.path, "Path copied") },
            { separator: true },
            { label: "Remove workspace", icon: "trash", danger: true, onSelect: () => removeWorkspace(element.dataset.path) },
          ],
          { align: "end" },
        );
      },
    },
  });
})();
