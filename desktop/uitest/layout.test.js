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
  const tabs = await ui.page.$$eval("#appbar .nav-tab[data-nav]", (els) => els.map((el) => el.innerText.trim()));
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

  // ---- the repository view at common desktop window sizes: everything fits,
  // nothing in the top bar or the repository header overlaps, every action
  // stays inside the window.
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
  // Pairs of visible controls in one bar whose boxes overlap, and controls
  // that leave the window.
  const collisions = () => ui.page.evaluate(() => {
    const shown = (el) => el.offsetParent !== null && !el.hidden && el.getBoundingClientRect().width > 0;
    const name = (el) => el.id || el.dataset.action || el.dataset.nav || el.className.split(" ")[0] || el.tagName;
    const groups = {
      appbar: document.querySelectorAll("#appbar .brand, #appbar .nav-tab, #globalSearchBox, #appbar .appbar-right > *"),
      header: document.querySelectorAll("#wsHead .ws-title > *, #wsHead .ws-tools > *:not(.ws-tools-sep)"),
    };
    const problems = [];
    for (const [group, list] of Object.entries(groups)) {
      const items = [...list].filter(shown).map((el) => ({ name: name(el), r: el.getBoundingClientRect() }));
      for (const item of items) {
        if (item.r.left < 0 || item.r.right > window.innerWidth + 0.5) problems.push(`${group}: ${item.name} leaves the window`);
      }
      for (let i = 0; i < items.length; i++) {
        for (let j = i + 1; j < items.length; j++) {
          const a = items[i].r, b = items[j].r;
          const x = Math.min(a.right, b.right) - Math.max(a.left, b.left);
          const y = Math.min(a.bottom, b.bottom) - Math.max(a.top, b.top);
          if (x > 1 && y > 1) problems.push(`${group}: ${items[i].name} overlaps ${items[j].name}`);
        }
      }
    }
    return problems;
  });
  const reachable = () => ui.page.evaluate(() => {
    const tabs = [...document.querySelectorAll("#appbar .nav-tab[data-nav]")];
    const more = document.getElementById("navMore");
    return tabs.every((tab) => !tab.hidden) || (!more.hidden && more.getBoundingClientRect().right <= window.innerWidth);
  });

  for (const [width, height] of [[1366, 768], [1280, 720], [1024, 768], [960, 640]]) {
    const size = `${width}x${height}`;
    await ui.page.setViewport({ width, height });
    await ui.settle(300);
    const box = await measure();
    expect(box.commit <= box.height && box.split <= box.height, `${size}: commit button is inside the window (${box.commit} of ${box.height})`);
    expect(box.pageScroll === 0, `${size}: the repository page itself does not scroll`);
    expect(box.listScrolls && box.diffScrolls, `${size}: the file list and the diff scroll on their own`);
    expect((await overflow()) <= 0, `${size}: nothing sticks out sideways`);
    const clashes = await collisions();
    expect(clashes.length === 0, `${size}: no overlapping controls${clashes.length ? ` (${clashes.join("; ")})` : ""}`);
    expect(await reachable(), `${size}: every main area is reachable from the top bar`);
    const tools = await ui.page.$$eval("#wsHead .ws-tools button", (els) => els.every((el) => { const r = el.getBoundingClientRect(); return r.right <= window.innerWidth && r.left >= 0 && r.width > 0; }));
    expect(tools, `${size}: branch, fetch, pull, push, and more stay reachable`);
    await ui.shot(`layout-repo-${size}`);
    for (const page of ["dashboard", "repositories", "identities", "workspaces", "doctor", "settings"]) {
      await ui.nav(page);
      await ui.settle(250);
      expect((await overflow()) <= 0, `${size}: ${page} fits sideways`);
      const pageClashes = await collisions();
      expect(pageClashes.length === 0, `${size}: ${page} top bar has no overlapping controls${pageClashes.length ? ` (${pageClashes.join("; ")})` : ""}`);
      await ui.shot(`layout-${page}-${size}`);
    }
    await ui.nav("repositories");
    await ui.click(`[data-action="open-repo"][data-path="${repo}"]`);
    await ui.page.waitForSelector(".diff tr.add");
    await ui.settle(200);
  }

  // ---- very narrow: the last tabs move into "More" instead of colliding
  await ui.page.setViewport({ width: 860, height: 600 });
  await ui.settle(300);
  const moreShown = await ui.page.$eval("#navMore", (el) => !el.hidden);
  expect(moreShown, "tabs that do not fit move into a More menu");
  expect((await collisions()).length === 0, "the narrow top bar has no overlapping controls");
  await ui.click("#navMore");
  await ui.page.waitForSelector(".popover .menu-item");
  await ui.shot("layout-nav-more-860");
  await ui.page.keyboard.press("Escape");
  await ui.nav("doctor");
  expect(await ui.page.evaluate(() => state.page === "doctor"), "a tab inside More still opens its page");
  expect(await ui.page.$eval("#navMore", (el) => el.classList.contains("active")), "More is marked when the open page is inside it");
  await ui.page.setViewport({ width: 1366, height: 699 });
  await ui.settle(300);
  expect(await ui.page.$eval("#navMore", (el) => el.hidden), "with room again every tab is back in the bar");

  // ---- panes: resize, collapse, and keyboard navigation in the file list
  await ui.nav("repositories");
  await ui.click(`[data-action="open-repo"][data-path="${repo}"]`);
  await ui.page.waitForSelector(".diff tr.add");
  await ui.settle(200);
  const sideWidth = () => ui.page.$eval("#sidePane", (el) => Math.round(el.getBoundingClientRect().width));
  const before = await sideWidth();
  const handle = await ui.page.$eval("#sideSplitter", (el) => { const r = el.getBoundingClientRect(); return { x: r.left + r.width / 2, y: r.top + r.height / 2 }; });
  await ui.page.mouse.move(handle.x, handle.y);
  await ui.page.mouse.down();
  await ui.page.mouse.move(handle.x + 120, handle.y, { steps: 6 });
  await ui.page.mouse.up();
  await ui.settle(200);
  const wider = await sideWidth();
  expect(wider >= before + 100, `dragging the splitter widens the file list (${before} -> ${wider})`);
  expect(await ui.page.evaluate(() => prefs.get("split.side", 0) > 0), "the new width is remembered");
  await ui.page.focus("#sideSplitter");
  await ui.page.keyboard.press("ArrowLeft");
  await ui.settle(100);
  expect((await sideWidth()) < wider, "the splitter also works with the keyboard");
  await ui.page.click("#sideSplitter", { clickCount: 2 });
  await ui.settle(150);
  expect(Math.abs((await sideWidth()) - 320) <= 2, "double-clicking the splitter restores the default width");
  const commitHeight = () => ui.page.$eval("#commitArea", (el) => Math.round(el.getBoundingClientRect().height));
  const commitBefore = await commitHeight();
  const rowHandle = await ui.page.$eval("#commitSplitter", (el) => { const r = el.getBoundingClientRect(); return { x: r.left + r.width / 2, y: r.top + r.height / 2 }; });
  await ui.page.mouse.move(rowHandle.x, rowHandle.y);
  await ui.page.mouse.down();
  await ui.page.mouse.move(rowHandle.x, rowHandle.y - 80, { steps: 6 });
  await ui.page.mouse.up();
  await ui.settle(150);
  expect((await commitHeight()) >= commitBefore + 60, "dragging up gives the commit message more room");
  await ui.shot("layout-repo-resized");

  // collapse the file list: the diff takes the width, commit stays one click away
  const diffWidth = () => ui.page.$eval("#diffPane", (el) => Math.round(el.getBoundingClientRect().width));
  const diffBefore = await diffWidth();
  await ui.click('#sidePane [data-action="side-toggle"]');
  expect((await diffWidth()) > diffBefore + 200, "hiding the file list gives the diff the room");
  expect((await ui.text("#railCount")) === "31/31", "the collapsed rail still shows how many files are selected");
  await ui.shot("layout-repo-collapsed");
  await ui.click('.side-rail [data-action="side-commit"]');
  expect(await ui.page.evaluate(() => document.activeElement?.id === "commitSummary"), "commit from the rail opens the panel at the summary");
  await ui.page.evaluate(() => document.activeElement.blur());
  await chord("b");
  expect(await ui.page.$eval("#changesLayout", (el) => el.classList.contains("side-collapsed")), `${"Ctrl"}+B hides the file list`);
  await chord("b");
  expect(await ui.page.$eval("#changesLayout", (el) => !el.classList.contains("side-collapsed")), `${"Ctrl"}+B shows it again`);

  // keyboard: arrows move through files, Space leaves one out of the commit
  await ui.click('.file-row[data-file="README.md"] .file-name');
  await ui.page.keyboard.press("ArrowDown");
  await ui.settle(300);
  const second = await ui.page.$eval(".file-row.selected", (el) => el.dataset.file);
  expect(second !== "README.md", "ArrowDown selects the next file");
  await ui.page.waitForFunction((f) => document.getElementById("diffPane").dataset.file === f, {}, second);
  await ui.page.keyboard.press(" ");
  await ui.settle(150);
  expect(await ui.page.$eval(".file-row.selected", (el) => el.classList.contains("is-excluded")), "Space leaves the selected file out and shows it as left out");
  expect((await ui.text("#selectionText")).startsWith("30 of 31"), "the selection count follows");
  await ui.page.keyboard.press(" ");
  await ui.settle(150);
  await ui.click('#diffPane [data-action="diff-step"][data-step="1"]');
  expect(await ui.page.$eval(".file-row.selected", (el, f) => el.dataset.file !== f, second), "the diff header steps to the next file");
  await ui.click('#diffPane [data-action="diff-wrap"]');
  expect(await ui.page.$eval("#diffPane .diff", (el) => el.classList.contains("wrap")), "long lines can be wrapped");
  await ui.click('#diffPane [data-action="diff-wrap"]');

  // the identity in the header explains itself
  await ui.click('#wsHead [data-action="ws-identity"]');
  await ui.page.waitForSelector(".identity-pop");
  const card = await ui.text(".identity-pop");
  expect(card.includes("Work Dev") && card.includes("bound"), "the header identity says who commits here and why");
  await ui.shot("layout-repo-identity-card");
  await ui.page.keyboard.press("Escape");
  await ui.click("#identitySwitch");
  expect((await ui.text(".popover")).includes("This repository"), "the identity switcher marks the identity of the open repository");
  await ui.page.keyboard.press("Escape");

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
