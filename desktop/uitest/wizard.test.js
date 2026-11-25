// First run: the setup guide from a clean machine to a working app.
const { run, expect } = require("./lib");
const fs = require("node:fs");
const path = require("node:path");
run(async (ui) => {
  await ui.open();
  expect((await ui.text("#onboarding")).includes("One computer, several Git identities"), "setup guide opens on a fresh machine");
  expect((await ui.text("#onboarding")).includes("is installed"), "git is detected");
  await ui.shot("w1-welcome");

  // restoring a backup is offered up front and can be left again
  await ui.click('#onboarding [data-ob="restore"]');
  expect((await ui.text("#onboarding")).includes("Restore from a backup"), "restore step shown");
  await ui.clickText("Restore", "#obRestoreForm button[type=submit]");
  expect((await ui.text("#onboarding")).includes("Choose the backup file"), "restore validates the backup file");
  await ui.shot("w1b-restore");
  await ui.click('#onboarding [data-ob="back-welcome"]');

  await ui.clickText("Get started", "#onboarding button");
  expect((await ui.text("#onboarding")).includes("Turn on push protection"), "push protection step shown");
  await ui.shot("w2-protection");

  await ui.clickText("Set up bgit", "#onboarding button");
  await ui.waitText("Add your first identity");
  expect(fs.existsSync(path.join(ui.home, ".bgit/hooks/pre-push")), "pre-push hook installed by the app");
  const hook = fs.readFileSync(path.join(ui.home, ".bgit/hooks/pre-push"), "utf8");
  expect(hook.includes("BGIT_HELPER="), "hook has a fallback helper path");
  expect(ui.git(ui.home, "config", "--global", "core.hooksPath").endsWith(".bgit/hooks"), "global hooks path set");

  // validation
  await ui.clickText("Add identity", "#obIdentityForm button[type=submit]");
  expect((await ui.text("#onboarding")).includes("Enter a short name"), "empty form shows a validation message");
  await ui.type("#obIdentityForm input[name=alias]", "work");
  await ui.type("#obIdentityForm input[name=githubUsername]", "work-gh-test");
  await ui.type("#obIdentityForm input[name=name]", "Work Dev");
  await ui.type("#obIdentityForm input[name=email]", "not-an-email");
  await ui.clickText("Add identity", "#obIdentityForm button[type=submit]");
  expect((await ui.text("#onboarding")).includes("valid email"), "bad email rejected");
  expect(await ui.page.$eval("#obIdentityForm input[name=alias]", (el) => el.value) === "work", "form keeps typed values after an error");
  await ui.type("#obIdentityForm input[name=email]", "dev@work.example");
  await ui.shot("w3-identity");
  await ui.clickText("Add identity", "#obIdentityForm button[type=submit]");
  await ui.waitText("Add your SSH key to GitHub");
  expect(fs.existsSync(path.join(ui.home, ".ssh/bgit_work-gh-test.pub")), "ssh key generated");
  expect((await ui.text("#onboarding")).includes("ssh-ed25519"), "public key shown");
  await ui.shot("w4-github");

  await ui.click('#onboarding [data-ob="open-github"]');
  expect((await ui.page.evaluate(() => window.__harnessClipboard)).startsWith("ssh-ed25519"), "key copied to clipboard");
  expect((await ui.page.evaluate(() => window.__harnessOpened.length)) === 1, "GitHub settings opened");

  await ui.click('#onboarding [data-ob="test"]');
  await ui.page.waitForFunction(() => [...document.querySelectorAll("#onboarding .banner strong")].some((el) => /^(Connected|Not connected yet)$/.test(el.innerText.trim())), { timeout: 40000 });
  await ui.settle();
  console.log("  connection result:", (await ui.text("#onboarding .ob-numbered .banner")).replace(/\n/g, " | "));
  await ui.shot("w4b-github-tested");

  await ui.click('#onboarding [data-ob="to-done"]');
  expect((await ui.text("#onboarding")).includes("You're ready"), "done step shown");
  await ui.shot("w5-done");
  await ui.click('#onboarding [data-ob="finish"]');
  await ui.settle(600);
  expect(await ui.page.$eval("#onboarding", (el) => el.hidden), "setup guide closes");
  expect((await ui.text("#identitySwitch")).includes("work"), "new identity is the default in the app");
  expect((await ui.text("#appbar .status-pill")).includes("Protected"), "top bar shows push protection is on");
  expect(ui.git(ui.home, "config", "--global", "user.email") === "dev@work.example", "global git identity set to the first identity");
  expect((await ui.text("#view")).includes("Bring in your first repository"), "dashboard guides to the first repository");
  await ui.shot("w6-dashboard-empty");

  // reopening the app must not show the guide again
  await ui.open();
  expect(await ui.page.$eval("#onboarding", (el) => el.hidden), "setup guide stays closed on next launch");

  // and it can be opened again from Settings
  await ui.nav("settings");
  await ui.click('#view [data-action="open-setup"]');
  expect(!(await ui.page.$eval("#onboarding", (el) => el.hidden)), "Settings reopens the setup guide");
}, { width: 1366, height: 699 });
