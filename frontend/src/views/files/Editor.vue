<template>
  <div id="editor-container">
    <header-bar>
      <action icon="close" :label="t('buttons.close')" @action="close()" />
      <title>{{ fileStore.req?.name ?? "" }}</title>

      <action
        icon="add"
        @action="increaseFontSize"
        :label="t('buttons.increaseFontSize')"
      />
      <span class="editor-font-size">{{ fontSize }}px</span>
      <action
        icon="remove"
        @action="decreaseFontSize"
        :label="t('buttons.decreaseFontSize')"
      />

      <button
        v-if="!readOnly"
        :disabled="!editor || saving"
        id="save-button"
        class="editor-save"
        :aria-label="t('buttons.save')"
        @click="save()"
      >
        <i class="material-icons" aria-hidden="true">save</i>
        <span>{{ t("buttons.save") }}</span>
      </button>

      <action
        icon="preview"
        :label="t('buttons.preview')"
        @action="preview()"
        v-show="isMarkdownFile"
      />
    </header-bar>

    <!-- preview container -->
    <div class="loading delayed" v-if="layoutStore.loading">
      <div class="spinner">
        <div class="bounce1"></div>
        <div class="bounce2"></div>
        <div class="bounce3"></div>
      </div>
    </div>
    <template v-else>
      <div class="editor-header">
        <Breadcrumbs base="/files" noLink />

        <div>
          <button
            :disabled="isSelectionEmpty"
            :aria-label="t('buttons.copy')"
            @click="executeEditorCommand('copy')"
          >
            <span><i class="material-icons">content_copy</i></span>
          </button>
          <button
            :disabled="isSelectionEmpty || readOnly"
            :aria-label="t('buttons.cut')"
            @click="executeEditorCommand('cut')"
          >
            <span><i class="material-icons">content_cut</i></span>
          </button>
          <button
            :disabled="readOnly"
            :aria-label="t('buttons.paste')"
            @click="executeEditorCommand('paste')"
          >
            <span><i class="material-icons">content_paste</i></span>
          </button>
          <button
            :aria-label="t('buttons.more')"
            @click="executeEditorCommand('openCommandPalette')"
          >
            <span><i class="material-icons">more_vert</i></span>
          </button>
        </div>
      </div>

      <div
        v-show="isPreview && isMarkdownFile"
        id="preview-container"
        class="md_preview"
        v-html="previewContent"
      ></div>
      <MonacoEditor
        v-show="!isPreview || !isMarkdownFile"
        id="editor"
        v-model="content"
        :filename="fileStore.req?.name || ''"
        :readOnly="readOnly"
        :fontSize="fontSize"
        :theme="getEditorTheme(authStore.user?.aceEditorTheme || '')"
        @ready="editorReady"
      />
    </template>
  </div>
</template>

<script setup lang="ts">
import type { editor as Monaco } from "monaco-editor";
import { files as api } from "@/api";
import buttons from "@/utils/buttons";
import url from "@/utils/url";
import DOMPurify from "dompurify";
import MonacoEditor from "@/components/files/MonacoEditor.vue";
import Breadcrumbs from "@/components/Breadcrumbs.vue";
import Action from "@/components/header/Action.vue";
import HeaderBar from "@/components/header/HeaderBar.vue";
import { useAuthStore } from "@/stores/auth";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";
import { getEditorTheme } from "@/utils/theme";
import { marked } from "marked";
import markedKatex from "marked-katex-extension";
import {
  computed,
  inject,
  onBeforeUnmount,
  onMounted,
  ref,
  shallowRef,
  watch,
} from "vue";
import { useI18n } from "vue-i18n";
import {
  onBeforeRouteLeave,
  onBeforeRouteUpdate,
  useRoute,
  useRouter,
} from "vue-router";
import { read, copy } from "@/utils/clipboard";

const $showError = inject<IToastError>("$showError")!;
const fileStore = useFileStore();
const authStore = useAuthStore();
const layoutStore = useLayoutStore();
const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const editor = shallowRef<Monaco.IStandaloneCodeEditor>();
const content = ref(fileStore.req?.content || "");
const fontSize = ref(
  Math.max(
    8,
    Math.min(48, Number(localStorage.getItem("editorFontSize")) || 14)
  )
);
const readOnly = computed(
  () => fileStore.req?.type === "textImmutable" || !authStore.user?.perm.modify
);
const saving = ref(false);
const isDirty = ref(false);
let savedVersion = 0;
const isPreview = ref(false);
const previewContent = ref("");
const isMarkdownFile = computed(() =>
  /\.(md|markdown)$/i.test(fileStore.req?.name || "")
);
const isSelectionEmpty = ref(true);
marked.use(markedKatex({ output: "mathml", throwOnError: false }));

watch([content, isPreview], async ([value, preview], _, onCleanup) => {
  let cancelled = false;
  onCleanup(() => {
    cancelled = true;
  });
  if (isMarkdownFile.value && preview) {
    try {
      const html = DOMPurify.sanitize(await marked(value));
      if (!cancelled) previewContent.value = html;
    } catch (error) {
      $showError(error instanceof Error ? error : String(error));
    }
  }
});

