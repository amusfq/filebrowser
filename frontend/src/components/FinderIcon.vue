<template>
  <svg
    class="finder-icon"
    width="20"
    height="20"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    stroke-width="1.6"
    stroke-linecap="round"
    stroke-linejoin="round"
    aria-hidden="true"
  >
    <path :d="paths[name] || paths.more" />
  </svg>
</template>

<script setup lang="ts">
defineProps<{ name: string }>();

const paths: Record<string, string> = {
  back: "m14 6-6 6 6 6",
  forward: "m10 6 6 6-6 6",
  up: "m6 14 6-6 6 6",
  sidebar: "M4 4h16v16H4z M9 4v16 M6.5 8h0 M6.5 12h0 M6.5 16h0",
  folder: "M3 7V5h7l2 2h9v12H3z",
  home: "m3 11 9-8 9 8 M5 10v11h5v-7h4v7h5V10",
  grid: "M4 4h6v6H4z M14 4h6v6h-6z M4 14h6v6H4z M14 14h6v6h-6z",
  list: "M8 6h12 M8 12h12 M8 18h12 M4 6h0 M4 12h0 M4 18h0",
  columns: "M3 4h18v16H3z M9 4v16 M15 4v16",
  gallery: "M3 4h18v12H3z M3 20h4 M10 20h4 M17 20h4",
  share: "M12 15V3 m-4 4 4-4 4 4 M7 11H4v10h16V11h-3",
  upload: "M12 16V3 m-4 4 4-4 4 4 M4 16v5h16v-5",
  download: "M12 3v13 m-4-4 4 4 4-4 M4 17v4h16v-4",
  newFolder: "M3 7V5h7l2 2h9v12H3z M12 10v6 M9 13h6",
  newFile: "M14 3H5v18h14V8z M14 3v5h5 M12 11v6 M9 14h6",
  search: "M16 10a6 6 0 1 1-12 0 6 6 0 0 1 12 0 m-2 4 6 6",
  more: "M5 12h0 M12 12h0 M19 12h0",
  info: "M12 22a10 10 0 1 0 0-20 10 10 0 0 0 0 20 M12 11v6 M12 7h0",
  settings:
    "M12 8a4 4 0 1 0 0 8 4 4 0 0 0 0-8 M10 3h4l1 3 3-1 2 3-2 3 2 3-2 3-3-1-1 3h-4l-1-3-3 1-2-3 2-3-2-3 2-3 3 1z",
  user: "M16 7a4 4 0 1 1-8 0 4 4 0 0 1 8 0 M4 21v-2a8 8 0 0 1 16 0v2",
  logout: "M9 3H4v18h5 M10 12h11 m-4-4 4 4-4 4",
  server: "M4 3h16v7H4z M4 14h16v7H4z M7 6.5h0 M7 17.5h0 M11 6.5h6 M11 17.5h6",
  help: "M12 22a10 10 0 1 0 0-20 10 10 0 0 0 0 20 M9 8a3 3 0 1 1 4 3v3 M12 17h0",
};
</script>
