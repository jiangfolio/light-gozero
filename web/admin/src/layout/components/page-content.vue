<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'

import { useTabsStore } from '@/stores/tabs'

const route = useRoute()
const tabsStore = useTabsStore()

const keepAliveIncludes = computed(() => {
  return tabsStore.tabs
    .filter((t) => {
      // 根据当前已打开页签；真正 keepAlive 由路由 meta 决定，这里用 name
      return Boolean(t.name)
    })
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
