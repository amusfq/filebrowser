<template>
  <div class="dashboard">
    <header-bar showMenu showLogo />

    <div id="nav">
      <div class="wrapper">
        <ul>
          <router-link
            :to="{
              path: '/settings/profile',
              query: { ...route.query, tab: 'profile', page: undefined },
            }"
            ><li :class="{ active: $route.path === '/settings/profile' }">
              {{ t("settings.profileSettings") }}
            </li></router-link
          >
          <router-link
            :to="{
              path: '/settings/shares',
              query: { ...route.query, tab: 'shares', page: undefined },
            }"
            v-if="user?.perm.share"
            ><li :class="{ active: $route.path === '/settings/shares' }">
              {{ t("settings.shareManagement") }}
            </li></router-link
          >
          <router-link
            :to="{
              path: '/settings/global',
              query: { ...route.query, tab: 'global', page: undefined },
            }"
            v-if="user?.perm.admin"
            ><li :class="{ active: $route.path === '/settings/global' }">
              {{ t("settings.globalSettings") }}
            </li></router-link
          >
          <router-link
            :to="{
              path: '/settings/users',
              query: { ...route.query, tab: 'users', page: undefined },
            }"
            v-if="user?.perm.admin"
            ><li
              :class="{
                active:
                  $route.path === '/settings/users' || $route.name === 'User',
              }"
            >
              {{ t("settings.userManagement") }}
            </li></router-link
          >
        </ul>
      </div>
    </div>

    <div class="card" v-if="user?.perm.admin">
      <div class="card-title">
        <h2>{{ t("settings.sunsetTitle") }}</h2>
      </div>

      <div class="card-content">
        <p>{{ t("settings.sunsetBody") }}</p>
        <p>
          <a
            href="https://github.com/filebrowser/filebrowser#security"
            target="_blank"
            rel="noopener noreferrer"
            >{{ t("settings.sunsetLink") }}</a
          >
        </p>
      </div>
    </div>

    <div v-if="loading">
      <h2 class="message delayed">
        <div class="spinner">
          <div class="bounce1"></div>
          <div class="bounce2"></div>
          <div class="bounce3"></div>
        </div>
        <span>{{ t("files.loading") }}</span>
      </h2>
    </div>

    <router-view></router-view>
  </div>
</template>

<script setup lang="ts">
import { useAuthStore } from "@/stores/auth";
import { useLayoutStore } from "@/stores/layout";
import HeaderBar from "@/components/header/HeaderBar.vue";
import { computed } from "vue";
import { useRoute } from "vue-router";
import { useI18n } from "vue-i18n";

const { t } = useI18n();
const route = useRoute();

const authStore = useAuthStore();
const layoutStore = useLayoutStore();

const user = computed(() => authStore.user);
const loading = computed(() => layoutStore.loading);
</script>
