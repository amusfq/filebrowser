<template>
  <div class="finder-columns">
    <section
      v-if="parentPath"
      class="finder-column finder-parent-column"
      :aria-label="t('finder.parentFolder')"
    >
      <p v-if="parentError" role="alert">{{ parentError }}</p>
      <router-link
        v-for="folder in parent?.items.filter((item) => item.isDir)"
        :key="folder.url"
        :to="{ path: folder.url, query: { ...route.query, page: undefined } }"
        :class="{
          'finder-parent-selected': folder.path === fileStore.req?.path,
        }"
      >
        <FinderFileIcon :isDir="true" :name="folder.name" :type="folder.type" />
        <span>{{ folder.name }}</span
        ><FinderIcon name="forward" />
      </router-link>
    </section>
    <section
      class="finder-column finder-current-column"
      :aria-label="fileStore.req?.name || t('sidebar.myFiles')"
    >
      <Item v-for="item in items" :key="item.url" v-bind="item" />
    </section>
    <section
      v-if="selected"
      class="finder-column finder-column-preview"
      :aria-label="t('buttons.info')"
    >
      <FinderFileIcon
        :isDir="selected.isDir"
        :name="selected.name"
        :type="selected.type"
      />
      <strong>{{ selected.name }}</strong>
      <span>{{
        selected.isDir ? t("files.folders") : filesize(selected.size)
      }}</span>
      <span>{{ dayjs(selected.modified).format("LL") }}</span>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch, onBeforeUnmount } from "vue";
import { useRoute } from "vue-router";
import { useI18n } from "vue-i18n";
import dayjs from "dayjs";
import { files as api } from "@/api";
import { StatusError } from "@/api/utils";
import { filesize } from "@/utils";
import { useFileStore } from "@/stores/file";
import FinderIcon from "@/components/FinderIcon.vue";
import FinderFileIcon from "./FinderFileIcon.vue";
import Item from "./ListingItem.vue";

defineProps<{ items: ResourceItem[] }>();
const route = useRoute();
const fileStore = useFileStore();
const { t } = useI18n();
const parent = ref<Resource | null>(null);
const parentError = ref("");
let controller = new AbortController();
const parentPath = computed(() => {
  const path = route.path.replace(/\/$/, "");
  return path === "/files" ? "" : path.slice(0, path.lastIndexOf("/") + 1);
});
const selected = computed(() => fileStore.req?.items[fileStore.selected[0]]);
watch(
  parentPath,
  async (path) => {
    controller.abort();
    controller = new AbortController();
    parent.value = null;
    parentError.value = "";
    if (!path) return;
    try {
      parent.value = await api.fetch(path, controller.signal);
    } catch (error) {
      if (error instanceof StatusError && error.is_canceled) return;
      parentError.value = t("finder.parentUnavailable");
    }
  },
  { immediate: true }
);
onBeforeUnmount(() => controller.abort());
</script>
