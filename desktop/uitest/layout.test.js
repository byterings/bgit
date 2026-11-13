// Small laptop screens, smooth re-drawing, and picking up changes made
// outside the app (for example with the bgit CLI).
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

  await ui.open();
  await ui.click("#sidebarNav [data-page=repositories]");
  await ui.waitText("storefront");

  // ---- re-drawing the same page must not replay the page animation
  const entered = await ui.page.evaluate(() => {
    transitionToPage("identities");
    return document.querySelector(".page-view").classList.contains("page-enter");
  });
  expect(entered, "a new page fades in");
  await ui.page.evaluate(() => render(currentStatus, currentDoctor));
  expect(!(await ui.page.$eval(".page-view", (el) => el.classList.contains("page-enter"))), "re-drawing the same page does not animate again");

  // ---- the repository view fits the window: nothing below the bottom edge
  await ui.click("#sidebarNav [data-page=repositories]");
  await ui.waitText("storefront");
  await ui.clickText("Open", "#app button.primary");
  await ui.waitText("31 changed files");
  await ui.page.waitForSelector(".diff tr.add");
  await ui.settle(200);
  const box = await ui.page.evaluate(() => {
    const bottom = (selector) => Math.round(document.querySelector(selector).getBoundingClientRect().bottom);
    const shell = document.getElementById("contentShell");
    return {
      height: window.innerHeight,
      commit: bottom("#commitButton"),
      split: bottom(".repo-split"),
      pageScroll: shell.scrollHeight - shell.clientHeight,
      listScrolls: document.querySelector(".file-rows").scrollHeight > document.querySelector(".file-rows").clientHeight,
      diffScrolls: document.querySelector(".diff-scroll").scrollHeight > document.querySelector(".diff-scroll").clientHeight,
    };
  });
  expect(box.commit <= box.height && box.split <= box.height, `commit button is inside the window (${box.commit} of ${box.height})`);
  expect(box.pageScroll === 0, "the repository page itself does not scroll");
  expect(box.listScrolls && box.diffScrolls, "the file list and the diff scroll on their own");
  await ui.shot("layout-repo-1366x768");

  // ---- a dialog taller than the window scrolls instead of being cut off
  await ui.click("#sidebarNav [data-page=identities]");
  await ui.clickText("Add Identity");
  await ui.page.setViewport({ width: 1366, height: 420 });
  await ui.settle(200);
  const dialog = await ui.page.$eval("#addDialog", (el) => {
    const rect = el.getBoundingClientRect();
    return { top: Math.round(rect.top), bottom: Math.round(rect.bottom), scrolls: el.scrollHeight > el.clientHeight, height: window.innerHeight };
  });
  expect(dialog.top >= 0 && dialog.bottom <= dialog.height && dialog.scrolls, "a tall dialog stays inside the window and scrolls");
  await ui.page.setViewport({ width: 1366, height: 699 });
  await ui.click("#cancelAdd");

  // ---- a change made outside the app shows up when the window is focused again
  expect((await ui.text("#topbar")).includes("work"), "top bar shows the active identity");
  // Stands in for `bgit use home` in a terminal: the backend state changes
  // without the page being told.
  await ui.call("ActivateIdentity", "home");
  await ui.page.evaluate(() => window.dispatchEvent(new Event("focus")));
  await ui.page.waitForFunction(() => document.getElementById("topbar").innerText.includes("home"), { timeout: 15000 });
  expect(true, "an identity switched outside the app appears after returning to the window");
  expect(!(await ui.page.$eval(".page-view", (el) => el.classList.contains("page-enter"))), "picking up outside changes does not animate the page");
}, { width: 1366, height: 699 });
