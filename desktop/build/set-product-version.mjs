// Writes the release version into desktop/wails.json so installers and app
// metadata carry it. Installer metadata only accepts numeric versions, so a
// pre-release suffix such as "-beta" is dropped.
//
//   node desktop/build/set-product-version.mjs 1.2.3-beta
import { readFileSync, writeFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const version = process.argv[2] || "";
const numeric = version.split("-")[0];
if (!/^\d+\.\d+\.\d+$/.test(numeric)) {
  console.error(`expected a version like 1.2.3 or 1.2.3-beta, got "${version}"`);
  process.exit(1);
}

const configPath = join(dirname(fileURLToPath(import.meta.url)), "..", "wails.json");
const config = JSON.parse(readFileSync(configPath, "utf8"));
config.info = { ...config.info, productVersion: numeric };
writeFileSync(configPath, JSON.stringify(config, null, 2) + "\n");
console.log(`productVersion set to ${numeric}`);
