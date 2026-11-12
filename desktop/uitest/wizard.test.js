const { run, expect } = require("./lib");
const fs = require("node:fs");
const path = require("node:path");
run(async (ui) => {
  await ui.open();
  expect((await ui.text("#onboarding")).includes("Welcome to bgit"), "wizard opens on a fresh machine");
  expect((await ui.text("#onboarding")).includes("is installed"), "git is detected");
  await ui.shot("w1-welcome");

  await ui.clickText("Get started");
  expect((await ui.text("#onboarding")).includes("Turn on push protection"), "setup step shown");
  await ui.shot("w2-setup");

  await ui.clickText("Set up bgit");
  await ui.waitText("Add your first identity");
  expect(fs.existsSync(path.join(ui.home, ".bgit/hooks/pre-push")), "pre-push hook installed by the app");
  const hook = fs.readFileSync(path.join(ui.home, ".bgit/hooks/pre-push"), "utf8");
  expect(hook.includes("BGIT_HELPER="), "hook has a fallback helper path");
  expect(ui.git(ui.home, "config", "--global", "core.hooksPath").endsWith(".bgit/hooks"), "global hooks path set");

  // validation
  await ui.clickText("Add identity");
  expect((await ui.text("#onboarding")).includes("Enter a short name"), "empty form shows a validation message");
  await ui.type("#obIdentityForm input[name=alias]", "work");
  await ui.type("#obIdentityForm input[name=githubUsername]", "work-gh-test");
  await ui.type("#obIdentityForm input[name=name]", "Work Dev");
  await ui.type("#obIdentityForm input[name=email]", "not-an-email");
  await ui.clickText("Add identity");
  expect((await ui.text("#onboarding")).includes("valid email"), "bad email rejected");
  expect(await ui.page.$eval("#obIdentityForm input[name=alias]", (el) => el.value) === "work", "form keeps typed values after an error");
  await ui.type("#obIdentityForm input[name=email]", "dev@work.example");
  await ui.shot("w3-identity");
  await ui.clickText("Add identity");
  await ui.waitText("Add your SSH key to GitHub");
  expect(fs.existsSync(path.join(ui.home, ".ssh/bgit_work-gh-test.pub")), "ssh key generated");
  expect((await ui.text("#onboarding")).includes("ssh-ed25519"), "public key shown");
  await ui.shot("w4-github");

  await ui.clickText("Copy key and open GitHub");
  expect((await ui.page.evaluate(() => window.__harnessClipboard)).startsWith("ssh-ed25519"), "key copied to clipboard");
  expect((await ui.page.evaluate(() => window.__harnessOpened.length)) === 1, "GitHub settings opened");

  await ui.clickText("Test connection");
  await ui.page.waitForFunction(() => document.querySelector("#onboarding .note") !== null, { timeout: 40000 });
  await ui.settle();
  console.log("  connection note:", (await ui.text("#onboarding .ob-numbered .note")).replace(/\n/g, " | "));
  await ui.shot("w4b-github-tested");

  await ui.clickText("Skip for now");
  expect((await ui.text("#onboarding")).includes("You're ready"), "done step shown");
  await ui.shot("w5-done");
  await ui.clickText("Open bgit");
  await ui.settle(600);
  expect(await ui.page.$eval("#onboarding", (el) => el.hidden), "wizard closes");
  expect((await ui.text("#topbar")).includes("work"), "new identity is active in the app");
  expect(ui.git(ui.home, "config", "--global", "user.email") === "dev@work.example", "global git identity set to the first identity");
  await ui.shot("w6-app");

  // reopening the app must not show the wizard again
  await ui.open();
  expect(await ui.page.$eval("#onboarding", (el) => el.hidden), "wizard stays closed on next launch");
});
