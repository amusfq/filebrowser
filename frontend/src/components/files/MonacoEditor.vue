<template>
  <div class="monaco-container">
    <div v-if="loading" class="monaco-message" role="status">
      {{ t("files.loading") }}
    </div>
    <div v-if="error" class="monaco-message" role="alert">{{ error }}</div>
    <div ref="container" class="monaco-host"></div>
  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import { editorLanguage } from "@/utils/editor";
import { monaco, prepareMonaco } from "@/utils/monaco";

const props = withDefaults(
  defineProps<{
    modelValue: string;
    filename: string;
    readOnly?: boolean;
    fontSize?: number;
    theme?: string;
  }>(),
  { readOnly: false, fontSize: 14, theme: "github-light" }
);
const emit = defineEmits<{
  "update:modelValue": [value: string];
  ready: [editor: monaco.editor.IStandaloneCodeEditor];
}>();
const { t } = useI18n();
const container = ref<HTMLDivElement>();
const loading = ref(true);
const error = ref("");
let disposed = false;
let editor: monaco.editor.IStandaloneCodeEditor | undefined;
let model: monaco.editor.ITextModel | undefined;

onMounted(async () => {
  try {
    await prepareMonaco();
    if (disposed || !container.value) return;
    model = monaco.editor.createModel(
      props.modelValue,
      editorLanguage(props.filename)
    );
    editor = monaco.editor.create(container.value, {
      model,
      theme: props.theme,
      readOnly: props.readOnly,
      domReadOnly: props.readOnly,
      fontSize: props.fontSize,
      fontFamily: "SFMono-Regular, Menlo, Monaco, Consolas, monospace",
      automaticLayout: true,
      minimap: { enabled: false },
      wordWrap: "on",
      scrollBeyondLastLine: false,
      padding: { top: 16, bottom: 16 },
      ariaLabel: props.filename,
    });
    editor.onDidChangeModelContent(() =>
      emit("update:modelValue", editor!.getValue())
    );
    emit("ready", editor);
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : String(cause);
  } finally {
    loading.value = false;
  }
});

watch(
  () => props.modelValue,
  (value) => {
    if (model && value !== model.getValue()) model.setValue(value);
  }
);
watch(
  () => props.filename,
  (name) => {
    if (model) monaco.editor.setModelLanguage(model, editorLanguage(name));
  }
);
watch(
  () => [props.readOnly, props.fontSize] as const,
  ([readOnly, fontSize]) => {
    editor?.updateOptions({ readOnly, domReadOnly: readOnly, fontSize });
  }
);
watch(
  () => props.theme,
  (theme) => {
    if (editor) monaco.editor.setTheme(theme);
  }
);
onBeforeUnmount(() => {
  disposed = true;
  editor?.dispose();
  model?.dispose();
});
</script>

<style scoped>
.monaco-container {
  position: relative;
  width: 100%;
  height: 100%;
  min-height: 240px;
}
.monaco-host {
  position: absolute;
  inset: 0;
}
.monaco-message {
  position: absolute;
  inset: 24px;
  z-index: 1;
  text-align: center;
  overflow-wrap: anywhere;
}
</style>
