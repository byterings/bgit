// Repository workspace: changes, commit, discard, push, history, branches,
// remotes, fetch and pull, merge conflicts, and repairing identity drift.
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
  const head = (selector) => `#wsHead [data-action="${selector}"]`;
  const tab = (name) => ui.click(`#wsTabs [data-tab="${name}"]`);

  // working tree changes of every kind
  ui.write(path.join(repo, "README.md"), "# Storefront\n\nline 2 changed\nline three\nline four\n");
  ui.write(path.join(repo, "src/new-feature.js"), "export const feature = true;\n");
  fs.rmSync(path.join(repo, "docs/old.md"));

  await ui.open();
  await ui.nav("repositories");
  await ui.waitText("storefront");
  expect((await ui.text("#view")).includes("3 changes"), "list shows the number of changes");
  await ui.click(`[data-action="open-repo"][data-path="${repo}"]`);
  await ui.waitText("3 of 3 files selected");
  await ui.page.waitForSelector(".diff tr.add");
  await ui.settle();
  const body = await ui.text("#fileRows");
  expect(body.includes("README.md") && body.includes("new-feature.js") && body.includes("old.md"), "changed files are listed");
  expect((await ui.text("#diffPane")).includes("line 2 changed"), "diff of the first file is shown");
  expect((await ui.text("#wsHead")).includes("work"), "header shows the repository identity");
  const identity = await ui.text("#commitIdentity");
  expect(identity.includes("Work Dev") && identity.includes("dev@work.example"), "commit panel shows who the commit will be recorded as");
  expect((await ui.text("#commitTarget")).toLowerCase().includes("commit to main"), "commit panel names the branch");
  await ui.shot("c1-changes");

  // select another file
  await ui.click('.file-row[data-file="src/new-feature.js"] .file-name');
  await ui.page.waitForFunction(() => document.getElementById("diffPane").innerText.includes("export const feature"));
  expect((await ui.page.$$("#diffPane .tok-keyword")).length > 0, "selecting a file shows its diff with syntax colours");
  await ui.type("#fileFilter", "readme");
  expect((await ui.page.$$("#fileRows .file-row")).length === 1, "the file filter narrows the list");
  await ui.type("#fileFilter", "");

  // commit button rules
  expect(await ui.page.$eval("#commitButton", (b) => b.disabled), "commit is disabled without a summary");
  // a background refresh while typing must not replace the input
  await ui.type("#commitSummary", "Update readme");
  await ui.focusWindow();
  await ui.settle(700);
  expect(await ui.page.evaluate(() => document.activeElement?.id === "commitSummary"), "the summary box keeps focus during a background refresh");
  await ui.page.keyboard.type(" and add feature");
  expect(await ui.page.$eval("#commitSummary", (i) => i.value) === "Update readme and add feature", "typed text survives a background refresh");
  expect(!(await ui.page.$eval("#commitButton", (b) => b.disabled)), "commit is enabled with a summary");
  // uncheck the deleted file -> partial commit
  await ui.click('.file-row[data-file="docs/old.md"] input[type=checkbox]');
  expect((await ui.text("#commitButton")).includes("Commit 2 files"), "button reflects a partial commit");
  await ui.click("#commitButton");
  // The file left out of the commit stays, still unselected.
  await ui.waitText("0 of 1 file selected");
  expect((await ui.text("#commitHint")).includes("Select at least one file"), "commit panel explains that nothing is selected");
  expect(ui.git(repo, "log", "-1", "--format=%s|%an|%ae") === "Update readme and add feature|Work Dev|dev@work.example", "commit created with the identity");
  expect(ui.git(repo, "show", "--format=", "--name-only", "HEAD") === "README.md\nsrc/new-feature.js", "only the selected files were committed");
  expect(await ui.page.$eval("#commitSummary", (i) => i.value) === "", "summary is cleared after commit");
  expect((await ui.text("#syncHint")).includes("1 commit not pushed yet"), "commit panel says a commit is waiting to be pushed");
  expect(await ui.page.$eval(head("ws-push"), (b) => b.classList.contains("btn-primary") && b.innerText.includes("1")), "push button is highlighted with the count");
  await ui.shot("c2-after-commit");

  // discard the remaining change
  await ui.click('#diffPane [data-action="discard-file"]');
  await ui.waitDialog("confirmDialog");
  expect((await ui.text("#confirmDialog")).includes("cannot be undone"), "discard warns that it cannot be undone");
  await ui.shot("c2b-discard-confirm");
  await ui.submitDialog("confirmDialog");
  await ui.waitText("Working tree is clean");
  expect(fs.existsSync(path.join(repo, "docs/old.md")), "discard restored the deleted file");
  await ui.shot("c2c-clean");

  // push
  await ui.click(head("ws-push"));
  await ui.waitText("Pushed to the remote");
  expect(ui.git(bare, "log", "-1", "--format=%s", "main") === "Update readme and add feature", "commit reached the remote");
  await ui.waitText("Up to date with origin/main");

  // history
  await tab("history");
  await ui.waitText("initial commit");
  await ui.page.waitForSelector("#commitDiff .diff");
  await ui.settle();
  expect((await ui.text("#commitDetail")).includes("Update readme and add feature"), "history shows commit details");
  expect((await ui.text("#commitDiff")).includes("line 2 changed"), "history shows the commit diff");
  await ui.click('#commitDetail .file-row[data-file="src/new-feature.js"]');
  await ui.page.waitForFunction(() => document.getElementById("commitDiff").innerText.includes("export const feature"));
  expect(true, "history shows the diff of each file in the commit");
  await ui.shot("c3-history");

  // new branch
  await tab("branches");
  await ui.waitText("Local branches");
  await ui.click('[data-action="branch-new"]');
  await ui.waitDialog("branchDialog");
  await ui.type("#branchName", "bad name");
  await ui.submitDialog("branchDialog");
  expect((await ui.dialogError("branchDialog")).includes("not a valid branch name"), "invalid branch name is explained");
  await ui.type("#branchName", "feature/checkout");
  await ui.shot("c4-new-branch-dialog");
  await ui.submitDialog("branchDialog");
  await ui.waitDialogClosed("branchDialog");
  expect(ui.git(repo, "branch", "--show-current") === "feature/checkout", "new branch is checked out");
  expect((await ui.text(head("ws-push"))).includes("Publish branch"), "new branch offers Publish");
  await ui.click(head("ws-push"));
  await ui.waitText("Published the branch");
  // switch back with the branch menu in the header
  await ui.click(head("ws-branch-menu"));
  await ui.page.waitForSelector(".popover .menu");
  await ui.shot("c5-branch-menu");
  await ui.menuItem("main");
  await ui.waitText("Switched to main");
  expect(ui.git(repo, "branch", "--show-current") === "main", "branch menu switches branches");

  // delete: a merged branch goes at once, an unmerged one asks again
  ui.git(repo, "checkout", "-q", "-b", "lonely");
  ui.write(path.join(repo, "lonely.txt"), "only here\n");
  ui.git(repo, "add", "-A");
  ui.git(repo, "commit", "-q", "-m", "Lonely commit");
  ui.git(repo, "checkout", "-q", "main");
  await ui.focusWindow();
  await ui.page.waitForSelector('[data-action="branch-delete"][data-name="lonely"]');
  await ui.shot("c6-branches");
  await ui.click('[data-action="branch-delete"][data-name="feature/checkout"]');
  await ui.waitDialog("confirmDialog");
  await ui.submitDialog("confirmDialog");
  await ui.page.waitForFunction(() => !document.querySelector('[data-action="branch-delete"][data-name="feature/checkout"]'), { timeout: 15000 });
  expect(!ui.git(repo, "branch", "--list", "feature/checkout"), "merged branch is deleted");
  await ui.settle(400);
  await ui.click('[data-action="branch-delete"][data-name="lonely"]');
  await ui.waitDialog("confirmDialog");
  await ui.submitDialog("confirmDialog");
  await ui.page.waitForFunction(() => document.querySelector("#confirmDialog h2")?.innerText.includes("not on any other branch"), { timeout: 15000 });
  expect(ui.git(repo, "branch", "--list", "lonely") !== "", "a branch with unique commits is not deleted without asking again");
  await ui.settle(300);
  await ui.shot("c7-delete-unmerged-confirm");
  await ui.submitDialog("confirmDialog");
  await ui.page.waitForFunction(() => !document.querySelector('[data-action="branch-delete"][data-name="lonely"]'), { timeout: 15000 });
  expect(!ui.git(repo, "branch", "--list", "lonely"), "the branch is deleted after the second confirmation");

  // remotes
  await tab("remotes");
  await ui.waitText("Routed through the right identity");
  expect((await ui.text("#wsBody")).includes("git@github.com-work-gh:acme/storefront.git"), "remote address is shown");
  await ui.shot("c8-remotes");

  // someone else pushes; fetch shows it, pull applies it
  await tab("changes");
  const other = path.join(ui.home, "other-clone");
  ui.git(ui.home, "-c", "core.hooksPath=/dev/null", "clone", "-q", bare, other);
  ui.write(path.join(other, "CHANGELOG.md"), "v1\n");
  ui.git(other, "add", "-A");
  ui.git(other, "-c", "user.name=Teammate", "-c", "user.email=mate@example.com", "commit", "-q", "-m", "Add changelog");
  ui.git(other, "-c", "core.hooksPath=/dev/null", "push", "-q", "origin", "main");
  await ui.click(head("ws-fetch"));
  await ui.waitText("Fetched from the remote");
  await ui.settle(300);
  expect((await ui.text(head("ws-pull"))).includes("1"), "pull button shows one incoming commit");
  expect((await ui.text("#wsSync")).includes("1 to pull"), "header says one commit is waiting");
  await ui.shot("c9-behind");
  await ui.click(head("ws-pull"));
  await ui.waitText("Pulled the latest commits");
  expect(fs.existsSync(path.join(repo, "CHANGELOG.md")), "pull brought the new file");

  // conflict: both sides edit the same line
  ui.write(path.join(other, "README.md"), "# Storefront\n\nTEAMMATE EDIT\nline three\nline four\n");
  ui.git(other, "add", "-A");
  ui.git(other, "-c", "user.name=Teammate", "-c", "user.email=mate@example.com", "commit", "-q", "-m", "Teammate edit");
  ui.git(other, "-c", "core.hooksPath=/dev/null", "push", "-q", "origin", "main");
  ui.write(path.join(repo, "README.md"), "# Storefront\n\nMY EDIT\nline three\nline four\n");
  await ui.focusWindow();
  await ui.waitText("1 of 1 file selected");
  await ui.type("#commitSummary", "My edit");
  // Ctrl+Enter commits from the keyboard
  await ui.page.keyboard.down("Control");
  await ui.page.keyboard.press("Enter");
  await ui.page.keyboard.up("Control");
  await ui.waitText("Working tree is clean");
  expect(ui.git(repo, "log", "-1", "--format=%s") === "My edit", "Ctrl+Enter commits");
  // push is rejected -> offered to pull -> conflict
  await ui.click(head("ws-push"));
  await ui.waitDialog("confirmDialog");
  expect((await ui.text("#confirmDialog")).includes("Pull before pushing"), "rejected push offers to pull");
  await ui.shot("c10-pull-before-push");
  await ui.submitDialog("confirmDialog");
  await ui.waitText("Merge in progress");
  await ui.settle(500);
  expect((await ui.text("#wsBanners")).includes("1 file has conflicts"), "conflict banner explains what to do");
  expect(await ui.page.$eval("#commitButton", (b) => b.disabled), "commit is disabled while conflicts remain");
  expect((await ui.text("#commitHint")).includes("Resolve the conflicts first"), "commit panel says why committing is not possible");
  await ui.shot("c11-conflict");
  await ui.click('#wsBanners [data-action="ws-abort-merge"]');
  await ui.waitDialog("confirmDialog");
  await ui.submitDialog("confirmDialog");
  await ui.waitText("Merge cancelled");
  expect(ui.git(repo, "status", "--porcelain") === "", "cancelling the merge restores a clean tree");

  // identity drift: push offers fix-and-push. First settle the divergence through git.
  ui.git(repo, "-c", "user.name=x", "-c", "user.email=x@x", "pull", "-q", "--no-rebase", "-X", "ours", "--no-edit");
  ui.git(repo, "config", "--local", "user.email", "drifted@example.com");
  await ui.focusWindow();
  await ui.waitText("does not match its identity");
  expect((await ui.page.$$('#wsTabs [data-tab="identity"] .dot')).length === 1, "the Identity tab carries a marker");
  await ui.click(head("ws-push"));
  await ui.waitDialog("confirmDialog");
  expect((await ui.text("#confirmDialog")).includes("Fix the repository before pushing"), "push with a mismatched identity offers fix and push");
  await ui.shot("c12-fix-and-push");
  await ui.submitDialog("confirmDialog");
  await ui.waitText("Pushed to the remote");
  expect(ui.git(repo, "config", "--local", "user.email") === "dev@work.example", "fix-and-push repaired the identity");
  expect(ui.git(repo, "status", "-sb").includes("...origin/main") && !ui.git(repo, "status", "-sb").includes("ahead"), "everything is pushed");

  // repository identity
  await tab("identity");
  await ui.waitText("Identity for this repository");
  const identityTab = await ui.text("#wsBody");
  expect(identityTab.includes("This repository is bound to it") && identityTab.includes("Switching the default identity does not affect it"), "identity tab explains where the identity comes from");
  await ui.shot("c13-repo-identity");
  await ui.click('#wsBody [data-action="identity-change"]');
  await ui.waitDialog("addRepoDialog");
  await ui.page.select("#addRepoAlias", "home");
  expect((await ui.text("#addRepoDialog [data-future]")).includes("Home Dev <dev@home.example>"), "changing identity previews who future commits are recorded as");
  await ui.shot("c14-change-identity-dialog");
  await ui.cancelDialog("addRepoDialog");
  expect(ui.git(repo, "config", "--local", "user.email") === "dev@work.example", "the identity only changes when confirmed");
}, { width: 1366, height: 699 });
