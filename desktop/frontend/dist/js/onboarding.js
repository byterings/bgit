// First-run wizard. It covers the app until Git is found, bgit setup has run,
// and an identity exists and is active, so a new user never needs the CLI.
//
// Uses the shared helpers defined in index.html (api, escapeHtml, iconSvg,
// notify, render, load, avatarBlock, copyTextValue).
(() => {
  registerIcons({
    github:
      '<svg viewBox="0 0 24 24" fill="none" stroke-width="2"><path d="M9 19c-5 1.500-5-2.500-7-3m14 6v-3.870a3.370 3.370 0 0 0-.940-2.610c3.140-.350 6.440-1.540 6.440-7A5.440 5.440 0 0 0 20 4.770 5.070 5.070 0 0 0 19.910 1S18.730.650 16 2.480a13.380 13.380 0 0 0-7 0C6.270.650 5.090 1 5.090 1A5.070 5.070 0 0 0 5 4.770a5.440 5.440 0 0 0-1.500 3.780c0 5.420 3.300 6.610 6.440 7A3.370 3.370 0 0 0 9 18.130V22"/></svg>',
    external:
      '<svg viewBox="0 0 24 24" fill="none" stroke-width="2"><path d="M15 3h6v6"/><path d="M10 14 21 3"/><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/></svg>',
    alert:
      '<svg viewBox="0 0 24 24" fill="none" stroke-width="2"><circle cx="12" cy="12" r="9"/><path d="M12 8v4"/><path d="M12 16h.01"/></svg>',
  });

  const STEPS = [
    { id: "welcome", label: "Welcome" },
    { id: "setup", label: "Protect pushes" },
    { id: "identity", label: "Add identity" },
    { id: "github", label: "Connect GitHub" },
    { id: "done", label: "Ready" },
  ];

  const wizard = {
    open: false,
    step: "welcome",
    busy: "",
    error: "",
    // Alias of the identity created or chosen in the wizard.
    alias: "",
    connection: null,
    restorePath: "",
    form: { alias: "", name: "", email: "", githubUsername: "", keyChoice: "generate", sshKeyPath: "" },
  };

  const root = () => document.getElementById("onboarding");

  // Called after every load(). Opens the wizard when setup is incomplete and
  // keeps an open wizard in sync with fresh state.
  window.syncOnboarding = (onboarding) => {
    if (!wizard.open) {
      if (!onboarding?.needed) {
        return;
      }
      wizard.open = true;
      wizard.step = firstStep(onboarding);
    }
    draw();
  };

  // Lets other screens (Settings) reopen the wizard on demand.
  window.openOnboarding = (step = "welcome") => {
    wizard.open = true;
    wizard.step = step;
    wizard.error = "";
    draw();
  };

  function firstStep(onboarding) {
    // A machine that was already set up (for example by the CLI) only needs
    // the missing part, not the whole introduction.
    if (onboarding.gitInstalled && onboarding.setupCompleted) {
      return "identity";
    }
    return "welcome";
  }

  function identities() {
    return currentStatus?.identities || [];
  }

  function wizardIdentity() {
    const list = identities();
    return list.find((identity) => identity.alias === wizard.alias) || list.find((identity) => identity.active) || list[0] || null;
  }

  function close() {
    wizard.open = false;
    root().hidden = true;
    root().innerHTML = "";
  }

  function draw() {
    const target = root();
    if (!wizard.open) {
      target.hidden = true;
      return;
    }
    target.hidden = false;
    target.innerHTML = `
      <aside class="ob-rail">
        <div class="brand">
          <strong>bgit</strong>
          <span>Identity manager</span>
        </div>
        ${rail()}
        <p class="ob-rail-foot">You can change everything later from the app.</p>
      </aside>
      <div class="ob-main">
        <section class="ob-card" aria-live="polite">${stepContent()}</section>
      </div>
    `;
    bind(target);
    const focusTarget = target.querySelector("[data-autofocus]");
    if (focusTarget) {
      focusTarget.focus();
    }
  }

  function rail() {
    if (wizard.step === "restore") {
      return `<ol class="ob-steps"><li class="current"><span class="ob-step-mark">1</span><span>Restore backup</span></li></ol>`;
    }
    const currentIndex = STEPS.findIndex((step) => step.id === wizard.step);
    return `
      <ol class="ob-steps">
        ${STEPS.map((step, index) => {
          const state = index < currentIndex ? "complete" : index === currentIndex ? "current" : "";
          const mark = index < currentIndex ? `<span class="icon" aria-hidden="true">${iconSvg("check")}</span>` : index + 1;
          return `<li class="${state}" ${state === "current" ? 'aria-current="step"' : ""}><span class="ob-step-mark">${mark}</span><span>${escapeHtml(step.label)}</span></li>`;
        }).join("")}
      </ol>
    `;
  }

  function errorNote() {
    return wizard.error ? `<div class="note error" role="alert">${escapeHtml(wizard.error)}</div>` : "";
  }

  function busyLabel(idle, busy) {
    return wizard.busy ? `<span class="spinner" aria-hidden="true"></span>${escapeHtml(busy)}` : escapeHtml(idle);
  }

  function disabledWhenBusy() {
    return wizard.busy ? "disabled" : "";
  }

  function stepContent() {
    switch (wizard.step) {
      case "setup":
        return setupStep();
      case "identity":
        return identities().length > 0 && !wizard.alias && !currentStatus?.activeAlias ? pickIdentityStep() : identityStep();
      case "github":
        return githubStep();
      case "done":
        return doneStep();
      case "restore":
        return restoreStep();
      default:
        return welcomeStep();
    }
  }

  function welcomeStep() {
    const state = currentOnboarding || {};
    const gitLine = state.gitInstalled
      ? `<li><span class="icon" aria-hidden="true">${iconSvg("check")}</span><span>Git ${escapeHtml(state.gitVersion || "")} is installed.</span></li>`
      : `<li class="bad"><span class="icon" aria-hidden="true">${iconSvg("alert")}</span><span>Git was not found on this computer. Install Git, then choose Check again.</span></li>`;
    return `
      <h1>Welcome to bgit</h1>
      <p>bgit makes sure every repository commits and pushes with the right Git identity: work, personal, or client. Setup takes about a minute.</p>
      <ul class="ob-list">${gitLine}</ul>
      ${errorNote()}
      <div class="ob-actions">
        <button type="button" data-ob="restore" ${disabledWhenBusy()}>Restore from a backup</button>
        <div class="ob-actions-end">
          ${
            state.gitInstalled
              ? `<button type="button" class="primary" data-ob="start" data-autofocus>Get started</button>`
              : `<button type="button" data-ob="get-git"><span class="icon" aria-hidden="true">${iconSvg("external")}</span>Download Git</button>
                 <button type="button" class="primary" data-ob="recheck" data-autofocus ${disabledWhenBusy()}>${busyLabel("Check again", "Checking")}</button>`
          }
        </div>
      </div>
    `;
  }

  function setupStep() {
    const state = currentOnboarding || {};
    const replacesHooks = state.hooksPath && !state.hooksPathManaged;
    if (state.setupCompleted) {
      return `
        <h1>Push protection is on</h1>
        <p>bgit is already set up on this computer.</p>
        <div class="ob-actions">
          <button type="button" data-ob="back-welcome">Back</button>
          <div class="ob-actions-end"><button type="button" class="primary" data-ob="to-identity" data-autofocus>Continue</button></div>
        </div>
      `;
    }
    return `
      <h1>Turn on push protection</h1>
      <p>bgit checks every push and stops it when the repository is about to use the wrong identity. To do that it makes three changes:</p>
      <ul class="ob-list">
        <li><span class="icon" aria-hidden="true">${iconSvg("check")}</span><span>Creates its settings folder, <code>.bgit</code>, in your home folder.</span></li>
        <li><span class="icon" aria-hidden="true">${iconSvg("check")}</span><span>Adds a safety check that Git runs before each push.</span></li>
        <li><span class="icon" aria-hidden="true">${iconSvg("check")}</span><span>Adds its own section to your SSH settings. The rest of the file is left as it is.</span></li>
      </ul>
      ${
        replacesHooks
          ? `<div class="note warn"><strong>Git already uses a custom hooks folder</strong>${escapeHtml(state.hooksPath)} will be replaced by the bgit hooks folder, so hooks in it stop running. bgit puts it back if you uninstall.</div>`
          : ""
      }
      ${errorNote()}
      <div class="ob-actions">
        <button type="button" data-ob="back-welcome" ${disabledWhenBusy()}>Back</button>
        <div class="ob-actions-end">
          <button type="button" class="primary" data-ob="run-setup" data-autofocus ${disabledWhenBusy()}>${busyLabel("Set up bgit", "Setting up")}</button>
        </div>
      </div>
    `;
  }

  function identityStep() {
    const form = wizard.form;
    const generate = form.keyChoice !== "existing";
    return `
      <h1>Add your first identity</h1>
      <p>An identity is one Git profile: the name and email on your commits, and the GitHub account you push with. You can add more later.</p>
      <form id="obIdentityForm" class="ob-form" novalidate>
        <div class="form-grid">
          <label>
            Short name for this identity
            <input name="alias" value="${escapeHtml(form.alias)}" placeholder="work" autocomplete="off" required data-autofocus />
          </label>
          <label>
            GitHub username
            <input name="githubUsername" value="${escapeHtml(form.githubUsername)}" placeholder="octocat" autocomplete="off" required />
          </label>
          <label>
            Name on commits
            <input name="name" value="${escapeHtml(form.name)}" placeholder="Jane Doe" autocomplete="name" required />
          </label>
          <label>
            Email on commits
            <input name="email" type="email" value="${escapeHtml(form.email)}" placeholder="jane@company.com" autocomplete="email" required />
          </label>
        </div>
        <div class="field-group">
          <span>SSH key</span>
          <div class="ob-choice">
            <label>
              <input type="radio" name="keyChoice" value="generate" ${generate ? "checked" : ""} />
              <span>Create a new key<small>Recommended. bgit creates a key just for this identity.</small></span>
            </label>
            <label>
              <input type="radio" name="keyChoice" value="existing" ${generate ? "" : "checked"} />
              <span>Use a key I already have<small>Choose the private key file, usually in the .ssh folder.</small></span>
            </label>
          </div>
          <div class="field-control" ${generate ? "hidden" : ""} id="obKeyPathRow">
            <input name="sshKeyPath" value="${escapeHtml(form.sshKeyPath)}" placeholder="Select a private key file" />
            <button type="button" data-ob="choose-key" title="Choose SSH private key" aria-label="Choose SSH private key">
              <span class="icon" aria-hidden="true">${iconSvg("folder")}</span>
            </button>
          </div>
        </div>
        ${errorNote()}
        <div class="ob-actions">
          <span></span>
          <div class="ob-actions-end">
            <button type="submit" class="primary" ${disabledWhenBusy()}>${busyLabel("Add identity", "Adding identity")}</button>
          </div>
        </div>
      </form>
    `;
  }

  // Shown when identities already exist (for example from the CLI or a
  // restored backup) but none is active.
  function pickIdentityStep() {
    return `
      <h1>Choose your default identity</h1>
      <p>This identity is used for repositories that are not bound to a specific one.</p>
      <div class="ob-pick">
        ${identities()
          .map(
            (identity) => `
              <button type="button" data-ob="activate" data-alias="${escapeHtml(identity.alias)}" ${disabledWhenBusy()}>
                ${avatarBlock(identity)}
                <span class="identity-details">
                  <strong>${escapeHtml(identity.alias)}</strong>
                  <span class="muted">${escapeHtml(identity.name)} · ${escapeHtml(identity.email)}</span>
                </span>
              </button>`,
          )
          .join("")}
      </div>
      ${errorNote()}
    `;
  }

  function githubStep() {
    const identity = wizardIdentity();
    const hasKey = identity?.sshPublicKeyStatus === "available" && identity.sshPublicKey;
    const result = wizard.connection;
    const resultNote = result
      ? `<div class="note ${result.passed ? "good" : "warn"}" role="status"><strong>${result.passed ? "Connected" : "Not connected yet"}</strong>${escapeHtml(result.message)}${result.fix && !result.passed ? ` ${escapeHtml(result.fix)}.` : ""}</div>`
      : "";
    if (!identity) {
      return `
        <h1>Connect GitHub</h1>
        <div class="note warn">No identity is available yet.</div>
        <div class="ob-actions"><button type="button" data-ob="to-identity">Back</button></div>
      `;
    }
    return `
      <h1>Add your SSH key to GitHub</h1>
      <p>GitHub needs the public key for <strong>${escapeHtml(identity.alias)}</strong> before you can clone or push as ${escapeHtml(identity.githubUsername)}.</p>
      ${
        hasKey
          ? `<ol class="ob-numbered">
              <li>
                <div>
                  <strong>Copy the public key</strong>
                  <div class="ob-key">
                    <code title="${escapeHtml(identity.sshPublicKey)}">${escapeHtml(identity.sshPublicKey)}</code>
                    <button type="button" data-ob="copy-key"><span class="icon" aria-hidden="true">${iconSvg("copy")}</span>Copy</button>
                  </div>
                </div>
              </li>
              <li>
                <div>
                  <strong>Paste it into GitHub</strong>
                  <span class="muted">Sign in as ${escapeHtml(identity.githubUsername)}, paste the key into the Key box, and choose Add SSH key.</span>
                  <div><button type="button" data-ob="open-github" data-autofocus><span class="icon" aria-hidden="true">${iconSvg("github")}</span>Copy key and open GitHub</button></div>
                </div>
              </li>
              <li>
                <div>
                  <strong>Check that it works</strong>
                  <div><button type="button" data-ob="test" ${disabledWhenBusy()}>${busyLabel("Test connection", "Testing connection")}</button></div>
                  ${resultNote}
                </div>
              </li>
            </ol>`
          : `<div class="note warn"><strong>No public key file found</strong>bgit expects a .pub file next to the private key. Add the key to GitHub yourself, or edit this identity later to choose another key.</div>`
      }
      ${errorNote()}
      <div class="ob-actions">
        <span></span>
        <div class="ob-actions-end">
          ${
            result?.passed
              ? `<button type="button" class="primary" data-ob="to-done">Continue</button>`
              : `<button type="button" data-ob="to-done">Skip for now</button>`
          }
        </div>
      </div>
    `;
  }

  function doneStep() {
    const identity = wizardIdentity();
    const state = currentOnboarding || {};
    const hasRepositories = Boolean(featurePages.repositories);
    return `
      <h1>You're ready</h1>
      <ul class="ob-list">
        ${identity ? `<li><span class="icon" aria-hidden="true">${iconSvg("check")}</span><span><strong>${escapeHtml(identity.alias)}</strong> is your active identity (${escapeHtml(identity.email)}).</span></li>` : ""}
        ${
          state.pushProtection
            ? `<li><span class="icon" aria-hidden="true">${iconSvg("check")}</span><span>Push protection is on for every repository on this computer.</span></li>`
            : `<li class="bad"><span class="icon" aria-hidden="true">${iconSvg("alert")}</span><span>Push protection is off. You can turn it on in Settings.</span></li>`
        }
        ${
          wizard.connection?.passed
            ? `<li><span class="icon" aria-hidden="true">${iconSvg("check")}</span><span>GitHub accepted your SSH key.</span></li>`
            : `<li class="bad"><span class="icon" aria-hidden="true">${iconSvg("alert")}</span><span>GitHub is not connected yet. Add the SSH key from the Identities page before you clone or push.</span></li>`
        }
      </ul>
      <p>${hasRepositories ? "Next, bring in a repository so bgit can keep it on the right identity." : "Open bgit to manage your identities and repositories."}</p>
      <div class="ob-actions">
        <button type="button" data-ob="finish">Open bgit</button>
        ${
          hasRepositories
            ? `<div class="ob-actions-end">
                <button type="button" data-ob="finish-add">Add existing repository</button>
                <button type="button" class="primary" data-ob="finish-clone" data-autofocus>Clone a repository</button>
              </div>`
            : ""
        }
      </div>
    `;
  }

  function restoreStep() {
    return `
      <h1>Restore from a backup</h1>
      <p>Choose a .bgit backup file made by bgit on another computer. Your identities, SSH keys, and repository settings are restored from it.</p>
      <form id="obRestoreForm" class="ob-form" novalidate>
        <div class="field-group">
          <span>Backup file</span>
          <div class="field-control">
            <input name="archivePath" value="${escapeHtml(wizard.restorePath)}" placeholder="Select a .bgit file" required />
            <button type="button" data-ob="choose-archive" title="Choose backup file" aria-label="Choose backup file">
              <span class="icon" aria-hidden="true">${iconSvg("folder")}</span>
            </button>
          </div>
        </div>
        <label>
          Backup password
          <input name="password" type="password" autocomplete="current-password" required />
        </label>
        ${errorNote()}
        <div class="ob-actions">
          <button type="button" data-ob="back-welcome" ${disabledWhenBusy()}>Back</button>
          <div class="ob-actions-end">
            <button type="submit" class="primary" ${disabledWhenBusy()}>${busyLabel("Restore", "Restoring")}</button>
          </div>
        </div>
      </form>
    `;
  }

  function go(step) {
    wizard.step = step;
    wizard.error = "";
    draw();
  }

  // Runs a backend call with the wizard in a busy state. Returns the result,
  // or undefined after showing the error.
  async function run(busyText, action) {
    wizard.busy = busyText;
    wizard.error = "";
    draw();
    try {
      return await action();
    } catch (error) {
      wizard.error = String(error);
      return undefined;
    } finally {
      wizard.busy = "";
      draw();
    }
  }

  async function refreshState() {
    const [status, onboarding] = await Promise.all([api().GetStatus(), api().GetOnboardingState()]);
    currentOnboarding = onboarding;
    render(status, currentDoctor);
  }

  function readIdentityForm(form) {
    const data = new FormData(form);
    wizard.form = {
      alias: String(data.get("alias") || "").trim(),
      name: String(data.get("name") || "").trim(),
      email: String(data.get("email") || "").trim(),
      githubUsername: String(data.get("githubUsername") || "").trim(),
      keyChoice: String(data.get("keyChoice") || "generate"),
      sshKeyPath: String(data.get("sshKeyPath") || "").trim(),
    };
    return wizard.form;
  }

  function identityFormProblem(form) {
    if (!form.alias) return "Enter a short name for this identity, such as work.";
    if (/\s/.test(form.alias)) return "The short name cannot contain spaces.";
    if (!form.githubUsername) return "Enter the GitHub username for this identity.";
    if (!form.name) return "Enter the name to show on commits.";
    if (!/^[^@\s]+@[^@\s]+\.[^@\s]+$/.test(form.email)) return "Enter a valid email address.";
    if (form.keyChoice === "existing" && !form.sshKeyPath) return "Choose the private key file to use.";
    return "";
  }

  async function submitIdentity(formElement) {
    const form = readIdentityForm(formElement);
    const problem = identityFormProblem(form);
    if (problem) {
      wizard.error = problem;
      draw();
      return;
    }
    const result = await run("Adding identity", () =>
      api().AddIdentity({
        alias: form.alias,
        name: form.name,
        email: form.email,
        githubUsername: form.githubUsername,
        sshKeyPath: form.keyChoice === "existing" ? form.sshKeyPath : "",
        generateSSHKey: form.keyChoice !== "existing",
      }),
    );
    if (!result) {
      return;
    }
    wizard.alias = form.alias;
    wizard.connection = null;
    await refreshState();
    go("github");
  }

  async function submitRestore(formElement) {
    const data = new FormData(formElement);
    wizard.restorePath = String(data.get("archivePath") || "").trim();
    const password = String(data.get("password") || "");
    if (!wizard.restorePath) {
      wizard.error = "Choose the backup file to restore.";
      draw();
      return;
    }
    if (!password) {
      wizard.error = "Enter the backup password.";
      draw();
      return;
    }
    const result = await run("Restoring", async () => {
      const imported = await api().ImportBackup({ archivePath: wizard.restorePath, password });
      // Make sure push protection is installed on this computer too.
      await api().RunSetup();
      return imported;
    });
    if (!result) {
      return;
    }
    await refreshState();
    notify({
      title: "Backup restored",
      message: `Restored ${result.usersCount || 0} ${result.usersCount === 1 ? "identity" : "identities"} and ${result.bindingsCount || 0} bound ${result.bindingsCount === 1 ? "repository" : "repositories"}.`,
      tone: "success",
    });
    wizard.alias = "";
    go(currentStatus?.activeAlias ? "done" : "identity");
  }

  async function copyKey(button) {
    const identity = wizardIdentity();
    if (identity?.sshPublicKey) {
      await copyTextValue(identity.sshPublicKey, button, `Copied SSH key for ${identity.alias}`);
    }
  }

  function bind(target) {
    const identityForm = target.querySelector("#obIdentityForm");
    if (identityForm) {
      identityForm.addEventListener("submit", (event) => {
        event.preventDefault();
        submitIdentity(identityForm);
      });
      identityForm.querySelectorAll("input[name='keyChoice']").forEach((radio) => {
        radio.addEventListener("change", () => {
          readIdentityForm(identityForm);
          target.querySelector("#obKeyPathRow").hidden = wizard.form.keyChoice !== "existing";
        });
      });
    }

    const restoreForm = target.querySelector("#obRestoreForm");
    if (restoreForm) {
      restoreForm.addEventListener("submit", (event) => {
        event.preventDefault();
        submitRestore(restoreForm);
      });
    }

    target.querySelectorAll("[data-ob]").forEach((button) => {
      button.addEventListener("click", () => handle(button.dataset.ob, button));
    });
  }

  async function handle(action, button) {
    switch (action) {
      case "start":
        go(currentOnboarding?.setupCompleted ? "identity" : "setup");
        break;
      case "back-welcome":
        go("welcome");
        break;
      case "to-identity":
        go("identity");
        break;
      case "to-done":
        go("done");
        break;
      case "restore":
        go("restore");
        break;
      case "get-git":
        await api().OpenExternalURL("https://git-scm.com/downloads");
        break;
      case "recheck":
        await run("Checking", refreshState);
        break;
      case "run-setup": {
        const state = await run("Setting up", () => api().RunSetup());
        if (state) {
          currentOnboarding = state;
          go(identities().length > 0 && currentStatus?.activeAlias ? "done" : "identity");
        }
        break;
      }
      case "choose-key": {
        const form = target("#obIdentityForm");
        readIdentityForm(form);
        const path = await api().ChooseSSHPrivateKeyPath();
        if (path) {
          wizard.form.sshKeyPath = path;
          draw();
        }
        break;
      }
      case "choose-archive": {
        const path = await api().ChooseImportArchivePath();
        if (path) {
          wizard.restorePath = path;
          draw();
        }
        break;
      }
      case "activate": {
        const alias = button.dataset.alias;
        const result = await run("Switching identity", () => api().ActivateIdentity(alias));
        if (result) {
          wizard.alias = alias;
          await refreshState();
          go("github");
        }
        break;
      }
      case "copy-key":
        await copyKey(button);
        break;
      case "open-github":
        await copyKey(button);
        try {
          await api().OpenGitHubKeySettings();
        } catch (error) {
          wizard.error = String(error);
          draw();
        }
        break;
      case "test": {
        const identity = wizardIdentity();
        const result = await run("Testing connection", () => api().TestIdentityConnection(identity.alias));
        if (result) {
          wizard.connection = result;
          draw();
        }
        break;
      }
      case "finish":
      case "finish-clone":
      case "finish-add":
        close();
        await load();
        if (action !== "finish" && featurePages.repositories) {
          // The repositories page opens the matching dialog when it renders.
          window.pendingRepositoryIntent = action === "finish-clone" ? "clone" : "add";
          transitionToPage("repositories");
        }
        break;
      default:
        break;
    }
  }

  function target(selector) {
    return root().querySelector(selector);
  }
})();
