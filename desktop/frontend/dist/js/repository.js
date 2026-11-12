// Repository view: everyday Git work for one repository. Changed files with
// diffs, commit, discard, branches, history, and fetch/pull/push.
//
// The page is drawn in regions (toolbar, banners, tabs, panes) that are
// updated on their own, so typing a commit message or scrolling a diff is not
// interrupted when the status refreshes.
//
// Uses the shared helpers defined in index.html.
(() => {
  const HISTORY_PAGE = 50;

  const kindLabels = {
    modified: { letter: "M", text: "Modified" },
    added: { letter: "A", text: "Added" },
    untracked: { letter: "A", text: "New file" },
    deleted: { letter: "D", text: "Deleted" },
    renamed: { letter: "R", text: "Renamed" },
    copied: { letter: "C", text: "Copied" },
    conflicted: { letter: "!", text: "Conflict" },
    typechange: { letter: "T", text: "Type changed" },
  };

  const view = freshView();

  function freshView() {
    return {
      path: "",
      repository: null,
      status: null,
      error: "",
      tab: "changes",
      busy: "",
      // Changes tab
      checked: new Set(),
      knownFiles: new Set(),
      selectedFile: "",
      diff: null,
      diffLoading: false,
      summary: "",
      description: "",
      mergeSummaryOffered: false,
      // History tab
      commits: null,
      historyDone: false,
      selectedCommit: "",
      commitFiles: null,
      commitFile: "",
      commitDiff: null,
      // Branch picker
      branches: null,
    };
  }

  const el = (id) => document.getElementById(id);
  const files = () => view.status?.files || [];

  // ---------- Opening and refreshing ----------

  window.openRepository = (path) => {
    Object.assign(view, freshView(), { path });
    transitionToPage("repo");
    refresh();
  };

  function applyState(state) {
    if (!state) {
      return;
    }
    view.repository = state.repository;
    view.status = state.status;
    view.error = "";

    const paths = new Set(files().map((file) => file.path));
    // New changes start selected; files that are gone drop out of the selection.
    paths.forEach((path) => {
      if (!view.knownFiles.has(path)) {
        view.checked.add(path);
      }
    });
    view.checked.forEach((path) => {
      if (!paths.has(path)) {
        view.checked.delete(path);
      }
    });
    view.knownFiles = paths;

    if (!paths.has(view.selectedFile)) {
      view.selectedFile = files()[0]?.path || "";
      view.diff = null;
    }

    // A merge needs a commit to finish; suggest its summary once.
    if (view.status.mergeInProgress && !view.mergeSummaryOffered) {
      view.mergeSummaryOffered = true;
      if (!view.summary.trim()) {
        view.summary = "Merge remote changes";
      }
    } else if (!view.status.mergeInProgress) {
      view.mergeSummaryOffered = false;
    }
  }

  // Reloads branch and file state. Quiet refreshes (window focus) do not
  // show errors as notifications.
  async function refresh({ quiet = false } = {}) {
    if (!view.path) {
      return;
    }
    const path = view.path;
    try {
      const state = await api().GetRepositoryState(path);
      if (path !== view.path) {
        return;
      }
      const previousFile = view.selectedFile;
      applyState(state);
      drawAll();
      if (view.selectedFile && (view.selectedFile !== previousFile || !view.diff || quiet)) {
        loadDiff();
      }
    } catch (error) {
      if (path !== view.path) {
        return;
      }
      view.error = String(error);
      if (!quiet) {
        drawAll();
      }
    }
  }

  window.addEventListener("focus", () => {
    if (currentPage === "repo" && !view.busy) {
      refresh({ quiet: true });
    }
  });

  // ---------- Page skeleton ----------

  function page() {
    return `
      <div id="repoView" class="repo-view">
        <div id="repoToolbar" class="repo-toolbar"></div>
        <div id="repoBanners"></div>
        <div id="repoTabs" class="repo-tabs" role="tablist"></div>
        <div id="repoBody" class="repo-body"></div>
      </div>
    `;
  }

  function bind() {
    if (!view.path) {
      transitionToPage("repositories", false);
      return;
    }
    el("repoToolbar").addEventListener("click", onToolbarClick);
    el("repoToolbar").addEventListener("change", onToolbarChange);
    el("repoBanners").addEventListener("click", onBannerClick);
    el("repoTabs").addEventListener("click", onTabClick);
    el("repoBody").addEventListener("click", onBodyClick);
    el("repoBody").addEventListener("change", onBodyChange);
    el("repoBody").addEventListener("input", onBodyInput);
    el("repoBody").addEventListener("submit", onBodySubmit);
    drawAll();
  }

  function drawAll() {
    if (currentPage !== "repo" || !el("repoView")) {
      return;
    }
    drawToolbar();
    drawBanners();
    drawTabs();
    drawBody();
  }

  // ---------- Toolbar ----------

  function drawToolbar() {
    const repo = view.repository;
    const status = view.status;
    const identity = (currentStatus?.identities || []).find((item) => item.alias === repo?.alias);
    const name = repo?.name || view.path.split(/[\\/]/).pop();
    const busy = Boolean(view.busy);
    const hasRemote = Boolean(repo?.remoteUrl);
    const canSync = hasRemote && status && !status.detached;

    el("repoToolbar").innerHTML = `
      <div class="repo-heading">
        <h1 class="truncate" title="${escapeHtml(view.path)}">${escapeHtml(name)}</h1>
        <div class="repo-heading-meta">
          ${identity ? avatarBlock(identity) : ""}
          <span class="truncate">${repo?.alias ? `${escapeHtml(repo.alias)} · ${escapeHtml(repo.identityEmail || "")}` : "No identity"}</span>
        </div>
      </div>
      <div class="repo-tools">
        <label class="branch-picker" title="Current branch">
          <span class="icon" aria-hidden="true">${iconSvg("branch")}</span>
          <span class="sr-only">Branch</span>
          <select id="repoBranch" ${busy || !status ? "disabled" : ""}>${branchOptions()}</select>
        </label>
        ${iconButton({ action: "repo-new-branch", icon: "plus", label: "New branch", disabled: busy || !status || status.unborn })}
        <span class="repo-tools-gap"></span>
        <button type="button" data-action="repo-fetch" ${busy || !hasRemote ? "disabled" : ""} title="Check the remote for new commits">
          ${syncLabel("fetch", "refresh", "Fetch", "Fetching")}
        </button>
        <button type="button" data-action="repo-pull" ${busy || !canSync || !status.hasUpstream ? "disabled" : ""} title="Download and apply new commits">
          ${syncLabel("pull", "download", "Pull", "Pulling", status?.behind)}
        </button>
        <button type="button" class="${status?.ahead > 0 || (canSync && !status.hasUpstream && !status.unborn) ? "primary" : ""}" data-action="repo-push" ${busy || !canSync || status.unborn ? "disabled" : ""} title="Upload your commits">
          ${syncLabel("push", "upload", status && canSync && !status.hasUpstream ? "Publish branch" : "Push", "Pushing", status?.ahead)}
        </button>
      </div>
    `;
  }

  function syncLabel(operation, icon, idle, working, count = 0) {
    if (view.busy === operation) {
      return `<span class="spinner" aria-hidden="true"></span>${escapeHtml(working)}`;
    }
    return `<span class="icon" aria-hidden="true">${iconSvg(icon)}</span>${escapeHtml(idle)}${count > 0 ? `<span class="count-pill">${count}</span>` : ""}`;
  }

  function branchOptions() {
    const status = view.status;
    if (!status) {
      return `<option>Loading</option>`;
    }
    const current = status.detached ? `Detached at ${status.headOid}` : status.branch;
    if (!view.branches) {
      return `<option selected>${escapeHtml(current)}</option>`;
    }
    const local = view.branches.filter((branch) => !branch.remote);
    const localNames = new Set(local.map((branch) => branch.name));
    // A remote branch that already has a local copy is not listed twice.
    const remote = view.branches.filter((branch) => branch.remote && !localNames.has(branch.name.split("/").slice(1).join("/")));
    const option = (branch) =>
      `<option value="${escapeHtml(branch.name)}" ${branch.current ? "selected" : ""}>${escapeHtml(branch.name)}</option>`;
    return `
      ${status.detached || status.unborn ? `<option selected disabled>${escapeHtml(current)}</option>` : ""}
      ${local.length ? `<optgroup label="Branches">${local.map(option).join("")}</optgroup>` : ""}
      ${remote.length ? `<optgroup label="On the remote">${remote.map(option).join("")}</optgroup>` : ""}
    `;
  }

  async function loadBranches() {
    try {
      view.branches = (await api().ListBranches(view.path)) || [];
      drawToolbar();
    } catch (error) {
      // The picker keeps showing the current branch.
    }
  }

  function onToolbarClick(event) {
    const button = event.target.closest("[data-action]");
    if (!button || button.disabled) {
      return;
    }
    switch (button.dataset.action) {
      case "repo-fetch":
        sync("fetch");
        break;
      case "repo-pull":
        sync("pull");
        break;
      case "repo-push":
        sync("push");
        break;
      case "repo-new-branch":
        newBranch();
        break;
      default:
        break;
    }
  }

  function onToolbarChange(event) {
    if (event.target.id === "repoBranch") {
      switchBranch(event.target.value);
    }
  }

  // ---------- Banners ----------

  function drawBanners() {
    const repo = view.repository;
    const status = view.status;
    const banners = [];

    if (view.error) {
      banners.push(`<div class="note error" role="alert">${escapeHtml(view.error)}</div>`);
    }
    if (repo && repo.health !== "ok" && (repo.problems || []).length > 0) {
      banners.push(`
        <div class="note ${repo.canFix ? "warn" : "error"} repo-banner">
          <span><strong>${repo.canFix ? "This repository does not match its identity" : "bgit cannot manage this repository"}</strong>${(repo.problems || []).map(escapeHtml).join(" ")}</span>
          ${repo.canFix ? `<button type="button" data-action="repo-fix"><span class="icon" aria-hidden="true">${iconSvg("wrench")}</span>Fix</button>` : ""}
        </div>
      `);
    }
    if (status?.mergeInProgress) {
      banners.push(`
        <div class="note warn repo-banner">
          <span><strong>Merge in progress</strong>${
            status.unresolvedConflicts > 0
              ? `${status.unresolvedConflicts} file${status.unresolvedConflicts === 1 ? " has" : "s have"} conflicts. Open ${status.unresolvedConflicts === 1 ? "it" : "them"} in your editor, fix the sections between <<<<<<< and >>>>>>>, and save. Then commit here.`
              : "All conflicts are resolved. Enter a summary and commit to finish the merge."
          }</span>
          <button type="button" data-action="repo-abort-merge" ${view.busy ? "disabled" : ""}>Cancel merge</button>
        </div>
      `);
    }
    if (status?.rebaseInProgress) {
      banners.push(`<div class="note warn"><strong>Rebase in progress</strong>Finish or cancel the rebase in a terminal before making changes here.</div>`);
    }
    if (status?.detached) {
      banners.push(`<div class="note"><strong>No branch is checked out</strong>You are looking at commit ${escapeHtml(status.headOid)}. Choose a branch to commit and sync.</div>`);
    }
    el("repoBanners").innerHTML = banners.join("");
  }

  function onBannerClick(event) {
    const button = event.target.closest("[data-action]");
    if (!button || button.disabled) {
      return;
    }
    if (button.dataset.action === "repo-fix") {
      fixRepository();
    }
    if (button.dataset.action === "repo-abort-merge") {
      abortMerge();
    }
  }

  // ---------- Tabs ----------

  function drawTabs() {
    const count = files().length;
    el("repoTabs").innerHTML = `
      <button type="button" role="tab" data-tab="changes" aria-selected="${view.tab === "changes"}" class="${view.tab === "changes" ? "active" : ""}">Changes${count > 0 ? `<span class="count-pill">${count}${view.status?.truncated ? "+" : ""}</span>` : ""}</button>
      <button type="button" role="tab" data-tab="history" aria-selected="${view.tab === "history"}" class="${view.tab === "history" ? "active" : ""}">History</button>
    `;
  }

  function onTabClick(event) {
    const button = event.target.closest("[data-tab]");
    if (!button || button.dataset.tab === view.tab) {
      return;
    }
    view.tab = button.dataset.tab;
    drawTabs();
    drawBody();
    if (view.tab === "history" && !view.commits) {
      loadHistory();
    }
  }

  function drawBody() {
    const body = el("repoBody");
    if (!view.status) {
      body.innerHTML = view.error ? "" : `<p class="muted repo-loading"><span class="spinner" aria-hidden="true"></span> Loading repository</p>`;
      return;
    }
    if (view.tab === "history") {
      body.innerHTML = `
        <div class="repo-split">
          <div id="historyList" class="repo-side"></div>
          <div id="commitDetail" class="repo-main-pane"></div>
        </div>
      `;
      drawHistory();
      drawCommitDetail();
      return;
    }
    body.innerHTML = `
      <div class="repo-split">
        <div class="repo-side">
          <div id="fileList" class="file-list"></div>
          <form id="commitForm" class="commit-box"></form>
        </div>
        <div id="diffPane" class="repo-main-pane"></div>
      </div>
    `;
    drawFiles();
    drawCommitBox();
    drawDiff("diffPane", view.diff, view.diffLoading, view.selectedFile, true);
    if (view.selectedFile && !view.diff && !view.diffLoading) {
      loadDiff();
    }
    if (!view.branches) {
      loadBranches();
    }
  }

  // ---------- Changes: file list ----------

  function fileRow(file, { selectable, selected, checked }) {
    const kind = kindLabels[file.kind] || kindLabels.modified;
    const title = file.origPath ? `${file.origPath} → ${file.path}` : file.path;
    const slash = file.path.lastIndexOf("/");
    const folder = slash >= 0 ? file.path.slice(0, slash + 1) : "";
    const name = slash >= 0 ? file.path.slice(slash + 1) : file.path;
    return `
      <div class="file-row ${selected ? "selected" : ""}" data-file="${escapeHtml(file.path)}" title="${escapeHtml(title)}">
        ${selectable ? `<input type="checkbox" data-check="${escapeHtml(file.path)}" ${checked ? "checked" : ""} aria-label="Include ${escapeHtml(file.path)} in the commit" />` : ""}
        <span class="file-name"><span class="muted">${escapeHtml(folder)}</span>${escapeHtml(name)}</span>
        <span class="file-kind ${escapeHtml(file.kind)}" title="${escapeHtml(kind.text)}">${kind.letter}<span class="sr-only">${escapeHtml(kind.text)}</span></span>
      </div>
    `;
  }

  function drawFiles() {
    const target = el("fileList");
    if (!target) {
      return;
    }
    const list = files();
    if (list.length === 0) {
      target.innerHTML = `<div class="pane-empty"><p>No changes.</p><p class="muted">Edit files in your editor. Changes appear here.</p></div>`;
      return;
    }
    const allChecked = list.every((file) => view.checked.has(file.path));
    target.innerHTML = `
      <div class="file-list-head">
        <label class="check">
          <input type="checkbox" id="checkAll" ${allChecked ? "checked" : ""} />
          ${list.length} changed file${list.length === 1 ? "" : "s"}${view.status.truncated ? " (more not shown)" : ""}
        </label>
        <button type="button" class="link" data-action="discard-all">Discard all</button>
      </div>
      <div class="file-rows">
        ${list.map((file) => fileRow(file, { selectable: true, selected: file.path === view.selectedFile, checked: view.checked.has(file.path) })).join("")}
      </div>
    `;
    const checkAll = el("checkAll");
    checkAll.indeterminate = !allChecked && view.checked.size > 0;
  }

  // ---------- Changes: commit box ----------

  function commitTarget() {
    const status = view.status;
    if (!status) {
      return "";
    }
    return status.detached ? `commit ${status.headOid}` : status.branch;
  }

  // Explains why committing is not possible right now, or "" when it is.
  function commitProblem() {
    const status = view.status;
    if (!view.repository?.alias) return "No identity applies to this repository.";
    if (status.rebaseInProgress) return "A rebase is in progress.";
    if (status.unresolvedConflicts > 0) return "Resolve the conflicts first.";
    if (files().length > 0 && view.checked.size === 0 && !status.mergeInProgress) return "Select at least one file.";
    return "";
  }

  function canCommit() {
    const status = view.status;
    if (!status || view.busy || !view.repository?.alias || status.rebaseInProgress || status.unresolvedConflicts > 0) {
      return false;
    }
    if (!view.summary.trim()) {
      return false;
    }
    return status.mergeInProgress || view.checked.size > 0;
  }

  function commitButtonLabel() {
    if (view.busy === "commit") {
      return `<span class="spinner" aria-hidden="true"></span>Committing`;
    }
    const count = view.checked.size;
    const partial = count > 0 && count < files().length && !view.status.mergeInProgress;
    return escapeHtml(`Commit ${partial ? `${count} file${count === 1 ? "" : "s"} ` : ""}to ${commitTarget()}`);
  }

  function drawCommitBox() {
    const form = el("commitForm");
    if (!form) {
      return;
    }
    const repo = view.repository;
    form.innerHTML = `
      <input name="summary" placeholder="Summary (required)" maxlength="200" autocomplete="off" value="${escapeHtml(view.summary)}" aria-label="Commit summary" />
      <textarea name="description" rows="3" placeholder="Description (optional)" aria-label="Commit description">${escapeHtml(view.description)}</textarea>
      <button type="submit" class="primary" id="commitButton" ${canCommit() ? "" : "disabled"}>${commitButtonLabel()}</button>
      <span class="commit-hint muted" id="commitHint">${escapeHtml(commitProblem() || (repo?.alias ? `Commits as ${repo.identityName} <${repo.identityEmail}>` : ""))}</span>
    `;
  }

  // Updates the commit button and hint without touching the inputs.
  function updateCommitControls() {
    const button = el("commitButton");
    if (!button) {
      return;
    }
    button.disabled = !canCommit();
    button.innerHTML = commitButtonLabel();
    const repo = view.repository;
    el("commitHint").textContent = commitProblem() || (repo?.alias ? `Commits as ${repo.identityName} <${repo.identityEmail}>` : "");
  }

  // ---------- Diff ----------

  async function loadDiff() {
    const file = view.selectedFile;
    if (!file) {
      return;
    }
    view.diffLoading = !view.diff || view.diff.path !== file;
    if (view.diffLoading) {
      drawDiff("diffPane", null, true, file, true);
    }
    try {
      const diff = await api().GetFileDiff({ path: view.path, file, commit: "" });
      if (file !== view.selectedFile) {
        return;
      }
      view.diff = diff;
    } catch (error) {
      if (file !== view.selectedFile) {
        return;
      }
      view.diff = { path: file, lines: [{ kind: "note", text: String(error) }], additions: 0, deletions: 0 };
    }
    view.diffLoading = false;
    if (view.tab === "changes") {
      drawDiff("diffPane", view.diff, false, file, true);
    }
  }

  function drawDiff(targetId, diff, loading, file, discardable) {
    const target = el(targetId);
    if (!target) {
      return;
    }
    if (!file) {
      target.innerHTML = `<div class="pane-empty"><p class="muted">${view.tab === "changes" ? "Select a changed file to see what changed." : "Select a file to see what changed."}</p></div>`;
      return;
    }
    // Keep the reading position when the same file is redrawn.
    const previousScroll = target.dataset.file === file ? target.querySelector(".diff-scroll")?.scrollTop || 0 : 0;
    target.dataset.file = file;

    const lines = diff?.lines || [];
    let content;
    if (loading) {
      content = `<div class="pane-empty"><p class="muted"><span class="spinner" aria-hidden="true"></span> Loading changes</p></div>`;
    } else if (lines.length === 0) {
      content = `<div class="pane-empty"><p class="muted">No text changes to show for this file.</p></div>`;
    } else {
      content = `
        <div class="diff-scroll">
          <table class="diff">
            <tbody>
              ${lines.map(diffRow).join("")}
            </tbody>
          </table>
          ${diff.truncated ? `<p class="muted diff-note">This change is long. Only the first part is shown.</p>` : ""}
        </div>
      `;
    }
    target.innerHTML = `
      <div class="diff-head">
        <strong class="truncate" title="${escapeHtml(file)}">${escapeHtml(file)}</strong>
        ${diff && !loading && !diff.binary ? `<span class="diff-stat"><span class="add">+${diff.additions || 0}</span> <span class="del">-${diff.deletions || 0}</span></span>` : ""}
        ${discardable ? `<button type="button" data-action="discard-file" data-file="${escapeHtml(file)}" ${view.busy ? "disabled" : ""}>Discard changes</button>` : ""}
      </div>
      ${content}
    `;
    const scroller = target.querySelector(".diff-scroll");
    if (scroller && previousScroll) {
      scroller.scrollTop = previousScroll;
    }
  }

  function diffRow(line) {
    if (line.kind === "hunk") {
      return `<tr class="hunk"><td class="num"></td><td class="num"></td><td class="code">${escapeHtml(line.text)}</td></tr>`;
    }
    if (line.kind === "note") {
      return `<tr class="note-row"><td class="num"></td><td class="num"></td><td class="code">${escapeHtml(line.text)}</td></tr>`;
    }
    const marker = line.kind === "add" ? "+" : line.kind === "del" ? "-" : " ";
    return `<tr class="${line.kind}"><td class="num">${line.oldLine || ""}</td><td class="num">${line.newLine || ""}</td><td class="code"><span class="marker" aria-hidden="true">${marker}</span>${escapeHtml(line.text)}</td></tr>`;
  }

  // ---------- History ----------

  async function loadHistory(more = false) {
    try {
      const skip = more ? view.commits.length : 0;
      const batch = (await api().GetHistory({ path: view.path, limit: HISTORY_PAGE, skip })) || [];
      view.commits = more ? view.commits.concat(batch) : batch;
      view.historyDone = batch.length < HISTORY_PAGE;
      if (!view.selectedCommit && view.commits.length > 0) {
        selectCommit(view.commits[0].oid);
      }
    } catch (error) {
      view.commits = view.commits || [];
      notify({ title: "Could not load history", message: String(error), tone: "error" });
    }
    drawHistory();
  }

  function relativeTime(seconds) {
    const diff = Date.now() / 1000 - seconds;
    if (diff < 60) return "just now";
    if (diff < 3600) return `${Math.floor(diff / 60)} min ago`;
    if (diff < 86400) return `${Math.floor(diff / 3600)} h ago`;
    if (diff < 86400 * 30) return `${Math.floor(diff / 86400)} d ago`;
    return new Date(seconds * 1000).toLocaleDateString();
  }

  function drawHistory() {
    const target = el("historyList");
    if (!target) {
      return;
    }
    if (!view.commits) {
      target.innerHTML = `<div class="pane-empty"><p class="muted"><span class="spinner" aria-hidden="true"></span> Loading history</p></div>`;
      return;
    }
    if (view.commits.length === 0) {
      target.innerHTML = `<div class="pane-empty"><p>No commits yet.</p></div>`;
      return;
    }
    target.innerHTML = `
      <div class="commit-rows">
        ${view.commits
          .map(
            (commit) => `
              <button type="button" class="commit-row ${commit.oid === view.selectedCommit ? "selected" : ""}" data-commit="${escapeHtml(commit.oid)}">
                <strong class="truncate">${escapeHtml(commit.summary)}</strong>
                <span class="muted truncate">${escapeHtml(commit.authorName)} · ${escapeHtml(relativeTime(commit.time))}</span>
                ${commit.unpushed ? `<span class="commit-flag" title="Not pushed yet"><span class="icon" aria-hidden="true">${iconSvg("upload")}</span><span class="sr-only">Not pushed yet</span></span>` : ""}
              </button>`,
          )
          .join("")}
      </div>
      ${view.historyDone ? "" : `<button type="button" class="history-more" data-action="history-more">Load older commits</button>`}
    `;
  }

  async function selectCommit(oid) {
    view.selectedCommit = oid;
    view.commitFiles = null;
    view.commitFile = "";
    view.commitDiff = null;
    drawHistory();
    drawCommitDetail();
    try {
      const list = (await api().GetCommitFiles({ path: view.path, commit: oid })) || [];
      if (oid !== view.selectedCommit) {
        return;
      }
      view.commitFiles = list;
      if (list.length > 0) {
        selectCommitFile(list[0].path);
        return;
      }
    } catch (error) {
      if (oid !== view.selectedCommit) {
        return;
      }
      view.commitFiles = [];
      notify({ title: "Could not load commit", message: String(error), tone: "error" });
    }
    drawCommitDetail();
  }

  async function selectCommitFile(file) {
    const oid = view.selectedCommit;
    view.commitFile = file;
    view.commitDiff = null;
    drawCommitDetail();
    try {
      const diff = await api().GetFileDiff({ path: view.path, file, commit: oid });
      if (oid !== view.selectedCommit || file !== view.commitFile) {
        return;
      }
      view.commitDiff = diff;
    } catch (error) {
      view.commitDiff = { path: file, lines: [{ kind: "note", text: String(error) }] };
    }
    drawDiff("commitDiff", view.commitDiff, false, file, false);
  }

  function drawCommitDetail() {
    const target = el("commitDetail");
    if (!target) {
      return;
    }
    const commit = (view.commits || []).find((item) => item.oid === view.selectedCommit);
    if (!commit) {
      target.innerHTML = `<div class="pane-empty"><p class="muted">Select a commit to see what it changed.</p></div>`;
      return;
    }
    const list = view.commitFiles;
    target.innerHTML = `
      <div class="commit-head">
        <strong>${escapeHtml(commit.summary)}</strong>
        ${commit.body ? `<p class="commit-body">${escapeHtml(commit.body)}</p>` : ""}
        <span class="muted">${escapeHtml(commit.authorName)} &lt;${escapeHtml(commit.authorEmail)}&gt; · ${escapeHtml(new Date(commit.time * 1000).toLocaleString())} · ${escapeHtml(commit.shortOid)}${commit.unpushed ? " · not pushed yet" : ""}</span>
      </div>
      <div class="commit-split">
        <div class="file-rows commit-files">
          ${
            list
              ? list.length > 0
                ? list.map((file) => fileRow(file, { selectable: false, selected: file.path === view.commitFile })).join("")
                : `<p class="muted pane-pad">This commit changed no files.</p>`
              : `<p class="muted pane-pad"><span class="spinner" aria-hidden="true"></span> Loading files</p>`
          }
        </div>
        <div id="commitDiff" class="commit-diff"></div>
      </div>
    `;
    drawDiff("commitDiff", view.commitDiff, Boolean(view.commitFile) && !view.commitDiff, view.commitFile, false);
  }

  // ---------- Body events ----------

  function onBodyClick(event) {
    const action = event.target.closest("[data-action]");
    if (action && !action.disabled) {
      switch (action.dataset.action) {
        case "discard-all":
          discard(files().map((file) => file.path));
          return;
        case "discard-file":
          discard([action.dataset.file]);
          return;
        case "history-more":
          loadHistory(true);
          return;
        default:
          break;
      }
    }

    const commitRow = event.target.closest("[data-commit]");
    if (commitRow) {
      if (commitRow.dataset.commit !== view.selectedCommit) {
        selectCommit(commitRow.dataset.commit);
      }
      return;
    }

    // Clicking the checkbox only changes the selection for the commit.
    if (event.target.closest("[data-check]") || event.target.id === "checkAll") {
      return;
    }
    const row = event.target.closest("[data-file]");
    if (!row) {
      return;
    }
    const file = row.dataset.file;
    if (view.tab === "history") {
      if (file !== view.commitFile) {
        selectCommitFile(file);
      }
      return;
    }
    if (file !== view.selectedFile) {
      view.selectedFile = file;
      view.diff = null;
      drawFiles();
      loadDiff();
    }
  }

  function onBodyChange(event) {
    if (event.target.id === "checkAll") {
      view.checked = event.target.checked ? new Set(files().map((file) => file.path)) : new Set();
      drawFiles();
      updateCommitControls();
      return;
    }
    const file = event.target.dataset?.check;
    if (file) {
      if (event.target.checked) {
        view.checked.add(file);
      } else {
        view.checked.delete(file);
      }
      const checkAll = el("checkAll");
      const total = files().length;
      checkAll.checked = view.checked.size === total;
      checkAll.indeterminate = view.checked.size > 0 && view.checked.size < total;
      updateCommitControls();
    }
  }

  function onBodyInput(event) {
    if (event.target.name === "summary") {
      view.summary = event.target.value;
      updateCommitControls();
    }
    if (event.target.name === "description") {
      view.description = event.target.value;
    }
  }

  function onBodySubmit(event) {
    if (event.target.id === "commitForm") {
      event.preventDefault();
      commit();
    }
  }

  // ---------- Actions ----------

  // Runs one Git action with the toolbar locked. Returns the result object,
  // or null when the backend call itself failed.
  async function act(busy, call) {
    view.busy = busy;
    drawToolbar();
    drawBanners();
    updateCommitControls();
    try {
      const result = await call();
      if (result?.state) {
        applyState(result.state);
      }
      return result;
    } catch (error) {
      notify({ title: "Action failed", message: String(error), tone: "error" });
      return null;
    } finally {
      view.busy = "";
      drawAll();
      if (view.tab === "changes" && view.selectedFile) {
        loadDiff();
      }
    }
  }

  function succeed(result, activityContext) {
    addActivity(result.message, activityContext, "success");
    notify({ title: "Done", message: result.message, tone: "success" });
    if (typeof invalidateRepositories === "function") {
      invalidateRepositories();
    }
  }

  function fail(title, result) {
    addActivity(title, result.message, "error");
    notify({ title, message: result.message, tone: result.kind === "conflict" ? "warn" : "error", timeout: 9000 });
  }

  // Commits and branch switches change what history shows.
  function resetHistory() {
    view.commits = null;
    view.selectedCommit = "";
    view.commitFiles = null;
    view.commitFile = "";
    view.commitDiff = null;
    if (view.tab === "history") {
      loadHistory();
    }
  }

  async function commit() {
    if (!canCommit()) {
      return;
    }
    const all = view.checked.size === files().length;
    const result = await act("commit", () =>
      api().CommitChanges({
        path: view.path,
        summary: view.summary.trim(),
        description: view.description.trim(),
        // An empty list means "everything", which also covers files that
        // changed after the list was drawn.
        files: all ? [] : [...view.checked],
      }),
    );
    if (!result) {
      return;
    }
    if (!result.ok) {
      fail("Commit failed", result);
      return;
    }
    view.summary = "";
    view.description = "";
    view.diff = null;
    drawCommitBox();
    succeed(result, view.repository?.name || "");
    resetHistory();
  }

  async function discard(paths) {
    if (paths.length === 0) {
      return;
    }
    const one = paths.length === 1;
    const confirmed = await confirmAction({
      title: one ? "Discard changes" : "Discard all changes",
      message: one ? `Discard your changes to ${paths[0]}?` : `Discard your changes to all ${paths.length} files?`,
      warning: "This cannot be undone. Changed files go back to the last commit and new files are deleted.",
      acceptLabel: "Discard",
      danger: true,
    });
    if (!confirmed) {
      return;
    }
    const result = await act("discard", () => api().DiscardChanges({ path: view.path, files: paths }));
    if (!result) {
      return;
    }
    if (!result.ok) {
      fail("Discard failed", result);
      return;
    }
    view.diff = null;
    succeed(result, view.repository?.name || "");
  }

  async function fixRepository() {
    view.busy = "fix";
    drawToolbar();
    try {
      const result = await api().FixRepository(view.path);
      notify({ title: "Done", message: result.message, tone: "success" });
      addActivity(result.message, view.repository?.name || "", "success");
      if (typeof invalidateRepositories === "function") {
        invalidateRepositories();
      }
      return true;
    } catch (error) {
      notify({ title: "Fix failed", message: String(error), tone: "error" });
      return false;
    } finally {
      view.busy = "";
      await refresh();
    }
  }

  async function abortMerge() {
    const confirmed = await confirmAction({
      title: "Cancel merge",
      message: "Cancel the merge and go back to how the branch was before pulling?",
      warning: "Any conflict resolutions you already made are lost.",
      acceptLabel: "Cancel merge",
      cancelLabel: "Keep merging",
      danger: true,
    });
    if (!confirmed) {
      return;
    }
    const result = await act("abort", () => api().AbortMerge(view.path));
    if (result?.ok) {
      succeed(result, view.repository?.name || "");
    } else if (result) {
      fail("Could not cancel the merge", result);
    }
  }

  const syncCalls = {
    fetch: () => api().FetchRepository(view.path),
    pull: () => api().PullRepository(view.path),
    push: () => api().PushRepository(view.path),
  };
  const syncFailureTitles = { fetch: "Fetch failed", pull: "Pull failed", push: "Push failed" };

  async function sync(operation) {
    const result = await act(operation, syncCalls[operation]);
    if (!result) {
      return;
    }
    if (result.ok) {
      succeed(result, view.repository?.name || "");
      view.branches = null;
      loadBranches();
      if (operation !== "fetch") {
        resetHistory();
      }
      return;
    }

    // Some failures have an obvious next step: offer it instead of only
    // reporting the error.
    if (operation === "push" && result.kind === "safety" && view.repository?.canFix) {
      const confirmed = await confirmAction({
        title: "Fix the repository before pushing",
        message: result.message,
        warning: result.detail || "",
        acceptLabel: "Fix and push",
      });
      if (confirmed && (await fixRepository())) {
        await sync("push");
      }
      return;
    }
    if (operation === "push" && result.kind === "rejected") {
      const confirmed = await confirmAction({
        title: "Pull before pushing",
        message: result.message,
        acceptLabel: "Pull now",
      });
      if (confirmed) {
        await sync("pull");
        if (view.status && !view.status.mergeInProgress && view.status.ahead > 0 && view.status.behind === 0) {
          await sync("push");
        }
      }
      return;
    }
    fail(syncFailureTitles[operation], result);
  }

  async function switchBranch(name) {
    const result = await act("branch", () => api().SwitchBranch({ path: view.path, name }));
    view.branches = null;
    loadBranches();
    if (!result) {
      return;
    }
    if (!result.ok) {
      fail("Could not switch branch", result);
      return;
    }
    view.diff = null;
    succeed(result, view.repository?.name || "");
    resetHistory();
  }

  function ensureBranchDialog() {
    if (el("branchDialog")) {
      return;
    }
    document.body.insertAdjacentHTML(
      "beforeend",
      `
      <dialog id="branchDialog">
        <form id="branchForm" class="dialog-body" method="dialog" novalidate>
          <h2>New branch</h2>
          <div class="form-stack">
            <label>
              Branch name
              <input name="name" placeholder="feature/my-change" autocomplete="off" spellcheck="false" required />
              <span class="field-hint" id="branchHint"></span>
            </label>
            <div id="branchError" class="note error" role="alert" hidden></div>
          </div>
          <div class="dialog-actions">
            <button type="button" id="branchCancel">Cancel</button>
            <button type="submit" class="primary">Create branch</button>
          </div>
        </form>
      </dialog>
    `,
    );
    const dialog = el("branchDialog");
    const form = el("branchForm");
    el("branchCancel").addEventListener("click", () => dialog.close());
    form.addEventListener("submit", async (event) => {
      event.preventDefault();
      const name = form.elements.name.value.trim();
      const errorBox = el("branchError");
      if (!name) {
        errorBox.textContent = "Enter a branch name.";
        errorBox.hidden = false;
        return;
      }
      errorBox.hidden = true;
      let result;
      try {
        result = await api().CreateBranch({ path: view.path, name });
      } catch (error) {
        errorBox.textContent = String(error);
        errorBox.hidden = false;
        return;
      }
      if (!result.ok) {
        errorBox.textContent = result.message;
        errorBox.hidden = false;
        return;
      }
      dialog.close();
      applyState(result.state);
      view.branches = null;
      drawAll();
      loadBranches();
      succeed(result, view.repository?.name || "");
      resetHistory();
    });
  }

  function newBranch() {
    ensureBranchDialog();
    const form = el("branchForm");
    form.reset();
    el("branchError").hidden = true;
    el("branchHint").textContent = `Starts from ${commitTarget()}. Your uncommitted changes come along.`;
    el("branchDialog").showModal();
    form.elements.name.focus();
  }

  registerPage({
    id: "repo",
    navParent: "repositories",
    render: page,
    bind,
  });
})();
