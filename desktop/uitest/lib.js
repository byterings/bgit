// Shared helpers for the desktop UI tests. Each test starts the dev harness
// (the real Go backend) on an isolated home folder, drives the frontend with
// headless Chrome, and reads the results back from disk and from Git.
//
// GitHub is replaced by local bare repositories reached through a stand-in
// SSH command, so clone, pull, and push run offline with the real Git protocol.
const { spawn, execFileSync } = require("node:child_process");
const fs = require("node:fs");
const net = require("node:net");
const os = require("node:os");
const path = require("node:path");
const puppeteer = require("puppeteer-core");

// Repository root, two levels above desktop/uitest.
const REPO = path.resolve(__dirname, "..", "..");
// Screenshots and the harness binary go to a temp folder, or BGIT_UITEST_OUT.
const OUT = process.env.BGIT_UITEST_OUT || path.join(os.tmpdir(), "bgit-uitest");
const SHOTS = path.join(OUT, "shots");
const CHROME = process.env.CHROME_PATH || findChrome();

function findChrome() {
  const candidates = [
    "/usr/bin/google-chrome",
    "/usr/bin/chromium",
    "/usr/bin/chromium-browser",
    "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
    "C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe",
  ];
  const found = candidates.find((candidate) => fs.existsSync(candidate));
  if (!found) {
    throw new Error("Chrome was not found. Set CHROME_PATH to a Chrome or Chromium executable.");
  }
  return found;
}

// Asks the system for a port nobody is using, so tests never collide with a
// harness that is already running.
function freePort() {
  return new Promise((resolve, reject) => {
    const probe = net.createServer();
    probe.once("error", reject);
    probe.listen(0, "127.0.0.1", () => {
      const { port } = probe.address();
      probe.close(() => resolve(port));
    });
  });
}

