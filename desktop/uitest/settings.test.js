// Workspaces, push protection, encrypted backup and restore, and removing bgit.
const { run, expect } = require("./lib");
const fs = require("node:fs");
const path = require("node:path");
run(async (ui) => {
  await ui.open();
  ui.git(ui.home, "config", "--global", "user.name", "Before Bgit");
  ui.git(ui.home, "config", "--global", "user.email", "before@example.com");
  await ui.seedIdentities();
  ui.addGitHubRepo("home-gh", "me", "blog");
  // a workspace folder containing an untouched repository
  const workspace = path.join(ui.home, "personal");
  const blog = path.join(workspace, "blog");
  fs.mkdirSync(blog, { recursive: true });
  ui.git(blog, "init", "-q", "-b", "main");
  ui.git(blog, "remote", "add", "origin", "https://github.com/me/blog.git");

  await ui.open();
  // ---- workspaces
  await ui.nav("workspaces");
  await ui.waitText("No workspaces yet");
  await ui.shot("s1-workspaces-empty");
  await ui.click('#view .page-head [data-action="add-workspace"]');
  await ui.waitDialog("workspaceDialog");
  await ui.submitDialog("workspaceDialog");
  expect((await ui.dialogError("workspaceDialog")).includes("Choose the workspace folder"), "workspace form validates the folder");
  await ui.type("#workspacePath", workspace);
  await ui.page.select("#workspaceAlias", "home");
  await ui.shot("s2-add-workspace-dialog");
  await ui.submitDialog("workspaceDialog");
  await ui.waitDialogClosed("workspaceDialog");
  await ui.waitText("1 repository found in this workspace");
  await ui.page.waitForSelector(`.workspace-card [data-action="open-repo"][data-path="${blog}"]`);
  expect(true, "the workspace lists the repository inside it");
  await ui.shot("s3-workspaces");
  // nested workspace is rejected with a reason
  await ui.click('#view .page-head [data-action="add-workspace"]');
  await ui.waitDialog("workspaceDialog");
  await ui.type("#workspacePath", blog);
  await ui.submitDialog("workspaceDialog");
  expect((await ui.dialogError("workspaceDialog")).includes("inside the workspace"), "nested workspace is rejected");
  await ui.cancelDialog("workspaceDialog");

  // the workspace repository shows up under Repositories and can be fixed
  await ui.nav("repositories");
  await ui.waitText("blog");
  await ui.waitText("needs attention");
  expect((await ui.page.$$(`.repo-row[data-path="${blog}"] [data-tip*="workspace"]`)).length === 1, "workspace repository is listed with its source");
  await ui.click(`[data-action="fix-repo"][data-path="${blog}"]`);
  await ui.waitGone("needs attention");
  expect(ui.git(blog, "config", "--local", "user.email") === "dev@home.example", "workspace repository now commits as the workspace identity");
  expect(ui.git(blog, "remote", "get-url", "origin") === "git@github.com-home-gh:me/blog.git", "workspace repository remote uses the workspace identity");

  // ---- settings: push protection off -> on
  ui.git(ui.home, "config", "--global", "core.hooksPath", path.join(ui.home, "other-hooks"));
  await ui.open();
  expect((await ui.text("#appbar .status-pill")).includes("Unprotected"), "top bar shows that pushes are not checked");
  await ui.click("#appbar .status-pill");
  await ui.waitText("Push protection is off");
  expect((await ui.text("#view")).includes("other-hooks"), "settings explain why protection is off");
  await ui.shot("s4-protection-off");
  await ui.click('#view [data-action="turn-on-protection"]');
  await ui.page.waitForSelector("#appbar .status-pill.is-on", { timeout: 15000 });
  expect(ui.git(ui.home, "config", "--global", "core.hooksPath").endsWith(".bgit/hooks"), "push protection is restored");
  await ui.shot("s5-protection-on");

  // ---- encrypted backup and restore
  await ui.click('[data-action="settings-section"][data-section="backup"]');
  const archive = path.join(ui.home, "backups", "test-backup.bgit");
  fs.mkdirSync(path.dirname(archive), { recursive: true });
  await ui.clickText("Export backup", "#exportForm button[type=submit]");
  expect((await ui.text("#exportError")).includes("Enter a password"), "export needs a password");
  await ui.type("#exportPath", archive);
  await ui.type("#exportPassword", "correct horse");
  await ui.type("#exportConfirm", "battery staple");
  await ui.clickText("Export backup", "#exportForm button[type=submit]");
  expect((await ui.text("#exportError")).includes("do not match"), "export checks the confirmation");
  await ui.type("#exportConfirm", "correct horse");
  await ui.shot("s6-backup");
  await ui.clickText("Export backup", "#exportForm button[type=submit]");
  await ui.waitText("Backup created");
  expect(fs.existsSync(archive), "backup file written");
  expect(!fs.readFileSync(archive).includes("dev@work.example"), "backup file is not readable as plain text");
  // wrong password is refused and nothing changes
  await ui.type("#importPath", archive);
  await ui.type("#importPassword", "wrong password");
  await ui.clickText("Restore backup", "#importForm button[type=submit]");
  await ui.waitDialog("confirmDialog");
  expect((await ui.text("#confirmDialog")).includes("replaced by the backup"), "restore says what it replaces");
  await ui.submitDialog("confirmDialog");
  await ui.page.waitForFunction(() => !document.getElementById("importError").hidden, { timeout: 30000 });
  console.log("  wrong password:", await ui.text("#importError"));
  expect((await ui.call("GetStatus")).identityCount === 2, "a failed restore changes nothing");
  await ui.type("#importPassword", "correct horse");
  await ui.clickText("Restore backup", "#importForm button[type=submit]");
  await ui.waitDialog("confirmDialog");
  await ui.submitDialog("confirmDialog");
  await ui.waitText("Backup restored");
  expect((await ui.text("#view")).includes("2 identities"), "restore reports what came back");
  expect((await ui.call("GetStatus")).identityCount === 2, "identities are present after restore");
  // A restored backup puts each identity's key at ~/.ssh/bgit_<identity>.
  expect(fs.existsSync(path.join(ui.home, ".ssh/bgit_work")), "restore brought the SSH key back");
  await ui.shot("s7-backup-restored");

  // ---- general
  await ui.click('[data-action="settings-section"][data-section="general"]');
  expect((await ui.text("#view")).includes("Keyboard shortcuts"), "general settings list the shortcuts");
  await ui.shot("s8-settings-general");

  // ---- remove bgit
  await ui.click('[data-action="settings-section"][data-section="advanced"]');
  await ui.click("#removeKeys");
  await ui.shot("s9-settings-advanced");
  await ui.click('#view [data-action="remove-bgit"]');
  await ui.waitDialog("confirmDialog");
  expect((await ui.text("#confirmDialog")).includes("cannot be recovered"), "removal warns about deleting keys");
  await ui.shot("s10-remove-confirm");
  await ui.submitDialog("confirmDialog");
  await ui.waitText("bgit was removed from this computer");
  await ui.settle(400);
  await ui.shot("s11-removed");
  expect(await ui.page.$eval("#onboarding", (el) => el.hidden), "the removal summary is not covered by the setup guide");
  expect(!fs.existsSync(path.join(ui.home, ".bgit")), "settings folder is deleted");
  expect(ui.git(blog, "remote", "get-url", "origin") === "git@github.com:me/blog.git", "repository remote is back to a standard GitHub address");
  expect(ui.git(ui.home, "config", "--global", "user.email") === "before@example.com", "previous Git identity is restored");
  expect(!fs.existsSync(path.join(ui.home, ".ssh/bgit_work")) && !fs.existsSync(path.join(ui.home, ".ssh/bgit_work.pub")), "the identity's key is deleted");
  expect((await ui.text("#view")).includes("SSH key files deleted"), "the summary lists the deleted key files");
  let hooks = "";
  try { hooks = ui.git(ui.home, "config", "--global", "core.hooksPath"); } catch (e) { hooks = ""; }
  expect(!hooks.includes(".bgit"), "hooks path no longer points at bgit (now: '" + hooks + "')");
  // starting again shows the setup guide
  await ui.click('#view [data-action="restart-setup"]');
  await ui.waitText("One computer, several Git identities");
  expect(!fs.existsSync(path.join(ui.home, ".bgit")), "opening the setup guide does not recreate settings");
}, { width: 1366, height: 699 });
