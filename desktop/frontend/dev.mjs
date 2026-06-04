import { createServer } from "node:http";
import { readFileSync } from "node:fs";

const html = readFileSync("dist/index.html");
const server = createServer((request, response) => {
  response.writeHead(200, { "content-type": "text/html; charset=utf-8" });
  response.end(html);
});

server.listen(34115, "127.0.0.1", () => {
  console.log("bgit desktop frontend listening on http://127.0.0.1:34115");
});
