// Settings page: push protection, setup guide, app information, and removing
// bgit from this computer.
//
// Uses the shared helpers defined in index.html.
(() => {
  const state = {
    info: null,
    // Summary shown after bgit was removed from this computer.
    removed: null,
  };

  async function loadInfo() {
    try {
      state.info = await api().GetAppInfo();
      if (currentPage === "settings") {
        render(currentStatus, currentDoctor);
      }
    } catch (error) {
      state.info = { version: "" };
    }
  }

  function protectionPanel() {
    const onboarding = currentOnboarding || {};
    let badge = `<span class="badge good">On</span>`;
    let copy = "Git runs a bgit check before every push and stops the push when a repository is about to use the wrong identity.";
    let action = "";

    if (!onboarding.gitInstalled) {
      badge = `<span class="badge error">Off</span>`;
      copy = "Git was not found on this computer. Install Git, then restart bgit.";
    } else if (!onboarding.pushProtection) {
      badge = `<span class="badge warn">Off</span>`;
      if (onboarding.hooksPath && !onboarding.hooksPathManaged) {
        copy = `Git is set to use hooks from ${onboarding.hooksPath}, so the bgit check does not run. Turning protection on replaces that setting; bgit restores it if you remove bgit.`;
      } else {
        copy = "The bgit check is not installed, so pushes are not checked.";
      }
      action = `<button type="button" class="primary" data-settings="protect">Turn on push protection</button>`;
    }

    return `
      <section class="panel">
        <div class="section-heading settings-heading">
          <h2>Push protection</h2>
          ${badge}
        </div>
        <p>${escapeHtml(copy)}</p>
        ${action ? `<div class="settings-actions">${action}</div>` : ""}
      </section>
    `;
  }

  function aboutPanel() {
    const version = state.info?.version;
    return `
      <section class="panel">
        <h2>About</h2>
        <div class="detail-list">
          <div class="detail-row"><span>Version</span><span>${escapeHtml(version || "Unknown")}</span></div>
          <div class="detail-row"><span>Identities</span><span>${currentStatus?.identityCount || 0}</span></div>
        </div>
        <div class="settings-actions">
          <button type="button" data-settings="guide">Run the setup guide again</button>
          <button type="button" data-settings="docs"><span class="icon" aria-hidden="true">${iconSvg("external")}</span>Documentation</button>
          <button type="button" data-settings="issues"><span class="icon" aria-hidden="true">${iconSvg("external")}</span>Report a problem</button>
        </div>
      </section>
    `;
  }

  function removePanel() {
    return `
      <section class="panel danger-zone">
        <h2>Remove bgit from this computer</h2>
        <p>Use this before uninstalling the app, so your repositories and Git keep working without bgit.</p>
        <ul class="plain-list">
          <li>Repository remotes go back to standard GitHub addresses.</li>
          <li>The bgit section is removed from your SSH settings.</li>
          <li>Push protection is removed and your previous Git hooks setting is restored.</li>
          <li>Your Git name and email go back to what they were before bgit, when known.</li>
          <li>bgit settings (identities, bindings, workspaces) are deleted.</li>
        </ul>
        <label class="check">
          <input type="checkbox" id="removeKeys" />
          Also delete the SSH keys that bgit created
        </label>
        <div class="settings-actions">
          <button type="button" class="danger-solid" data-settings="remove">Remove bgit</button>
        </div>
      </section>
    `;
  }

  function removedPage() {
    const result = state.removed;
    const list = (title, items) =>
      items.length > 0
        ? `<h3>${escapeHtml(title)}</h3><ul class="plain-list">${items.map((item) => `<li>${escapeHtml(item)}</li>`).join("")}</ul>`
        : "";
    return `
      ${pageHeading("bgit was removed from this computer", "You can now uninstall the bgit app like any other app.")}
      <section class="panel">
        <div class="note ${result.warnings.length > 0 ? "warn" : "good"}">
          <strong>${result.warnings.length > 0 ? "Finished with warnings" : "Everything was restored"}</strong>
          ${escapeHtml(result.gitIdentityNote || "")}
        </div>
        ${list("Needs your attention", result.warnings)}
        ${list(`Repository remotes restored (${result.restoredRepos.length})`, result.restoredRepos)}
        ${list("SSH key files deleted", result.removedKeys)}
        <div class="settings-actions">
          <button type="button" class="primary" data-settings="restart">Set up bgit again</button>
        </div>
      </section>
    `;
  }

  function page() {
    if (state.removed) {
      return removedPage();
    }
    return `
      ${pageHeading("Settings", "Push protection, setup, and removing bgit.")}
      ${protectionPanel()}
      ${aboutPanel()}
      ${removePanel()}
    `;
  }

  async function turnOnProtection(button) {
    button.disabled = true;
    button.innerHTML = `<span class="spinner" aria-hidden="true"></span>Turning on`;
    try {
      currentOnboarding = await api().RunSetup();
      notify({ title: "Done", message: "Push protection is on", tone: "success" });
      addActivity("Push protection turned on", "Settings", "success");
    } catch (error) {
      notify({ title: "Could not turn on push protection", message: String(error), tone: "error" });
    }
    await load();
  }

  async function removeBgit(button) {
    const removeKeys = document.getElementById("removeKeys").checked;
    const confirmed = await confirmAction({
      title: "Remove bgit from this computer",
      message: "Remove all bgit settings and restore your repositories and Git configuration?",
      warning: removeKeys
        ? "The SSH keys bgit created will be deleted. GitHub will reject them from then on, and they cannot be recovered. Make a backup first if you may need your identities again."
        : "Your identities are deleted from bgit. SSH key files are kept. Make a backup first if you may need your identities again.",
      acceptLabel: "Remove bgit",
      danger: true,
    });
    if (!confirmed) {
      return;
    }
    button.disabled = true;
    button.innerHTML = `<span class="spinner" aria-hidden="true"></span>Removing`;
    try {
      state.removed = await api().UninstallBgit({ removeKeys });
      const [status, doctor, onboarding] = await Promise.all([api().GetStatus(), api().GetDoctorStatus(), api().GetOnboardingState()]);
      currentOnboarding = onboarding;
      render(status, doctor);
    } catch (error) {
      notify({ title: "Could not remove bgit", message: String(error), tone: "error" });
      button.disabled = false;
      button.textContent = "Remove bgit";
    }
  }

  function bind(target) {
    if (!state.info) {
      loadInfo();
    }
    target.querySelectorAll("[data-settings]").forEach((button) => {
      button.addEventListener("click", async () => {
        switch (button.dataset.settings) {
          case "protect":
            await turnOnProtection(button);
            break;
          case "guide":
            openOnboarding("welcome");
            break;
          case "docs":
            await api().OpenExternalURL("https://bgitcli.com/docs");
            break;
          case "issues":
            await api().OpenExternalURL("https://github.com/byterings/bgit/issues");
            break;
          case "remove":
            await removeBgit(button);
            break;
          case "restart":
            state.removed = null;
            await load();
            break;
          default:
            break;
        }
      });
    });
  }

  // While the removal summary is showing, the page must stay visible even
  // though bgit is no longer configured.
  window.settingsShowsRemoval = () => Boolean(state.removed);

  registerPage({ id: "settings", render: page, bind });
})();
