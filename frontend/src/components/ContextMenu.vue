<template>
  <div
    class="context-menu"
    ref="contextMenu"
    v-show="show"
    :style="{
      top: `${top}px`,
      left: `${left}px`,
    }"
  >
    <slot />
  </div>
</template>

<script setup lang="ts">
import { ref, watch, computed, onUnmounted } from "vue";
import { useElementBounding, useWindowSize } from "@vueuse/core";

const emit = defineEmits(["hide"]);
const props = defineProps<{ show: boolean; pos: { x: number; y: number } }>();
const contextMenu = ref<HTMLElement | null>(null);
const { width, height } = useElementBounding(contextMenu);
const { width: viewportWidth, height: viewportHeight } = useWindowSize();

const top = computed(() => {
  const y =
    props.pos.y + height.value > viewportHeight.value - 8
      ? props.pos.y - height.value
      : props.pos.y;
  return Math.max(8, Math.min(y, viewportHeight.value - height.value - 8));
});

const left = computed(() => {
  return Math.max(
    8,
    Math.min(props.pos.x, viewportWidth.value - width.value - 8)
  );
});

const hideContextMenu = () => {
  emit("hide");
};

watch(
  () => props.show,
  (val) => {
    if (val) {
      document.addEventListener("click", hideContextMenu);
    } else {
      document.removeEventListener("click", hideContextMenu);
    }
  }
);

onUnmounted(() => {
  document.removeEventListener("click", hideContextMenu);
});
</script>
