<template>
  <footer class="finder-status" aria-label="Pagination">
    <span role="status">{{
      selected
        ? t("finder.selected", { count: selected, total })
        : t("finder.items", { count: total })
    }}</span>
    <div v-if="total > pageSize" class="finder-pagination">
      <button
        class="finder-tool"
        :disabled="page === 1"
        :aria-label="t('finder.previousPage')"
        @click="go(page - 1)"
      >
        <FinderIcon name="back" />
      </button>
      <span>{{ t("finder.page", { page, total: pages }) }}</span>
      <button
        class="finder-tool"
        :disabled="page === pages"
        :aria-label="t('finder.nextPage')"
        @click="go(page + 1)"
      >
        <FinderIcon name="forward" />
      </button>
    </div>
    <span class="finder-status-name">File Browser</span>
  </footer>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import FinderIcon from "@/components/FinderIcon.vue";
import { listPage, pageSize } from "@/utils/finder";

const props = defineProps<{
  total: number;
  selected?: number;
  queryKey?: string;
}>();
const route = useRoute();
const router = useRouter();
const { t } = useI18n();
const pages = computed(() => Math.max(1, Math.ceil(props.total / pageSize)));
const page = computed(() =>
  listPage(route.query[props.queryKey || "page"], props.total)
);
const go = (value: number) =>
  router.push({
    path: route.path,
    query: {
      ...route.query,
      [props.queryKey || "page"]: value === 1 ? undefined : String(value),
    },
  });
</script>
