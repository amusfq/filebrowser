<template>
  <div class="dashboard remote-storage">
    <header-bar showMenu showLogo />
    <div class="card">
      <div class="card-title">
        <h2>{{ t("remote.title") }}</h2>
      </div>
      <div class="card-content">
        <label for="remote-server">{{ t("remote.server") }}</label>
        <select
          id="remote-server"
          class="input"
          :value="serverID"
          @change="selectServer"
        >
          <option value="">{{ t("remote.selectServer") }}</option>
          <option v-for="server in servers" :key="server.id" :value="server.id">
            {{ server.name }} — {{ server.username }}@{{ server.host }}
          </option>
        </select>
        <button class="button" :disabled="busy" @click="adding = !adding">
          {{ t("remote.add") }}
        </button>
        <button
          v-if="selectedServer"
          class="button button--red"
          :disabled="busy"
          @click="removeServer"
        >
          {{ t("remote.remove") }}
        </button>
        <p v-if="error" role="alert">{{ error }}</p>
        <p v-if="notice" role="status">{{ notice }}</p>
        <form v-if="adding" @submit.prevent="addServer">
          <label for="remote-name">{{ t("remote.name") }}</label>
          <input
            id="remote-name"
            v-model="form.name"
            class="input"
            required
            :disabled="busy"
          />
          <label for="remote-host">{{ t("remote.host") }}</label>
          <input
            id="remote-host"
            v-model="form.host"
            class="input"
            required
            :disabled="busy"
          />
          <label for="remote-port">{{ t("remote.port") }}</label>
          <input
            id="remote-port"
            v-model.number="form.port"
            class="input"
            type="number"
            min="1"
            max="65535"
            required
            :disabled="busy"
          />
          <label for="remote-user">{{ t("remote.username") }}</label>
          <input
            id="remote-user"
            v-model="form.username"
            class="input"
            autocomplete="off"
            required
            :disabled="busy"
          />
          <label for="remote-root">{{ t("remote.root") }}</label>
          <input
            id="remote-root"
            v-model="form.root"
            class="input"
            placeholder="/srv/files"
            required
            :disabled="busy"
          />
          <label for="remote-fingerprint">{{ t("remote.fingerprint") }}</label>
          <input
            id="remote-fingerprint"
            v-model="form.fingerprint"
            class="input"
            placeholder="SHA256:…"
            required
            :disabled="busy"
            aria-describedby="remote-host-key-help"
          />
          <p id="remote-host-key-help">{{ t("remote.fingerprintHelp") }}</p>
          <label
            ><input
              v-model="form.keyAuth"
              type="checkbox"
              :disabled="busy"
              @change="form.credential = ''"
            />
            {{ t("remote.keyAuth") }}</label
          >
          <label for="remote-credential">{{
            t(form.keyAuth ? "remote.privateKey" : "remote.password")
          }}</label>
          <textarea
            v-if="form.keyAuth"
            id="remote-credential"
            v-model="form.credential"
            class="input"
            rows="5"
            required
            :disabled="busy"
            spellcheck="false"
          ></textarea>
          <input
            v-else
            id="remote-credential"
            v-model="form.credential"
            class="input"
            type="password"
            autocomplete="new-password"
            required
            :disabled="busy"
          />
          <p>{{ t("remote.adminOnly") }}</p>
          <button type="submit" class="button button--blue" :disabled="busy">
            {{ t(busy ? "remote.connecting" : "remote.connectSave") }}
          </button>
        </form>
      </div>
    </div>
    <div v-if="selectedServer" class="card">
      <div class="card-title remote-path">
        <h2>{{ selectedServer.name }}: {{ folder }}</h2>
        <button
          class="button"
          :disabled="folder === '/' || loading"
          @click="openFolder(parentFolder)"
        >
          {{ t("remote.up") }}
        </button>
        <button class="button" :disabled="loading || busy" @click="loadFiles">
          {{ t("remote.refresh") }}
        </button>
      </div>
      <div class="card-content">
        <label for="remote-search">{{ t("remote.search") }}</label>
        <input
          id="remote-search"
          class="input"
          type="search"
          :value="search"
          @input="searchFiles"
        />
        <template v-if="authStore.user?.perm.create">
          <label for="remote-upload">{{ t("remote.upload") }}</label>
          <input
            id="remote-upload"
            type="file"
            :disabled="busy || loading"
            @change="uploadFile"
          />
          <p>{{ t("remote.noOverwrite") }}</p>
        </template>
        <p v-if="loading" role="status">{{ t("files.loading") }}</p>
        <table v-else>
          <thead>
            <tr>
              <th>{{ t("remote.name") }}</th>
              <th>{{ t("remote.size") }}</th>
              <th>{{ t("remote.modified") }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="entry in visibleItems" :key="entry.name">
              <td>
                <button
                  v-if="entry.isDir"
                  class="remote-folder"
                  @click="openFolder(entryPath(entry.name))"
                >
                  <FinderIcon name="folder" /> {{ entry.name }}
                </button>
                <a
                  v-else-if="authStore.user?.perm.download"
                  :href="api.downloadURL(serverID, entryPath(entry.name))"
                  >{{ entry.name }}</a
                >
                <span v-else>{{ entry.name }}</span>
              </td>
              <td>{{ entry.isDir ? "—" : prettyBytes(entry.size) }}</td>
              <td>{{ new Date(entry.modified).toLocaleString() }}</td>
            </tr>
            <tr v-if="!filteredItems.length">
              <td colspan="3">{{ t("remote.empty") }}</td>
            </tr>
          </tbody>
        </table>
      </div>
      <FinderStatus :total="filteredItems.length" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import prettyBytes from "pretty-bytes";
import HeaderBar from "@/components/header/HeaderBar.vue";
import FinderIcon from "@/components/FinderIcon.vue";
import FinderStatus from "@/components/files/FinderStatus.vue";
import { useAuthStore } from "@/stores/auth";
import { listPage, pageSize } from "@/utils/finder";
import * as api from "@/api/remote";

const route = useRoute();
const router = useRouter();
const { t } = useI18n();
const authStore = useAuthStore();
const servers = ref<api.RemoteServer[]>([]);
const items = ref<api.RemoteEntry[]>([]);
const error = ref("");
const notice = ref("");
const busy = ref(false);
const loading = ref(false);
const adding = ref(false);
const form = reactive({
  name: "",
  host: "",
  port: 22,
  username: "",
  root: "/",
  fingerprint: "",
  keyAuth: false,
  credential: "",
});
const serverID = computed(() =>
  typeof route.query.server === "string" ? route.query.server : ""
);
const folder = computed(() =>
  typeof route.query.path === "string" ? route.query.path : "/"
);
const search = computed(() =>
  typeof route.query.search === "string" ? route.query.search : ""
);
const selectedServer = computed(() =>
  servers.value.find((server) => server.id === serverID.value)
);
const parentFolder = computed(
  () => folder.value.replace(/\/?[^/]+\/?$/, "") || "/"
);
const filteredItems = computed(() =>
  items.value.filter((entry) =>
    entry.name.toLocaleLowerCase().includes(search.value.toLocaleLowerCase())
  )
);
const visibleItems = computed(() => {
  const page = listPage(route.query.page, filteredItems.value.length);
  return filteredItems.value.slice((page - 1) * pageSize, page * pageSize);
});
const entryPath = (name: string) =>
  `${folder.value.replace(/\/$/, "")}/${name}`;
let controller: AbortController | undefined;

function selectServer(event: Event) {
  router.push({
    query: {
      ...route.query,
      server: (event.target as HTMLSelectElement).value || undefined,
      path: undefined,
      page: undefined,
      search: undefined,
    },
  });
}
function openFolder(path: string) {
  router.push({
    query: { ...route.query, path, page: undefined, search: undefined },
  });
}
function searchFiles(event: Event) {
  router.push({
    query: {
      ...route.query,
      search: (event.target as HTMLInputElement).value || undefined,
      page: undefined,
    },
  });
}
async function loadFiles() {
  controller?.abort();
  const request = new AbortController();
  controller = request;
  items.value = [];
  error.value = "";
  loading.value = !!serverID.value;
  if (!serverID.value) return;
  try {
    items.value = await api.files(serverID.value, folder.value, request.signal);
  } catch (err) {
    if (!request.signal.aborted) error.value = String(err);
  } finally {
    if (!request.signal.aborted) loading.value = false;
  }
}
async function addServer() {
  busy.value = true;
  error.value = "";
  notice.value = "";
  try {
    const server = await api.add(form);
    servers.value.push(server);
    form.credential = "";
    adding.value = false;
    await router.push({
      query: {
        ...route.query,
        server: server.id,
        path: undefined,
        page: undefined,
        search: undefined,
      },
    });
  } catch (err) {
    error.value = String(err);
  } finally {
    busy.value = false;
  }
}
async function removeServer() {
  if (!window.confirm(t("remote.removeConfirm"))) return;
  const id = serverID.value;
  busy.value = true;
  try {
    await api.remove(id);
    servers.value = servers.value.filter((server) => server.id !== id);
    await router.push({
      query: {
        ...route.query,
        server: undefined,
        path: undefined,
        page: undefined,
        search: undefined,
      },
    });
  } catch (err) {
    error.value = String(err);
  } finally {
    busy.value = false;
  }
}
async function uploadFile(event: Event) {
  const input = event.target as HTMLInputElement;
  const file = input.files?.[0];
  if (!file) return;
  const id = serverID.value;
  const path = folder.value;
  busy.value = true;
  error.value = "";
  notice.value = "";
  try {
    await api.upload(id, path, file);
    if (id === serverID.value && path === folder.value) {
      await loadFiles();
      notice.value = t("remote.uploaded", { name: file.name });
    }
  } catch (err) {
    error.value = String(err);
  } finally {
    busy.value = false;
    input.value = "";
  }
}
watch([serverID, folder], () => {
  notice.value = "";
  loadFiles();
});
onMounted(async () => {
  try {
    servers.value = await api.list();
    await loadFiles();
  } catch (err) {
    error.value = String(err);
  }
});
onUnmounted(() => controller?.abort());
</script>

<style scoped>
.remote-storage .card {
  margin: 1rem;
}
.remote-storage label {
  display: block;
  margin-top: 1rem;
}
.remote-storage .button {
  margin: 0.75rem 0.5rem 0 0;
}
.remote-path {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.5rem;
}
.remote-path h2 {
  flex: 1;
  overflow-wrap: anywhere;
}
.remote-folder {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  border: 0;
  background: transparent;
  color: inherit;
  cursor: pointer;
  text-align: left;
}
.remote-storage td {
  overflow-wrap: anywhere;
}
.remote-storage table {
  width: 100%;
}
.remote-storage .finder-status {
  position: static;
}
</style>
