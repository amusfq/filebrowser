import { describe, expect, it } from "vitest";
import { editorLanguage, editorTheme } from "../editor";

describe("Monaco language detection", () => {
  it.each([
    ["app.TSX", "tsx"],
    ["src/main.rs", "rust"],
    ["main.go", "go"],
    ["Dockerfile", "dockerfile"],
    ["Dockerfile.production", "dockerfile"],
    [".env.local", "ini"],
    [".bashrc", "bash"],
    ["config.yml", "yaml"],
    ["App.vue", "vue"],
    ["notes.unknown", "plaintext"],
    ["README", "plaintext"],
  ])("detects %s as %s", (name, language) => {
    expect(editorLanguage(name)).toBe(language);
  });
});

describe("editor theme compatibility", () => {
  it("follows the app theme for automatic and unknown preferences", () => {
    expect(editorTheme("", true)).toBe("github-dark");
    expect(editorTheme("", false)).toBe("github-light");
    expect(editorTheme("unknown", true)).toBe("github-dark");
  });
  it("retains supported themes and maps saved Ace preferences", () => {
    expect(editorTheme("monokai", false)).toBe("monokai");
    expect(editorTheme("ace/theme/chrome", true)).toBe("github-light");
    expect(editorTheme("ace/theme/twilight", false)).toBe("github-dark");
    expect(editorTheme("solarized_dark", false)).toBe("solarized-dark");
  });
});
