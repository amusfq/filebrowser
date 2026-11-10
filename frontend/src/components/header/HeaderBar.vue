<template>
  <header class="finder-toolbar" @keydown.esc="layoutStore.closeHovers()">
    <button
      v-if="showMenu"
      class="finder-tool menu-button"
      :aria-label="t('buttons.toggleSidebar')"
      :title="t('buttons.toggleSidebar')"
      @click="layoutStore.showHover('sidebar')"
    >
      <FinderIcon name="sidebar" />
    </button>
    <div v-if="showLogo" class="finder-history">
      <button
        class="finder-tool"
        :aria-label="t('finder.back')"
        :title="t('finder.back')"
        :disabled="!canGoBack"
        @click="router.back()"
      >
        <FinderIcon name="back" />
      </button>
      <button
        class="finder-tool"
        :aria-label="t('finder.forward')"
        :title="t('finder.forward')"
        :disabled="!canGoForward"
        @click="router.forward()"
      >
        <FinderIcon name="forward" />
      </button>
    </div>
    <div v-if="showLogo" class="finder-window-title">
      <strong>{{ windowTitle }}</strong>
      <span>{{
        isShare
          ? t("finder.shared")
          : fileStore.isFiles
            ? t("finder.location")
            : t("sidebar.settings")
      }}</span>
    </div>
    <slot />

    <div
      id="dropdown"
      v-if="ifActionsSlot"
      :class="{ active: layoutStore.currentPromptName === 'more' }"
    >
      <slot name="actions" />
    </div>

    <button
      v-if="ifActionsSlot"
      id="more"
      class="finder-tool"
      :aria-label="t('buttons.more')"
      :title="t('buttons.more')"
      :aria-expanded="layoutStore.currentPromptName === 'more'"
      @click="
        layoutStore.currentPromptName === 'more'
          ? layoutStore.closeHovers()
          : layoutStore.showHover('more')
      "
    >
      <FinderIcon name="more" />
    </button>

    <div
      class="overlay"
      v-show="layoutStore.currentPromptName == 'more'"
      @click="layoutStore.closeHovers"
    />
  </header>
</template>

<script setup lang="ts">
import { useLayoutStore } from "@/stores/layout";
import { useFileStore } from "@/stores/file";
import FinderIcon from "@/components/FinderIcon.vue";
import { computed, useSlots } from "vue";
import { useI18n } from "vue-i18n";
import { useRoute, useRouter } from "vue-router";

defineProps<{
  showLogo?: boolean;
  showMenu?: boolean;
}>();

const layoutStore = useLayoutStore();
const fileStore = useFileStore();
const route = useRoute();
const router = useRouter();
const slots = useSlots();

const { t } = useI18n();

const ifActionsSlot = computed(() => (slots.actions ? true : false));
const isShare = computed(() => route.path.startsWith("/share/"));
const windowTitle = computed(() =>
  isShare.value
    ? fileStore.req?.name || t("finder.shared")
    : fileStore.isFiles
      ? fileStore.req?.name || t("sidebar.myFiles")
      : t("sidebar.settings")
);
const canGoBack = computed(() => {
  void route.fullPath;
  return Boolean(window.history.state?.back);
});
const canGoForward = computed(() => {
  void route.fullPath;
  return Boolean(window.history.state?.forward);
});
</script>

<style></style>
