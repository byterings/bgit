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
  await ui.click("#sidebarNav [data-page=workspaces]");
  await ui.waitText("No workspaces yet");
  await ui.clickText("Add workspace", "#app button");
  await ui.page.waitForSelector("#workspaceDialog[open]");
  await ui.clickText("Add workspace", "#workspaceForm button[type=submit]");
  expect((await ui.text("#workspaceError")).includes("Choose the workspace folder"), "workspace form validates the folder");
  await ui.type("#workspaceForm input[name=path]", workspace);
  await ui.page.select("#workspaceForm select[name=alias]", "home");
  await ui.clickText("Add workspace", "#workspaceForm button[type=submit]");
  await ui.page.waitForFunction(() => !document.getElementById("workspaceDialog").open, { timeout: 15000 });
  await ui.waitText("1 repository");
  await ui.shot("s1-workspaces");
  // nested workspace is rejected with a reason
  await ui.clickText("Add workspace", "#app button");
  await ui.type("#workspaceForm input[name=path]", blog);
  await ui.clickText("Add workspace", "#workspaceForm button[type=submit]");
  await ui.page.waitForFunction(() => !document.getElementById("workspaceError").hidden, { timeout: 15000 });
  expect((await ui.text("#workspaceError")).includes("inside the workspace"), "nested workspace is rejected");
  await ui.click("#workspaceCancel");

  // the workspace repository shows up under Repositories and can be fixed
  await ui.clickText("View repositories", "#app button");
  await ui.waitText("blog");
  await ui.waitText("needs attention");
  expect((await ui.text("#app")).includes("From workspace"), "workspace repository is listed with its source");
  await ui.clickText("Fix it", "#app button");
  await ui.page.waitForFunction(() => !document.body.innerText.includes("needs attention"), { timeout: 15000 });
  expect(ui.git(blog, "config", "--local", "user.email") === "dev@home.example", "workspace repository now commits as the workspace identity");
  expect(ui.git(blog, "remote", "get-url", "origin") === "git@github.com-home-gh:me/blog.git", "workspace repository remote uses the workspace identity");

  // ---- settings: push protection off -> on
  ui.git(ui.home, "config", "--global", "core.hooksPath", path.join(ui.home, "other-hooks"));
  await ui.open();
  await ui.click("#sidebarNav [data-page=settings]");
  await ui.waitText("Turn on push protection");
  expect((await ui.text("#app")).includes("other-hooks"), "settings explain why protection is off");
  await ui.shot("s2-settings-off");
  await ui.clickText("Turn on push protection", "#app button");
  await ui.page.waitForFunction(() => !document.body.innerText.includes("Turn on push protection"), { timeout: 15000 });
  expect(ui.git(ui.home, "config", "--global", "core.hooksPath").endsWith(".bgit/hooks"), "push protection is restored");
  await ui.settle(400);
  await ui.shot("s3-settings");

  // ---- remove bgit
  await ui.click("#removeKeys");
  await ui.clickText("Remove bgit", "#app button.danger-solid");
  await ui.page.waitForSelector("#confirmDialog[open]");
  expect((await ui.text("#confirmDialog")).includes("cannot be recovered"), "removal warns about deleting keys");
  await ui.click("#confirmAccept");
  await ui.waitText("bgit was removed from this computer");
  await ui.settle(400);
  await ui.shot("s4-removed");
  expect(await ui.page.$eval("#onboarding", (el) => el.hidden), "the removal summary is not covered by the wizard");
  expect(!fs.existsSync(path.join(ui.home, ".bgit")), "settings folder is deleted");
  expect(ui.git(blog, "remote", "get-url", "origin") === "git@github.com:me/blog.git", "repository remote is back to a standard GitHub address");
  expect(ui.git(ui.home, "config", "--global", "user.email") === "before@example.com", "previous Git identity is restored");
  expect(!fs.existsSync(path.join(ui.home, ".ssh/bgit_work-gh")), "generated key is deleted");
  let hooks = "";
  try { hooks = ui.git(ui.home, "config", "--global", "core.hooksPath"); } catch (e) { hooks = ""; }
  expect(!hooks.includes(".bgit"), "hooks path no longer points at bgit (now: '" + hooks + "')");
  // starting again shows the wizard
  await ui.clickText("Set up bgit again", "#app button");
  await ui.waitText("Welcome to bgit");
  expect(!fs.existsSync(path.join(ui.home, ".bgit")), "opening the wizard does not recreate settings");
});
