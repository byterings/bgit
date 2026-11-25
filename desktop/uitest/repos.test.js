// Repository list: clone, add existing, filter, fix, change identity, remove, search.
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
  await ui.nav("repositories");
  await ui.waitText("No repositories yet");
  await ui.shot("r1-empty");

  // ---- clone
  await ui.click('#view .page-head [data-action="clone-repo"]');
  await ui.waitDialog("cloneDialog");
  await ui.submitDialog("cloneDialog");
  expect((await ui.dialogError("cloneDialog")).includes("Enter the repository address"), "clone validates an empty address");
  await ui.type("#cloneUrl", "https://github.com/acme/storefront");
  await ui.type("#cloneFolder", projects);
  expect((await ui.text("#cloneDialog [data-target]")).includes(path.join(projects, "storefront")), "clone shows the target folder");
  expect((await ui.text("#cloneDialog [data-future]")).includes("Work Dev <dev@work.example>"), "clone says who future commits will be recorded as");
  await ui.shot("r2-clone-dialog");
  await ui.submitDialog("cloneDialog");
  await ui.waitDialogClosed("cloneDialog");
  await ui.page.waitForSelector("#wsHead h1");
  const cloned = path.join(projects, "storefront");
  expect((await ui.text("#wsHead h1")).trim() === "storefront", "the cloned repository opens");
  expect(fs.existsSync(path.join(cloned, "README.md")), "repository cloned to disk");
  expect(ui.git(cloned, "remote", "get-url", "origin") === "git@github.com-work-gh:acme/storefront.git", "clone uses the identity host alias");
  expect(ui.git(cloned, "config", "--local", "user.email") === "dev@work.example", "clone has the identity's commit email");

  // wrong identity gets a clear auth error and the dialog stays open
  await ui.click('#wsHead [data-action="ws-back"]');
  await ui.waitText("Each repository commits and pushes");
  await ui.click('#view .page-head [data-action="clone-repo"]');
  await ui.waitDialog("cloneDialog");
  await ui.type("#cloneUrl", "acme/storefront");
  await ui.page.select("#cloneAlias", "home");
  await ui.type("#cloneFolder", ui.home);
  await ui.submitDialog("cloneDialog");
  expect((await ui.dialogError("cloneDialog")).includes("GitHub did not accept this identity"), "clone with an identity without access explains the auth failure");
  await ui.shot("r3-clone-auth-error");
  await ui.cancelDialog("cloneDialog");

  // ---- add existing
  await ui.click('#view .page-head [data-action="add-repo"]');
  await ui.waitDialog("addRepoDialog");
  await ui.type("#addRepoPath", existing);
  await ui.page.select("#addRepoAlias", "home");
  await ui.shot("r4-add-existing-dialog");
  await ui.submitDialog("addRepoDialog");
  await ui.waitDialogClosed("addRepoDialog");
  await ui.waitText("old-project");
  expect(ui.git(existing, "remote", "get-url", "origin") === "git@github.com-home-gh:me/old-project.git", "existing repo remote now uses the identity key");
  expect(ui.git(existing, "config", "--local", "user.email") === "dev@home.example", "existing repo has the identity's commit email");

  // ---- break a repo, see it flagged, filter to it, fix it
  ui.git(cloned, "config", "--local", "user.email", "someone@else.example");
  ui.git(cloned, "remote", "set-url", "origin", "https://github.com/acme/storefront.git");
  await ui.focusWindow();
  await ui.waitText("needs attention");
  expect((await ui.text("#view")).includes("Needs fix"), "broken repository is flagged");
  expect((await ui.page.$$('#appbar [data-nav="repositories"] .dot')).length === 1, "the Repositories tab carries a marker");
  await ui.shot("r5-list-attention");
  await ui.click('[data-action="set-filter"][data-filter="attention"]');
  expect((await ui.page.$$("#repoRows .repo-row")).length === 1, "filter shows only repositories needing attention");
  await ui.type("#repoSearch", "old");
  await ui.waitText("Nothing matches");
  expect(await ui.page.evaluate(() => document.activeElement.id === "repoSearch"), "the filter box keeps focus while typing");
  await ui.click('[data-action="clear-filters"]');
  expect((await ui.page.$$("#repoRows .repo-row")).length === 2, "clearing filters shows everything again");
  await ui.click('#view [data-action="fix-all-repos"]');
  await ui.waitGone("needs attention");
  expect(ui.git(cloned, "config", "--local", "user.email") === "dev@work.example", "fix restores the commit email");
  expect(ui.git(cloned, "remote", "get-url", "origin") === "git@github.com-work-gh:acme/storefront.git", "fix restores the remote");
  await ui.shot("r6-list-ok");

  // ---- change identity from the context menu (right click)
  await ui.page.click(`.repo-row[data-path="${cloned}"] .repo-main`, { button: "right" });
  await ui.page.waitForSelector(".popover .menu");
  await ui.shot("r7-context-menu");
  await ui.menuItem("Change identity");
  await ui.waitDialog("addRepoDialog");
  expect((await ui.text("#addRepoDialog h2")).includes("Change the identity of storefront"), "the dialog names the repository");
  expect(await ui.page.$eval("#addRepoPath", (el) => el.readOnly), "the folder cannot be changed while changing identity");
  await ui.cancelDialog("addRepoDialog");
  expect(ui.git(cloned, "config", "--local", "user.email") === "dev@work.example", "cancelling leaves the identity alone");

  // ---- remove from the row menu
  await ui.click(`[data-action="repo-menu"][data-path="${existing}"]`);
  await ui.menuItem("Remove from bgit");
  await ui.waitDialog("confirmDialog");
  expect((await ui.text("#confirmDialog")).includes("not deleted"), "removal explains that files are kept");
  await ui.submitDialog("confirmDialog");
  await ui.waitDialogClosed("confirmDialog");
  await ui.page.waitForFunction(() => !document.querySelector("#view").innerText.includes("old-project"), { timeout: 15000 });
  expect(fs.existsSync(existing), "removing from bgit keeps the folder");

  // ---- search from the top bar
  await ui.page.keyboard.down("Control");
  await ui.page.keyboard.press("k");
  await ui.page.keyboard.up("Control");
  expect(await ui.page.evaluate(() => document.activeElement.id === "globalSearch"), "Ctrl+K focuses search");
  await ui.page.keyboard.type("store");
  await ui.page.waitForSelector(".popover.search-pop [data-search-path]");
  await ui.shot("r8-global-search");
  await ui.page.keyboard.press("Enter");
  await ui.page.waitForSelector("#wsHead h1");
  expect((await ui.text("#wsHead h1")).trim() === "storefront", "search opens the repository");
}, { width: 1366, height: 699 });
