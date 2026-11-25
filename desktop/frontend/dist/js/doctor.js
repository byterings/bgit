// Doctor: a diagnostic view of everything bgit depends on. Checks from the
// backend are shown by category, together with two categories the app already
// knows about (push protection and repositories). Nothing here is made up:
// every row comes from a real check.
(() => {
  // Titles and "why it matters" for the backend's check sections.
  const SECTIONS = {
    Config: { title: "Configuration", iconName: "settings", why: "bgit reads identities, bound repositories, and workspaces from its configuration file." },
    "SSH Setup": { title: "SSH keys", iconName: "key", why: "Each identity signs in to GitHub with its own SSH key. A missing or unreadable key blocks cloning and pushing." },
    "SSH Agent": { title: "SSH agent", iconName: "lock", why: "The agent holds keys that have a passphrase. Keys without a passphrase work without it." },
    "Git Config": { title: "Git identity", iconName: "user", why: "Repositories without an identity of their own commit with the default identity, so Git must be set to it." },
  };
  const TONE = { pass: "success", warn: "warning", fail: "danger" };
  const LEVEL_ICON = { pass: "check", warn: "alert", fail: "alert" };

  let running = false;

  // Categories derived from state the app already has.
  function protectionCategory() {
    const onboarding = state.onboarding || {};
    const checks = [];
    if (!onboarding.gitInstalled) {
      checks.push({ level: "fail", message: "Git was not found on this computer", fix: "Install Git from git-scm.com, then restart bgit" });
    } else if (onboarding.pushProtection) {
      checks.push({ level: "pass", message: "The push check is installed and can run" });
    } else if (onboarding.hooksPath && !onboarding.hooksPathManaged) {
      checks.push({ level: "warn", message: `Git uses another hooks folder (${onboarding.hooksPath}), so the bgit check does not run`, fix: "Turn push protection on in Settings" });
    } else {
      checks.push({ level: "warn", message: "The push check is not installed", fix: "Turn push protection on in Settings" });
    }
    return {
      title: "Push protection",
      iconName: "shield",
      why: "Before every push, bgit checks that the repository is using its own identity. Without it a push can go out under the wrong account.",
      checks,
      action: onboarding.pushProtection ? "" : `<button type="button" class="btn btn-sm" data-action="navigate" data-page="settings" data-section="protection">Open Settings</button>`,
    };
  }

  function repositoriesCategory() {
    const repos = repositories();
    if (!state.repos) {
      return null;
    }
    const checks = [];
    const healthy = repos.filter((repo) => repo.health === "ok").length;
    if (repos.length === 0) {
      checks.push({ level: "pass", message: "No repositories are managed yet" });
    } else if (healthy > 0) {
      checks.push({ level: "pass", message: `${plural(healthy, "repository matches", "repositories match")} their identity` });
    }
    for (const repo of repos.filter((item) => item.health !== "ok")) {
      checks.push({
        level: repo.health === "attention" ? "warn" : "fail",
        message: `${repo.name}: ${(repo.problems || []).join(" ")}`,
        fix: repo.canFix ? "Use Fix on the Repositories page" : "",
      });
    }
    const fixable = repos.filter((repo) => repo.canFix).length;
    return {
      title: "Repositories",
      iconName: "repo",
      why: "A repository whose commit identity or remote address does not match its identity has its pushes blocked.",
      checks,
      action: fixable > 0 ? `<button type="button" class="btn btn-sm btn-primary" data-action="fix-all-repos">${icon("wrench", "sm")}Fix ${fixable === 1 ? "it" : `all ${fixable}`}</button>` : "",
    };
  }

  function categories() {
    const fromBackend = (state.doctor?.sections || []).map((section) => {
      const known = SECTIONS[section.title] || { title: section.title, iconName: "info", why: "" };
      return { ...known, checks: section.checks || [], action: "" };
    });
    return [protectionCategory(), ...fromBackend, repositoriesCategory()].filter(Boolean);
  }

  function worst(checks) {
    if (checks.some((check) => check.level === "fail")) return "fail";
    if (checks.some((check) => check.level === "warn")) return "warn";
    return "pass";
  }

  // Backend hints are either a sentence or "Run: <command>". Commands are
  // shown as code so they can be read and copied exactly.
  function fixText(fix) {
    const command = fix.match(/^Run: (.+)$/i) || fix.match(/^(chmod \d+ .+)$/);
    return command ? `How to fix: run <code class="selectable">${escapeHtml(command[1])}</code>` : `How to fix: ${escapeHtml(fix)}`;
  }

  function checkRow(check) {
    return `
      <div class="check-row">
        <span class="check-icon is-${TONE[check.level] || "warning"}">${icon(LEVEL_ICON[check.level] || "alert")}</span>
        <div>
          <strong>${escapeHtml(check.message)}</strong>
          ${check.fix ? `<span class="muted">${fixText(check.fix)}</span>` : ""}
        </div>
      </div>`;
  }

  function categoryCard(category) {
    const level = worst(category.checks);
    const issues = category.checks.filter((check) => check.level !== "pass");
    const passed = category.checks.filter((check) => check.level === "pass");
    const label = level === "pass" ? "Passed" : level === "warn" ? plural(issues.length, "warning", "warnings") : plural(issues.filter((check) => check.level === "fail").length, "problem", "problems");
    return `
      <section class="card doctor-card">
        <div class="card-head">
          <h2>${icon(category.iconName)}${escapeHtml(category.title)}</h2>
          <div class="cluster">${category.action || ""}${badge(label, TONE[level], true)}</div>
        </div>
        <div class="doctor-body">
          ${category.why ? `<p class="doctor-why">${escapeHtml(category.why)}</p>` : ""}
          <div class="check-rows">${issues.map(checkRow).join("")}</div>
          ${
            passed.length > 0
              ? issues.length === 0
                ? `<div class="check-rows">${passed.map(checkRow).join("")}</div>`
                : `<details class="passed"><summary>${plural(passed.length, "check", "checks")} passed</summary><div class="check-rows">${passed.map(checkRow).join("")}</div></details>`
              : ""
          }
        </div>
      </section>`;
  }

  function page() {
    const list = categories();
    const all = list.flatMap((category) => category.checks);
    const failed = all.filter((check) => check.level === "fail").length;
    const warnings = all.filter((check) => check.level === "warn").length;
    const passed = all.filter((check) => check.level === "pass").length;
    const overall = failed > 0 ? "fail" : warnings > 0 ? "warn" : "pass";
    const headline = failed > 0 ? `${plural(failed, "problem needs", "problems need")} attention` : warnings > 0 ? `${plural(warnings, "warning", "warnings")} to review` : "Everything looks good";
    return `
      <div class="page">
        <header class="page-head">
          <div class="page-title">
            <h1>Doctor</h1>
            <p>Checks the pieces bgit relies on: Git, SSH keys, push protection, and your repositories.</p>
          </div>
          <div class="cluster">
            <span class="faint">${state.doctorRunAt ? `Checked ${timeAgo(state.doctorRunAt)}` : "Not checked yet"}</span>
            <button type="button" class="btn btn-primary" data-action="run-doctor" ${running ? "disabled" : ""}>${running ? `<span class="spinner" aria-hidden="true"></span>Checking` : `${icon("activity")}Run checks`}</button>
          </div>
        </header>
        <section class="card doctor-summary is-${TONE[overall]}">
          <span class="doctor-summary-icon">${icon(overall === "pass" ? "shield-check" : "alert", "lg")}</span>
          <div class="grow">
            <h2>${escapeHtml(headline)}</h2>
            <span class="muted">${plural(all.length, "check", "checks")} across ${plural(list.length, "area", "areas")}</span>
          </div>
          <div class="cluster">
            ${badge(`${passed} passed`, "success", true)}
            ${badge(plural(warnings, "warning", "warnings"), warnings > 0 ? "warning" : "", true)}
            ${badge(plural(failed, "problem", "problems"), failed > 0 ? "danger" : "", true)}
          </div>
        </section>
        <div class="doctor-grid">${list.map(categoryCard).join("")}</div>
      </div>`;
  }

  async function runChecks() {
    running = true;
    render();
    try {
      const [doctor, onboarding] = await Promise.all([api().GetDoctorStatus(), api().GetOnboardingState()]);
      state.doctor = doctor;
      state.onboarding = onboarding;
      state.doctorRunAt = Date.now();
      await loadRepositories();
      const tone = doctor.errors > 0 ? "danger" : doctor.warnings > 0 ? "warning" : "success";
      addActivity("Doctor finished", `${plural(doctor.errors || 0, "problem", "problems")}, ${plural(doctor.warnings || 0, "warning", "warnings")}`, tone);
      notify({
        title: "Checks finished",
        message: doctor.errors > 0 ? "Some problems need attention." : doctor.warnings > 0 ? "There are warnings to review." : "Everything looks good.",
        tone,
      });
    } catch (error) {
      notify({ title: "Checks could not run", message: errorText(error), tone: "danger" });
    }
    running = false;
    render();
  }

  registerPage({
    id: "doctor",
    render: page,
    enter() {
      loadRepositories();
    },
    actions: {
      "run-doctor": () => runChecks(),
    },
  });
})();
