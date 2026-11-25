// Dashboard and Doctor: real numbers, attention banners, and health checks.
const { run, expect } = require("./lib");
const fs = require("node:fs");
const path = require("node:path");
run(async (ui) => {
  await ui.open();
  await ui.seedIdentities();
  await ui.open();
  expect((await ui.text("#view")).includes("Bring in your first repository"), "an empty dashboard guides to the first repository");

  ui.addGitHubRepo("work-gh", "acme", "storefront");
  const projects = path.join(ui.home, "projects");
  fs.mkdirSync(projects, { recursive: true });
  const cloned = await ui.call("CloneRepository", { url: "acme/storefront", parentDir: projects, folderName: "", alias: "work" });
  const repo = cloned.repository.path;
  const second = path.join(projects, "notes");
  fs.mkdirSync(second, { recursive: true });
  ui.git(second, "init", "-q", "-b", "main");
  ui.git(second, "remote", "add", "origin", "https://github.com/home-gh/notes.git");
  await ui.call("AddRepository", { path: second, alias: "home" });
  ui.write(path.join(repo, "README.md"), "changed\n");
  ui.git(second, "config", "--local", "user.email", "wrong@example.com");

  await ui.open();
  await ui.waitText("storefront");
  const dashboard = await ui.text("#view");
  expect(dashboard.includes("Welcome back, Work"), "dashboard greets the default identity");
  expect(dashboard.includes("1 repository needs attention"), "dashboard reports the repository that needs attention");
  const stats = await ui.page.$$eval("#view .stat", (els) => els.map((el) => el.innerText.replace(/\s+/g, " ").trim()));
  console.log("  stats:", stats.join(" | "));
  expect(stats.some((s) => /^2\b/.test(s)) && stats.length >= 3, "stats come from the real repository list");
  await ui.shot("d1-dashboard");

  // a stat opens the list with that filter
  await ui.click('#view .stat[data-filter="changes"]');
  await ui.page.waitForSelector("#repoRows .repo-row");
  expect((await ui.page.$$("#repoRows .repo-row")).length === 1 && (await ui.text("#repoRows")).includes("storefront"), "a stat opens the filtered repository list");
  await ui.click('[data-action="clear-filters"], [data-action="set-filter"][data-filter="all"]');

  // ---- doctor
  await ui.nav("doctor");
  await ui.waitText("Push protection");
  const doctor = await ui.text("#view");
  for (const title of ["Configuration", "SSH keys", "SSH agent", "Git identity", "Push protection", "Repositories"]) {
    expect(doctor.includes(title), `doctor has a ${title} section`);
  }
  expect(doctor.includes("2 identities configured"), "counts read naturally");
  expect(!doctor.includes("(s)"), "no '(s)' wording");
  await ui.shot("d2-doctor");
  await ui.click('#view [data-action="run-doctor"]');
  await ui.waitText("Checks finished");
  await ui.click('#view [data-action="fix-all-repos"]');
  await ui.page.waitForFunction(() => !document.querySelector('#view [data-action="fix-all-repos"]'), { timeout: 15000 });
  expect(ui.git(second, "config", "--local", "user.email") === "dev@home.example", "doctor fixes the repository");

  // ---- activity in this session
  await ui.click('#appbar [data-action="appbar-activity"]');
  await ui.page.waitForSelector(".popover .activity-pop");
  expect((await ui.text(".popover .activity-pop")).includes("Doctor finished"), "activity lists what was done in this session");
  await ui.shot("d3-activity");
  await ui.page.keyboard.press("Escape");

  // ---- back on the dashboard everything is healthy
  await ui.nav("dashboard");
  await ui.waitGone("needs attention");
  await ui.shot("d4-dashboard-healthy");
}, { width: 1366, height: 699 });
