// Settings: general information and shortcuts, appearance, push protection,
// encrypted backup and restore, and removing bgit from this computer.
(() => {
  const SECTIONS = [
    { id: "general", label: "General", iconName: "settings" },
    { id: "appearance", label: "Appearance", iconName: "sun" },
    { id: "protection", label: "Push protection", iconName: "shield" },
    { id: "backup", label: "Backup & Restore", iconName: "archive" },
    { id: "advanced", label: "Advanced", iconName: "wrench" },
  ];
  const SHORTCUTS = [
    ["Search repositories", [MOD_KEY, "K"]],
    ["Go to Dashboard … Doctor", [MOD_KEY, "1 – 5"]],
    ["Open Settings", [MOD_KEY, ","]],
    ["Commit (in a repository)", [MOD_KEY, "Enter"]],
    ["Hide or show the file list (in a repository)", [MOD_KEY, "B"]],
    ["Move through changed files (file list focused)", ["↑", "↓"]],
    ["Include or leave out the selected file", ["Space"]],
    ["Filter the repository list", ["/"]],
    ["Close a menu or dialog", ["Esc"]],
  ];

  const view = {
    section: "general",
    exportPath: "",
    importPath: "",
    lastExport: "",
    lastImport: null,
    // Summary shown after bgit was removed from this computer.
    removed: null,
  };

  // While the removal summary is showing, the page must stay visible even
  // though bgit is no longer configured, and the setup guide must not cover it.
  window.settingsShowsRemoval = () => Boolean(view.removed);

  function row(title, text, control) {
    return `
      <div class="setting-row">
        <div class="setting-text"><strong>${escapeHtml(title)}</strong>${text ? `<span class="muted">${text}</span>` : ""}</div>
        <div class="setting-control">${control}</div>
      </div>`;
  }

  // ---------- Sections ----------

  function general() {
    return `
      <section class="card">
        <div class="card-head"><h2>About</h2></div>
        <div class="setting-rows">
          ${row("Version", "", `<span class="mono selectable">${escapeHtml(state.appInfo?.version || "Unknown")}</span>`)}
          ${row("Setup guide", "Walk through setup again: Git check, push protection, and identities.", `<button type="button" class="btn" data-action="open-setup">Open setup guide</button>`)}
          ${row(
            "Help",
            "Documentation and issue tracker open in your browser.",
            `<div class="cluster">
              <button type="button" class="btn" data-action="open-link" data-url="https://bgitcli.com/docs">${icon("external", "sm")}Documentation</button>
              <button type="button" class="btn" data-action="open-link" data-url="https://github.com/byterings/bgit/issues">${icon("external", "sm")}Report a problem</button>
            </div>`,
          )}
        </div>
      </section>
      <section class="card">
        <div class="card-head"><h2>${icon("keyboard")}Keyboard shortcuts</h2></div>
        <div class="setting-rows">
          ${SHORTCUTS.map(([label, keys]) => row(label, "", keys.map((key) => `<span class="kbd">${escapeHtml(key)}</span>`).join(" "))).join("")}
        </div>
      </section>`;
  }

  function appearance() {
    const choice = prefs.get("theme", "dark");
    const option = (id, label, iconName, text) => `
      <button type="button" class="theme-option ${choice === id ? "selected" : ""}" data-action="set-theme" data-theme-choice="${id}" aria-pressed="${choice === id}">
        ${icon(iconName, "lg")}
        <strong>${label}</strong>
        <span class="muted">${text}</span>
      </button>`;
    return `
      <section class="card">
        <div class="card-head"><h2>Theme</h2></div>
        <div class="card-body">
          <div class="theme-options">
            ${option("dark", "Dark", "moon", "Easy on the eyes")}
            ${option("light", "Light", "sun", "For bright rooms")}
            ${option("system", "System", "monitor", "Follow the computer")}
          </div>
        </div>
      </section>`;
  }

  function protection() {
    const onboarding = state.onboarding || {};
    let tone = "success";
    let title = "Push protection is on";
    let text = "Git runs a bgit check before every push and stops the push when a repository is about to use the wrong identity.";
    let action = "";
    if (!onboarding.gitInstalled) {
      tone = "danger";
      title = "Git was not found";
      text = "Install Git, then restart bgit.";
    } else if (!onboarding.pushProtection) {
      tone = "warning";
      title = "Push protection is off";
      text =
        onboarding.hooksPath && !onboarding.hooksPathManaged
          ? `Git is set to use hooks from ${onboarding.hooksPath}, so the bgit check does not run. Turning protection on replaces that setting; bgit puts it back if you remove bgit.`
          : "The bgit check is not installed, so pushes are not checked.";
      action = `<button type="button" class="btn btn-primary" data-action="turn-on-protection">${icon("shield-check")}Turn on push protection</button>`;
    }
    return `
      <section class="card">
        <div class="card-head"><h2>Push protection</h2>${badge(tone === "success" ? "On" : "Off", tone, true)}</div>
        <div class="card-body stack">
          ${banner({ tone, iconName: tone === "success" ? "shield-check" : "shield-off", title, text, actions: action })}
          <dl class="props">
            <div><dt>How it works</dt><dd><span>A Git hook runs for every repository on this computer before each push.</span></dd></div>
            <div><dt>Git hooks folder</dt><dd><span class="mono selectable truncate-start"><span>${escapeHtml(onboarding.hooksPath || "Not set")}</span></span></dd></div>
          </dl>
          <p class="field-hint">While bgit's hooks folder is in use, Git ignores hooks stored inside individual repositories.</p>
        </div>
      </section>`;
  }

  function backup() {
    const last = view.lastImport;
    return `
      <div class="section-intro">
        ${icon("lock")}
        <p>A backup holds your identities, their SSH keys, bound repositories, and workspaces in one <code>.bgit</code> file, encrypted with your password (Argon2id and AES-256-GCM). <strong>The password cannot be recovered.</strong></p>
      </div>
      <div class="backup-grid">
        <section class="card">
          <div class="card-head"><h2>Export a backup</h2></div>
          <form class="card-body stack" id="exportForm" novalidate>
            <div class="field">
              <label for="exportPath">Save to</label>
              <div class="input-group">
                <input class="input mono" id="exportPath" name="outputPath" value="${escapeHtml(view.exportPath)}" placeholder="Default backup folder" autocomplete="off" spellcheck="false" />
                <button type="button" class="btn" data-action="choose-export">${icon("folder")}Browse</button>
              </div>
            </div>
            <div class="field-row">
              <div class="field"><label for="exportPassword">Password</label><input class="input" id="exportPassword" name="password" type="password" autocomplete="new-password" /></div>
              <div class="field"><label for="exportConfirm">Confirm password</label><input class="input" id="exportConfirm" name="confirmPassword" type="password" autocomplete="new-password" /></div>
            </div>
            <div class="dialog-error" id="exportError" role="alert" hidden></div>
            ${view.lastExport ? banner({ tone: "success", title: "Backup created", text: view.lastExport }) : ""}
            <div><button type="submit" class="btn btn-primary">${icon("download")}Export backup</button></div>
          </form>
        </section>
        <section class="card">
          <div class="card-head"><h2>Restore from a backup</h2></div>
          <form class="card-body stack" id="importForm" novalidate>
            <div class="field">
              <label for="importPath">Backup file</label>
              <div class="input-group">
                <input class="input mono" id="importPath" name="archivePath" value="${escapeHtml(view.importPath)}" placeholder="Select a .bgit file" autocomplete="off" spellcheck="false" />
                <button type="button" class="btn" data-action="choose-import">${icon("folder")}Browse</button>
              </div>
            </div>
            <div class="field"><label for="importPassword">Password</label><input class="input" id="importPassword" name="password" type="password" autocomplete="current-password" /></div>
            <div class="dialog-error" id="importError" role="alert" hidden></div>
            ${
              last
                ? banner({
                    tone: "success",
                    title: "Backup restored",
                    text: `${plural(last.usersCount || 0, "identity", "identities")}, ${plural(last.bindingsCount || 0, "bound repository", "bound repositories")}, ${plural(last.workspacesCount || 0, "workspace", "workspaces")}.${last.activeUser ? ` Default identity: ${last.activeUser}.` : ""}`,
                  })
                : ""
            }
            <div><button type="submit" class="btn">${icon("upload")}Restore backup</button></div>
            <p class="field-hint">Restoring replaces the identities, bound repositories, and workspaces on this computer.</p>
          </form>
        </section>
      </div>`;
  }

  function advanced() {
    return `
      <section class="card danger-card">
        <div class="card-head"><h2>Remove bgit from this computer</h2></div>
        <div class="card-body stack">
          <p>Do this before uninstalling the app, so your repositories and Git keep working without bgit.</p>
          <ul class="plain-list">
            <li>Repository remotes go back to standard GitHub addresses.</li>
            <li>The bgit section is removed from your SSH settings.</li>
            <li>Push protection is removed and your previous Git hooks setting is restored.</li>
            <li>Your Git name and email go back to what they were before bgit, when known.</li>
            <li>bgit settings (identities, bound repositories, workspaces) are deleted.</li>
          </ul>
          <label class="check"><input type="checkbox" id="removeKeys" />Also delete the SSH keys that bgit created</label>
          <div><button type="button" class="btn btn-danger" data-action="remove-bgit">${icon("trash")}Remove bgit</button></div>
        </div>
      </section>`;
  }

  function removedPage() {
    const result = view.removed;
    const list = (title, items) =>
      items.length > 0 ? `<div><h3>${escapeHtml(title)}</h3><ul class="plain-list mono-list">${items.map((item) => `<li>${escapeHtml(item)}</li>`).join("")}</ul></div>` : "";
    return `
      <div class="page narrow">
        <header class="page-head">
          <div class="page-title">
            <h1>bgit was removed from this computer</h1>
            <p>You can now uninstall the bgit app like any other app.</p>
          </div>
        </header>
        <section class="card">
          <div class="card-body stack">
            ${banner({
              tone: result.warnings.length > 0 ? "warning" : "success",
              title: result.warnings.length > 0 ? "Finished with warnings" : "Everything was restored",
              text: result.gitIdentityNote || "",
            })}
            ${list("Needs your attention", result.warnings)}
            ${list(`Repository remotes restored (${result.restoredRepos.length})`, result.restoredRepos)}
            ${list("SSH key files deleted", result.removedKeys)}
            <div><button type="button" class="btn btn-primary" data-action="restart-setup">Set up bgit again</button></div>
          </div>
        </section>
      </div>`;
  }

  function page() {
    if (view.removed) {
      return removedPage();
    }
    const content = { general, appearance, protection, backup, advanced }[view.section] || general;
    return `
      <div class="page">
        <header class="page-head">
          <div class="page-title">
            <h1>Settings</h1>
            <p>Preferences for this app, push protection, backups, and removal.</p>
          </div>
        </header>
        <div class="settings-layout">
          <nav class="settings-nav" aria-label="Settings sections">
            ${SECTIONS.map(
              (section) => `
                <button type="button" class="settings-nav-item ${view.section === section.id ? "active" : ""}" data-action="settings-section" data-section="${section.id}" ${view.section === section.id ? 'aria-current="page"' : ""}>
                  ${icon(section.iconName)}${escapeHtml(section.label)}
                </button>`,
            ).join("")}
          </nav>
          <div class="stack">${content()}</div>
        </div>
      </div>`;
  }

  // ---------- Actions ----------

  function showFormError(id, message) {
    const box = document.getElementById(id);
    box.textContent = message || "";
    box.hidden = !message;
  }

  async function exportBackup(form) {
    const request = {
      password: form.elements.password.value,
      confirmPassword: form.elements.confirmPassword.value,
      outputPath: form.elements.outputPath.value.trim(),
    };
    if (!request.password) {
      showFormError("exportError", "Enter a password for the backup.");
      return;
    }
    if (request.password !== request.confirmPassword) {
      showFormError("exportError", "The two passwords do not match.");
      return;
    }
    showFormError("exportError", "");
    try {
      const result = await withBusy(form.querySelector("button[type=submit]"), "Exporting", () => api().ExportBackup(request));
      view.lastExport = result.archivePath || "";
      view.exportPath = "";
      addActivity("Encrypted backup created", result.archivePath || "");
      notify({ title: "Backup created", message: result.archivePath || "" });
      render();
    } catch (error) {
      showFormError("exportError", errorText(error));
    }
  }

  async function importBackup(form) {
    const request = { archivePath: form.elements.archivePath.value.trim(), password: form.elements.password.value };
    if (!request.archivePath) {
      showFormError("importError", "Choose the backup file to restore.");
      return;
    }
    if (!request.password) {
      showFormError("importError", "Enter the backup password.");
      return;
    }
    showFormError("importError", "");
    const confirmed = await confirmAction({
      title: "Restore this backup?",
      message: "The identities, bound repositories, and workspaces on this computer are replaced by the backup.",
      warning: "What is here now is not kept. Export a backup first if you may need it.",
      acceptLabel: "Restore",
      danger: true,
    });
    if (!confirmed) {
      return;
    }
    try {
      const result = await withBusy(form.querySelector("button[type=submit]"), "Restoring", () => api().ImportBackup(request));
      view.lastImport = result;
      view.importPath = "";
      addActivity("Backup restored", request.archivePath);
      notify({ title: "Backup restored" });
      await load();
    } catch (error) {
      showFormError("importError", errorText(error));
    }
  }

  async function removeBgit(button) {
    const removeKeys = document.getElementById("removeKeys").checked;
    const confirmed = await confirmAction({
      title: "Remove bgit from this computer?",
      message: "All bgit settings are removed and your repositories and Git configuration are restored.",
      warning: removeKeys
        ? "The SSH keys bgit created will be deleted. GitHub will reject them from then on, and they cannot be recovered. Export a backup first if you may need your identities again."
        : "Your identities are deleted from bgit. SSH key files are kept. Export a backup first if you may need your identities again.",
      acceptLabel: "Remove bgit",
      danger: true,
    });
    if (!confirmed) {
      return;
    }
    try {
      view.removed = await withBusy(button, "Removing", () => api().UninstallBgit({ removeKeys }));
      const [status, doctor, onboarding] = await Promise.all([api().GetStatus(), api().GetDoctorStatus(), api().GetOnboardingState()]);
      state.status = status;
      state.doctor = doctor;
      state.onboarding = onboarding;
      state.repos = null;
      render();
    } catch (error) {
      notify({ title: "Could not remove bgit", message: errorText(error), tone: "danger" });
    }
  }

  registerPage({
    id: "settings",
    // The removal summary is shown after bgit's configuration is gone.
    allowUnconfigured: true,
    render: () => (state.status.configured || view.removed ? page() : unconfiguredView()),
    enter(params) {
      if (params.section && SECTIONS.some((section) => section.id === params.section)) {
        view.section = params.section;
      }
    },
    actions: {
      "settings-section"(element) {
        view.section = element.dataset.section;
        render();
      },
      "set-theme"(element) {
        prefs.set("theme", element.dataset.themeChoice);
        applyTheme();
        render();
      },
      async "turn-on-protection"(element) {
        try {
          state.onboarding = await withBusy(element, "Turning on", () => api().RunSetup());
          addActivity("Push protection turned on", "Settings");
          notify({ title: "Push protection is on" });
        } catch (error) {
          notify({ title: "Could not turn on push protection", message: errorText(error), tone: "danger" });
        }
        await load();
      },
      async "choose-export"() {
        try {
          const path = await api().ChooseExportArchivePath();
          if (path) {
            view.exportPath = path;
            document.getElementById("exportPath").value = path;
          }
        } catch (error) {
          showFormError("exportError", errorText(error));
        }
      },
      async "choose-import"() {
        try {
          const path = await api().ChooseImportArchivePath();
          if (path) {
            view.importPath = path;
            document.getElementById("importPath").value = path;
          }
        } catch (error) {
          showFormError("importError", errorText(error));
        }
      },
      "remove-bgit"(element) {
        removeBgit(element);
      },
      async "restart-setup"() {
        view.removed = null;
        state.page = "dashboard";
        await load();
      },
    },
    on: {
      input(target) {
        if (target.id === "exportPath") view.exportPath = target.value;
        if (target.id === "importPath") view.importPath = target.value;
        const form = target.closest("form");
        if (form?.id === "exportForm") showFormError("exportError", "");
        if (form?.id === "importForm") showFormError("importError", "");
      },
      submit(target, event) {
        event.preventDefault();
        if (target.id === "exportForm") exportBackup(target);
        if (target.id === "importForm") importBackup(target);
      },
    },
  });
})();
