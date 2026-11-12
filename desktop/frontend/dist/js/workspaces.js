// Workspaces page: folders where every repository inside uses one identity.
//
// Uses the shared helpers defined in index.html.
(() => {
  function identities() {
    return currentStatus?.identities || [];
  }

  function page(status) {
    const workspaces = status.workspaces || [];
    const canAdd = identities().length > 0;
    const actions = `<button type="button" class="primary" data-ws="add" ${canAdd ? "" : "disabled"}><span class="icon" aria-hidden="true">${iconSvg("plus")}</span>Add workspace</button>`;
    return `
      ${pageHeading("Workspaces", "Every repository inside a workspace folder uses that workspace's identity.", actions)}
      ${
        workspaces.length === 0
          ? `<section class="panel empty">
              <div class="empty-stack">
                <p>No workspaces yet. Add a folder such as "work" or "clients" so everything inside it uses one identity.</p>
                ${canAdd ? "" : `<p class="muted">Add an identity first.</p>`}
              </div>
            </section>`
          : `<section class="repo-list">${workspaces.map(row).join("")}</section>`
      }
    `;
  }

  function row(workspace) {
    const identity = identities().find((item) => item.alias === workspace.alias);
    const problem = !workspace.exists
      ? "This folder no longer exists."
      : workspace.missingIdentity
        ? `Identity '${workspace.alias}' no longer exists.`
        : "";
    const count = workspace.repositoryCount || 0;
    return `
      <article class="repo-record workspace-record ${problem ? "blocked" : ""}">
        <div class="repo-main">
          <div class="repo-title">
            <strong class="truncate" title="${escapeHtml(workspace.path)}">${escapeHtml(workspace.path.split(/[\\/]/).filter(Boolean).pop() || workspace.path)}</strong>
            ${problem ? `<span class="badge error">Problem</span>` : ""}
          </div>
          <span class="muted truncate repo-location" title="${escapeHtml(workspace.path)}">${escapeHtml(workspace.path)}</span>
        </div>
        <div class="repo-identity">
          ${identity ? avatarBlock(identity) : `<span class="avatar" aria-hidden="true">?</span>`}
          <div class="identity-details">
            <strong class="truncate">${escapeHtml(workspace.alias)}</strong>
            <span class="muted truncate">${escapeHtml(workspace.identityEmail || "")}</span>
          </div>
        </div>
        <div class="repo-branch">
          <span>${workspace.exists ? `${count} repositor${count === 1 ? "y" : "ies"}` : "Folder not found"}</span>
        </div>
        <div class="actions repo-actions">
          ${iconButton({ action: "ws-remove", icon: "trash", label: `Remove workspace ${workspace.path}`, path: workspace.path, danger: true })}
          ${workspace.exists && count > 0 ? `<button type="button" data-page-jump="repositories">View repositories</button>` : ""}
        </div>
        ${problem ? `<ul class="repo-problems"><li>${escapeHtml(problem)}</li></ul>` : ""}
      </article>
    `;
  }

  function ensureDialog() {
    if (document.getElementById("workspaceDialog")) {
      return;
    }
    document.body.insertAdjacentHTML(
      "beforeend",
      `
      <dialog id="workspaceDialog">
        <form id="workspaceForm" class="dialog-body" method="dialog" novalidate>
          <h2>Add a workspace</h2>
          <div class="form-stack">
            <div class="field-group">
              <span>Folder</span>
              <div class="field-control">
                <input name="path" placeholder="Choose a folder" required />
                <button type="button" id="workspaceChoose" title="Choose folder" aria-label="Choose folder">
                  <span class="icon" aria-hidden="true">${iconSvg("folder")}</span>
                </button>
              </div>
              <span class="field-hint">Repositories in this folder and its subfolders use the identity below.</span>
            </div>
            <label>
              Identity
              <select name="alias" required></select>
            </label>
            <div id="workspaceError" class="note error" role="alert" hidden></div>
          </div>
          <div class="dialog-actions">
            <button type="button" id="workspaceCancel">Cancel</button>
            <button type="submit" class="primary" id="workspaceSubmit">Add workspace</button>
          </div>
        </form>
      </dialog>
    `,
    );

    const dialog = document.getElementById("workspaceDialog");
    const form = document.getElementById("workspaceForm");
    const errorBox = document.getElementById("workspaceError");
    const showError = (message) => {
      errorBox.textContent = message;
      errorBox.hidden = !message;
    };

    document.getElementById("workspaceCancel").addEventListener("click", () => dialog.close());
    document.getElementById("workspaceChoose").addEventListener("click", async () => {
      try {
        const folder = await api().ChooseWorkspaceFolder();
        if (folder) {
          form.elements.path.value = folder;
        }
      } catch (error) {
        showError(String(error));
      }
    });
    form.addEventListener("submit", async (event) => {
      event.preventDefault();
      const request = { path: form.elements.path.value.trim(), alias: form.elements.alias.value };
      if (!request.path) {
        showError("Choose the workspace folder.");
        return;
      }
      showError("");
      const submit = document.getElementById("workspaceSubmit");
      submit.disabled = true;
      try {
        const result = await api().AddWorkspace(request);
        dialog.close();
        finish(result);
        const added = (result.status?.workspaces || []).find((workspace) => workspace.path === request.path);
        if (added?.repositoryCount > 0) {
          notify({
            title: "Repositories found",
            message: `${added.repositoryCount} repositor${added.repositoryCount === 1 ? "y" : "ies"} in this workspace. Open Repositories to check that they match the identity.`,
            tone: "warn",
            timeout: 9000,
          });
        }
      } catch (error) {
        showError(String(error));
      } finally {
        submit.disabled = false;
      }
    });
  }

  function finish(result) {
    if (typeof invalidateRepositories === "function") {
      invalidateRepositories();
    }
    addActivity(result.message, "Workspaces", "success");
    notify({ title: "Done", message: result.message, tone: "success" });
    render(result.status || currentStatus, currentDoctor);
  }

  function bind(target) {
    ensureDialog();

    target.querySelectorAll("[data-ws='add']").forEach((button) => {
      button.addEventListener("click", () => {
        const form = document.getElementById("workspaceForm");
        form.reset();
        form.elements.alias.innerHTML = identities()
          .map(
            (identity) =>
              `<option value="${escapeHtml(identity.alias)}" ${identity.active ? "selected" : ""}>${escapeHtml(identity.alias)} (${escapeHtml(identity.email)})</option>`,
          )
          .join("");
        document.getElementById("workspaceError").hidden = true;
        document.getElementById("workspaceDialog").showModal();
      });
    });

    target.querySelectorAll("[data-action='ws-remove']").forEach((button) => {
      button.addEventListener("click", async () => {
        const confirmed = await confirmAction({
          title: "Remove workspace",
          message: "Stop using this folder as a workspace?",
          warning: "The folder and its repositories are not deleted. Repositories that are not bound go back to your default identity.",
          acceptLabel: "Remove",
          danger: true,
        });
        if (!confirmed) {
          return;
        }
        button.disabled = true;
        try {
          finish(await api().RemoveWorkspace(button.dataset.path));
        } catch (error) {
          notify({ title: "Action failed", message: String(error), tone: "error" });
          button.disabled = false;
        }
      });
    });
  }

  registerPage({ id: "workspaces", render: page, bind });
})();
