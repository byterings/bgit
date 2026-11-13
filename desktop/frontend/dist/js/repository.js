// Repository workspace: everyday Git work for one repository.
// Header (branch, sync), then tabs: Changes, History, Branches, Remotes,
// Identity. The page is drawn in regions that update on their own, so typing
// a commit message or reading a diff is never interrupted by a refresh.
(() => {
  const HISTORY_PAGE = 50;
  const TABS = [
    { id: "changes", label: "Changes" },
    { id: "history", label: "History" },
    { id: "branches", label: "Branches" },
    { id: "remotes", label: "Remotes" },
    { id: "identity", label: "Identity" },
  ];
  const KIND = {
    modified: { letter: "M", text: "Modified" },
    added: { letter: "A", text: "Added" },
    untracked: { letter: "A", text: "New file" },
    deleted: { letter: "D", text: "Deleted" },
    renamed: { letter: "R", text: "Renamed" },
    copied: { letter: "C", text: "Copied" },
    conflicted: { letter: "!", text: "Conflict" },
    typechange: { letter: "T", text: "Type changed" },
  };

  let view = freshView();
  let drawnTab = "";

  function freshView() {
    return {
      path: "",
      repository: null,
      status: null,
      error: "",
      tab: "changes",
      busy: "",
      // Changes
      checked: new Set(),
      knownFiles: new Set(),
      selectedFile: "",
      fileFilter: "",
      diff: null,
      diffLoading: false,
      summary: "",
      description: "",
      mergeSummaryOffered: false,
      // History
      commits: null,
      historyDone: false,
      selectedCommit: "",
      commitFiles: null,
      commitFile: "",
      commitDiff: null,
      // Branches and remotes
      branches: null,
      remotes: null,
      // Identity
      connection: null,
      testing: false,
    };
  }

  const el = (id) => document.getElementById(id);
  const files = () => view.status?.files || [];
  const repoName = () => view.repository?.name || baseName(view.path);
  const repoIdentity = () => identityByAlias(view.repository?.alias);

  // ---------- Opening and refreshing ----------

  // known is the repository as a list already shows it; with it the header is
  // drawn at once and only the file list waits for Git.
  // The identity of the open repository, for the identity switcher.
  window.currentRepositoryAlias = () => view.repository?.alias || "";

  window.openRepository = (path, known = null) => {
    view = freshView();
    view.path = path;
    view.repository = known || repositoryByPath(path);
    drawnTab = "";
    recentRepositories.touch(path);
    navigate("repo");
    refresh();
  };

  function applyState(repoState) {
    if (!repoState) {
      return;
    }
    view.repository = repoState.repository;
    view.status = repoState.status;
    view.error = "";

    const paths = new Set(files().map((file) => file.path));
    // New changes start selected; files that are gone leave the selection.
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

  // Reloads branch and file state. Quiet refreshes (window focus) keep the
  // page as it is when the request fails.
  async function refresh({ quiet = false } = {}) {
    if (!view.path) {
      return;
    }
    const path = view.path;
    try {
      const repoState = await api().GetRepositoryState(path);
      if (path !== view.path) {
        return;
      }
      const previousFile = view.selectedFile;
      applyState(repoState);
      drawAll();
      if (view.tab === "changes" && view.selectedFile && (view.selectedFile !== previousFile || !view.diff || quiet)) {
        loadDiff();
      }
    } catch (error) {
      if (path !== view.path) {
        return;
      }
      view.error = errorText(error);
      if (!quiet) {
        drawAll();
      }
    }
  }

  // ---------- Page skeleton ----------

  function page() {
    return `
      <div class="ws">
        <div class="ws-head" id="wsHead"></div>
        <div class="ws-tabbar"><div class="tabs" id="wsTabs" role="tablist"></div></div>
        <div class="ws-banners" id="wsBanners"></div>
        <div class="ws-body" id="wsBody"></div>
      </div>`;
  }

  function bind() {
    if (!view.path) {
      navigate("repositories", {}, { remember: false });
      return;
    }
    drawnTab = "";
    drawAll();
  }

  function drawAll() {
    if (state.page !== "repo" || !el("wsHead")) {
      return;
    }
    drawHead();
    drawTabs();
    drawBanners();
    drawBody();
  }

  // ---------- Header ----------

  // The branch's relation to its remote, as a short status: tone, icon, and
  // text. "detail" is hidden in narrow windows.
  function syncState() {
    const status = view.status;
    const repo = view.repository;
    if (!status) return { tone: "", iconName: "", text: "", detail: "" };
    if (!repo?.remoteUrl) return { tone: "", iconName: "globe", text: "No remote", detail: "", tip: "This repository has no remote to push to" };
    if (status.detached) return { tone: "info", iconName: "commit", text: "Detached", detail: ` at ${status.headOid}`, tip: "No branch is checked out" };
    if (status.unborn) return { tone: "", iconName: "commit", text: "No commits yet", detail: "" };
    if (!status.hasUpstream) return { tone: "info", iconName: "upload", text: "Not published", detail: "", tip: "This branch is not on the remote yet" };
    const parts = [];
    if (status.ahead > 0) parts.push(`${icon("arrow-up", "sm")}${status.ahead} to push`);
    if (status.behind > 0) parts.push(`${icon("arrow-down", "sm")}${status.behind} to pull`);
    if (parts.length > 0) {
      return { tone: status.behind > 0 ? "warning" : "accent", html: parts.join(""), tip: `Compared with ${status.upstream}` };
    }
    return { tone: "success", iconName: "check", text: "Up to date", detail: ` with ${status.upstream}`, tip: `Nothing to push or pull. Fetch to check the remote again.` };
  }

  function syncButton(operation, iconName, label, workingLabel, count, extraClass, disabled, tip) {
    const working = view.busy === operation;
    return `
      <button type="button" class="btn sync-btn ${extraClass}" data-action="ws-${operation}" ${disabled ? "disabled" : ""} data-tip="${escapeHtml(tip)}" data-tip-align="end" aria-label="${escapeHtml(label)}">
        ${working ? `<span class="spinner" aria-hidden="true"></span><span class="btn-label">${workingLabel}</span>` : `${icon(iconName)}<span class="btn-label">${label}</span>${count > 0 ? `<span class="count">${count}</span>` : ""}`}
      </button>`;
  }

  // How the repository got its identity, in two words for the header.
  const SOURCE_LABEL = { binding: "Bound", workspace: "Workspace", remote: "From remote" };

  function drawHead() {
    const repo = view.repository;
    const status = view.status;
    const busy = Boolean(view.busy);
    const hasRemote = Boolean(repo?.remoteUrl);
    const canSync = hasRemote && status && !status.detached;
    const publish = canSync && !status.hasUpstream && !status.unborn;
    const branchLabel = !status ? "…" : status.detached ? status.headOid : status.branch;
    const sync = syncState();
    const identity = repoIdentity();
    const source = SOURCE_LABEL[repo?.identitySource] || "Default";

    el("wsHead").innerHTML = `
      <div class="ws-title">
        <button type="button" class="icon-btn" data-action="ws-back" aria-label="Back to repositories" data-tip="Back to repositories" data-tip-align="start">${icon("arrow-left")}</button>
        <h1 class="truncate" title="${escapeHtml(view.path)}">${escapeHtml(repoName())}</h1>
        <button type="button" class="ws-identity ${identity ? "" : "is-missing"} ${repo && repo.health !== "ok" ? "is-attention" : ""}" data-action="ws-identity" aria-haspopup="dialog" aria-expanded="false" aria-label="Repository identity: ${escapeHtml(repo?.alias || "none")}">
          <span class="ws-identity-label">Commits as</span>
          ${identity ? avatar(identity, "xs") : icon("alert", "sm")}
          <strong class="truncate">${escapeHtml(repo?.alias || "No identity")}</strong>
          ${identity ? `<span class="ws-identity-source">${escapeHtml(source)}</span>` : ""}
          ${icon("chevron-down", "sm")}
        </button>
      </div>
      <div class="ws-tools">
        <button type="button" class="btn ws-branch" data-action="ws-branch-menu" aria-haspopup="menu" aria-expanded="false" aria-label="Current branch: ${escapeHtml(branchLabel)}" ${busy || !status ? "disabled" : ""}>
          ${icon(status?.detached ? "commit" : "branch")}<span class="truncate">${escapeHtml(branchLabel)}</span>${icon("chevron-down", "sm")}
        </button>
        <span class="ws-sync ${sync.tone ? `is-${sync.tone}` : ""}" id="wsSync" ${sync.tip ? `data-tip="${escapeHtml(sync.tip)}"` : ""} role="status">
          ${sync.html || `${sync.iconName ? icon(sync.iconName, "sm") : ""}<span class="truncate">${escapeHtml(sync.text)}<span class="ws-sync-detail">${escapeHtml(sync.detail || "")}</span></span>`}
        </span>
        <span class="ws-tools-sep" aria-hidden="true"></span>
        ${syncButton("fetch", "refresh", "Fetch", "Fetching", 0, "", busy || !hasRemote, "Check the remote for new commits")}
        ${syncButton("pull", "download", "Pull", "Pulling", status?.behind || 0, (status?.behind || 0) > 0 ? "btn-attention" : "", busy || !canSync || !status.hasUpstream, "Download and apply new commits")}
        ${syncButton("push", "upload", publish ? "Publish branch" : "Push", "Pushing", status?.ahead || 0, (status?.ahead || 0) > 0 || publish ? "btn-primary" : "", busy || !canSync || status.unborn, publish ? "Upload this branch for the first time" : "Upload your commits")}
        <button type="button" class="icon-btn bordered" data-action="ws-more" aria-label="Repository actions" aria-haspopup="menu" data-tip="More" data-tip-align="end">${icon("more")}</button>
      </div>`;
  }

  // The identity card under the header pill: which identity commits here,
  // why, and how it relates to the default identity.
  function openIdentityCard(button) {
    const repo = view.repository;
    const identity = repoIdentity();
    const fallback = activeIdentity();
    const why =
      repo?.identitySource === "binding"
        ? "This repository is bound to it."
        : repo?.identitySource === "workspace"
          ? `It is inside the workspace ${baseName(repo.sourcePath || "")}.`
          : repo?.identitySource === "remote"
            ? "The remote address names it."
            : "It is the default identity. This repository has no identity of its own.";
    const fixed = repo?.identitySource === "binding" || repo?.identitySource === "workspace";
    const defaultLine = !fallback
      ? ""
      : fallback.alias === repo?.alias
        ? `<p class="muted">${fixed ? "It is also your default identity." : "Switching the default identity changes this repository too."}</p>`
        : `<p class="muted">Your default identity is <strong>${escapeHtml(fallback.alias)}</strong>. It does not apply here.</p>`;
    const problems = repo && repo.health !== "ok" ? (repo.problems || []).join(" ") : "";
    const html = `
      <div class="repo-id-card">
        ${
          identity
            ? `<div class="repo-id-card-head">
                ${avatar(identity, "lg")}
                <div class="grow"><strong class="truncate">${escapeHtml(identity.alias)}</strong><span class="muted truncate">${escapeHtml(identity.name)} · ${escapeHtml(identity.email)}</span></div>
                ${badge(SOURCE_LABEL[repo?.identitySource] || "Default", "accent")}
              </div>
              <p>Commits here are recorded as <strong>${escapeHtml(identity.name)}</strong> and pushed with the GitHub account <strong>${escapeHtml(identity.githubUsername)}</strong>.</p>
              <p class="muted">Why: ${escapeHtml(why)}</p>
              ${defaultLine}`
            : `<p><strong>No identity applies to this repository.</strong></p><p class="muted">${escapeHtml(problems || "Bind it to an identity before committing.")}</p>`
        }
        ${problems && identity ? banner({ tone: repo.canFix ? "warning" : "danger", title: repo.canFix ? "Does not match its identity" : "Cannot be managed", text: problems }) : ""}
        <div class="cluster">
          ${problems && repo.canFix ? `<button type="button" class="btn btn-sm btn-primary" data-card="fix">${icon("wrench", "sm")}Fix</button>` : ""}
          <button type="button" class="btn btn-sm ${problems && repo.canFix ? "" : "btn-primary"}" data-card="change">${icon("users", "sm")}Change identity</button>
          <button type="button" class="btn btn-sm btn-ghost" data-card="details">Details</button>
        </div>
      </div>`;
    openPopover(button, html, {
      width: 340,
      className: "identity-pop",
      onOpen(element) {
        element.addEventListener("click", (event) => {
          const choice = event.target.closest("[data-card]")?.dataset.card;
          if (!choice) return;
          closePopover();
          if (choice === "fix") fixThisRepository();
          if (choice === "change") changeIdentity();
          if (choice === "details") setTab("identity");
        });
        element.querySelector("[data-card]")?.focus({ preventScroll: true });
      },
    });
  }

  // ---------- Tabs ----------

  function drawTabs() {
    const count = files().length;
    const needsFix = view.repository && view.repository.health !== "ok";
    el("wsTabs").innerHTML = TABS.map((tab) => {
      let extra = "";
      if (tab.id === "changes" && count > 0) {
        extra = `<span class="count">${count}${view.status?.truncated ? "+" : ""}</span>`;
      }
      if (tab.id === "identity" && needsFix) {
        extra = `<span class="dot is-warning"></span>`;
      }
      return `<button type="button" role="tab" class="tab ${view.tab === tab.id ? "active" : ""}" aria-selected="${view.tab === tab.id}" data-action="ws-tab" data-tab="${tab.id}">${tab.label}${extra}</button>`;
    }).join("");
  }

  function setTab(tab) {
    if (tab === view.tab) {
      return;
    }
    view.tab = tab;
    drawTabs();
    drawBanners();
    drawBody();
    if (tab === "history" && !view.commits) loadHistory();
    // Branches and remotes may have changed outside the app; what is already
    // known shows at once and is refreshed underneath.
    if (tab === "branches") loadBranches();
    if (tab === "remotes") loadRemotes();
  }

  // ---------- Banners ----------

  function drawBanners() {
    const repo = view.repository;
    const status = view.status;
    const items = [];
    if (view.error) {
      items.push(
        banner({
          tone: "danger",
          title: "This repository could not be read",
          text: view.error,
          actions: `<button type="button" class="btn btn-sm" data-action="ws-retry">${icon("refresh", "sm")}Try again</button>`,
        }),
      );
    }
    // The Identity tab shows the same information in detail.
    if (repo && repo.health !== "ok" && (repo.problems || []).length > 0 && view.tab !== "identity") {
      items.push(
        banner({
          tone: repo.canFix ? "warning" : "danger",
          title: repo.canFix ? "This repository does not match its identity" : "bgit cannot manage this repository",
          text: (repo.problems || []).join(" "),
          actions: repo.canFix ? `<button type="button" class="btn btn-sm btn-primary" data-action="ws-fix">${icon("wrench", "sm")}Fix</button>` : "",
        }),
      );
    }
    if (status?.mergeInProgress) {
      items.push(
        banner({
          tone: "warning",
          iconName: "branch",
          title: "Merge in progress",
          text:
            status.unresolvedConflicts > 0
              ? `${plural(status.unresolvedConflicts, "file has", "files have")} conflicts. Open ${status.unresolvedConflicts === 1 ? "it" : "them"} in your editor, fix the sections between <<<<<<< and >>>>>>>, and save. Then commit here.`
              : "All conflicts are resolved. Enter a summary and commit to finish the merge.",
          actions: `<button type="button" class="btn btn-sm" data-action="ws-abort-merge" ${view.busy ? "disabled" : ""}>Cancel merge</button>`,
        }),
      );
    }
    if (status?.rebaseInProgress) {
      items.push(banner({ tone: "warning", title: "Rebase in progress", text: "Finish or cancel the rebase in a terminal before making changes here." }));
    }
    if (status?.detached) {
      items.push(banner({ tone: "info", title: "No branch is checked out", text: `You are looking at commit ${status.headOid}. Choose a branch to commit and sync.` }));
    }
    el("wsBanners").innerHTML = items.join("");
  }

  // ---------- Body ----------

  function drawBody() {
    const body = el("wsBody");
    if (!view.status) {
      drawnTab = "";
      body.innerHTML = view.error ? "" : loadingLine("Reading the repository");
      return;
    }
    const rebuild = drawnTab !== view.tab;
    drawnTab = view.tab;
    switch (view.tab) {
      case "history":
        if (rebuild) {
          body.innerHTML = `<div class="history" id="historyLayout"><div class="pane history-list" id="historyList"></div><div class="splitter x" id="historySplitter" aria-label="Resize the commit list"></div><div class="pane commit-detail" id="commitDetail"></div></div>`;
          const layout = el("historyLayout");
          bindSplitter(el("historySplitter"), {
            container: layout,
            variable: "--history-w",
            key: "history",
            fallback: 340,
            min: 240,
            max: () => Math.max(260, Math.min(620, layout.clientWidth * 0.5)),
          });
        }
        drawHistory();
        drawCommitDetail();
        break;
      case "branches":
        drawBranches();
        break;
      case "remotes":
        drawRemotes();
        break;
      case "identity":
        drawIdentity();
        break;
      default:
        drawChanges(rebuild);
    }
  }

  // ---------- Changes ----------

  function drawChanges(rebuild) {
    const body = el("wsBody");
    // When the panes already exist only their contents are updated: rebuilding
    // them would replace the commit inputs while the user is typing.
    if (rebuild || !el("fileRows")) {
      const collapsed = prefs.get("ws.sideCollapsed", false);
      body.innerHTML = `
        <div class="changes ${collapsed ? "side-collapsed" : ""}" id="changesLayout">
          <aside class="pane side-pane" id="sidePane" aria-label="Changed files and commit">
            <div class="pane-head">
              <label class="check"><input type="checkbox" id="checkAll" aria-label="Select all files" /><span id="selectionText"></span></label>
              <span class="pane-head-tools">
                <button type="button" class="icon-btn sm" data-action="files-menu" aria-label="File actions" aria-haspopup="menu" data-tip="File actions" data-tip-align="end">${icon("more")}</button>
                <button type="button" class="icon-btn sm" data-action="side-toggle" aria-label="Hide the file list" data-tip="Hide the file list (${MOD_KEY}+B)" data-tip-align="end">${icon("sidebar")}</button>
              </span>
            </div>
            <label class="search pane-search">${icon("search", "sm")}<input class="input sm" id="fileFilter" type="text" placeholder="Filter files" value="${escapeHtml(view.fileFilter)}" autocomplete="off" spellcheck="false" aria-label="Filter changed files" /></label>
            <div class="file-rows" id="fileRows" tabindex="0" aria-label="Changed files. Arrow keys move, Space includes or leaves out a file."></div>
            <div class="splitter y" id="commitSplitter" aria-label="Resize the commit area"></div>
            <div class="commit-area" id="commitArea">
              <form id="commitForm" class="commit-form" novalidate>
                <span class="commit-target truncate" id="commitTarget"></span>
                <div class="commit-identity" id="commitIdentity"></div>
                <div class="commit-summary">
                  <input class="input" id="commitSummary" name="summary" maxlength="200" placeholder="Summary (required)" aria-label="Summary" autocomplete="off" value="${escapeHtml(view.summary)}" />
                  <span class="commit-count" id="summaryCount" aria-live="polite"></span>
                </div>
                <textarea class="textarea" id="commitDescription" name="description" placeholder="Description (optional)" aria-label="Description (optional)">${escapeHtml(view.description)}</textarea>
                <button type="submit" class="btn btn-primary btn-lg btn-block" id="commitButton"></button>
                <span class="commit-hint" id="commitHint"></span>
              </form>
              <div class="sync-hint" id="syncHint"></div>
            </div>
          </aside>
          <div class="splitter x" id="sideSplitter" aria-label="Resize the file list"></div>
          <nav class="side-rail" aria-label="Collapsed file list">
            <button type="button" class="icon-btn" data-action="side-toggle" aria-label="Show the file list" data-tip="Show the file list (${MOD_KEY}+B)" data-tip-align="start">${icon("sidebar-open")}</button>
            <span class="rail-count" id="railCount" data-tip="Files selected for the commit" data-tip-align="start"></span>
            <button type="button" class="icon-btn" data-action="side-commit" aria-label="Write the commit" data-tip="Commit" data-tip-align="start">${icon("check")}</button>
          </nav>
          <section class="pane diff-pane" id="diffPane" aria-label="Changes in the selected file"></section>
        </div>`;
      const layout = el("changesLayout");
      bindSplitter(el("sideSplitter"), {
        container: layout,
        variable: "--side-w",
        key: "side",
        fallback: 320,
        min: 240,
        max: () => Math.max(260, Math.min(560, layout.clientWidth * 0.5)),
      });
      bindSplitter(el("commitSplitter"), {
        container: el("sidePane"),
        variable: "--commit-h",
        key: "commit",
        fallback: 236,
        min: 196,
        max: () => Math.max(220, el("sidePane").clientHeight - 150),
        axis: "y",
        reverse: true,
      });
    }
    drawFiles();
    syncCommitInputs();
    updateCommitControls();
    drawDiff("diffPane", view.diff, view.diffLoading, view.selectedFile, true);
    if (view.selectedFile && !view.diff && !view.diffLoading) {
      loadDiff();
    }
  }

  function visibleFiles() {
    const needle = view.fileFilter.trim().toLowerCase();
    return needle ? files().filter((file) => file.path.toLowerCase().includes(needle)) : files();
  }

  function fileRow(file, { selectable, selected, checked }) {
    const kind = KIND[file.kind] || KIND.modified;
    const slash = file.path.lastIndexOf("/");
    const folder = slash >= 0 ? file.path.slice(0, slash + 1) : "";
    const name = slash >= 0 ? file.path.slice(slash + 1) : file.path;
    const title = file.origPath ? `${file.origPath} → ${file.path}` : file.path;
    return `
      <div class="file-row ${selected ? "selected" : ""} ${selectable && !checked ? "is-excluded" : ""}" data-action="select-file" data-file="${escapeHtml(file.path)}" ${selectable ? 'data-context="file"' : ""} title="${escapeHtml(title)}${selectable && !checked ? " (left out of the commit)" : ""}" ${selected ? 'aria-current="true"' : ""}>
        ${selectable ? `<input type="checkbox" data-check="${escapeHtml(file.path)}" ${checked ? "checked" : ""} aria-label="Include ${escapeHtml(file.path)} in the commit" />` : ""}
        <span class="file-name truncate-start"><span><span class="faint">${escapeHtml(folder)}</span>${escapeHtml(name)}</span></span>
        <span class="file-status ${escapeHtml(file.kind)}" title="${escapeHtml(kind.text)}">${kind.letter}</span>
      </div>`;
  }

  function drawFiles() {
    const rows = el("fileRows");
    if (!rows) {
      return;
    }
    const all = files();
    const visible = visibleFiles();
    if (all.length === 0) {
      rows.innerHTML = `<p class="pane-note">Files you change appear here.</p>`;
    } else if (visible.length === 0) {
      rows.innerHTML = `<p class="pane-note">No file matches “${escapeHtml(view.fileFilter)}”.</p>`;
    } else {
      rows.innerHTML = visible.map((file) => fileRow(file, { selectable: true, selected: file.path === view.selectedFile, checked: view.checked.has(file.path) })).join("");
    }
    const checkAll = el("checkAll");
    checkAll.disabled = all.length === 0;
    checkAll.checked = all.length > 0 && view.checked.size === all.length;
    checkAll.indeterminate = view.checked.size > 0 && view.checked.size < all.length;
    el("selectionText").textContent =
      all.length === 0 ? "No changed files" : `${view.checked.size} of ${plural(all.length, "file", "files")} selected${view.status.truncated ? " (more not shown)" : ""}`;
    const rail = el("railCount");
    if (rail) {
      rail.textContent = all.length === 0 ? "0" : `${view.checked.size}/${all.length}`;
    }
  }

  // Keyboard use of the file list: arrows move the selection, Space includes
  // or leaves out the selected file.
  function fileListKeys(event) {
    const visible = visibleFiles();
    if (visible.length === 0) {
      return;
    }
    const index = visible.findIndex((file) => file.path === view.selectedFile);
    if (event.key === "ArrowDown" || event.key === "ArrowUp") {
      event.preventDefault();
      const next = visible[Math.max(0, Math.min(visible.length - 1, index + (event.key === "ArrowDown" ? 1 : -1)))];
      selectFile(next.path);
    } else if (event.key === " " && view.selectedFile) {
      event.preventDefault();
      toggleChecked(view.selectedFile, !view.checked.has(view.selectedFile));
    }
  }

  function selectFile(path) {
    if (!path || path === view.selectedFile) {
      return;
    }
    view.selectedFile = path;
    view.diff = null;
    drawFiles();
    el("fileRows")?.querySelector(".file-row.selected")?.scrollIntoView({ block: "nearest" });
    loadDiff();
  }

  // Moves through the changed files from the diff header.
  function stepFile(step) {
    const list = visibleFiles().length > 0 ? visibleFiles() : files();
    const index = list.findIndex((file) => file.path === view.selectedFile);
    const next = list[index + step];
    if (next) {
      selectFile(next.path);
    }
  }

  function toggleChecked(path, include) {
    if (include) {
      view.checked.add(path);
    } else {
      view.checked.delete(path);
    }
    drawFiles();
    updateCommitControls();
  }

  function setSideCollapsed(collapsed) {
    prefs.set("ws.sideCollapsed", collapsed);
    el("changesLayout")?.classList.toggle("side-collapsed", collapsed);
    drawDiff("diffPane", view.diff, view.diffLoading, view.selectedFile, true);
  }

  // Why committing is not possible right now, or "" when it is.
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

  // Copies message text that the app changed (cleared after a commit, or
  // suggested for a merge) into the inputs, without disturbing typing.
  function syncCommitInputs() {
    const form = el("commitForm");
    if (!form) {
      return;
    }
    for (const name of ["summary", "description"]) {
      const input = form.elements[name];
      if (input.value !== view[name] && document.activeElement !== input) {
        input.value = view[name];
      }
    }
  }

  // Updates everything around the commit inputs without touching the inputs.
  function updateCommitControls() {
    const button = el("commitButton");
    if (!button) {
      return;
    }
    const status = view.status;
    const identity = repoIdentity();
    const count = view.checked.size;
    const target = status.detached ? `commit ${status.headOid}` : status.branch;

    el("commitTarget").textContent = status.mergeInProgress ? `Finish merge on ${target}` : `Commit to ${target}`;
    el("commitIdentity").innerHTML = identity
      ? `${avatar(identity, "sm")}<div><strong class="truncate">${escapeHtml(identity.name)}</strong><span class="muted truncate">${escapeHtml(identity.email)}</span></div><button type="button" class="badge is-accent badge-btn" data-action="ws-identity" data-tip="Commits here use the identity '${escapeHtml(identity.alias)}'" data-tip-align="end">${escapeHtml(identity.alias)}</button>`
      : `${icon("alert")}<div><strong>No identity</strong><span class="muted">Bind this repository to an identity first.</span></div><button type="button" class="btn btn-sm" data-action="identity-change">Choose</button>`;
    el("commitIdentity").classList.toggle("is-missing", !identity);

    const length = view.summary.length;
    const counter = el("summaryCount");
    counter.textContent = length > 50 ? `${length}/72` : "";
    counter.classList.toggle("is-long", length > 72);

    button.disabled = !canCommit();
    button.innerHTML =
      view.busy === "commit"
        ? `<span class="spinner" aria-hidden="true"></span>Committing`
        : status.mergeInProgress
          ? `${icon("check")}Commit merge`
          : `${icon("check")}Commit ${count > 0 ? plural(count, "file", "files") : ""}`;
    el("commitHint").innerHTML = commitProblem()
      ? `<span class="is-problem">${escapeHtml(commitProblem())}</span>`
      : `<span class="kbd">${MOD_KEY}</span><span class="kbd">Enter</span> to commit`;

    const hint = el("syncHint");
    const parts = [];
    if (status.ahead > 0 && status.hasUpstream) {
      parts.push(`<div class="sync-hint-row">${icon("upload", "sm")}<span>${plural(status.ahead, "commit", "commits")} not pushed yet</span><button type="button" class="link-btn" data-action="ws-push">Push</button></div>`);
    }
    if (status.behind > 0) {
      parts.push(`<div class="sync-hint-row">${icon("download", "sm")}<span>${plural(status.behind, "commit", "commits")} to pull</span><button type="button" class="link-btn" data-action="ws-pull">Pull</button></div>`);
    }
    if (view.repository?.remoteUrl && !status.hasUpstream && !status.unborn && !status.detached) {
      parts.push(`<div class="sync-hint-row">${icon("upload", "sm")}<span>This branch is not on the remote yet</span><button type="button" class="link-btn" data-action="ws-push">Publish</button></div>`);
    }
    hint.innerHTML = parts.join("");
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
      view.diff = { path: file, lines: [{ kind: "note", text: errorText(error) }], additions: 0, deletions: 0 };
    }
    view.diffLoading = false;
    if (view.tab === "changes") {
      drawDiff("diffPane", view.diff, false, file, true);
    }
  }

  // Pairs removed lines with the added lines that replace them, so the part
  // of a line that actually changed can be marked. Returns a map from line
  // index to the [start, end) range of changed characters.
  function changedRanges(lines) {
    const ranges = new Map();
    let i = 0;
    while (i < lines.length) {
      if (lines[i].kind !== "del") {
        i += 1;
        continue;
      }
      const delStart = i;
      while (i < lines.length && lines[i].kind === "del") i += 1;
      const addStart = i;
      while (i < lines.length && lines[i].kind === "add") i += 1;
      const pairs = Math.min(addStart - delStart, i - addStart);
      for (let n = 0; n < pairs; n += 1) {
        const before = lines[delStart + n].text;
        const after = lines[addStart + n].text;
        let prefix = 0;
        while (prefix < before.length && prefix < after.length && before[prefix] === after[prefix]) prefix += 1;
        let suffix = 0;
        while (suffix < before.length - prefix && suffix < after.length - prefix && before[before.length - 1 - suffix] === after[after.length - 1 - suffix]) suffix += 1;
        // Only mark when enough of the line stayed the same to make the
        // marked part meaningful.
        const kept = prefix + suffix;
        if (kept >= 3 && kept >= Math.min(before.length, after.length) * 0.3) {
          if (before.length - suffix > prefix) ranges.set(delStart + n, [prefix, before.length - suffix]);
          if (after.length - suffix > prefix) ranges.set(addStart + n, [prefix, after.length - suffix]);
        }
      }
    }
    return ranges;
  }

  // Wraps the characters [start, end) of highlighted HTML in <mark>, closing
  // and re-opening the mark around tags so the markup stays well formed.
  function markRange(html, start, end) {
    let out = "";
    let position = 0;
    let open = false;
    for (let i = 0; i < html.length; ) {
      if (html[i] === "<") {
        const close = html.indexOf(">", i) + 1;
        out += open ? `</mark>${html.slice(i, close)}<mark>` : html.slice(i, close);
        i = close;
        continue;
      }
      if (position === start && !open) {
        out += "<mark>";
        open = true;
      }
      if (position === end && open) {
        out += "</mark>";
        open = false;
      }
      const length = html[i] === "&" ? html.indexOf(";", i) + 1 - i : 1;
      out += html.slice(i, i + length);
      i += length;
      position += 1;
    }
    return open ? `${out}</mark>` : out;
  }

  function diffRow(line, family, range) {
    if (line.kind === "hunk") {
      return `<tr class="hunk"><td class="num"></td><td class="num"></td><td class="code">${escapeHtml(line.text)}</td></tr>`;
    }
    if (line.kind === "note") {
      return `<tr class="note"><td class="num"></td><td class="num"></td><td class="code">${escapeHtml(line.text)}</td></tr>`;
    }
    const marker = line.kind === "add" ? "+" : line.kind === "del" ? "−" : " ";
    let code = highlightCode(line.text, family);
    if (range) {
      code = markRange(code, range[0], range[1]);
    }
    return `<tr class="${line.kind}"><td class="num">${line.oldLine || ""}</td><td class="num">${line.newLine || ""}</td><td class="code"><span class="marker" aria-hidden="true">${marker}</span>${code}</td></tr>`;
  }

  function drawDiff(targetId, diff, loading, file, discardable) {
    const target = el(targetId);
    if (!target) {
      return;
    }
    if (!file) {
      const clean = files().length === 0 && targetId === "diffPane";
      target.innerHTML = emptyState({
        iconName: clean ? "check" : "file",
        title: clean ? "Working tree is clean" : "No file selected",
        text: clean
          ? view.status?.ahead > 0
            ? `There is nothing to commit. ${plural(view.status.ahead, "commit is", "commits are")} waiting to be pushed.`
            : "There is nothing to commit. Changes you save in your editor appear here."
          : "Select a file to see what changed.",
        actions:
          clean && view.status?.ahead > 0 && view.status?.hasUpstream
            ? `<button type="button" class="btn btn-primary" data-action="ws-push" ${view.busy ? "disabled" : ""}>${icon("upload")}Push ${plural(view.status.ahead, "commit", "commits")}</button>`
            : "",
        compact: true,
      });
      target.dataset.file = "";
      return;
    }
    // Keep the reading position when the same file is re-drawn.
    const previousScroll = target.dataset.file === file ? target.querySelector(".diff-scroll")?.scrollTop || 0 : 0;
    target.dataset.file = file;

    const working = targetId === "diffPane";
    const list = working ? visibleFiles() : view.commitFiles || [];
    const change = (working ? files() : list).find((item) => item.path === file);
    const kind = KIND[change?.kind] || KIND.modified;
    const lines = diff?.lines || [];
    const family = highlightFamily(file);
    const wrap = prefs.get("diff.wrap", false);
    let content;
    if (loading) {
      content = loadingLine("Loading changes");
    } else if (diff?.binary) {
      content = emptyState({ iconName: "file", title: "Binary file", text: "This file is not text, so there is no line-by-line view.", compact: true });
    } else if (lines.length === 0) {
      content = `<p class="pane-note">No text changes to show for this file.</p>`;
    } else {
      const ranges = changedRanges(lines);
      content = `
        <div class="diff-scroll">
          <table class="diff ${wrap ? "wrap" : ""}"><tbody>${lines.map((line, index) => diffRow(line, family, ranges.get(index))).join("")}</tbody></table>
          ${diff.truncated ? `<p class="pane-note">This change is long. Only the first part is shown.</p>` : ""}
        </div>`;
    }
    const index = list.findIndex((item) => item.path === file);
    const stepper =
      working && list.length > 1
        ? `<span class="diff-step">
            <button type="button" class="icon-btn sm" data-action="diff-step" data-step="-1" aria-label="Previous file" data-tip="Previous file" ${index <= 0 ? "disabled" : ""}>${icon("chevron-up", "sm")}</button>
            <span class="faint">${index + 1}/${list.length}</span>
            <button type="button" class="icon-btn sm" data-action="diff-step" data-step="1" aria-label="Next file" data-tip="Next file" ${index < 0 || index >= list.length - 1 ? "disabled" : ""}>${icon("chevron-down", "sm")}</button>
          </span>`
        : "";
    target.innerHTML = `
      <div class="diff-head">
        <span class="file-status ${escapeHtml(change?.kind || "modified")}" title="${escapeHtml(kind.text)}">${kind.letter}</span>
        <span class="diff-path mono truncate-start grow selectable" title="${escapeHtml(file)}"><span>${escapeHtml(file)}</span></span>
        ${change?.origPath ? `<span class="faint truncate">from ${escapeHtml(change.origPath)}</span>` : ""}
        ${diff && !loading && !diff.binary ? `<span class="diff-stat mono"><span class="add">+${diff.additions || 0}</span><span class="del">−${diff.deletions || 0}</span></span>` : ""}
        ${stepper}
        <button type="button" class="icon-btn sm" data-action="diff-wrap" aria-pressed="${wrap}" aria-label="Wrap long lines" data-tip="Wrap long lines" data-tip-align="end">${icon("wrap", "sm")}</button>
        ${discardable ? `<button type="button" class="btn btn-sm btn-danger-quiet" data-action="discard-file" data-file="${escapeHtml(file)}" ${view.busy ? "disabled" : ""}>Discard</button>` : ""}
      </div>
      ${content}`;
    const scroller = target.querySelector(".diff-scroll");
    if (scroller && previousScroll) {
      scroller.scrollTop = previousScroll;
    }
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
      notify({ title: "Could not load history", message: errorText(error), tone: "danger" });
    }
    drawHistory();
  }

  function drawHistory() {
    const target = el("historyList");
    if (!target) {
      return;
    }
    if (!view.commits) {
      target.innerHTML = loadingLine("Loading history");
      return;
    }
    if (view.commits.length === 0) {
      target.innerHTML = emptyState({ iconName: "commit", title: "No commits yet", text: "Commits you make appear here.", compact: true });
      return;
    }
    target.innerHTML = `
      <div class="pane-head"><span class="eyebrow">${view.status?.detached ? "History" : `History of ${escapeHtml(view.status?.branch || "")}`}</span></div>
      <div class="commit-rows">
        ${view.commits
          .map(
            (commit) => `
              <button type="button" class="commit-row ${commit.oid === view.selectedCommit ? "selected" : ""}" data-action="select-commit" data-commit="${escapeHtml(commit.oid)}">
                <span class="commit-row-main">
                  <strong class="truncate">${escapeHtml(commit.summary)}</strong>
                  <span class="muted truncate">${escapeHtml(commit.authorName)} · ${escapeHtml(timeAgo(commit.time * 1000))}</span>
                </span>
                <span class="commit-row-side">
                  <span class="mono faint">${escapeHtml(commit.shortOid)}</span>
                  ${commit.unpushed ? `<span class="badge is-accent" data-tip="Not pushed yet" data-tip-align="end">${icon("upload", "sm")}</span>` : ""}
                </span>
              </button>`,
          )
          .join("")}
        ${view.historyDone ? "" : `<button type="button" class="btn btn-sm history-more" data-action="history-more">Load older commits</button>`}
      </div>`;
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
      notify({ title: "Could not load the commit", message: errorText(error), tone: "danger" });
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
      view.commitDiff = { path: file, lines: [{ kind: "note", text: errorText(error) }] };
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
      target.innerHTML = view.commits && view.commits.length === 0 ? "" : emptyState({ iconName: "commit", title: "No commit selected", text: "Select a commit to see what it changed.", compact: true });
      return;
    }
    const list = view.commitFiles;
    const author = { name: commit.authorName, alias: commit.authorName };
    target.innerHTML = `
      <div class="commit-head">
        <div class="spread">
          <h2 class="selectable">${escapeHtml(commit.summary)}</h2>
          ${commit.unpushed ? badge("Not pushed", "accent") : ""}
        </div>
        ${commit.body ? `<p class="commit-body selectable">${escapeHtml(commit.body)}</p>` : ""}
        <div class="commit-meta">
          ${avatar(author, "sm")}
          <span class="truncate"><strong>${escapeHtml(commit.authorName)}</strong> <span class="muted">${escapeHtml(commit.authorEmail)}</span></span>
          <span class="muted">${escapeHtml(new Date(commit.time * 1000).toLocaleString())}</span>
          <button type="button" class="btn btn-sm btn-ghost mono" data-action="copy" data-value="${escapeHtml(commit.oid)}" data-copied="Commit id copied" data-tip="Copy the full commit id" data-tip-align="end">${escapeHtml(commit.shortOid)}${icon("copy", "sm")}</button>
        </div>
      </div>
      <div class="commit-split">
        <div class="file-rows commit-files" id="commitFiles">
          ${
            list
              ? list.length > 0
                ? list.map((file) => fileRow(file, { selectable: false, selected: file.path === view.commitFile })).join("")
                : `<p class="pane-note">This commit changed no files.</p>`
              : loadingLine("Loading files")
          }
        </div>
        <div class="splitter x" id="commitFilesSplitter" aria-label="Resize the file list of the commit"></div>
        <div class="commit-diff" id="commitDiff"></div>
      </div>`;
    bindSplitter(el("commitFilesSplitter"), {
      container: target,
      variable: "--commit-files-w",
      key: "commitFiles",
      fallback: 240,
      min: 160,
      max: () => Math.max(180, Math.min(480, target.clientWidth * 0.5)),
    });
    drawDiff("commitDiff", view.commitDiff, Boolean(view.commitFile) && !view.commitDiff, view.commitFile, false);
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

  // ---------- Branches ----------

  async function loadBranches() {
    try {
      view.branches = (await api().ListBranches(view.path)) || [];
    } catch (error) {
      view.branches = [];
      notify({ title: "Could not load branches", message: errorText(error), tone: "danger" });
    }
    if (view.tab === "branches") {
      drawBranches();
    }
    return view.branches;
  }

  function branchGroups() {
    const all = view.branches || [];
    const local = all.filter((branch) => !branch.remote);
    const localNames = new Set(local.map((branch) => branch.name));
    // A remote branch that already has a local copy is not listed twice.
    const remoteOnly = all.filter((branch) => branch.remote && !localNames.has(branch.name.split("/").slice(1).join("/")));
    return { local, remoteOnly };
  }

  function drawBranches() {
    const body = el("wsBody");
    if (!view.branches) {
      body.innerHTML = loadingLine("Loading branches");
      return;
    }
    const { local, remoteOnly } = branchGroups();
    const busy = Boolean(view.busy);
    body.innerHTML = `
      <div class="ws-scroll">
        <div class="ws-page">
          <section class="card">
            <div class="card-head">
              <h2>Local branches <span class="faint">${local.length}</span></h2>
              <button type="button" class="btn btn-sm btn-primary" data-action="branch-new" ${busy || view.status.unborn ? "disabled" : ""}>${icon("plus", "sm")}New branch</button>
            </div>
            <div class="branch-rows">
              ${
                local.length === 0
                  ? `<p class="pane-note">No branches yet. The first commit creates one.</p>`
                  : local
                      .map(
                        (branch) => `
                          <div class="branch-row ${branch.current ? "current" : ""}">
                            ${icon("branch")}
                            <span class="branch-name"><strong class="mono truncate">${escapeHtml(branch.name)}</strong>${branch.current ? badge("Current", "accent") : ""}</span>
                            <span class="muted truncate">${branch.upstream ? `Tracks ${escapeHtml(branch.upstream)}` : "Not on the remote"}</span>
                            <span class="faint">${branch.commitTime ? escapeHtml(timeAgo(branch.commitTime * 1000)) : ""}</span>
                            <span class="branch-actions">
                              ${branch.current ? "" : `<button type="button" class="btn btn-sm" data-action="branch-switch" data-name="${escapeHtml(branch.name)}" ${busy ? "disabled" : ""}>Switch</button>`}
                              <button type="button" class="icon-btn sm" data-action="branch-delete" data-name="${escapeHtml(branch.name)}" aria-label="Delete ${escapeHtml(branch.name)}" data-tip="${branch.current ? "Switch to another branch to delete this one" : "Delete branch"}" data-tip-align="end" ${branch.current || busy ? "disabled" : ""}>${icon("trash", "sm")}</button>
                            </span>
                          </div>`,
                      )
                      .join("")
              }
            </div>
          </section>
          ${
            remoteOnly.length > 0
              ? `<section class="card">
                  <div class="card-head"><h2>On the remote only <span class="faint">${remoteOnly.length}</span></h2></div>
                  <div class="branch-rows">
                    ${remoteOnly
                      .map(
                        (branch) => `
                          <div class="branch-row">
                            ${icon("globe")}
                            <span class="branch-name"><strong class="mono truncate">${escapeHtml(branch.name)}</strong></span>
                            <span class="muted truncate">No local copy yet</span>
                            <span class="faint">${branch.commitTime ? escapeHtml(timeAgo(branch.commitTime * 1000)) : ""}</span>
                            <span class="branch-actions"><button type="button" class="btn btn-sm" data-action="branch-switch" data-name="${escapeHtml(branch.name)}" ${busy ? "disabled" : ""}>Check out</button></span>
                          </div>`,
                      )
                      .join("")}
                  </div>
                </section>`
              : ""
          }
        </div>
      </div>`;
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
    succeed(result);
    resetHistory();
  }

  async function newBranch() {
    const status = view.status;
    const from = status.detached ? `commit ${status.headOid}` : status.branch;
    const result = await openDialog({
      id: "branchDialog",
      title: "New branch",
      description: `Starts from ${from}. Your uncommitted changes come along.`,
      submitLabel: "Create branch",
      busyLabel: "Creating",
      body: `
        <div class="field">
          <label for="branchName">Branch name</label>
          <input class="input mono" id="branchName" name="name" placeholder="feature/my-change" autocomplete="off" spellcheck="false" autofocus />
        </div>`,
      async onSubmit(form) {
        const name = form.elements.name.value.trim();
        if (!name) throw new Error("Enter a branch name.");
        const created = await api().CreateBranch({ path: view.path, name });
        if (!created.ok) throw new Error(created.message);
        return created;
      },
    });
    if (result?.state) {
      applyState(result.state);
      view.branches = null;
      drawAll();
      loadBranches();
      succeed(result);
      resetHistory();
    }
  }

  async function deleteBranch(name) {
    const confirmed = await confirmAction({
      title: `Delete branch ${name}?`,
      message: "The branch is deleted on this computer. The remote is not changed.",
      acceptLabel: "Delete branch",
      danger: true,
    });
    if (!confirmed) {
      return;
    }
    let result = await act("branch", () => api().DeleteBranch({ path: view.path, name, force: false }));
    if (result && !result.ok && result.kind === "unmerged") {
      // Git refused because the commits exist nowhere else. Losing them must
      // be a separate, explicit decision.
      const force = await confirmAction({
        title: `${name} has commits that are not on any other branch`,
        message: "Deleting it loses those commits.",
        warning: "This cannot be undone.",
        acceptLabel: "Delete anyway",
        danger: true,
      });
      if (!force) {
        return;
      }
      result = await act("branch", () => api().DeleteBranch({ path: view.path, name, force: true }));
    }
    view.branches = null;
    loadBranches();
    if (!result) {
      return;
    }
    if (result.ok) {
      succeed(result);
    } else {
      fail("Could not delete the branch", result);
    }
  }

  // ---------- Remotes ----------

  async function loadRemotes() {
    try {
      view.remotes = (await api().ListRemotes(view.path)) || [];
    } catch (error) {
      view.remotes = [];
      notify({ title: "Could not load remotes", message: errorText(error), tone: "danger" });
    }
    if (view.tab === "remotes") {
      drawRemotes();
    }
  }

  // How bgit sees the origin remote, from the repository's health check.
  function originRouting() {
    const repo = view.repository;
    const remoteProblem = (repo?.problems || []).find((problem) => /remote/i.test(problem));
    if (remoteProblem) {
      return { tone: repo.canFix ? "warning" : "danger", text: remoteProblem, fix: repo.canFix };
    }
    if (repo?.remoteUrl) {
      return { tone: "success", text: `Pushes and pulls go through the SSH key of '${repo.alias}'.`, fix: false };
    }
    return null;
  }

  function drawRemotes() {
    const body = el("wsBody");
    if (!view.remotes) {
      body.innerHTML = loadingLine("Loading remotes");
      return;
    }
    const routing = originRouting();
    body.innerHTML = `
      <div class="ws-scroll">
        <div class="ws-page">
          <section class="card">
            <div class="card-head"><h2>Remotes <span class="faint">${view.remotes.length}</span></h2></div>
            ${
              view.remotes.length === 0
                ? emptyState({ iconName: "globe", title: "No remote yet", text: "This repository only exists on this computer. Add a remote with Git to push it to GitHub.", compact: true })
                : `<div class="remote-rows">
                    ${view.remotes
                      .map((remote) => {
                        const isOrigin = remote.name === "origin";
                        const samePush = !remote.pushUrl || remote.pushUrl === remote.fetchUrl;
                        return `
                          <div class="remote-row">
                            <div class="remote-name">${icon("globe")}<strong>${escapeHtml(remote.name)}</strong>${isOrigin ? badge("Used for push and pull", "accent") : ""}</div>
                            <dl class="props">
                              <div><dt>${samePush ? "Address" : "Fetch address"}</dt><dd><span class="mono selectable truncate">${escapeHtml(remote.fetchUrl)}</span><button type="button" class="icon-btn sm" data-action="copy" data-value="${escapeHtml(remote.fetchUrl)}" data-copied="Address copied" aria-label="Copy address" data-tip="Copy">${icon("copy", "sm")}</button></dd></div>
                              ${samePush ? "" : `<div><dt>Push address</dt><dd><span class="mono selectable truncate">${escapeHtml(remote.pushUrl)}</span><button type="button" class="icon-btn sm" data-action="copy" data-value="${escapeHtml(remote.pushUrl)}" data-copied="Address copied" aria-label="Copy push address" data-tip="Copy">${icon("copy", "sm")}</button></dd></div>`}
                            </dl>
                            ${
                              isOrigin && routing
                                ? banner({
                                    tone: routing.tone,
                                    iconName: routing.tone === "success" ? "shield-check" : "alert",
                                    title: routing.tone === "success" ? "Routed through the right identity" : "Not routed through this repository's identity",
                                    text: routing.text,
                                    actions: routing.fix ? `<button type="button" class="btn btn-sm btn-primary" data-action="ws-fix">${icon("wrench", "sm")}Fix</button>` : "",
                                  })
                                : ""
                            }
                          </div>`;
                      })
                      .join("")}
                  </div>`
            }
            <div class="card-foot"><p>bgit manages the address of <strong>origin</strong> so GitHub sees the SSH key of this repository's identity. Other remotes are left exactly as they are.</p></div>
          </section>
        </div>
      </div>`;
  }

  // ---------- Identity ----------

  function identityChecks() {
    const repo = view.repository;
    const identity = repoIdentity();
    const problems = repo?.problems || [];
    const commitProblemText = problems.find((problem) => /commit/i.test(problem));
    const remoteProblemText = problems.find((problem) => /remote/i.test(problem));
    const checks = [];

    checks.push(
      commitProblemText
        ? { tone: "warning", title: "Commit identity", text: commitProblemText }
        : identity
          ? { tone: "success", title: "Commit identity", text: `Git records commits here as ${identity.name} <${identity.email}>.` }
          : { tone: "danger", title: "Commit identity", text: problems[0] || "No identity applies to this repository." },
    );
    if (remoteProblemText) {
      checks.push({ tone: repo.canFix ? "warning" : "danger", title: "Remote address", text: remoteProblemText });
    } else if (repo?.remoteUrl) {
      checks.push({ tone: "success", title: "Remote address", text: `origin goes through the SSH key of '${repo.alias}'.` });
    } else {
      checks.push({ tone: "", title: "Remote address", text: "No remote is configured, so there is nothing to push to yet." });
    }
    if (identity) {
      const keyOk = identity.sshKeyStatus === "available";
      checks.push({
        tone: keyOk ? "success" : "warning",
        title: "SSH key",
        text: keyOk ? "The identity's SSH key is on this computer." : "The identity has no usable SSH key. Edit the identity to choose one.",
      });
      const test = view.connection;
      checks.push({
        tone: test ? (test.passed ? "success" : "warning") : "",
        title: "GitHub connection",
        text: test ? `${test.message}${test.fix && !test.passed ? ` ${test.fix}.` : ""}` : "Not tested in this session.",
        action: `<button type="button" class="btn btn-sm" data-action="identity-test" ${keyOk && !view.testing ? "" : "disabled"}>${view.testing ? `<span class="spinner" aria-hidden="true"></span>Testing` : "Test connection"}</button>`,
      });
    }
    return checks;
  }

  // How the default identity relates to this repository's identity.
  function defaultIdentityNote() {
    const fallback = activeIdentity();
    const repo = view.repository;
    if (!fallback) {
      return `<span class="muted">None</span>`;
    }
    if (fallback.alias === repo?.alias) {
      return `${identityChip(fallback.alias)}<span class="muted truncate">same as this repository</span>`;
    }
    return `${identityChip(fallback.alias)}<span class="muted truncate">not used here</span>`;
  }

  function drawIdentity() {
    const body = el("wsBody");
    const repo = view.repository;
    const identity = repoIdentity();
    const source =
      repo?.identitySource === "binding"
        ? "This repository is bound to it"
        : repo?.identitySource === "workspace"
          ? `The workspace folder ${repo.sourcePath || ""}`
          : repo?.identitySource === "remote"
            ? "The remote address names this identity"
            : "It is the default identity";
    const iconFor = { success: "check", warning: "alert", danger: "alert", "": "info" };
    body.innerHTML = `
      <div class="ws-scroll">
        <div class="ws-page two-col">
          <section class="card">
            <div class="card-head"><h2>Identity for this repository</h2></div>
            <div class="card-body stack">
              ${
                identity
                  ? `<div class="identity-summary">
                      ${avatar(identity, "lg")}
                      <div><strong>${escapeHtml(identity.alias)}</strong><span class="muted">${escapeHtml(identity.name)} · ${escapeHtml(identity.email)}</span></div>
                    </div>
                    <p class="identity-lead">New commits here are recorded as <strong>${escapeHtml(identity.name)} &lt;${escapeHtml(identity.email)}&gt;</strong> and pushed with the GitHub account <strong>${escapeHtml(identity.githubUsername)}</strong>.</p>
                    <dl class="props">
                      <div><dt>Name on commits</dt><dd><span class="selectable truncate">${escapeHtml(identity.name)}</span></dd></div>
                      <div><dt>Email on commits</dt><dd><span class="selectable truncate">${escapeHtml(identity.email)}</span></dd></div>
                      <div><dt>GitHub user</dt><dd><span class="selectable truncate">${escapeHtml(identity.githubUsername)}</span></dd></div>
                      <div><dt>Chosen because</dt><dd><span class="truncate">${escapeHtml(source)}</span></dd></div>
                      <div><dt>Default identity</dt><dd>${defaultIdentityNote()}</dd></div>
                    </dl>`
                  : banner({ tone: "danger", title: "No identity applies here", text: (repo?.problems || [])[0] || "Bind this repository to an identity." })
              }
              <div class="cluster">
                <button type="button" class="btn btn-primary" data-action="identity-change">${icon("users")}Change identity</button>
                ${identity ? `<button type="button" class="btn" data-action="navigate" data-page="identities" data-alias="${escapeHtml(identity.alias)}">Open identity</button>` : ""}
                ${repo?.source === "binding" ? `<button type="button" class="btn btn-danger-quiet" data-action="identity-remove">Remove from bgit</button>` : ""}
              </div>
              <p class="field-hint">The identity of this repository only changes when you change it here.${repo?.identitySource === "binding" || repo?.identitySource === "workspace" ? " Switching the default identity does not affect it." : " It currently follows the default identity; bind it to keep one identity."}</p>
            </div>
          </section>
          <section class="card">
            <div class="card-head">
              <h2>Checks</h2>
              ${repo?.canFix ? `<button type="button" class="btn btn-sm btn-primary" data-action="ws-fix">${icon("wrench", "sm")}Fix repository</button>` : ""}
            </div>
            <div class="check-rows">
              ${identityChecks()
                .map(
                  (check) => `
                    <div class="check-row">
                      <span class="check-icon ${check.tone ? `is-${check.tone}` : ""}">${icon(iconFor[check.tone])}</span>
                      <div><strong>${escapeHtml(check.title)}</strong><span class="muted">${escapeHtml(check.text)}</span></div>
                      ${check.action || ""}
                    </div>`,
                )
                .join("")}
            </div>
          </section>
        </div>
      </div>`;
  }

  async function testConnection() {
    view.testing = true;
    drawIdentity();
    try {
      view.connection = await api().TestIdentityConnection(view.repository.alias);
    } catch (error) {
      view.connection = { passed: false, message: errorText(error), fix: "" };
    }
    view.testing = false;
    if (view.tab === "identity") {
      drawIdentity();
    }
  }

  // ---------- Actions ----------

  // Runs one Git action with the workspace locked. Returns the result object,
  // or null when the backend call itself failed.
  async function act(busy, call) {
    view.busy = busy;
    drawHead();
    drawBanners();
    updateCommitControls();
    try {
      const result = await call();
      if (result?.state) {
        applyState(result.state);
      }
      return result;
    } catch (error) {
      notify({ title: "Action failed", message: errorText(error), tone: "danger" });
      return null;
    } finally {
      view.busy = "";
      drawAll();
      if (view.tab === "changes" && view.selectedFile) {
        loadDiff();
      }
    }
  }

  function succeed(result) {
    addActivity(result.message, repoName());
    notify({ title: result.message });
    loadRepositories();
  }

  function fail(title, result) {
    addActivity(title, result.message, "danger");
    notify({ title, message: result.message, tone: result.kind === "conflict" ? "warning" : "danger", timeout: 9000 });
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
    syncCommitInputs();
    updateCommitControls();
    succeed(result);
    resetHistory();
  }

  async function discard(paths) {
    if (paths.length === 0) {
      return;
    }
    const one = paths.length === 1;
    const confirmed = await confirmAction({
      title: one ? `Discard changes to ${baseName(paths[0])}?` : `Discard changes to all ${paths.length} files?`,
      message: "Changed files go back to the last commit and new files are deleted.",
      warning: "This cannot be undone.",
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
    succeed(result);
  }

  async function fixThisRepository() {
    view.busy = "fix";
    drawHead();
    let fixed = false;
    try {
      const result = await api().FixRepository(view.path);
      addActivity(result.message, repoName());
      notify({ title: result.message });
      applyStatus(result.status);
      loadRepositories();
      fixed = true;
    } catch (error) {
      notify({ title: "Could not fix the repository", message: errorText(error), tone: "danger" });
    }
    view.busy = "";
    view.remotes = null;
    await refresh();
    if (view.tab === "remotes") {
      loadRemotes();
    }
    return fixed;
  }

  async function abortMerge() {
    const confirmed = await confirmAction({
      title: "Cancel the merge?",
      message: "The branch goes back to how it was before pulling.",
      warning: "Conflict resolutions you already made are lost.",
      acceptLabel: "Cancel merge",
      cancelLabel: "Keep merging",
      danger: true,
    });
    if (!confirmed) {
      return;
    }
    const result = await act("abort", () => api().AbortMerge(view.path));
    if (result?.ok) {
      succeed(result);
    } else if (result) {
      fail("Could not cancel the merge", result);
    }
  }

  const SYNC_CALLS = {
    fetch: () => api().FetchRepository(view.path),
    pull: () => api().PullRepository(view.path),
    push: () => api().PushRepository(view.path),
  };
  const SYNC_FAILURES = { fetch: "Fetch failed", pull: "Pull failed", push: "Push failed" };

  async function sync(operation) {
    const result = await act(operation, SYNC_CALLS[operation]);
    if (!result) {
      return;
    }
    if (result.ok) {
      succeed(result);
      view.branches = null;
      if (view.tab === "branches") {
        loadBranches();
      }
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
      if (confirmed && (await fixThisRepository())) {
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
    fail(SYNC_FAILURES[operation], result);
  }

  async function openBranchMenu(button) {
    const branches = view.branches || (await loadBranches());
    const { local, remoteOnly } = branchGroups();
    const items = [{ heading: "Switch branch" }];
    for (const branch of local.slice(0, 12)) {
      items.push({ label: branch.name, icon: "branch", checked: branch.current, onSelect: () => !branch.current && switchBranch(branch.name) });
    }
    if (remoteOnly.length > 0) {
      items.push({ heading: "On the remote" });
      for (const branch of remoteOnly.slice(0, 6)) {
        items.push({ label: branch.name, icon: "globe", onSelect: () => switchBranch(branch.name) });
      }
    }
    items.push({ separator: true });
    items.push({ label: "New branch", icon: "plus", disabled: view.status.unborn, onSelect: newBranch });
    items.push({ label: branches.length > 18 ? "All branches" : "Manage branches", icon: "layers", onSelect: () => setTab("branches") });
    openMenu(button, items, { width: 280 });
  }

  function fileMenuItems(path) {
    return [
      { label: "Discard changes", icon: "trash", danger: true, onSelect: () => discard([path]) },
      { separator: true },
      { label: "Copy path", icon: "copy", onSelect: () => copyText(path, "Path copied") },
    ];
  }

  async function changeIdentity() {
    const updated = await openAddRepositoryDialog(view.repository);
    if (updated) {
      view.connection = null;
      view.remotes = null;
      await refresh();
      if (view.tab === "remotes") {
        loadRemotes();
      }
    }
  }

  async function removeFromBgit() {
    if (await removeRepository(view.repository)) {
      navigate("repositories", {}, { remember: false });
    }
  }

  registerPage({
    id: "repo",
    nav: "repositories",
    // The panes scroll on their own, so the page fills the window.
    fill: true,
    render: page,
    bind,
    // Files were probably edited while the window was in the background.
    focus() {
      if (!view.busy) {
        refresh({ quiet: true });
        if (view.tab === "branches") loadBranches();
        if (view.tab === "remotes") loadRemotes();
      }
    },
    keys(event) {
      if (view.tab !== "changes") {
        return;
      }
      const mod = event.ctrlKey || event.metaKey;
      if (mod && event.key === "Enter") {
        event.preventDefault();
        commit();
      } else if (mod && !event.shiftKey && event.key.toLowerCase() === "b") {
        event.preventDefault();
        setSideCollapsed(!el("changesLayout")?.classList.contains("side-collapsed"));
      } else if (!mod && event.target.id === "fileRows") {
        fileListKeys(event);
      }
    },
    actions: {
      "ws-back"() {
        navigate("repositories", {}, { remember: false });
      },
      "ws-tab"(element) {
        setTab(element.dataset.tab);
      },
      "ws-branch-menu"(element) {
        openBranchMenu(element);
      },
      "ws-identity"(element) {
        openIdentityCard(element);
      },
      "ws-more"(element) {
        const repo = view.repository;
        openMenu(
          element,
          [
            { label: "Show in file manager", icon: "folder", onSelect: () => revealRepository(view.path) },
            { label: "Copy path", icon: "copy", onSelect: () => copyText(view.path, "Path copied") },
            { separator: true },
            { label: "Change identity", icon: "users", onSelect: () => changeIdentity() },
            repo?.source === "binding" ? { label: "Remove from bgit", icon: "trash", danger: true, onSelect: () => removeFromBgit() } : null,
          ],
          { align: "end" },
        );
      },
      "ws-fetch": () => sync("fetch"),
      "ws-pull": () => sync("pull"),
      "ws-push": () => sync("push"),
      "ws-fix": () => fixThisRepository(),
      "ws-retry"(element) {
        withBusy(element, "Reading", () => refresh());
      },
      "ws-abort-merge": () => abortMerge(),
      "select-file"(element, event) {
        // The checkbox only changes what goes into the commit.
        if (event.target.matches("input[type=checkbox]")) {
          return;
        }
        const file = element.dataset.file;
        if (view.tab === "history") {
          if (file !== view.commitFile) {
            selectCommitFile(file);
          }
          return;
        }
        selectFile(file);
        // Arrow keys continue from the row that was clicked.
        el("fileRows")?.focus({ preventScroll: true });
      },
      "side-toggle"() {
        setSideCollapsed(!el("changesLayout")?.classList.contains("side-collapsed"));
      },
      "side-commit"() {
        setSideCollapsed(false);
        el("commitSummary")?.focus();
      },
      "diff-step"(element) {
        stepFile(Number(element.dataset.step));
      },
      "diff-wrap"() {
        prefs.set("diff.wrap", !prefs.get("diff.wrap", false));
        document.querySelectorAll(".diff").forEach((table) => table.classList.toggle("wrap", prefs.get("diff.wrap", false)));
        document.querySelectorAll('[data-action="diff-wrap"]').forEach((button) => button.setAttribute("aria-pressed", String(prefs.get("diff.wrap", false))));
      },
      "files-menu"(element) {
        openMenu(
          element,
          [{ label: "Discard all changes", icon: "trash", danger: true, disabled: files().length === 0, onSelect: () => discard(files().map((file) => file.path)) }],
          { align: "end" },
        );
      },
      "discard-file"(element) {
        discard([element.dataset.file]);
      },
      "select-commit"(element) {
        if (element.dataset.commit !== view.selectedCommit) {
          selectCommit(element.dataset.commit);
        }
      },
      "history-more": () => loadHistory(true),
      "branch-new": () => newBranch(),
      "branch-switch"(element) {
        switchBranch(element.dataset.name);
      },
      "branch-delete"(element) {
        deleteBranch(element.dataset.name);
      },
      "identity-test": () => testConnection(),
      "identity-change": () => changeIdentity(),
      "identity-remove": () => removeFromBgit(),
    },
    on: {
      input(target) {
        if (target.id === "commitSummary") {
          view.summary = target.value;
          updateCommitControls();
        }
        if (target.id === "commitDescription") {
          view.description = target.value;
        }
        if (target.id === "fileFilter") {
          view.fileFilter = target.value;
          drawFiles();
        }
      },
      change(target) {
        if (target.id === "checkAll") {
          view.checked = target.checked ? new Set(files().map((file) => file.path)) : new Set();
          drawFiles();
          updateCommitControls();
          return;
        }
        const file = target.dataset?.check;
        if (file) {
          toggleChecked(file, target.checked);
        }
      },
      submit(target, event) {
        if (target.id === "commitForm") {
          event.preventDefault();
          commit();
        }
      },
    },
    context(element, event) {
      if (element.dataset.context === "file") {
        openContextMenu(event, fileMenuItems(element.dataset.file));
      }
    },
  });
})();
