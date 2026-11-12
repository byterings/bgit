// Runs every *.test.js in this folder, one after another, and reports which
// ones failed. Each test starts its own harness on its own isolated home.
const { spawnSync } = require("node:child_process");
const fs = require("node:fs");
const path = require("node:path");

const tests = fs.readdirSync(__dirname).filter((file) => file.endsWith(".test.js")).sort();
const failed = [];

for (const test of tests) {
  console.log(`\n== ${test}`);
  const result = spawnSync(process.execPath, [path.join(__dirname, test)], { stdio: "inherit" });
  if (result.status !== 0) {
    failed.push(test);
  }
}

console.log(failed.length === 0 ? `\nAll ${tests.length} UI tests passed.` : `\nFailed: ${failed.join(", ")}`);
process.exit(failed.length === 0 ? 0 : 1);