async function start({ port, home, width = 1180, height = 760, build = true } = {}) {
  port = port || (await freePort());
  fs.mkdirSync(SHOTS, { recursive: true });
  home = home || fs.mkdtempSync(path.join(os.tmpdir(), "bgit-ui-home-"));
  const bin = path.join(OUT, process.platform === "win32" ? "harness.exe" : "harness");
  if (build) {
    execFileSync("go", ["build", "-tags", "devharness", "-o", bin, "./desktop"], { cwd: REPO, stdio: "inherit" });
  }
  const env = { ...process.env };
  delete env.SSH_AUTH_SOCK;
  // Offline stand-in for github.com: Git's SSH transport runs against local
  // bare repositories under <home>/fake-github/hosts/<host alias>/.
  const fakeRoot = path.join(home, "fake-github");
  fs.mkdirSync(path.join(fakeRoot, "hosts"), { recursive: true });
  const fakeSsh = path.join(fakeRoot, "ssh");
  fs.writeFileSync(fakeSsh, `#!/bin/sh
while [ $# -gt 0 ]; do
  case "$1" in
    -o) shift 2 ;;
    -*) shift ;;
    *) break ;;
  esac
done
host="\${1#git@}"
shift
if [ ! -d "${fakeRoot}/hosts/$host" ]; then
  echo "git@github.com: Permission denied (publickey)." >&2
  exit 255
fi
cd "${fakeRoot}/hosts/$host" || exit 255
exec sh -c "$*"
`, { mode: 0o755 });
  env.GIT_SSH_COMMAND = fakeSsh;
  const server = spawn(bin, ["-addr", `127.0.0.1:${port}`, "-dist", path.join(REPO, "desktop/frontend/dist"), "-home", home], { cwd: REPO, env, stdio: ["ignore", "pipe", "pipe"] });
  let serverLog = "";
  server.stdout.on("data", (d) => (serverLog += d));
  server.stderr.on("data", (d) => (serverLog += d));
  // Ready means our own harness process printed its listening line; a
  // harness that exits early (for example on a port clash) fails the test.
  await new Promise((resolve, reject) => {
    const started = Date.now();
    server.once("exit", (code) => reject(new Error(`harness exited with code ${code}: ${serverLog}`)));
    const tick = () => {
      if (serverLog.includes(`bgit dev harness on http://127.0.0.1:${port}`)) return resolve();
      if (Date.now() - started > 15000) return reject(new Error("harness did not start: " + serverLog));
      setTimeout(tick, 100);
    };
    tick();
  });

  const browser = await puppeteer.launch({ executablePath: CHROME, headless: true, args: ["--no-sandbox", "--disable-gpu", `--window-size=${width},${height}`] });
  const page = await browser.newPage();
  await page.setViewport({ width, height });
  const problems = [];
  page.on("console", (msg) => { if (msg.type() === "error") problems.push("console: " + msg.text()); });
  page.on("pageerror", (err) => problems.push("pageerror: " + err.message));
  page.on("requestfailed", (req) => problems.push("requestfailed: " + req.url()));

  const url = `http://127.0.0.1:${port}/`;
  const ctx = {
    home, page, browser, problems, url,
    async open() { await page.goto(url, { waitUntil: "networkidle0" }); await ctx.settle(); },
    async settle(ms = 350) { await new Promise((r) => setTimeout(r, ms)); },
    // Screenshots are taken without notifications and with the pointer parked,
    // so they show the screen and not a passing toast or tooltip.
    async shot(name, { keepToasts = false } = {}) {
      const size = page.viewport();
      await page.mouse.move(2, size.height - 2);
      if (!keepToasts) await page.evaluate(() => { document.getElementById("toasts").innerHTML = ""; });
      await ctx.settle(250);
      const file = path.join(SHOTS, name + ".png");
      await page.screenshot({ path: file });
      return file;
    },
    // Main navigation in the top bar: dashboard, repositories, identities, workspaces, doctor, settings.
    // A tab that does not fit the window is reached through the "More" menu.
    async nav(pageId) {
      const hidden = await page.$eval(`#appbar [data-nav="${pageId}"]`, (el) => el.hidden).catch(() => false);
      if (!hidden) {
        await ctx.click(`#appbar [data-nav="${pageId}"]`);
        return;
      }
      await ctx.click("#navMore");
      const label = { dashboard: "Dashboard", repositories: "Repositories", identities: "Identities", workspaces: "Workspaces", doctor: "Doctor" }[pageId];
      await ctx.menuItem(label);
    },
    async waitDialog(id) { await page.waitForSelector(`#${id}[open]`, { timeout: 10000 }); await ctx.settle(200); },
    async submitDialog(id) { await ctx.click(`#${id} [data-dialog-submit]`); },
    async cancelDialog(id) { await ctx.click(`#${id} [data-dialog-cancel]`); await ctx.waitDialogClosed(id); },
    // Dialogs are removed from the page when they close.
    async waitDialogClosed(id, timeout = 30000) { await page.waitForFunction((i) => !document.getElementById(i), { timeout }, id); await ctx.settle(250); },
    async dialogError(id, timeout = 30000) {
      await page.waitForFunction((i) => { const box = document.querySelector(`#${i} [data-dialog-error]`); return box && !box.hidden; }, { timeout }, id);
      return ctx.text(`#${id} [data-dialog-error]`);
    },
    async menuItem(label) { await ctx.clickText(label, ".popover .menu-item"); },
    // What the app does when its window comes back to the front.
    async focusWindow() { await page.evaluate(() => { lastFocusRefresh = 0; window.dispatchEvent(new Event("focus")); }); await ctx.settle(150); },
    async waitGone(text, timeout = 20000) { await page.waitForFunction((t) => !document.body.innerText.includes(t), { timeout }, text); await ctx.settle(150); },
    async text(selector = "body") { return page.$eval(selector, (el) => el.innerText); },
    async click(selector) { await page.waitForSelector(selector, { visible: true, timeout: 8000 }); await page.click(selector); await ctx.settle(); },
    async clickText(text, scope = "button") {
      const handle = await page.evaluateHandle((t, s) => [...document.querySelectorAll(s)].find((el) => el.offsetParent !== null && el.innerText.trim().includes(t)), text, scope);
      const el = handle.asElement();
      if (!el) throw new Error(`no visible ${scope} with text "${text}"`);
      await el.click(); await ctx.settle();
    },
    async type(selector, value) { await page.waitForSelector(selector, { visible: true, timeout: 8000 }); await page.click(selector, { clickCount: 3 }); await page.keyboard.press("Backspace"); await page.type(selector, value); },
    async waitText(text, timeout = 20000) { await page.waitForFunction((t) => document.body.innerText.includes(t), { timeout }, text); await ctx.settle(150); },
    async call(name, ...args) { return page.evaluate((n, a) => window.go.main.App[n](...a), name, args); },
    // Creates a bare repository reachable as git@github.com-<user>:<owner>/<name>.git, with one commit.
    addGitHubRepo(githubUser, owner, name, files = { "README.md": "line one\nline two\nline three\n" }) {
      const bare = path.join(fakeRoot, "hosts", "github.com-" + githubUser, owner, name + ".git");
      fs.mkdirSync(path.dirname(bare), { recursive: true });
      ctx.git(fakeRoot, "init", "-q", "--bare", "-b", "main", bare);
      const work = path.join(fakeRoot, "seed-" + githubUser + "-" + name);
      fs.mkdirSync(work, { recursive: true });
      ctx.git(work, "init", "-q", "-b", "main");
      for (const [file, content] of Object.entries(files)) {
        fs.mkdirSync(path.dirname(path.join(work, file)), { recursive: true });
        fs.writeFileSync(path.join(work, file), content);
      }
      ctx.git(work, "add", "-A");
      ctx.git(work, "-c", "user.name=Seed User", "-c", "user.email=seed@example.com", "commit", "-q", "-m", "initial commit");
      ctx.git(work, "-c", "core.hooksPath=/dev/null", "push", "-q", bare, "main");
      return bare;
    },
    // Sets up bgit with identities "work" (active) and "home" without the wizard.
    async seedIdentities() {
      await ctx.call("RunSetup");
      await ctx.call("AddIdentity", { alias: "work", name: "Work Dev", email: "dev@work.example", githubUsername: "work-gh", sshKeyPath: "", generateSSHKey: true });
      await ctx.call("AddIdentity", { alias: "home", name: "Home Dev", email: "dev@home.example", githubUsername: "home-gh", sshKeyPath: "", generateSSHKey: true });
    },
    write(file, content) { fs.mkdirSync(path.dirname(file), { recursive: true }); fs.writeFileSync(file, content); },
    git(dir, ...args) { return execFileSync("git", args, { cwd: dir, env: { ...process.env, HOME: home, GIT_CONFIG_GLOBAL: path.join(home, ".gitconfig"), GIT_CONFIG_NOSYSTEM: "1", GIT_TERMINAL_PROMPT: "0", GIT_SSH_COMMAND: fakeSsh } }).toString().trim(); },
    async stop() { await browser.close(); server.kill(); },
  };
  return ctx;
}

function expect(condition, message) {
  if (!condition) throw new Error("EXPECT FAILED: " + message);
  console.log("  ok: " + message);
}

// Runs one scenario and exits non-zero when it fails.
function run(scenario, options) {
  (async () => {
    const ui = await start(options);
    try {
      await scenario(ui);
      const problems = ui.problems.filter((problem) => !problem.includes(".png") && !problem.includes("404"));
      if (problems.length > 0) {
        throw new Error("page reported problems:\n" + problems.join("\n"));
      }
    } finally {
      await ui.stop();
    }
  })().catch((error) => {
    console.error(error);
    process.exit(1);
  });
}

module.exports = { start, run, expect, REPO, SHOTS };
