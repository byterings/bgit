const { run, expect } = require("./lib");
const fs = require("node:fs");
const path = require("node:path");
run(async (ui) => {
  await ui.open();
  await ui.seedIdentities();
  const bare = ui.addGitHubRepo("work-gh", "acme", "storefront", {
    "README.md": "# Storefront\n\nline two\nline three\n",
    "src/app.js": "function main() {\n  return 1;\n}\n",
    "docs/old.md": "old doc\n",
  });
  const projects = path.join(ui.home, "projects");
  fs.mkdirSync(projects, { recursive: true });
  const cloned = await ui.call("CloneRepository", { url: "acme/storefront", parentDir: projects, folderName: "", alias: "work" });
  const repo = cloned.repository.path;

  // working tree changes of every kind
  ui.write(path.join(repo, "README.md"), "# Storefront\n\nline 2 changed\nline three\nline four\n");
  ui.write(path.join(repo, "src/new-feature.js"), "export const feature = true;\n");
  fs.rmSync(path.join(repo, "docs/old.md"));

  await ui.open();
  await ui.click("#sidebarNav [data-page=repositories]");
  await ui.waitText("storefront");
  expect((await ui.text("#app")).includes("3 changes"), "list shows the number of changes");
  await ui.clickText("Open", "#app button.primary");
  await ui.waitText("3 changed files");
  await ui.page.waitForSelector(".diff tr.add");
  await ui.settle();
  const body = await ui.text("#repoBody");
  expect(body.includes("README.md") && body.includes("new-feature.js") && body.includes("old.md"), "changed files are listed");
  expect((await ui.text("#diffPane")).includes("line 2 changed"), "diff of the first file is shown");
  expect((await ui.text("#repoToolbar")).includes("work · dev@work.example"), "toolbar shows the repository identity");
  await ui.shot("c1-changes");

  // select another file
  await ui.click('.file-row[data-file="src/new-feature.js"] .file-name');
  await ui.page.waitForFunction(() => document.getElementById("diffPane").innerText.includes("export const feature"));
  expect(true, "selecting a file shows its diff");

  // commit button rules
  expect(await ui.page.$eval("#commitButton", (b) => b.disabled), "commit is disabled without a summary");
  // a background refresh while typing must not replace the input
  await ui.type("#commitForm input[name=summary]", "Update readme");
  await ui.page.evaluate(() => window.dispatchEvent(new Event("focus")));
  await ui.settle(700);
  expect(await ui.page.evaluate(() => document.activeElement?.name === "summary"), "the summary box keeps focus during a background refresh");
  await ui.page.keyboard.type(" and add feature");
  expect(await ui.page.$eval("#commitForm input[name=summary]", (i) => i.value) === "Update readme and add feature", "typed text survives a background refresh");
  expect(!(await ui.page.$eval("#commitButton", (b) => b.disabled)), "commit is enabled with a summary");
  // uncheck the deleted file -> partial commit
  await ui.click('.file-row[data-file="docs/old.md"] input[type=checkbox]');
  expect((await ui.text("#commitButton")).includes("Commit 2 files to main"), "button reflects a partial commit");
  await ui.click("#commitButton");
  await ui.waitText("1 changed file");
  expect(ui.git(repo, "log", "-1", "--format=%s|%an|%ae") === "Update readme and add feature|Work Dev|dev@work.example", "commit created with the identity");
  expect(ui.git(repo, "show", "--format=", "--name-only", "HEAD") === "README.md\nsrc/new-feature.js", "only the selected files were committed");
  expect(await ui.page.$eval("#commitForm input[name=summary]", (i) => i.value) === "", "summary is cleared after commit");
  expect((await ui.text("#repoToolbar")).includes("Push"), "push button present");
  await ui.settle(400);
  await ui.shot("c2-after-commit");

  // discard the remaining change
  await ui.clickText("Discard changes", "#diffPane button");
  await ui.page.waitForSelector("#confirmDialog[open]");
  await ui.click("#confirmAccept");
  await ui.waitText("No changes.");
  expect(fs.existsSync(path.join(repo, "docs/old.md")), "discard restored the deleted file");

  // push
  await ui.clickText("Push", "#repoToolbar button");
  await ui.waitText("Pushed to the remote");
  expect(ui.git(bare, "log", "-1", "--format=%s", "main") === "Update readme and add feature", "commit reached the remote");

  // history
  await ui.click('#repoTabs [data-tab="history"]');
  await ui.waitText("initial commit");
  await ui.page.waitForSelector("#commitDiff .diff");
  await ui.settle();
  expect((await ui.text("#commitDetail")).includes("Update readme and add feature"), "history shows commit details");
  expect((await ui.text("#commitDiff")).includes("line 2 changed"), "history shows the commit diff");
  await ui.shot("c3-history");

  // new branch
  await ui.click('#repoTabs [data-tab="changes"]');
  await ui.click('[data-action="repo-new-branch"]');
  await ui.page.waitForSelector("#branchDialog[open]");
  await ui.type("#branchForm input[name=name]", "bad name");
  await ui.clickText("Create branch", "#branchForm button[type=submit]");
  await ui.page.waitForFunction(() => !document.getElementById("branchError").hidden);
  expect((await ui.text("#branchError")).includes("not a valid branch name"), "invalid branch name is explained");
  await ui.type("#branchForm input[name=name]", "feature/checkout");
  await ui.clickText("Create branch", "#branchForm button[type=submit]");
  await ui.page.waitForFunction(() => !document.getElementById("branchDialog").open);
  await ui.settle(400);
  expect(ui.git(repo, "branch", "--show-current") === "feature/checkout", "new branch is checked out");
  expect((await ui.text("#repoToolbar")).includes("Publish branch"), "new branch offers Publish");
  await ui.clickText("Publish branch", "#repoToolbar button");
  await ui.waitText("Published the branch");
  // switch back with the picker
  await ui.page.select("#repoBranch", "main");
  await ui.waitText("Switched to main");
  expect(ui.git(repo, "branch", "--show-current") === "main", "branch picker switches branches");

  // someone else pushes; fetch shows it, pull applies it
  const other = path.join(ui.home, "other-clone");
  ui.git(ui.home, "-c", "core.hooksPath=/dev/null", "clone", "-q", bare, other);
  ui.write(path.join(other, "CHANGELOG.md"), "v1\n");
  ui.git(other, "add", "-A");
  ui.git(other, "-c", "user.name=Teammate", "-c", "user.email=mate@example.com", "commit", "-q", "-m", "Add changelog");
  ui.git(other, "-c", "core.hooksPath=/dev/null", "push", "-q", "origin", "main");
  await ui.clickText("Fetch", "#repoToolbar button");
  await ui.waitText("Fetched from the remote");
  await ui.settle(300);
  expect((await ui.text('[data-action="repo-pull"]')).includes("1"), "pull button shows one incoming commit");
  await ui.shot("c4-behind");
  await ui.clickText("Pull", "#repoToolbar button");
  await ui.waitText("Pulled the latest commits");
  expect(fs.existsSync(path.join(repo, "CHANGELOG.md")), "pull brought the new file");

  // conflict: both sides edit the same line
  ui.write(path.join(other, "README.md"), "# Storefront\n\nTEAMMATE EDIT\nline three\nline four\n");
  ui.git(other, "add", "-A");
  ui.git(other, "-c", "user.name=Teammate", "-c", "user.email=mate@example.com", "commit", "-q", "-m", "Teammate edit");
  ui.git(other, "-c", "core.hooksPath=/dev/null", "push", "-q", "origin", "main");
  ui.write(path.join(repo, "README.md"), "# Storefront\n\nMY EDIT\nline three\nline four\n");
  await ui.page.evaluate(() => window.dispatchEvent(new Event("focus")));
  await ui.waitText("1 changed file");
  await ui.type("#commitForm input[name=summary]", "My edit");
  await ui.click("#commitButton");
  await ui.waitText("No changes.");
  // push is rejected -> offered to pull -> conflict
  await ui.clickText("Push", "#repoToolbar button");
  await ui.page.waitForSelector("#confirmDialog[open]", { timeout: 20000 });
  expect((await ui.text("#confirmDialog")).includes("Pull before pushing"), "rejected push offers to pull");
  await ui.click("#confirmAccept");
  await ui.waitText("Merge in progress");
  await ui.settle(500);
  expect((await ui.text("#repoBanners")).includes("1 file has conflicts"), "conflict banner explains what to do");
  expect(await ui.page.$eval("#commitButton", (b) => b.disabled), "commit is disabled while conflicts remain");
  await ui.shot("c5-conflict");
  await ui.clickText("Cancel merge", "#repoBanners button");
  await ui.page.waitForSelector("#confirmDialog[open]");
  await ui.click("#confirmAccept");
  await ui.waitText("Merge cancelled");
  expect(ui.git(repo, "status", "--porcelain") === "", "cancelling the merge restores a clean tree");

  // identity drift: push offers fix-and-push... first resolve by taking theirs through git, then test drift
  ui.git(repo, "-c", "user.name=x", "-c", "user.email=x@x", "pull", "-q", "--no-rebase", "-X", "ours", "--no-edit");
  ui.git(repo, "config", "--local", "user.email", "drifted@example.com");
  await ui.page.evaluate(() => window.dispatchEvent(new Event("focus")));
  await ui.waitText("does not match its identity");
  await ui.clickText("Push", "#repoToolbar button");
  await ui.page.waitForSelector("#confirmDialog[open]", { timeout: 20000 });
  expect((await ui.text("#confirmDialog")).includes("Fix the repository before pushing"), "push with a mismatched identity offers fix and push");
  await ui.shot("c6-fix-and-push");
  await ui.click("#confirmAccept");
  await ui.waitText("Pushed to the remote");
  expect(ui.git(repo, "config", "--local", "user.email") === "dev@work.example", "fix-and-push repaired the identity");
  expect(ui.git(repo, "status", "-sb").includes("...origin/main") && !ui.git(repo, "status", "-sb").includes("ahead"), "everything is pushed");
}, { width: 1280, height: 820 });
