// Identities: the identity manager. A list of identities on the left and the
// selected identity's Git configuration, GitHub connection, SSH key, and
// repositories on the right. Also owns the add / edit / delete dialogs.
(() => {
  const view = {
    alias: "",
    // Connection test results from this session, by alias.
    tests: {},
    testing: "",
  };

  const KEY_STATUS = {
    available: { label: "SSH key ready", tone: "success" },
    missing: { label: "SSH key missing", tone: "danger" },
    not_configured: { label: "No SSH key", tone: "warning" },
    unknown: { label: "SSH key unreadable", tone: "warning" },
  };

  function selected() {
    return identityByAlias(view.alias) || activeIdentity() || identities()[0] || null;
  }

  function reposFor(alias) {
    return repositories().filter((repo) => repo.alias === alias);
  }

  function workspacesFor(alias) {
    return (state.status?.workspaces || []).filter((workspace) => workspace.alias === alias);
  }

  // A public key is safe to show, but long: keep its type and its ends.
  function shortKey(publicKey) {
    const [type, body = "", comment = ""] = publicKey.trim().split(/\s+/);
    const middle = body.length > 28 ? `${body.slice(0, 14)}…${body.slice(-10)}` : body;
    return [type, middle, comment].filter(Boolean).join(" ");
  }

  // ---------- List ----------

  function card(identity, isSelected) {
    const key = KEY_STATUS[identity.sshKeyStatus] || KEY_STATUS.unknown;
    const count = reposFor(identity.alias).length;
    return `
      <button type="button" class="identity-card ${isSelected ? "selected" : ""}" data-action="select-identity" data-alias="${escapeHtml(identity.alias)}" aria-pressed="${isSelected}">
        ${avatar(identity, "lg")}
        <span class="identity-card-text">
          <span class="identity-card-title"><strong class="truncate">${escapeHtml(identity.alias)}</strong>${identity.active ? badge("Default", "accent") : ""}</span>
          <span class="truncate">${escapeHtml(identity.name)}</span>
          <span class="muted truncate">${escapeHtml(identity.email)}</span>
          <span class="identity-card-meta">
            <span class="dot is-${key.tone}" data-tip="${key.label}" data-tip-align="start"></span>
            <span class="faint truncate">${icon("github", "sm")}${escapeHtml(identity.githubUsername || "No GitHub user")}</span>
            <span class="faint">${plural(count, "repo", "repos")}</span>
          </span>
        </span>
      </button>`;
  }

  // ---------- Detail ----------

  function section(title, content, actions = "") {
    return `
      <section class="detail-section">
        <div class="section-head"><h3>${escapeHtml(title)}</h3>${actions}</div>
        ${content}
      </section>`;
  }

  function connectionBlock(identity) {
    const test = view.tests[identity.alias];
    const testing = view.testing === identity.alias;
    let status = `<span class="muted">Not tested in this session.</span>`;
    if (test) {
      status = test.passed
        ? banner({ tone: "success", title: "Connected", text: test.message })
        : banner({ tone: "warning", title: "Not connected", text: `${test.message}${test.fix ? ` ${test.fix}.` : ""}` });
    }
    return `
      <dl class="props">
        <div><dt>GitHub user</dt><dd><span class="selectable truncate">${escapeHtml(identity.githubUsername || "Not set")}</span>${
          identity.githubUsername
            ? `<button type="button" class="icon-btn sm" data-action="open-link" data-url="https://github.com/${escapeHtml(identity.githubUsername)}" aria-label="Open GitHub profile" data-tip="Open profile on GitHub">${icon("external", "sm")}</button>`
            : ""
        }</dd></div>
      </dl>
      <div class="connection-status">${status}</div>
      <div class="cluster">
        <button type="button" class="btn btn-sm" data-action="test-connection" data-alias="${escapeHtml(identity.alias)}" ${identity.sshKeyStatus === "available" && !testing ? "" : "disabled"}>
          ${testing ? `<span class="spinner" aria-hidden="true"></span>Testing` : `${icon("activity", "sm")}Test connection`}
        </button>
      </div>`;
  }

  function sshBlock(identity) {
    const key = KEY_STATUS[identity.sshKeyStatus] || KEY_STATUS.unknown;
    const hasPublic = identity.sshPublicKeyStatus === "available" && identity.sshPublicKey;
    return `
      <dl class="props">
        <div><dt>Status</dt><dd>${badge(key.label, key.tone, true)}</dd></div>
        ${identity.sshKeyPath ? `<div><dt>Key file</dt><dd><span class="mono selectable truncate-start"><span>${escapeHtml(identity.sshKeyPath)}</span></span></dd></div>` : ""}
        <div><dt>Public key</dt><dd>${
          hasPublic
            ? `<span class="mono truncate" data-tip="This is the public half of the key. It is safe to share." data-tip-align="start">${escapeHtml(shortKey(identity.sshPublicKey))}</span>
               <button type="button" class="icon-btn sm" data-action="copy" data-value="${escapeHtml(identity.sshPublicKey)}" data-copied="Public key copied" aria-label="Copy public key" data-tip="Copy public key">${icon("copy", "sm")}</button>`
            : `<span class="muted">${identity.sshKeyStatus === "available" ? "No .pub file next to the private key" : "Not available"}</span>`
        }</dd></div>
      </dl>
      <div class="cluster">
        <button type="button" class="btn btn-sm" data-action="copy-open-github" data-alias="${escapeHtml(identity.alias)}" ${hasPublic ? "" : "disabled"}>${icon("github", "sm")}Copy key and open GitHub</button>
      </div>
      <p class="field-hint">${icon("lock", "sm")} The private key stays on this computer and is never shown.</p>`;
  }

  function reposBlock(identity) {
    const repos = reposFor(identity.alias);
    const workspaces = workspacesFor(identity.alias);
    if (repos.length === 0 && workspaces.length === 0) {
      return `<p class="muted">No repository uses this identity yet.${identity.active ? " As the default identity, it applies to repositories that are not managed by bgit." : ""}</p>`;
    }
    return `
      <div class="mini-list">
        ${repos
          .map(
            (repo) => `
              <button type="button" class="mini-row" data-action="open-repo" data-path="${escapeHtml(repo.path)}" ${repo.exists ? "" : "disabled"}>
                ${icon("repo", "sm")}<strong class="truncate">${escapeHtml(repo.name)}</strong>
                <span class="faint truncate">${escapeHtml(repo.branch || "")}</span>
                ${repo.health === "ok" ? "" : badge(repo.health === "attention" ? "Needs fix" : repo.health === "missing" ? "Missing" : "Blocked", repo.health === "attention" ? "warning" : "danger")}
              </button>`,
          )
          .join("")}
        ${workspaces
          .map(
            (workspace) => `
              <button type="button" class="mini-row" data-action="navigate" data-page="workspaces">
                ${icon("layers", "sm")}<strong class="truncate">${escapeHtml(baseName(workspace.path))}</strong>
                <span class="faint truncate">Workspace · ${plural(workspace.repositoryCount || 0, "repository", "repositories")}</span>
              </button>`,
          )
          .join("")}
      </div>`;
  }

  function detail(identity) {
    const repoCount = reposFor(identity.alias).length;
    return `
      <article class="card identity-detail">
        <header class="identity-hero">
          ${avatar(identity, "xl")}
          <div class="identity-hero-text">
            <div class="identity-hero-title"><h1 class="truncate">${escapeHtml(identity.alias)}</h1>${identity.active ? badge("Default identity", "accent") : ""}</div>
            <span class="muted truncate">${escapeHtml(identity.name)} · ${escapeHtml(identity.email)}</span>
          </div>
          <div class="cluster">
            ${identity.active ? "" : `<button type="button" class="btn" data-action="set-default" data-alias="${escapeHtml(identity.alias)}" data-tip="Use for repositories without an identity of their own" data-tip-align="end">${icon("star")}Set as default</button>`}
            <button type="button" class="btn" data-action="edit-identity" data-alias="${escapeHtml(identity.alias)}">${icon("edit")}Edit</button>
            <button type="button" class="icon-btn bordered" data-action="identity-menu" data-alias="${escapeHtml(identity.alias)}" aria-label="More actions" aria-haspopup="menu">${icon("more")}</button>
          </div>
        </header>
        <div class="detail-grid">
          ${section(
            "Git configuration",
            `<dl class="props">
              <div><dt>Name on commits</dt><dd><span class="selectable truncate">${escapeHtml(identity.name)}</span></dd></div>
              <div><dt>Email on commits</dt><dd><span class="selectable truncate">${escapeHtml(identity.email)}</span><button type="button" class="icon-btn sm" data-action="copy" data-value="${escapeHtml(identity.email)}" data-copied="Email copied" aria-label="Copy email" data-tip="Copy email">${icon("copy", "sm")}</button></dd></div>
              <div><dt>Role</dt><dd>${
                identity.active
                  ? `<span>Default for repositories without their own identity</span>`
                  : `<span class="muted">Used only where a repository or workspace selects it</span>`
              }</dd></div>
            </dl>`,
          )}
          ${section("GitHub connection", connectionBlock(identity))}
          ${section("SSH key", sshBlock(identity))}
          ${section(`Repositories (${repoCount})`, reposBlock(identity))}
        </div>
        <footer class="card-foot identity-foot">
          ${icon("archive", "sm")}<span class="muted">Identities and their SSH keys are included in encrypted backups.</span>
          <button type="button" class="link-btn" data-action="navigate" data-page="settings" data-section="backup">Export or import a backup</button>
        </footer>
      </article>`;
  }

  function page() {
    const list = identities();
    const head = `
      <header class="page-head">
        <div class="page-title">
          <h1>Identities</h1>
          <p>An identity is a Git name, email, and SSH key. Each repository uses exactly one.</p>
        </div>
        <div class="cluster">
          <button type="button" class="btn" data-action="navigate" data-page="settings" data-section="backup">${icon("archive")}Import or export</button>
          <button type="button" class="btn btn-primary" data-action="add-identity">${icon("plus")}Add identity</button>
        </div>
      </header>`;
    if (list.length === 0) {
      return `<div class="page">${head}<section class="card">${emptyState({
        iconName: "users",
        title: "No identities yet",
        text: "Add one for each Git account you use: work, personal, a client. bgit keeps their names, emails, and SSH keys apart.",
        actions: `<button type="button" class="btn btn-primary" data-action="add-identity">${icon("plus")}Add identity</button>`,
      })}</section></div>`;
    }
    const current = selected();
    return `
      <div class="page">
        ${head}
        <div class="identity-layout">
          <div class="identity-list" role="list">${list.map((identity) => card(identity, identity.alias === current.alias)).join("")}</div>
          ${detail(current)}
        </div>
      </div>`;
  }

  // ---------- Actions ----------

  async function testConnection(alias) {
    view.testing = alias;
    render();
    try {
      const result = await api().TestIdentityConnection(alias);
      view.tests[alias] = result;
      addActivity(result.passed ? `GitHub accepted the key for '${alias}'` : `GitHub connection failed for '${alias}'`, "Identities", result.passed ? "success" : "warning");
    } catch (error) {
      view.tests[alias] = { passed: false, message: errorText(error), fix: "" };
    }
    view.testing = "";
    render();
  }

  async function copyAndOpenGitHub(alias) {
    const identity = identityByAlias(alias);
    if (!identity?.sshPublicKey) {
      return;
    }
    await copyText(identity.sshPublicKey, "Public key copied. Paste it into GitHub.");
    try {
      await api().OpenGitHubKeySettings();
    } catch (error) {
      notify({ title: "Could not open GitHub", message: errorText(error), tone: "danger" });
    }
  }

  async function deleteIdentity(alias) {
    const identity = identityByAlias(alias);
    if (!identity) {
      return;
    }
    const repos = reposFor(alias).filter((repo) => repo.source === "binding").length;
    const workspaces = workspacesFor(alias).length;
    const effects = [
      repos > 0 ? `${plural(repos, "repository", "repositories")} will no longer be bound to an identity.` : "",
      workspaces > 0 ? `${plural(workspaces, "workspace", "workspaces")} using this identity will be removed.` : "",
      identity.active ? "It is the default identity; choose another default afterwards." : "",
    ].filter(Boolean);
    const result = await openDialog({
      id: "deleteIdentityDialog",
      title: `Delete identity '${alias}'?`,
      description: "The identity is removed from bgit. Repository files are not touched.",
      submitLabel: "Delete identity",
      busyLabel: "Deleting",
      danger: true,
      body: `
        ${effects.length > 0 ? banner({ tone: "warning", title: "This also changes", text: effects.join(" ") }) : ""}
        ${
          identity.sshKeyPath
            ? `<label class="check"><input type="checkbox" name="deleteKeys" />Also delete this identity's SSH key files</label>
               <span class="field-hint">Only do this if the key is not used anywhere else. Deleted keys cannot be recovered.</span>`
            : ""
        }`,
      onSubmit: (form) => api().DeleteIdentity({ alias, deleteKeys: Boolean(form.elements.deleteKeys?.checked) }),
    });
    if (result?.status) {
      view.alias = "";
      delete view.tests[alias];
      finishAction(result, { activity: "Identities" });
    }
  }

  registerPage({
    id: "identities",
    render: page,
    enter(params) {
      if (params.alias) {
        view.alias = params.alias;
      }
    },
    actions: {
      "select-identity"(element) {
        view.alias = element.dataset.alias;
        render();
      },
      "set-default"(element) {
        setDefaultIdentity(element.dataset.alias);
      },
      "edit-identity"(element) {
        openIdentityDialog(element.dataset.alias);
      },
      "test-connection"(element) {
        testConnection(element.dataset.alias);
      },
      "copy-open-github"(element) {
        copyAndOpenGitHub(element.dataset.alias);
      },
      "identity-menu"(element) {
        const alias = element.dataset.alias;
        const identity = identityByAlias(alias);
        openMenu(
          element,
          [
            { label: "Test connection", icon: "activity", disabled: identity?.sshKeyStatus !== "available", onSelect: () => testConnection(alias) },
            { label: "Copy public key", icon: "copy", disabled: !identity?.sshPublicKey, onSelect: () => copyText(identity.sshPublicKey, "Public key copied") },
            { separator: true },
            { label: "Delete identity", icon: "trash", danger: true, onSelect: () => deleteIdentity(alias) },
          ],
          { align: "end" },
        );
      },
    },
  });

  registerActions({
    "add-identity"() {
      openIdentityDialog();
    },
  });

  // ---------- Add / edit dialog ----------

  // Without an alias this adds an identity; with one it edits that identity.
  window.openIdentityDialog = async (alias = "") => {
    const identity = alias ? identityByAlias(alias) : null;
    const editing = Boolean(identity);
    const result = await openDialog({
      id: "identityDialog",
      wide: true,
      title: editing ? `Edit identity '${identity.alias}'` : "Add an identity",
      description: editing
        ? "Changes apply to future commits in every repository that uses this identity."
        : "One Git profile: the name and email on your commits, and the GitHub account you push with.",
      submitLabel: editing ? "Save changes" : "Add identity",
      busyLabel: editing ? "Saving" : "Adding",
      body: `
        <div class="field-row">
          <div class="field">
            <label for="identityAlias">Short name</label>
            <input class="input" id="identityAlias" name="alias" value="${escapeHtml(identity?.alias || "")}" placeholder="work" autocomplete="off" spellcheck="false" ${editing ? "readonly" : "autofocus"} />
            <span class="field-hint">${editing ? "The short name cannot be changed." : "How you refer to it, such as work or personal."}</span>
          </div>
          <div class="field">
            <label for="identityGitHub">GitHub username</label>
            <input class="input" id="identityGitHub" name="githubUsername" value="${escapeHtml(identity?.githubUsername || "")}" placeholder="octocat" autocomplete="off" spellcheck="false" />
          </div>
        </div>
        <div class="field-row">
          <div class="field">
            <label for="identityName">Name on commits</label>
            <input class="input" id="identityName" name="name" value="${escapeHtml(identity?.name || "")}" placeholder="Jane Doe" autocomplete="off" ${editing ? "autofocus" : ""} />
          </div>
          <div class="field">
            <label for="identityEmail">Email on commits</label>
            <input class="input" id="identityEmail" name="email" type="email" value="${escapeHtml(identity?.email || "")}" placeholder="jane@company.com" autocomplete="off" spellcheck="false" />
          </div>
        </div>
        <div class="field">
          <span class="field-label">SSH key</span>
          ${
            editing
              ? ""
              : `<div class="choice-list">
                  <label class="choice"><input type="radio" name="keyChoice" value="generate" checked /><span><strong>Create a new key</strong><span class="muted">Recommended. bgit creates a key just for this identity.</span></span></label>
                  <label class="choice"><input type="radio" name="keyChoice" value="existing" /><span><strong>Use a key I already have</strong><span class="muted">Choose the private key file, usually in the .ssh folder.</span></span></label>
                </div>`
          }
          <div class="input-group" data-key-row ${editing ? "" : "hidden"}>
            <input class="input mono" name="sshKeyPath" value="${escapeHtml(identity?.sshKeyPath || "")}" placeholder="Select a private key file" autocomplete="off" spellcheck="false" />
            <button type="button" class="btn" data-browse>${icon("folder")}Browse</button>
          </div>
        </div>`,
      onOpen({ form, setError }) {
        form.querySelectorAll("input[name=keyChoice]").forEach((radio) => {
          radio.addEventListener("change", () => {
            form.querySelector("[data-key-row]").hidden = form.elements.keyChoice.value !== "existing";
          });
        });
        form.querySelector("[data-browse]").addEventListener("click", async () => {
          try {
            const path = await api().ChooseSSHPrivateKeyPath();
            if (path) {
              form.elements.sshKeyPath.value = path;
            }
          } catch (error) {
            setError(errorText(error));
          }
        });
      },
      async onSubmit(form) {
        const values = {
          alias: form.elements.alias.value.trim(),
          name: form.elements.name.value.trim(),
          email: form.elements.email.value.trim(),
          githubUsername: form.elements.githubUsername.value.trim(),
          sshKeyPath: form.elements.sshKeyPath.value.trim(),
        };
        if (!values.alias) throw new Error("Enter a short name for this identity, such as work.");
        if (/\s/.test(values.alias)) throw new Error("The short name cannot contain spaces.");
        if (!values.githubUsername) throw new Error("Enter the GitHub username for this identity.");
        if (!values.name) throw new Error("Enter the name to show on commits.");
        if (!/^[^@\s]+@[^@\s]+\.[^@\s]+$/.test(values.email)) throw new Error("Enter a valid email address.");
        if (editing) {
          return api().UpdateIdentity(values);
        }
        const generate = form.elements.keyChoice.value !== "existing";
        if (!generate && !values.sshKeyPath) throw new Error("Choose the private key file to use.");
        return api().AddIdentity({ ...values, sshKeyPath: generate ? "" : values.sshKeyPath, generateSSHKey: generate });
      },
    });
    if (result?.status) {
      const savedAlias = editing ? identity.alias : result.status.identities?.[result.status.identities.length - 1]?.alias;
      view.alias = savedAlias || view.alias;
      delete view.tests[savedAlias];
      finishAction(result, { activity: "Identities" });
      if (!editing) {
        // Straight to the new identity, where the next step (adding its key to
        // GitHub) is one click away.
        navigate("identities", { alias: view.alias });
      }
    }
    return result;
  };
})();
