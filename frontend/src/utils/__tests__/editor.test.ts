import { describe, expect, it } from "vitest";
import { editorLanguage, editorTheme, sharedEditPath } from "../editor";

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

describe("shared-file edit links", () => {
  const shares = [{ hash: "owned", path: "/My Files", userID: 7 }];
  it("opens the owning user's file with special characters encoded", () => {
    expect(sharedEditPath(shares, "owned", 7, "/notes #1.ts")).toBe(
      "/files/My%20Files/notes%20%231.ts"
    );
  });
  it("supports single-file shares and the shared root folder", () => {
    expect(
      sharedEditPath(
        [{ hash: "file", path: "/app.ts", userID: 7 }],
        "file",
        7,
        ""
      )
    ).toBe("/files/app.ts");
    expect(
      sharedEditPath(
        [{ hash: "root", path: "/", userID: 7 }],
        "root",
        7,
        "/app.ts"
      )
    ).toBe("/files/app.ts");
  });
  it("does not map shares belonging to another user, including in admin results", () => {
    expect(sharedEditPath(shares, "owned", 8, "/app.ts")).toBe("");
    expect(sharedEditPath(shares, "missing", 7, "/app.ts")).toBe("");
  });
});
