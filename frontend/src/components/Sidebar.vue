<template>
  <div
    v-show="active"
    @click="closeHovers"
    class="overlay finder-sidebar-overlay"
  ></div>
  <nav
    class="finder-sidebar"
    :class="{ active }"
    :aria-label="$t('buttons.toggleSidebar')"
  >
    <div class="finder-sidebar-title">
      <div class="finder-traffic-lights" aria-hidden="true">
        <span></span><span></span><span></span>
      </div>
      <span class="finder-brand">File Browser</span>
    </div>
    <template v-if="isLoggedIn">
      <section class="finder-sidebar-section">
        <h2>{{ $t("finder.favorites") }}</h2>
        <button
          class="action"
          :class="{ 'finder-current': $route.path.startsWith('/files') }"
          @click="toRoot"
          :title="$t('sidebar.myFiles')"
        >
          <FinderIcon name="folder" /><span>{{ $t("sidebar.myFiles") }}</span>
        </button>
        <button
          v-if="user.perm.share"
          class="action"
          :class="{ 'finder-current': $route.path === '/settings/shares' }"
          @click="toShares"
        >
          <FinderIcon name="share" /><span>{{ $t("finder.shared") }}</span>
        </button>
        <button
          class="action"
          :class="{ 'finder-current': $route.path === '/settings/profile' }"
          @click="toAccountSettings"
        >
          <FinderIcon name="home" /><span>{{ user.username }}</span>
        </button>
      </section>
      <section class="finder-sidebar-section">
        <h2>{{ $t("finder.locations") }}</h2>
        <button class="action" @click="toRoot">
          <FinderIcon name="server" /><span>{{ serverName }}</span>
        </button>
        <button
          v-if="user.perm.admin"
          class="action"
          :class="{
            'finder-current':
              $route.path.startsWith('/settings') &&
              !['/settings/profile', '/settings/shares'].includes($route.path),
          }"
          @click="toGlobalSettings"
        >
          <FinderIcon name="settings" /><span>{{
            $t("sidebar.settings")
          }}</span>
        </button>
      </section>
      <section
        v-if="user.perm.create && isFiles"
        class="finder-sidebar-section"
      >
        <h2>{{ $t("finder.actions") }}</h2>
        <button class="action" @click="showHover('newDir')">
          <FinderIcon name="newFolder" /><span>{{
            $t("sidebar.newFolder")
          }}</span>
        </button>
        <button class="action" @click="showHover('newFile')">
          <FinderIcon name="newFile" /><span>{{ $t("sidebar.newFile") }}</span>
        </button>
      </section>
    </template>
    <section v-else class="finder-sidebar-section">
      <router-link v-if="!hideLoginButton" class="action" to="/login"
        ><FinderIcon name="user" /><span>{{
          $t("sidebar.login")
        }}</span></router-link
      >
      <router-link v-if="signup" class="action" to="/login"
        ><FinderIcon name="newFile" /><span>{{
          $t("sidebar.signup")
        }}</span></router-link
      >
    </section>
    <div class="finder-sidebar-bottom">
      <div class="finder-disk" v-if="isFiles && !disableUsedPercentage">
        <progress-bar :val="usage.usedPercentage" size="small" />
        <p>
          {{ $t("sidebar.diskUsed", { used: usage.used, total: usage.total }) }}
        </p>
      </div>
      <button class="action" @click="help">
        <FinderIcon name="help" /><span>{{ $t("sidebar.help") }}</span>
      </button>
      <button
        v-if="isLoggedIn && canLogout"
        class="action"
        id="logout"
        @click="logout"
      >
        <FinderIcon name="logout" /><span>{{ $t("sidebar.logout") }}</span>
      </button>
      <p class="finder-version">
        <span v-if="disableExternal">File Browser</span
        ><a
          v-else
          href="https://github.com/filebrowser/filebrowser"
          target="_blank"
          rel="noopener noreferrer"
          >File Browser</a
        >
        {{ version }}
      </p>
    </div>
  </nav>
</template>

<script>
import { reactive } from "vue";
import { mapActions, mapState } from "pinia";
import { useAuthStore } from "@/stores/auth";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";

import * as auth from "@/utils/auth";
import {
  version,
  signup,
  hideLoginButton,
  disableExternal,
  disableUsedPercentage,
  noAuth,
  logoutPage,
  loginPage,
} from "@/utils/constants";
import { files as api } from "@/api";
import FinderIcon from "@/components/FinderIcon.vue";
import ProgressBar from "@/components/ProgressBar.vue";
import prettyBytes from "pretty-bytes";

const USAGE_DEFAULT = { used: "0 B", total: "0 B", usedPercentage: 0 };

export default {
  name: "sidebar",
  setup() {
    const usage = reactive(USAGE_DEFAULT);
    return { usage, usageAbortController: new AbortController() };
  },
  components: {
    ProgressBar,
    FinderIcon,
  },
  inject: ["$showError"],
  computed: {
    ...mapState(useAuthStore, ["user", "isLoggedIn"]),
    ...mapState(useFileStore, ["isFiles", "reload"]),
    ...mapState(useLayoutStore, ["currentPromptName"]),
    active() {
      return this.currentPromptName === "sidebar";
    },
    serverName: () => window.location.hostname,
    signup: () => signup,
    hideLoginButton: () => hideLoginButton,
    version: () => version,
    disableExternal: () => disableExternal,
    disableUsedPercentage: () => disableUsedPercentage,
    canLogout: () => !noAuth && (loginPage || logoutPage !== "/login"),
  },
  methods: {
    ...mapActions(useLayoutStore, ["closeHovers", "showHover"]),
    abortOngoingFetchUsage() {
      this.usageAbortController.abort();
    },
    async fetchUsage() {
      const path = this.$route.path.endsWith("/")
        ? this.$route.path
        : this.$route.path + "/";
      let usageStats = USAGE_DEFAULT;
      if (this.disableUsedPercentage) {
        return Object.assign(this.usage, usageStats);
      }
      try {
        this.abortOngoingFetchUsage();
        this.usageAbortController = new AbortController();
        const usage = await api.usage(path, this.usageAbortController.signal);
        usageStats = {
          used: prettyBytes(usage.used, { binary: true }),
          total: prettyBytes(usage.total, { binary: true }),
          usedPercentage: Math.round((usage.used / usage.total) * 100),
        };
      } finally {
        return Object.assign(this.usage, usageStats);
      }
    },
    toRoot() {
      this.$router.push({
        path: "/files/",
        query: { ...this.$route.query, page: undefined, tab: undefined },
      });
      this.closeHovers();
    },
    toAccountSettings() {
      this.$router.push({
        path: "/settings/profile",
        query: { ...this.$route.query, page: undefined, tab: "profile" },
      });
      this.closeHovers();
    },
    toGlobalSettings() {
      this.$router.push({
        path: "/settings/global",
        query: { ...this.$route.query, page: undefined, tab: "global" },
      });
      this.closeHovers();
    },
    toShares() {
      this.$router.push({
        path: "/settings/shares",
        query: { ...this.$route.query, page: undefined, tab: "shares" },
      });
      this.closeHovers();
    },
    help() {
      this.showHover("help");
    },
    logout: auth.logout,
  },
  watch: {
    "$route.path": {
      handler(path) {
        if (path.includes("/files")) {
          this.fetchUsage();
        }
      },
      immediate: true,
    },
  },
  unmounted() {
    this.abortOngoingFetchUsage();
  },
};
</script>
