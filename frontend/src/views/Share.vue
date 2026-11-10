<template>
  <div class="finder-share">
    <header-bar showMenu showLogo>
      <action
        v-if="req?.isDir && !error"
        icon="check_circle"
        :label="t('buttons.selectMultiple')"
        @action="toggleMultipleSelection"
      />
      <button
        v-if="req && !error"
        class="finder-tool"
        :aria-label="t('buttons.copyDownloadLinkToClipboard')"
        :title="t('buttons.copyDownloadLinkToClipboard')"
        @click="copyToClipboard(isSingleFile() ? linkSelected() : link)"
      >
        <FinderIcon name="share" />
      </button>
      <action
        v-if="fileStore.selectedCount && !error"
        icon="file_download"
        :label="t('buttons.download')"
        @action="download"
        :counter="fileStore.selectedCount"
      />
    </header-bar>

    <breadcrumbs :base="'/share/' + hash" />

    <div v-if="layoutStore.loading" class="share-loading" role="status">
      <div class="spinner">
        <div class="bounce1"></div>
        <div class="bounce2"></div>
        <div class="bounce3"></div>
      </div>
      <span>{{ t("files.loading") }}</span>
    </div>
    <div v-else-if="error">
      <template v-if="error.status === 401">
        <form
          class="card floating"
          id="password"
          @submit.prevent="fetchData"
          role="dialog"
          aria-modal="true"
          :aria-label="t('login.password')"
        >
          <div
            v-if="attemptedPasswordLogin"
            class="share__wrong__password"
            role="alert"
          >
            {{ t("login.wrongCredentials") }}
          </div>
          <div class="card-title">
            <h2>{{ t("login.password") }}</h2>
          </div>
          <div class="card-content">
            <input
              v-focus
              class="input input--block"
              type="password"
              autocomplete="current-password"
              :aria-label="t('login.password')"
              :placeholder="t('login.password')"
              v-model="password"
            />
          </div>
          <div class="card-action">
            <button type="submit" class="button button--flat">
              {{ t("buttons.submit") }}
            </button>
          </div>
        </form>
        <div class="overlay" />
      </template>
      <errors v-else :errorCode="error.status" />
    </div>
    <div v-else-if="req" class="share-workspace">
      <section
        v-if="req.isDir"
        id="shareList"
        class="share-content"
        :aria-label="t('files.files')"
      >
        <div class="share-section-heading">
          <FinderIcon name="folder" />
          <h1>{{ req.name }}</h1>
        </div>
        <div v-if="req.items.length" id="listing" class="list file-icons">
          <item
            v-for="item in visibleItems"
            :key="item.url"
            :index="item.index"
            :name="item.name"
            :isDir="item.isDir"
            :url="item.url"
            :modified="item.modified"
            :type="item.type"
            :size="item.size"
            readOnly
          />
          <div :class="{ active: fileStore.multiple }" id="multiple-selection">
            <p>{{ t("files.multipleSelectionEnabled") }}</p>
            <button
              class="finder-tool"
              :aria-label="t('buttons.clear')"
              @click="fileStore.multiple = false"
            >
              <FinderIcon name="back" />
            </button>
          </div>
        </div>
        <h2 v-else class="message">
          <FinderIcon name="folder" /><span>{{ t("files.lonely") }}</span>
        </h2>
      </section>
      <section
        v-else
        class="share-content share-quicklook"
        :aria-label="t('sidebar.preview')"
      >
        <div class="share-preview-stage">
          <video
            v-if="req.type === 'video'"
            :key="previewLink"
            :src="previewLink"
            controls
            playsinline
            preload="metadata"
            :aria-label="req.name"
          >
            <a :href="link">{{ t("buttons.download") }}</a>
          </video>
          <img
            v-else-if="req.type === 'image'"
            :src="previewLink"
            :alt="req.name"
          />
          <template v-else>
            <FinderFileIcon :isDir="false" :name="req.name" :type="req.type" />
            <audio
              v-if="req.type === 'audio'"
              :src="previewLink"
              controls
              preload="metadata"
              :aria-label="req.name"
            />
          </template>
        </div>
        <div class="share-preview-caption">
          <h1>{{ req.name }}</h1>
          <p>
            {{ humanSize }} <span aria-hidden="true">·</span> {{ humanTime }}
          </p>
        </div>
      </section>

      <aside class="share-inspector" :aria-label="t('prompts.fileInfo')">
        <div class="share-file-summary">
          <FinderFileIcon
            :isDir="req.isDir"
            :name="req.name"
            :type="req.type"
          />
          <h2>{{ req.name }}</h2>
          <p>
            {{
              req.isDir
                ? t("download.downloadFolder")
                : t("download.downloadFile")
            }}
          </p>
        </div>
        <div class="share-download-actions">
          <a :href="link" class="share-download"
            ><FinderIcon name="download" />{{ t("buttons.download") }}</a
          >
          <a
            v-if="!req.isDir"
            :href="inlineLink"
            target="_blank"
            rel="noopener noreferrer"
            class="share-open"
            >{{ t("buttons.openFile") }}<FinderIcon name="forward"
          /></a>
        </div>
        <dl class="share-metadata">
          <div>
            <dt>{{ t("prompts.size") }}</dt>
            <dd>
              {{
                req.isDir
                  ? t("finder.items", { count: req.items.length })
                  : humanSize
              }}
            </dd>
          </div>
          <div :title="modTime">
            <dt>{{ t("prompts.lastModified") }}</dt>
            <dd>{{ humanTime }}</dd>
          </div>
        </dl>
        <div v-if="selectedItem" class="share-selection-preview">
          <h3>{{ t("sidebar.preview") }}</h3>
          <img
            v-if="selectedItem.type === 'image'"
            :src="previewLink"
            :alt="selectedItem.name"
          />
          <video
            v-else-if="selectedItem.type === 'video'"
            :src="previewLink"
            controls
            playsinline
            preload="metadata"
            :aria-label="selectedItem.name"
          />
          <audio
            v-else-if="selectedItem.type === 'audio'"
            :src="previewLink"
            controls
            preload="metadata"
            :aria-label="selectedItem.name"
          />
          <FinderFileIcon
            v-else
            :isDir="selectedItem.isDir"
            :name="selectedItem.name"
            :type="selectedItem.type"
          />
          <p>{{ selectedItem.name }}</p>
        </div>
        <div class="share-qr">
          <qrcode-vue
            :value="link"
            :size="112"
            level="M"
            render-as="svg"
            :aria-label="t('buttons.download')"
          />
        </div>
      </aside>
    </div>
    <FinderStatus
      v-if="req && !error && !layoutStore.loading"
      :total="req.isDir ? req.items.length : 1"
      :selected="fileStore.selectedCount"
    />
  </div>
