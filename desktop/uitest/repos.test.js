const { run, expect } = require("./lib");
const fs = require("node:fs");
const path = require("node:path");
run(async (ui) => {
  await ui.open();
  await ui.seedIdentities();
  ui.addGitHubRepo("work-gh", "acme", "storefront");
  const projects = path.join(ui.home, "projects");
  fs.mkdirSync(projects, { recursive: true });
  // an existing local repo with an https remote
  const existing = path.join(ui.home, "old-project");
  fs.mkdirSync(existing, { recursive: true });
  ui.git(existing, "init", "-q", "-b", "main");
  ui.git(existing, "remote", "add", "origin", "https://github.com/me/old-project.git");

  await ui.open();
  await ui.click("#sidebarNav [data-page=repositories]");
  await ui.waitText("No repositories yet");
  await ui.shot("r1-empty");

  // ---- clone
  await ui.clickText("Clone repository", "#app button.primary");
  await ui.page.waitForSelector("#cloneDialog[open]");
  await ui.clickText("Clone", "#cloneForm button[type=submit]");
  expect((await ui.text("#cloneError")).includes("Enter the repository address"), "clone validates an empty address");
  await ui.type("#cloneForm input[name=url]", "https://github.com/acme/storefront");
  await ui.type("#cloneForm input[name=parentDir]", projects);
  expect((await ui.text("#cloneTarget")).includes(path.join(projects, "storefront")), "clone shows the target folder");
  await ui.shot("r2-clone-dialog");
  await ui.clickText("Clone", "#cloneForm button[type=submit]");
  await ui.page.waitForFunction(() => !document.getElementById("cloneDialog").open, { timeout: 30000 });
  await ui.settle(600);
  const cloned = path.join(projects, "storefront");
  expect(fs.existsSync(path.join(cloned, "README.md")), "repository cloned to disk");
  expect(ui.git(cloned, "remote", "get-url", "origin") === "git@github.com-work-gh:acme/storefront.git", "clone uses the identity host alias");
  expect(ui.git(cloned, "config", "--local", "user.email") === "dev@work.example", "clone has the identity's commit email");

  // wrong identity gets a clear auth error and the dialog stays open
  await ui.click("#sidebarNav [data-page=repositories]");
  await ui.waitText("storefront");
  await ui.clickText("Clone repository", "#app button.primary");
  await ui.type("#cloneForm input[name=url]", "acme/storefront");
  await ui.page.select("#cloneForm select[name=alias]", "home");
  await ui.type("#cloneForm input[name=parentDir]", path.join(ui.home));
  await ui.clickText("Clone", "#cloneForm button[type=submit]");
  await ui.page.waitForFunction(() => !document.getElementById("cloneError").hidden, { timeout: 30000 });
  expect((await ui.text("#cloneError")).includes("GitHub did not accept this identity"), "clone with an identity without access explains the auth failure");
  await ui.shot("r3-clone-auth-error");
  await ui.clickText("Cancel", "#cloneForm button");

  // ---- add existing
  await ui.clickText("Add existing", "#app button");
  await ui.page.waitForSelector("#addRepoDialog[open]");
  await ui.type("#addRepoForm input[name=path]", existing);
  await ui.page.select("#addRepoForm select[name=alias]", "home");
  await ui.clickText("Add repository", "#addRepoForm button[type=submit]");
  await ui.page.waitForFunction(() => !document.getElementById("addRepoDialog").open, { timeout: 15000 });
  await ui.waitText("old-project");
  expect(ui.git(existing, "remote", "get-url", "origin") === "git@github.com-home-gh:me/old-project.git", "existing repo remote now uses the identity key");
  expect(ui.git(existing, "config", "--local", "user.email") === "dev@home.example", "existing repo has the identity's commit email");

  // ---- break a repo, see it flagged, fix it
  ui.git(cloned, "config", "--local", "user.email", "someone@else.example");
  ui.git(cloned, "remote", "set-url", "origin", "https://github.com/acme/storefront.git");
  await ui.click("#sidebarNav [data-page=dashboard]");
  await ui.click("#sidebarNav [data-page=repositories]");
  await ui.waitText("needs attention");
  expect((await ui.text("#app")).includes("Needs fix"), "broken repository is flagged");
  await ui.shot("r4-list-attention");
  await ui.clickText("Fix it", "#app button");
  await ui.page.waitForFunction(() => !document.body.innerText.includes("needs attention"), { timeout: 15000 });
  expect(ui.git(cloned, "config", "--local", "user.email") === "dev@work.example", "fix restores the commit email");
  expect(ui.git(cloned, "remote", "get-url", "origin") === "git@github.com-work-gh:acme/storefront.git", "fix restores the remote");
  await ui.settle(500);
  await ui.shot("r5-list-ok");

  // ---- remove
  await ui.click(`[data-action="repos-remove"][data-path="${existing}"]`);
  await ui.page.waitForSelector("#confirmDialog[open]");
  await ui.click("#confirmAccept");
  await ui.page.waitForFunction(() => !document.querySelector("#app").innerText.includes("old-project"), { timeout: 15000 });
  expect(fs.existsSync(existing), "removing from bgit keeps the folder");
});
