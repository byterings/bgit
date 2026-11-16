const { run, expect } = require("./lib");
run(async (ui) => {
  await ui.open();
  // fast path through the wizard using the backend directly
  await ui.call("RunSetup");
  await ui.call("AddIdentity", { alias: "work", name: "Work Dev", email: "dev@work.example", githubUsername: "work-gh-test", sshKeyPath: "", generateSSHKey: true });
  await ui.open();
  expect(await ui.page.$eval("#onboarding", (el) => el.hidden), "no wizard once setup and identity exist");
  await ui.click("#sidebarNav [data-page=identities]");
  await ui.clickText("Add Identity");
  await ui.type("#addForm input[name=alias]", "home");
  await ui.type("#addForm input[name=githubUsername]", "home-gh-test");
  await ui.type("#addForm input[name=name]", "Home Dev");
  await ui.type("#addForm input[name=email]", "dev@home.example");
  await ui.click("#addForm input[name=generateSSHKey]");
  await ui.clickText("Add Identity", "#addForm button");
  await ui.page.waitForSelector("#viewDialog[open]", { timeout: 15000 });
  await ui.settle();
  expect((await ui.text("#viewDialog")).includes("home"), "details dialog opens for the new identity");
  await ui.clickText("Copy key and open GitHub", "#viewDialog button");
  expect((await ui.page.evaluate(() => window.__harnessClipboard)).startsWith("ssh-ed25519"), "key copied from the details dialog");
  await ui.clickText("Test connection", "#viewDialog button");
  await ui.page.waitForFunction(() => !document.getElementById("viewConnectionResult").hidden, { timeout: 40000 });
  console.log("  result:", (await ui.text("#viewConnectionResult")).replace(/\n/g, " | "));
  await ui.shot("i1-identity-details");
});
