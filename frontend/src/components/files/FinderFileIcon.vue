<template>
  <svg
    v-if="isDir"
    class="finder-file-icon finder-folder"
    viewBox="0 0 80 64"
    fill="none"
    aria-hidden="true"
  >
    <path
      d="M3 12a5 5 0 0 1 5-5h21l7 7h36a5 5 0 0 1 5 5v36H3z"
      fill="#279bd5"
    />
    <path
      d="M3 22a4 4 0 0 1 4-4h66a4 4 0 0 1 4 4v33a5 5 0 0 1-5 5H8a5 5 0 0 1-5-5z"
      fill="#74cef4"
    />
    <path d="M7 19h66" stroke="#b5ecff" stroke-width="2" />
    <path d="M4 55a4 4 0 0 0 4 4h64a4 4 0 0 0 4-4" stroke="#40a9da" />
  </svg>
  <svg
    v-else
    class="finder-file-icon finder-document"
    :data-kind="type"
    viewBox="0 0 64 72"
    fill="none"
    aria-hidden="true"
  >
    <path
      d="M12 3h27l14 14v49a3 3 0 0 1-3 3H12a3 3 0 0 1-3-3V6a3 3 0 0 1 3-3Z"
      fill="var(--surfacePrimary)"
      stroke="#b8bcc4"
    />
    <path
      d="M39 3v14h14"
      fill="#e9ebef"
      stroke="#b8bcc4"
      stroke-linejoin="round"
    />
    <rect x="4" y="34" width="46" height="22" rx="3" fill="currentColor" />
    <text
      x="27"
      y="49"
      text-anchor="middle"
      fill="white"
      font-size="11"
      font-family="-apple-system, BlinkMacSystemFont, sans-serif"
      font-weight="600"
    >
      {{ label }}
    </text>
  </svg>
</template>

<script setup lang="ts">
import { computed } from "vue";
const props = defineProps<{ isDir: boolean; name: string; type: string }>();
const label = computed(() => {
  const extension = props.name.includes(".")
    ? props.name.split(".").pop()
    : "FILE";
  return extension?.slice(0, 4).toUpperCase();
});
</script>