</template>

<script setup lang="ts">
import { pub as api } from "@/api";
import { filesize } from "@/utils";
import dayjs from "dayjs";
import FinderIcon from "@/components/FinderIcon.vue";
import FinderFileIcon from "@/components/files/FinderFileIcon.vue";
import FinderStatus from "@/components/files/FinderStatus.vue";
import { listPage, pageSize } from "@/utils/finder";
import HeaderBar from "@/components/header/HeaderBar.vue";
import Action from "@/components/header/Action.vue";
import Breadcrumbs from "@/components/Breadcrumbs.vue";
import Errors from "@/views/Errors.vue";
import QrcodeVue from "qrcode.vue";
import Item from "@/components/files/ListingItem.vue";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";
import { computed, inject, onMounted, onBeforeUnmount, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { useI18n } from "vue-i18n";
import { StatusError } from "@/api/utils";
import { copy } from "@/utils/clipboard";

const error = ref<StatusError | null>(null);
const password = ref<string>("");
const attemptedPasswordLogin = ref<boolean>(false);
const hash = ref<string>("");
const token = ref<string>("");

const $showError = inject<IToastError>("$showError")!;
const $showSuccess = inject<IToastSuccess>("$showSuccess")!;

const { t } = useI18n({});

const route = useRoute();
const fileStore = useFileStore();
const layoutStore = useLayoutStore();

watch(
  () => route.path,
  () => {
    fetchData();
  }
);

const req = computed(() => fileStore.req);
const visibleItems = computed(() => {
  const items = req.value?.items || [];
  const page = listPage(route.query.page, items.length);
  return items.slice((page - 1) * pageSize, page * pageSize);
});

// Define computes

const selectedItem = computed(() =>
  req.value?.isDir && fileStore.selectedCount === 1
    ? req.value.items[fileStore.selected[0]]
    : null
);
const preview = computed(() => selectedItem.value || req.value);
const link = computed(() => (req.value ? api.getDownloadURL(req.value) : ""));
const previewLink = computed(() => {
  if (!req.value) return "";
  return api.getDownloadURL(
    {
      ...req.value,
      path: preview.value?.path || req.value.path,
    },
    true
  );
});
const inlineLink = computed(() =>
  req.value ? api.getDownloadURL(req.value, true) : ""
);
const humanSize = computed(() => {
  if (req.value) {
    return req.value.isDir
      ? req.value.items.length
      : filesize(req.value.size ?? 0);
  } else {
    return "";
  }
});
const humanTime = computed(() => dayjs(req.value?.modified).fromNow());
const modTime = computed(() =>
  req.value
    ? new Date(Date.parse(req.value.modified)).toLocaleString()
    : new Date().toLocaleString()
);

// Functions
const fetchData = async () => {
  fileStore.reload = false;
  fileStore.selected = [];
  fileStore.multiple = false;
  layoutStore.closeHovers();

  // Set loading to true and reset the error.
  layoutStore.loading = true;
  error.value = null;
  if (password.value !== "") {
    attemptedPasswordLogin.value = true;
  }

  let url = route.path;
  if (url === "") url = "/";
  if (url[0] !== "/") url = "/" + url;

  try {
    const file = await api.fetch(url, password.value);
    file.hash = hash.value;

    token.value = file.token || "";

    fileStore.updateRequest(file);
    document.title = `${file.name} - ${document.title}`;
  } catch (err) {
    if (err instanceof Error) {
      error.value = err;
    }
  } finally {
    layoutStore.loading = false;
  }
};

const keyEvent = (event: KeyboardEvent) => {
  if (event.key === "Escape") {
    // If we're on a listing, unselect all
    // files and folders.
    if (fileStore.selectedCount > 0) {
      fileStore.selected = [];
    }
  }
};

const toggleMultipleSelection = () => {
  fileStore.toggleMultiple();
};

const isSingleFile = () =>
  Boolean(selectedItem.value && !selectedItem.value.isDir);

const download = () => {
  if (!req.value) return false;

  if (isSingleFile()) {
    api.download(
      null,
      hash.value,
      token.value,
      req.value.items[fileStore.selected[0]].path
    );
    return true;
  }

  layoutStore.showHover({
    prompt: "download",
    confirm: (format: DownloadFormat) => {
      if (req.value === null) return false;
      layoutStore.closeHovers();

      const files: string[] = [];

      for (const i of fileStore.selected) {
        files.push(req.value.items[i].path);
      }

      api.download(format, hash.value, token.value, ...files);
      return true;
    },
  });

  return true;
};

const linkSelected = () => {
  return isSingleFile() && req.value
    ? api.getDownloadURL({
        ...req.value,
        hash: hash.value,
        path: req.value.items[fileStore.selected[0]].path,
      })
    : "";
};

const copyToClipboard = (text: string) => {
  copy({ text }).then(
    () => {
      // clipboard successfully set
      $showSuccess(t("success.linkCopied"));
    },
    () => {
      // clipboard write failed
      copy({ text }, { permission: true }).then(
        () => {
          // clipboard successfully set
          $showSuccess(t("success.linkCopied"));
        },
        (e) => {
          // clipboard write failed
          $showError(e);
        }
      );
    }
  );
};

onMounted(async () => {
  // Created
  hash.value = route.params.path[0];
  window.addEventListener("keydown", keyEvent);
  await fetchData();
});

onBeforeUnmount(() => {
  // Destroyed
  window.removeEventListener("keydown", keyEvent);
});
</script>
