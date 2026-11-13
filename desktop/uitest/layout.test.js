// Window sizes, smooth re-drawing, themes, keyboard shortcuts, and picking up
// changes made outside the app (for example with the bgit CLI).
const { run, expect } = require("./lib");
const fs = require("node:fs");
const path = require("node:path");

// The inside of a maximised window on a 1366x768 screen.
run(async (ui) => {
  await ui.open();
  await ui.seedIdentities();
  ui.addGitHubRepo("work-gh", "acme", "storefront", { "README.md": "# Storefront\n\nline two\n" });
  const projects = path.join(ui.home, "projects");
  fs.mkdirSync(projects, { recursive: true });
  const cloned = await ui.call("CloneRepository", { url: "acme/storefront", parentDir: projects, folderName: "", alias: "work" });
  const repo = cloned.repository.path;
  ui.write(path.join(repo, "README.md"), "# Storefront\n\n" + Array.from({ length: 120 }, (_, i) => `changed line ${i}`).join("\n") + "\n");
  for (let i = 0; i < 30; i++) {
    ui.write(path.join(repo, `src/file-${i}.js`), `export const value${i} = ${i};\n`);
  }
  const overflow = () => ui.page.evaluate(() => document.documentElement.scrollWidth - window.innerWidth);

  await ui.open();
  await ui.nav("repositories");
  await ui.waitText("storefront");

  // ---- navigation is in the top bar, in the agreed order
  const tabs = await ui.page.$$eval("#appbar .nav-tab", (els) => els.map((el) => el.innerText.trim()));
  expect(tabs.join(",") === "Dashboard,Repositories,Identities,Workspaces,Doctor", "top navigation lists the five areas");
  expect((await ui.page.$$("aside.sidebar, #sidebarNav")).length === 0, "there is no sidebar");

  // ---- re-drawing the same page must not replay the page animation
  const entered = await ui.page.evaluate(() => {
    navigate("identities");
    return document.querySelector(".page-root").classList.contains("page-enter");
  });
  expect(entered, "a new page fades in");
  await ui.page.evaluate(() => render());
  expect(!(await ui.page.$eval(".page-root", (el) => el.classList.contains("page-enter"))), "re-drawing the same page does not animate again");

  // ---- keyboard shortcuts
  const chord = async (key) => { await ui.page.keyboard.down("Control"); await ui.page.keyboard.press(key); await ui.page.keyboard.up("Control"); await ui.settle(200); };
  await chord("4");
  expect(await ui.page.evaluate(() => state.page === "workspaces"), "Ctrl+4 opens Workspaces");
  await chord(",");
  expect(await ui.page.evaluate(() => state.page === "settings"), "Ctrl+, opens Settings");
  await chord("2");
  expect(await ui.page.evaluate(() => state.page === "repositories"), "Ctrl+2 opens Repositories");
  await ui.page.keyboard.press("/");
  expect(await ui.page.evaluate(() => document.activeElement.id === "repoSearch"), "/ focuses the repository filter");
  await ui.page.evaluate(() => document.activeElement.blur());

  // ---- the repository view fits the window: nothing below the bottom edge
  await ui.click(`[data-action="open-repo"][data-path="${repo}"]`);
  await ui.waitText("31 of 31 files selected");
  await ui.page.waitForSelector(".diff tr.add");
  await ui.settle(200);
  const measure = () => ui.page.evaluate(() => {
    const bottom = (selector) => Math.round(document.querySelector(selector).getBoundingClientRect().bottom);
    const view = document.getElementById("view");
    return {
      height: window.innerHeight,
      commit: bottom("#commitButton"),
      split: bottom(".changes"),
      pageScroll: view.scrollHeight - view.clientHeight,
      listScrolls: document.querySelector("#fileRows").scrollHeight > document.querySelector("#fileRows").clientHeight,
      diffScrolls: document.querySelector(".diff-scroll").scrollHeight > document.querySelector(".diff-scroll").clientHeight,
    };
  });
  let box = await measure();
  expect(box.commit <= box.height && box.split <= box.height, `commit button is inside the window (${box.commit} of ${box.height})`);
  expect(box.pageScroll === 0, "the repository page itself does not scroll");
  expect(box.listScrolls && box.diffScrolls, "the file list and the diff scroll on their own");
  expect((await overflow()) <= 0, "nothing sticks out sideways at 1366 wide");
  await ui.shot("layout-repo-1366x768");

  // ---- a narrow window keeps every part reachable
  await ui.page.setViewport({ width: 960, height: 640 });
  await ui.settle(300);
  box = await measure();
  expect(box.commit <= box.height, `commit button is inside a 960x640 window (${box.commit} of ${box.height})`);
  expect((await overflow()) <= 0, "nothing sticks out sideways at 960 wide");
  const navVisible = await ui.page.$$eval("#appbar .nav-tab", (els) => els.every((el) => { const r = el.getBoundingClientRect(); return r.width > 0 && r.right <= window.innerWidth; }));
  expect(navVisible, "navigation stays visible in a narrow window");
  const toolsVisible = await ui.page.$eval('#wsHead [data-action="ws-push"]', (el) => { const r = el.getBoundingClientRect(); return r.right <= window.innerWidth && r.left >= 0; });
  expect(toolsVisible, "push stays reachable in a narrow window");
  await ui.shot("layout-repo-960x640");
  for (const page of ["dashboard", "repositories", "identities", "workspaces", "doctor", "settings"]) {
    await ui.nav(page);
    await ui.settle(250);
    expect((await overflow()) <= 0, `${page} fits a 960 wide window`);
    await ui.shot(`layout-${page}-960x640`);
  }
  await ui.page.setViewport({ width: 1366, height: 699 });

  // ---- a dialog taller than the window scrolls instead of being cut off
  await ui.nav("identities");
  await ui.click('#view .page-head [data-action="add-identity"]');
  await ui.waitDialog("identityDialog");
  await ui.page.setViewport({ width: 1366, height: 420 });
  await ui.settle(300);
  const dialog = await ui.page.$eval("#identityDialog", (el) => {
    const rect = el.getBoundingClientRect();
    const submit = el.querySelector("[data-dialog-submit]").getBoundingClientRect();
    const body = el.querySelector(".dialog-body");
    return { top: Math.round(rect.top), bottom: Math.round(rect.bottom), submitBottom: Math.round(submit.bottom), scrolls: body.scrollHeight > body.clientHeight || el.scrollHeight > el.clientHeight, height: window.innerHeight };
  });
  expect(dialog.top >= 0 && dialog.bottom <= dialog.height && dialog.scrolls, "a tall dialog stays inside the window and scrolls");
  expect(dialog.submitBottom <= dialog.height, "the dialog's buttons stay visible");
  await ui.shot("layout-dialog-short-window");
  await ui.page.setViewport({ width: 1366, height: 699 });
  await ui.page.keyboard.press("Escape");
  await ui.waitDialogClosed("identityDialog");
  expect(true, "Escape closes a dialog");

  // ---- themes
  await ui.nav("settings");
  await ui.click('[data-action="settings-section"][data-section="appearance"]');
  expect(await ui.page.evaluate(() => document.documentElement.dataset.theme === "dark"), "dark is the default theme");
  await ui.click('[data-action="set-theme"][data-theme-choice="light"]');
  expect(await ui.page.evaluate(() => document.documentElement.dataset.theme === "light"), "light theme applies at once");
  await ui.shot("layout-light-settings");
  await ui.open();
  expect(await ui.page.evaluate(() => document.documentElement.dataset.theme === "light"), "the theme is remembered");
  await ui.shot("layout-light-dashboard");
  await ui.nav("repositories");
  await ui.click(`[data-action="open-repo"][data-path="${repo}"]`);
  await ui.page.waitForSelector(".diff tr.add");
  await ui.shot("layout-light-repo");
  await ui.nav("identities");
  await ui.shot("layout-light-identities");
  await ui.nav("doctor");
  await ui.shot("layout-light-doctor");
  await ui.page.evaluate(() => { prefs.set("theme", "dark"); applyTheme(); });

  // ---- a change made outside the app shows up when the window is focused again
  expect((await ui.text("#identitySwitch")).includes("work"), "top bar shows the default identity");
  // Stands in for `bgit use home` in a terminal: the backend state changes
  // without the page being told.
  await ui.call("ActivateIdentity", "home");
  await ui.focusWindow();
  await ui.page.waitForFunction(() => document.getElementById("identitySwitch").innerText.includes("home"), { timeout: 15000 });
  expect(true, "an identity switched outside the app appears after returning to the window");
  expect(!(await ui.page.$eval(".page-root", (el) => el.classList.contains("page-enter"))), "picking up outside changes does not animate the page");
}, { width: 1366, height: 699 });
