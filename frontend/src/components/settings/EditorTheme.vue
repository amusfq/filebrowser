<template>
  <select name="editorTheme" v-on:change="change" :value="aceEditorTheme">
    <option
      v-for="theme in editorThemes"
      :value="theme.value"
      :key="theme.value"
    >
      {{ theme.label }}
    </option>
  </select>
</template>

<script setup lang="ts">
import { type SelectHTMLAttributes } from "vue";
import { editorThemes } from "@/utils/editor";

defineProps<{
  aceEditorTheme: string;
}>();

const emit = defineEmits<{
  (e: "update:aceEditorTheme", val: string | null): void;
}>();

const change = (event: Event) => {
  emit("update:aceEditorTheme", (event.target as SelectHTMLAttributes)?.value);
};
</script>