const editorReady = (instance: Monaco.IStandaloneCodeEditor) => {
  editor.value = instance;
  savedVersion = instance.getModel()!.getAlternativeVersionId();
  instance.onDidChangeModelContent(() => {
    isDirty.value =
      instance.getModel()!.getAlternativeVersionId() !== savedVersion;
  });
  instance.onDidChangeCursorSelection(() => {
    isSelectionEmpty.value = instance.getSelection()?.isEmpty() ?? true;
  });
  instance.focus();
};

const executeEditorCommand = async (name: string) => {
  const instance = editor.value;
  if (!instance) return;
  try {
    if (name === "paste") {
      if (readOnly.value) return;
      const text = await read();
      if (editor.value === instance) {
        instance.executeEdits("clipboard", [
          { range: instance.getSelection()!, text, forceMoveMarkers: true },
        ]);
      }
    } else if (name === "copy" || name === "cut") {
      const selection = instance.getSelection()!;
      const model = instance.getModel()!;
      const version = model.getVersionId();
      await copy({ text: model.getValueInRange(selection) });
      if (
        name === "cut" &&
        !readOnly.value &&
        model.getVersionId() === version
      ) {
        instance.executeEdits("clipboard", [
          { range: selection, text: "", forceMoveMarkers: true },
        ]);
      }
    } else {
      instance.focus();
      await instance.getAction("editor.action.quickCommand")?.run();
      return;
    }
    instance.focus();
  } catch (error) {
    $showError(error instanceof Error ? error : String(error));
  }
};

const save = async () => {
  const model = editor.value?.getModel();
  if (!model || readOnly.value || saving.value) return false;
  const version = model.getAlternativeVersionId();
  const text = model.getValue();
  saving.value = true;
  buttons.loading("save");
  try {
    await api.put(route.path, text);
    if (editor.value?.getModel() === model) {
      savedVersion = version;
      isDirty.value = model.getAlternativeVersionId() !== savedVersion;
      buttons.success("save");
    }
    return true;
  } catch (error) {
    buttons.done("save");
    $showError(error instanceof Error ? error : String(error));
    return false;
  } finally {
    saving.value = false;
  }
};

const confirmNavigation = () => {
  if (!isDirty.value) return true;
  return new Promise<boolean>((resolve) => {
    const leave = () => {
      resolve(true);
      layoutStore.closeHovers();
    };
    layoutStore.showHover({
      prompt: "discardEditorChanges",
      confirm: (event: Event) => {
        event.preventDefault();
        leave();
      },
      saveAction: async () => {
        if ((await save()) && !isDirty.value) leave();
      },
      close: async () => {
        resolve(false);
        return "";
      },
    });
  });
};
onBeforeRouteLeave(confirmNavigation);
onBeforeRouteUpdate((to, from) => {
  if (to.path !== from.path || to.query.edit !== from.query.edit)
    return confirmNavigation();
  return true;
});

const keyEvent = (event: KeyboardEvent) => {
  if (event.defaultPrevented || layoutStore.currentPromptName) return;
  if (event.code === "Escape") close();
  if ((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === "s") {
    event.preventDefault();
    save();
  }
};
const handlePageChange = (event: BeforeUnloadEvent) => {
  if (isDirty.value) {
    event.preventDefault();
    event.returnValue = true;
  }
};
onMounted(() => {
  window.addEventListener("keydown", keyEvent);
  window.addEventListener("beforeunload", handlePageChange);
});
onBeforeUnmount(() => {
  window.removeEventListener("keydown", keyEvent);
  window.removeEventListener("beforeunload", handlePageChange);
  editor.value = undefined;
});
watch(fontSize, (size) => localStorage.setItem("editorFontSize", String(size)));
const increaseFontSize = () => {
  fontSize.value = Math.min(48, fontSize.value + 1);
};
const decreaseFontSize = () => {
  fontSize.value = Math.max(8, fontSize.value - 1);
};
const close = () =>
  router.push({
    path: url.removeLastDir(route.path) + "/",
    query: { ...route.query, edit: undefined },
  });
const preview = () => {
  isPreview.value = !isPreview.value;
};
</script>

<style scoped>
.editor-save {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
  padding: 7px 12px;
  border: 0;
  border-radius: 5px;
  color: white;
  background: var(--blue);
  cursor: pointer;
  font-size: 13px;
}
.editor-save:disabled {
  opacity: 0.5;
  cursor: default;
}
.editor-save i {
  font-size: 18px;
}

.editor-font-size {
  margin: 0 0.5em;
  color: var(--fg);
}

.editor-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.editor-header > div > button {
  background: transparent;
  color: var(--action);
  border: none;
  outline: none;
  opacity: 0.8;
  cursor: pointer;
}

.editor-header > div > button:hover:not(:disabled) {
  opacity: 1;
}

.editor-header > div > button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.editor-header > div > button > span > i {
  font-size: 1.2rem;
}
</style>
