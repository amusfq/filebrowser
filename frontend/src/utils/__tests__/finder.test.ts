import { describe, expect, it } from "vitest";
import { finderView, listPage, listQuery, settingsLocation } from "../finder";

describe("Finder URL state", () => {
  it("restores independent view and settings tabs from the URL", () => {
    expect(finderView({ view: "gallery", tab: "users" })).toBe(
      "mosaic gallery"
    );
    expect(finderView({ view: "unknown" }, "list")).toBe("list");
    expect(finderView({ view: "columns" })).toBe("columns");
    expect(
      settingsLocation("/settings/profile", { tab: "users", q: "kept" })
    ).toEqual({
      path: "/settings/users",
      query: { tab: "users", q: "kept" },
      replace: true,
    });
    expect(
      settingsLocation("/settings/users/12", { tab: "users" })
    ).toBeUndefined();
    expect(
      settingsLocation("/settings/users", { tab: "users" })
    ).toBeUndefined();
  });

  it("validates and clamps the selected page", () => {
    expect(listPage("2", 250)).toBe(2);
    expect(listPage("999", 250)).toBe(3);
    expect(listPage("2", 0)).toBe(1);
    for (const invalid of ["0", "-1", "1.5", "abc", ["2"], null, undefined]) {
      expect(listPage(invalid, 250)).toBe(1);
    }
  });

  it("resets the page while retaining the active tabs and unrelated parameters", () => {
    expect(
      listQuery(
        { page: "3", view: "list", tab: "shares", custom: "keep" },
        { q: "new" }
      )
    ).toEqual({
      page: undefined,
      view: "list",
      tab: "shares",
      custom: "keep",
      q: "new",
    });
  });
});
