import { createServer } from "node:http";
import { readFile } from "node:fs/promises";
import { extname, join, normalize } from "node:path";

// Static server for `wails dev`. The frontend has no build step: it is plain
// HTML, CSS, and JS served straight from dist/.
const root = "dist";
const types = {
  ".html": "text/html; charset=utf-8",
  ".css": "text/css; charset=utf-8",
  ".js": "application/javascript; charset=utf-8",
  ".svg": "image/svg+xml",
  ".png": "image/png",
};

const server = createServer(async (request, response) => {
  const urlPath = decodeURIComponent(new URL(request.url, "http://localhost").pathname);
  const relative = normalize(urlPath === "/" ? "index.html" : urlPath.slice(1));
  if (relative.startsWith("..")) {
    response.writeHead(403);
    response.end();
    return;
  }
  try {
    const body = await readFile(join(root, relative));
    response.writeHead(200, {
      "content-type": types[extname(relative)] || "application/octet-stream",
      "cache-control": "no-store",
    });
    response.end(body);
  } catch {
    response.writeHead(404);
    response.end("not found");
  }
});

server.listen(34115, "127.0.0.1", () => {
  console.log("bgit desktop frontend listening on http://127.0.0.1:34115");
});
