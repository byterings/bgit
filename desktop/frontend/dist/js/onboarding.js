// First-run setup guide. It covers the app until Git is found, push
// protection is on, and an identity exists and is the default, so a new user
// never needs the command line.
(() => {
  const STEPS = [
    { id: "welcome", label: "Welcome" },
    { id: "setup", label: "Push protection" },
    { id: "identity", label: "Identity" },
    { id: "github", label: "GitHub" },
    { id: "done", label: "Ready" },
  ];

  const wizard = {
    open: false,
    step: "welcome",
    busy: "",
    error: "",
    // Alias of the identity created or chosen in the guide.
    alias: "",
    connection: null,
    restorePath: "",
    form: { alias: "", name: "", email: "", githubUsername: "", keyChoice: "generate", sshKeyPath: "" },
  };

  const root = () => document.getElementById("onboarding");

  // Called after every load(). Opens the guide when setup is incomplete and
  // keeps an open guide in step with fresh state.
  window.syncOnboarding = (onboarding) => {
    if (!wizard.open) {
      if (!onboarding?.needed) {
        return;
      }
      wizard.open = true;
      // A machine that was already set up (for example with the CLI) only
      // needs the missing part, not the whole introduction.
      wizard.step = onboarding.gitInstalled && onboarding.setupCompleted ? "identity" : "welcome";
    }
    draw();
  };

  // Lets Settings reopen the guide on demand.
  window.openOnboarding = (step = "welcome") => {
    wizard.open = true;
    wizard.step = step;
    wizard.error = "";
    draw();
  };

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
    const index = STEPS.findIndex((step) => step.id === wizard.step);
    target.innerHTML = `
      <header class="ob-top">
        <div class="brand">${BRAND_MARK}<span>bgit</span></div>
        ${
          wizard.step === "restore"
            ? `<span class="muted">Restore from a backup</span>`
            : `<ol class="ob-steps">
                ${STEPS.map((step, position) => {
                  const stateClass = position < index ? "complete" : position === index ? "current" : "";
                  return `<li class="${stateClass}" ${stateClass === "current" ? 'aria-current="step"' : ""}><span class="ob-step-mark">${position < index ? icon("check", "sm") : position + 1}</span><span class="ob-step-label">${escapeHtml(step.label)}</span></li>`;
                }).join("")}
              </ol>`
        }
        <span class="faint ob-progress">${wizard.step === "restore" ? "" : `Step ${index + 1} of ${STEPS.length}`}</span>
      </header>
      <div class="ob-main"><section class="ob-card" aria-live="polite">${stepContent()}</section></div>`;
    bind(target);
    target.querySelector("[data-autofocus]")?.focus();
  }

  function errorNote() {
    return wizard.error ? `<div class="dialog-error" role="alert">${escapeHtml(wizard.error)}</div>` : "";
  }

  function busyLabel(idle, busy) {
    return wizard.busy ? `<span class="spinner" aria-hidden="true"></span>${escapeHtml(busy)}` : idle;
  }

  const off = () => (wizard.busy ? "disabled" : "");

  function why(text) {
    return `<p class="ob-why">${icon("info", "sm")}<span>${escapeHtml(text)}</span></p>`;
  }

  function stepContent() {
    switch (wizard.step) {
      case "setup":
        return setupStep();
      case "identity":
        return identities().length > 0 && !wizard.alias && !state.status?.activeAlias ? pickIdentityStep() : identityStep();
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
    const onboarding = state.onboarding || {};
    return `
      <div class="ob-head">
        <h1>One computer, several Git identities</h1>
        <p>bgit keeps work, personal, and client accounts apart. Every repository commits and pushes with the identity you chose for it.</p>
      </div>
      <ul class="ob-points">
        <li>${icon("users")}<div><strong>An identity per account</strong><span>Name, email, and SSH key, kept together.</span></div></li>
        <li>${icon("repo")}<div><strong>A repository uses exactly one</strong><span>Clone, commit, and push without switching anything.</span></div></li>
        <li>${icon("shield-check")}<div><strong>Every push is checked</strong><span>A push under the wrong account is stopped before it leaves.</span></div></li>
      </ul>
      ${
        onboarding.gitInstalled
          ? banner({ tone: "success", title: `Git ${onboarding.gitVersion || ""} is installed`, text: "Setup takes about a minute." })
          : banner({ tone: "danger", title: "Git was not found on this computer", text: "Install Git, then choose Check again." })
      }
      ${errorNote()}
      <div class="ob-actions">
        <button type="button" class="btn" data-ob="restore" ${off()}>${icon("archive")}Restore from a backup</button>
        <div class="cluster">
          ${
            onboarding.gitInstalled
              ? `<button type="button" class="btn btn-primary btn-lg" data-ob="start" data-autofocus>Get started</button>`
              : `<button type="button" class="btn" data-ob="get-git">${icon("external")}Download Git</button>
                 <button type="button" class="btn btn-primary btn-lg" data-ob="recheck" data-autofocus ${off()}>${busyLabel("Check again", "Checking")}</button>`
          }
        </div>
      </div>`;
  }

  function setupStep() {
    const onboarding = state.onboarding || {};
    const replacesHooks = onboarding.hooksPath && !onboarding.hooksPathManaged;
    if (onboarding.setupCompleted && onboarding.pushProtection) {
      return `
        <div class="ob-head"><h1>Push protection is on</h1><p>bgit is already set up on this computer.</p></div>
        <div class="ob-actions">
          <button type="button" class="btn" data-ob="back-welcome">Back</button>
          <button type="button" class="btn btn-primary btn-lg" data-ob="to-identity" data-autofocus>Continue</button>
        </div>`;
    }
    return `
      <div class="ob-head">
        <h1>Turn on push protection</h1>
        <p>Before each push, bgit checks that the repository is using its own identity, and stops the push if it is not.</p>
      </div>
      ${why("Without this check a commit can be pushed under the wrong name or to the wrong account, and that is hard to undo.")}
      <ul class="ob-points">
        <li>${icon("folder")}<div><strong>A settings folder</strong><span>bgit creates <code>.bgit</code> in your home folder.</span></div></li>
        <li>${icon("shield-check")}<div><strong>A check before each push</strong><span>Git runs it for every repository on this computer.</span></div></li>
        <li>${icon("key")}<div><strong>Its own section in your SSH settings</strong><span>The rest of the file is left as it is.</span></div></li>
      </ul>
      ${
        replacesHooks
          ? banner({ tone: "warning", title: "Git already uses a custom hooks folder", text: `${onboarding.hooksPath} will be replaced by the bgit hooks folder, so hooks in it stop running. bgit puts it back if you remove bgit.` })
          : ""
      }
      ${errorNote()}
      <div class="ob-actions">
        <button type="button" class="btn" data-ob="back-welcome" ${off()}>Back</button>
        <button type="button" class="btn btn-primary btn-lg" data-ob="run-setup" data-autofocus ${off()}>${busyLabel("Set up bgit", "Setting up")}</button>
      </div>`;
  }

  function identityStep() {
    const form = wizard.form;
    const generate = form.keyChoice !== "existing";
    return `
      <div class="ob-head">
        <h1>Add your first identity</h1>
        <p>One Git profile: the name and email on your commits, and the GitHub account you push with. You can add more later.</p>
      </div>
      <form id="obIdentityForm" class="ob-form" novalidate>
        <div class="field-row">
          <div class="field">
            <label for="obAlias">Short name</label>
            <input class="input" id="obAlias" name="alias" value="${escapeHtml(form.alias)}" placeholder="work" autocomplete="off" spellcheck="false" data-autofocus />
          </div>
          <div class="field">
            <label for="obGitHub">GitHub username</label>
            <input class="input" id="obGitHub" name="githubUsername" value="${escapeHtml(form.githubUsername)}" placeholder="octocat" autocomplete="off" spellcheck="false" />
          </div>
        </div>
        <div class="field-row">
          <div class="field">
            <label for="obName">Name on commits</label>
            <input class="input" id="obName" name="name" value="${escapeHtml(form.name)}" placeholder="Jane Doe" autocomplete="off" />
          </div>
          <div class="field">
            <label for="obEmail">Email on commits</label>
            <input class="input" id="obEmail" name="email" type="email" value="${escapeHtml(form.email)}" placeholder="jane@company.com" autocomplete="off" spellcheck="false" />
          </div>
        </div>
        <div class="field">
          <span class="field-label">SSH key</span>
          <div class="choice-list">
            <label class="choice"><input type="radio" name="keyChoice" value="generate" ${generate ? "checked" : ""} /><span><strong>Create a new key</strong><span class="muted">Recommended. bgit creates a key just for this identity.</span></span></label>
            <label class="choice"><input type="radio" name="keyChoice" value="existing" ${generate ? "" : "checked"} /><span><strong>Use a key I already have</strong><span class="muted">Choose the private key file, usually in the .ssh folder.</span></span></label>
          </div>
          <div class="input-group" id="obKeyPathRow" ${generate ? "hidden" : ""}>
            <input class="input mono" name="sshKeyPath" value="${escapeHtml(form.sshKeyPath)}" placeholder="Select a private key file" autocomplete="off" spellcheck="false" />
            <button type="button" class="btn" data-ob="choose-key">${icon("folder")}Browse</button>
          </div>
        </div>
        ${errorNote()}
        <div class="ob-actions">
          <span></span>
          <button type="submit" class="btn btn-primary btn-lg" ${off()}>${busyLabel("Add identity", "Adding identity")}</button>
        </div>
      </form>`;
  }

  // Shown when identities already exist (from the CLI or a restored backup)
  // but none is the default.
  function pickIdentityStep() {
    return `
      <div class="ob-head">
        <h1>Choose your default identity</h1>
        <p>The default is used for repositories that do not have an identity of their own.</p>
      </div>
      <div class="ob-pick">
        ${identities()
          .map(
            (identity) => `
              <button type="button" class="identity-card" data-ob="activate" data-alias="${escapeHtml(identity.alias)}" ${off()}>
                ${avatar(identity, "lg")}
                <span class="identity-card-text">
                  <strong>${escapeHtml(identity.alias)}</strong>
                  <span class="muted">${escapeHtml(identity.name)} · ${escapeHtml(identity.email)}</span>
                </span>
              </button>`,
          )
          .join("")}
      </div>
      ${errorNote()}`;
  }

  function githubStep() {
    const identity = wizardIdentity();
    if (!identity) {
      return `
        <div class="ob-head"><h1>Connect GitHub</h1></div>
        ${banner({ tone: "warning", title: "No identity is available yet" })}
        <div class="ob-actions"><button type="button" class="btn" data-ob="to-identity">Back</button></div>`;
    }
    const hasKey = identity.sshPublicKeyStatus === "available" && identity.sshPublicKey;
    const result = wizard.connection;
    return `
      <div class="ob-head">
        <h1>Add your SSH key to GitHub</h1>
        <p>GitHub needs the public key of <strong>${escapeHtml(identity.alias)}</strong> before you can clone or push as ${escapeHtml(identity.githubUsername)}.</p>
      </div>
      ${why("The key is how GitHub knows which account is pushing. Each identity has its own, so accounts never mix.")}
      ${
        hasKey
          ? `<ol class="ob-numbered">
              <li><div>
                <strong>Copy the public key</strong>
                <div class="ob-key"><code class="truncate" title="${escapeHtml(identity.sshPublicKey)}">${escapeHtml(identity.sshPublicKey)}</code><button type="button" class="btn btn-sm" data-ob="copy-key">${icon("copy", "sm")}Copy</button></div>
              </div></li>
              <li><div>
                <strong>Paste it into GitHub</strong>
                <span class="muted">Sign in as ${escapeHtml(identity.githubUsername)}, paste the key into the Key box, and choose Add SSH key.</span>
                <div><button type="button" class="btn" data-ob="open-github" data-autofocus>${icon("github")}Copy key and open GitHub</button></div>
              </div></li>
              <li><div>
                <strong>Check that it works</strong>
                <div><button type="button" class="btn" data-ob="test" ${off()}>${busyLabel("Test connection", "Testing connection")}</button></div>
                ${result ? banner({ tone: result.passed ? "success" : "warning", title: result.passed ? "Connected" : "Not connected yet", text: `${result.message}${result.fix && !result.passed ? ` ${result.fix}.` : ""}` }) : ""}
              </div></li>
            </ol>`
          : banner({ tone: "warning", title: "No public key file found", text: "bgit expects a .pub file next to the private key. Add the key to GitHub yourself, or edit this identity later to choose another key." })
      }
      ${errorNote()}
      <div class="ob-actions">
        <span></span>
        ${result?.passed ? `<button type="button" class="btn btn-primary btn-lg" data-ob="to-done">Continue</button>` : `<button type="button" class="btn" data-ob="to-done">Skip for now</button>`}
      </div>`;
  }

  function doneStep() {
    const identity = wizardIdentity();
    const onboarding = state.onboarding || {};
    const line = (ok, text) => `<li class="${ok ? "ok" : "todo"}">${icon(ok ? "check" : "alert")}<span>${text}</span></li>`;
    return `
      <div class="ob-head">
        <h1>You're ready</h1>
        <p>Next, bring in a repository so bgit can keep it on the right identity.</p>
      </div>
      <ul class="ob-checklist">
        ${identity ? line(true, `<strong>${escapeHtml(identity.alias)}</strong> is your default identity (${escapeHtml(identity.email)}).`) : ""}
        ${line(Boolean(onboarding.pushProtection), onboarding.pushProtection ? "Push protection is on for every repository on this computer." : "Push protection is off. You can turn it on in Settings.")}
        ${line(Boolean(wizard.connection?.passed), wizard.connection?.passed ? "GitHub accepted your SSH key." : "GitHub is not connected yet. Add the SSH key from the Identities page before you clone or push.")}
      </ul>
      <div class="ob-actions">
        <button type="button" class="btn" data-ob="finish">Open bgit</button>
        <div class="cluster">
          <button type="button" class="btn" data-ob="finish-add">${icon("folder")}Add existing repository</button>
          <button type="button" class="btn btn-primary btn-lg" data-ob="finish-clone" data-autofocus>${icon("download")}Clone a repository</button>
        </div>
      </div>`;
  }

  function restoreStep() {
    return `
      <div class="ob-head">
        <h1>Restore from a backup</h1>
        <p>Choose a .bgit backup made by bgit on another computer. Your identities, SSH keys, and repository settings are restored from it.</p>
      </div>
      <form id="obRestoreForm" class="ob-form" novalidate>
        <div class="field">
          <label for="obArchive">Backup file</label>
          <div class="input-group">
            <input class="input mono" id="obArchive" name="archivePath" value="${escapeHtml(wizard.restorePath)}" placeholder="Select a .bgit file" autocomplete="off" spellcheck="false" data-autofocus />
            <button type="button" class="btn" data-ob="choose-archive">${icon("folder")}Browse</button>
          </div>
        </div>
        <div class="field">
          <label for="obPassword">Backup password</label>
          <input class="input" id="obPassword" name="password" type="password" autocomplete="current-password" />
        </div>
        ${errorNote()}
        <div class="ob-actions">
          <button type="button" class="btn" data-ob="back-welcome" ${off()}>Back</button>
          <button type="submit" class="btn btn-primary btn-lg" ${off()}>${busyLabel("Restore", "Restoring")}</button>
        </div>
      </form>`;
  }

  function go(step) {
    wizard.step = step;
    wizard.error = "";
    draw();
  }

  // Runs a backend call with the guide in a busy state. Returns the result,
  // or undefined after showing the error.
  async function run(busyText, action) {
    wizard.busy = busyText;
    wizard.error = "";
    draw();
    try {
      return await action();
    } catch (error) {
      wizard.error = errorText(error);
      return undefined;
    } finally {
      wizard.busy = "";
      draw();
    }
  }

  async function refreshState() {
    const [status, onboarding] = await Promise.all([api().GetStatus(), api().GetOnboardingState()]);
    state.status = status;
    state.onboarding = onboarding;
    render();
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
      message: `${plural(result.usersCount || 0, "identity", "identities")} and ${plural(result.bindingsCount || 0, "bound repository", "bound repositories")}.`,
    });
    wizard.alias = "";
    go(state.status?.activeAlias ? "done" : "identity");
  }

  async function copyKey() {
    const identity = wizardIdentity();
    if (identity?.sshPublicKey) {
      await copyText(identity.sshPublicKey, `Public key of ${identity.alias} copied`);
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
        go(state.onboarding?.setupCompleted && state.onboarding?.pushProtection ? "identity" : "setup");
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
        const onboarding = await run("Setting up", () => api().RunSetup());
        if (onboarding) {
          state.onboarding = onboarding;
          go(identities().length > 0 && state.status?.activeAlias ? "done" : "identity");
        }
        break;
      }
      case "choose-key": {
        readIdentityForm(root().querySelector("#obIdentityForm"));
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
        await copyKey();
        break;
      case "open-github":
        await copyKey();
        try {
          await api().OpenGitHubKeySettings();
        } catch (error) {
          wizard.error = errorText(error);
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
        if (action !== "finish") {
          navigate("repositories");
          if (action === "finish-clone") {
            openCloneDialog();
          } else {
            openAddRepositoryDialog();
          }
        }
        break;
      default:
        break;
    }
  }
})();
