// Identity manager: add, inspect, connect, edit, switch the default, delete.
const { run, expect } = require("./lib");
const fs = require("node:fs");
const path = require("node:path");
run(async (ui) => {
  await ui.open();
  // fast path through the setup guide using the backend directly
  await ui.call("RunSetup");
  await ui.call("AddIdentity", { alias: "work", name: "Work Dev", email: "dev@work.example", githubUsername: "work-gh-test", sshKeyPath: "", generateSSHKey: true });
  await ui.open();
  expect(await ui.page.$eval("#onboarding", (el) => el.hidden), "no setup guide once setup and identity exist");
  await ui.nav("identities");
  await ui.waitText("Git configuration");
  expect((await ui.text(".identity-detail")).includes("Default identity"), "the default identity is marked");
  expect(!(await ui.text(".identity-detail")).includes("BEGIN OPENSSH PRIVATE KEY"), "private key is never shown");

  // ---- add
  await ui.click('#view .page-head [data-action="add-identity"]');
  await ui.waitDialog("identityDialog");
  await ui.submitDialog("identityDialog");
  expect((await ui.dialogError("identityDialog")).includes("Enter a short name"), "identity form validates");
  await ui.type("#identityDialog input[name=alias]", "home");
  await ui.type("#identityDialog input[name=githubUsername]", "home-gh-test");
  await ui.type("#identityDialog input[name=name]", "Home Dev");
  await ui.type("#identityDialog input[name=email]", "dev@home.example");
  await ui.shot("i1-add-identity-dialog");
  await ui.submitDialog("identityDialog");
  await ui.waitDialogClosed("identityDialog");
  expect((await ui.text(".identity-detail h1")).trim() === "home", "the new identity is selected");
  expect(fs.existsSync(path.join(ui.home, ".ssh/bgit_home-gh-test")), "a key was created for the new identity");
  expect((await ui.text("#identitySwitch")).includes("work"), "adding an identity does not change the default");

  // ---- connect
  await ui.click('.identity-detail [data-action="copy-open-github"]');
  expect((await ui.page.evaluate(() => window.__harnessClipboard)).startsWith("ssh-ed25519"), "public key copied");
  expect((await ui.page.evaluate(() => window.__harnessOpened.length)) === 1, "GitHub key settings opened");
  await ui.click('.identity-detail [data-action="test-connection"]');
  await ui.page.waitForSelector(".connection-status .banner", { timeout: 40000 });
  console.log("  result:", (await ui.text(".connection-status")).replace(/\n/g, " | "));
  await ui.shot("i2-identity-detail");

  // ---- edit
  await ui.click('.identity-detail [data-action="edit-identity"]');
  await ui.waitDialog("identityDialog");
  expect(await ui.page.$eval("#identityDialog input[name=alias]", (el) => el.readOnly), "the short name cannot be edited");
  await ui.type("#identityDialog input[name=name]", "Home Developer");
  await ui.shot("i3-edit-identity-dialog");
  await ui.submitDialog("identityDialog");
  await ui.waitDialogClosed("identityDialog");
  expect((await ui.text(".identity-detail")).includes("Home Developer"), "edited name is shown");

  // ---- default identity: from the detail view, then from the top bar
  await ui.click('.identity-detail [data-action="set-default"]');
  await ui.page.waitForFunction(() => document.getElementById("identitySwitch").innerText.includes("home"), { timeout: 15000 });
  expect(ui.git(ui.home, "config", "--global", "user.email") === "dev@home.example", "global git identity follows the default identity");
  await ui.click("#identitySwitch");
  await ui.page.waitForSelector(".popover.identity-menu");
  expect((await ui.text(".popover.identity-menu")).includes("Bound repositories and workspaces keep theirs"), "the switcher explains what switching does not change");
  await ui.shot("i4-identity-switcher");
  await ui.menuItem("dev@work.example");
  await ui.page.waitForFunction(() => document.getElementById("identitySwitch").innerText.includes("work"), { timeout: 15000 });
  expect(ui.git(ui.home, "config", "--global", "user.email") === "dev@work.example", "switching in the top bar changes the default identity");

  // ---- delete
  await ui.click('.identity-card[data-alias="home"]');
  await ui.click('.identity-detail [data-action="identity-menu"]');
  await ui.menuItem("Delete identity");
  await ui.waitDialog("deleteIdentityDialog");
  await ui.shot("i5-delete-identity-dialog");
  await ui.submitDialog("deleteIdentityDialog");
  await ui.waitDialogClosed("deleteIdentityDialog");
  expect((await ui.page.$$(".identity-card")).length === 1, "the identity is gone from the list");
  expect(fs.existsSync(path.join(ui.home, ".ssh/bgit_home-gh-test")), "key files are kept unless asked otherwise");
}, { width: 1366, height: 699 });
