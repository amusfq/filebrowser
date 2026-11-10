export const editorThemes = [
  { value: "", label: "System" },
  { value: "github-light", label: "GitHub Light" },
  { value: "github-dark", label: "GitHub Dark" },
  { value: "monokai", label: "Monokai" },
  { value: "solarized-light", label: "Solarized Light" },
  { value: "solarized-dark", label: "Solarized Dark" },
];

export function editorTheme(preference: string, dark: boolean) {
  const name = preference.replace(/^ace\/theme\//, "");
  if (name && editorThemes.some((theme) => theme.value === name)) return name;
  if (
    [
      "chrome",
      "clouds",
      "dawn",
      "github",
      "textmate",
      "tomorrow",
      "xcode",
    ].includes(name)
  )
    return "github-light";
  if (name === "solarized_light") return "solarized-light";
  if (name === "solarized_dark") return "solarized-dark";
  if (
    ["twilight", "tomorrow_night", "tomorrow_night_blue", "dracula"].includes(
      name
    )
  )
    return "github-dark";
  return dark ? "github-dark" : "github-light";
}

export const editorLanguages = {
  javascript: ["js", "mjs", "cjs"],
  jsx: ["jsx"],
  typescript: ["ts", "mts", "cts"],
  tsx: ["tsx"],
  json: ["json", "jsonc", "jsonl", "webmanifest"],
  html: ["html", "htm"],
  css: ["css"],
  scss: ["scss"],
  markdown: ["md", "markdown", "mdown"],
  yaml: ["yaml", "yml"],
  bash: ["sh", "bash", "zsh", "bashrc", "zshrc"],
  python: ["py", "pyw"],
  go: ["go"],
  rust: ["rs"],
  sql: ["sql"],
  vue: ["vue"],
  svelte: ["svelte"],
  php: ["php"],
  java: ["java"],
  c: ["c", "h"],
  cpp: ["cpp", "cc", "cxx", "hpp"],
  csharp: ["cs"],
  ruby: ["rb", "rake"],
  toml: ["toml"],
  ini: ["ini", "conf", "cfg", "env"],
  dockerfile: ["dockerfile"],
  xml: ["xml", "svg", "xsl"],
  diff: ["diff", "patch"],
};

export function editorLanguage(filename: string) {
  const name = filename.split("/").pop()?.toLowerCase() || "";
  if (name === "dockerfile" || name.startsWith("dockerfile."))
    return "dockerfile";
  if (name === "makefile") return "plaintext";
  if (name === ".env" || name.startsWith(".env.")) return "ini";
  const extension = name.split(".").pop() || "";
  return (
    Object.entries(editorLanguages).find(([, extensions]) =>
      extensions.includes(extension)
    )?.[0] || "plaintext"
  );
}
