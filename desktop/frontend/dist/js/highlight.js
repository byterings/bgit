// A small syntax highlighter for diff lines. It works one line at a time and
// only knows comments, strings, numbers, keywords, and calls, which is enough
// to make a diff readable without shipping a full grammar per language.

const HIGHLIGHT_FAMILIES = {
  c: ["js", "jsx", "mjs", "cjs", "ts", "tsx", "go", "java", "c", "h", "cc", "cpp", "hpp", "cs", "rs", "php", "swift", "kt", "kts", "scala", "dart", "json", "css", "scss", "less"],
  hash: ["py", "rb", "sh", "bash", "zsh", "yml", "yaml", "toml", "ini", "conf", "pl", "r", "tf", "dockerfile", "makefile", "mk", "gitignore", "env"],
};

const HIGHLIGHT_KEYWORDS = new Set(
  (
    "abstract as async await break case catch class const continue def default defer del do elif else end enum export extends " +
    "false final finally fn for from func function go if impl implements import in interface is lambda let match mod module " +
    "new nil None not null of or package pass private protected pub public raise return select self static struct super switch " +
    "then this throw trait true True False try type typeof undefined unless until use val var void when where while with yield " +
    "and begin chan elsif ensure fallthrough goto map range rescue require union unsafe"
  ).split(" "),
);

const HIGHLIGHT_C =
  /(?<comment>\/\/.*$|\/\*.*?\*\/)|(?<string>"(?:\\.|[^"\\])*"|'(?:\\.|[^'\\])*'|`(?:\\.|[^`\\])*`)|(?<number>\b\d[\d_.]*\b)|(?<word>[A-Za-z_$][\w$]*)(?<call>(?=\s*\())?/g;
const HIGHLIGHT_HASH =
  /(?<comment>#.*$)|(?<string>"(?:\\.|[^"\\])*"|'(?:\\.|[^'\\])*')|(?<number>\b\d[\d_.]*\b)|(?<word>[A-Za-z_][\w]*)(?<call>(?=\s*\())?/g;

function highlightFamily(path) {
  const name = String(path || "").split("/").pop().toLowerCase();
  const extension = name.includes(".") ? name.split(".").pop() : name;
  if (HIGHLIGHT_FAMILIES.c.includes(extension)) return "c";
  if (HIGHLIGHT_FAMILIES.hash.includes(extension)) return "hash";
  return "";
}

// Returns escaped HTML for one line of code.
function highlightCode(text, family) {
  if (!family || text.length > 400) {
    return escapeHtml(text);
  }
  const pattern = family === "hash" ? HIGHLIGHT_HASH : HIGHLIGHT_C;
  let html = "";
  let last = 0;
  for (const match of text.matchAll(pattern)) {
    const groups = match.groups;
    let kind = "";
    if (groups.comment !== undefined) kind = "comment";
    else if (groups.string !== undefined) kind = "string";
    else if (groups.number !== undefined) kind = "number";
    else if (HIGHLIGHT_KEYWORDS.has(groups.word)) kind = "keyword";
    else if (groups.call !== undefined) kind = "title";
    html += escapeHtml(text.slice(last, match.index));
    html += kind ? `<span class="tok-${kind}">${escapeHtml(match[0])}</span>` : escapeHtml(match[0]);
    last = match.index + match[0].length;
  }
  return html + escapeHtml(text.slice(last));
}
