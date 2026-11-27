<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'

import { useTabsStore } from '@/stores/tabs'

const route = useRoute()
const tabsStore = useTabsStore()

/** 仅缓存标记了 keepAlive 的已开页签，避免无缓存菜单被误缓存导致脏数据 */
const keepAliveIncludes = computed(() => {
  return tabsStore.tabs
    .filter((t) => Boolean(t.name) && Boolean(t.keepAlive))
    .map((t) => t.name)
})
</script>

<template>
  <router-view v-slot="{ Component }">
    <keep-alive :include="keepAliveIncludes">
      <component :is="Component" :key="route.name" />
    </keep-alive>
  </router-view>
</template>
