<script setup lang="ts">
import { computed, h, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { useAppConfigStore } from '@/stores/app-config'
import { resolveMenuFullPath } from '@/stores/permission'

import AppLogo from '../app-logo.vue'
import MenuIcon from '../menu-icon.vue'
import { hasSideMenu, selectedTopMenu } from './shared'

const props = defineProps<{
  collapsed?: boolean
}>()

const route = useRoute()
const router = useRouter()
const appConfig = useAppConfigStore()

const topMenu = selectedTopMenu
const openKeys = ref<string[]>([])

const showSecond = computed(
  () => Boolean(topMenu.value?.children?.some((c: any) => c.type === 'menu' || c.type === 'directory')),
)

watch(
  showSecond,
  (v) => {
    hasSideMenu.value = v
  },
  { immediate: true },
)

type MenuItem = {
  key: string
  label: string
  icon?: () => ReturnType<typeof h>
  children?: MenuItem[]
}

function toItems(nodes: any[], parentPath: string): MenuItem[] {
  return nodes
    .filter((n) => n.type === 'menu' || n.type === 'directory')
    .map((node) => {
      const fullPath = resolveMenuFullPath(node, parentPath)
      if (node.type === 'directory' && node.children?.length) {
        return {
          key: `dir:${node.id}`,
          label: node.name,
          icon: () => h(MenuIcon, { name: node.icon }),
          children: toItems(node.children, fullPath),
        }
      }
      return {
        key: fullPath,
        label: node.name,
        icon: () => h(MenuIcon, { name: node.icon }),
      }
    })
}

const menuItems = computed(() => {
  if (!topMenu.value) return []
  const base = resolveMenuFullPath(topMenu.value, '')
  return toItems(topMenu.value.children || [], base)
})

const selectedKeys = computed(() => [route.path])
const logoHeight = computed(() => (appConfig.state.compact ? '40px' : '46px'))

watch(
  topMenu,
  (menu) => {
    if (!menu?.children?.length) {
      openKeys.value = []
      return
    }
    openKeys.value = (menu.children || [])
      .filter((c: any) => c.type === 'directory')
      .map((c: any) => `dir:${c.id}`)
  },
  { immediate: true },
)

function onClick(info: { key: string | number }) {
  const key = String(info.key)
  if (key.startsWith('dir:')) return
  router.push(key)
}
</script>

<template>
  <div v-show="showSecond" class="recursion-container">
    <AppLogo
      v-if="!collapsed"
      dark
      class="top-expand-logo"
      :style="{ height: logoHeight, minHeight: logoHeight }"
    />
    <AppLogo
      v-else
      dark
      collapsed
      class="top-expand-logo min"
      :style="{ height: logoHeight, minHeight: logoHeight }"
    />
    <AMenu
      mode="inline"
      theme="light"
      :inline-collapsed="collapsed"
      :selected-keys="selectedKeys"
      :open-keys="openKeys"
      :items="menuItems"
      @click="onClick"
      @open-change="(keys: string[]) => (openKeys = keys)"
    />
  </div>
</template>

<style scoped lang="less">
.recursion-container {
  height: 100%;
  background: #fff;
}

.top-expand-logo {
  background-color: #001529 !important;
  width: 100%;
}

:deep(.ant-menu) {
  border-inline-end: none !important;
}
</style>
