import type { LocationQuery } from "vue-router";

export const finderViews = [
  { key: "icons", mode: "mosaic", icon: "grid", label: "Icons" },
  { key: "list", mode: "list", icon: "list", label: "List" },
  { key: "columns", mode: "columns", icon: "columns", label: "Columns" },
  { key: "gallery", mode: "mosaic gallery", icon: "gallery", label: "Gallery" },
] as const;

export const pageSize = 100;

export function finderView(
  query: LocationQuery,
  fallback: ViewModeType = "mosaic"
) {
  return finderViews.find((view) => view.key === query.view)?.mode || fallback;
}

export function listPage(value: LocationQuery[string], total: number) {
  const page =
    typeof value === "string" && /^[1-9]\d*$/.test(value) ? Number(value) : 1;
  return Math.min(page, Math.max(1, Math.ceil(total / pageSize)));
}

export function listQuery(query: LocationQuery, changes: LocationQuery) {
  return { ...query, page: undefined, ...changes };
}

export function settingsLocation(path: string, query: LocationQuery) {
  const match = path.match(/^\/settings\/(profile|shares|global|users)\/?$/);
  if (!match) return;
  const tab =
    typeof query.tab === "string" &&
    ["profile", "shares", "global", "users"].includes(query.tab)
      ? query.tab
      : match[1];
  const target = `/settings/${tab}`;
  if (path !== target || query.tab !== tab)
    return { path: target, query: { ...query, tab }, replace: true };
}
