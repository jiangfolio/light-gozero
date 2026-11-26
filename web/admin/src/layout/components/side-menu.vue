<script setup lang="ts">
import { computed, h, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import type { ResourceNode } from '@/api/system/user'
import { useAppConfigStore } from '@/stores/app-config'
import { resolveMenuFullPath, usePermissionStore } from '@/stores/permission'

import AppLogo from './app-logo.vue'
import MenuIcon from './menu-icon.vue'

defineProps<{ collapsed?: boolean }>()

const route = useRoute()
const router = useRouter()
const permissionStore = usePermissionStore()
const appConfig = useAppConfigStore()
const openKeys = ref<string[]>([])

type MenuItem = {
  key: string
  label: string
  icon?: () => ReturnType<typeof h>
  children?: MenuItem[]
}

function toItems(nodes: ResourceNode[], parentPath = ''): MenuItem[] {
  return nodes.map((node) => {
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

const menuItems = computed<MenuItem[]>(() => [
  {
    key: '/home',
    label: '首页',
    icon: () => h(MenuIcon, { name: 'Home' }),
  },
  ...toItems(permissionStore.menuTree),
])

const selectedKeys = computed(() => [route.path])
const isDark = computed(() => appConfig.state.menuTheme === 'dark')

function collectOpenKeys(path: string) {
  const keys: string[] = []
  const walk = (nodes: ResourceNode[], parentPath: string) => {
    for (const n of nodes) {
      const full = resolveMenuFullPath(n, parentPath)
      if (path === full || path.startsWith(`${full}/`)) {
        if (n.type === 'directory') keys.push(`dir:${n.id}`)
        if (n.children?.length) walk(n.children, full)
      }
    }
  }
  walk(permissionStore.menuTree, '')
  return keys
}

watch(
  () => route.path,
  (path) => {
    openKeys.value = collectOpenKeys(path)
  },
  { immediate: true },
)

function onOpenChange(keys: string[]) {
  if (!appConfig.state.menuAccordion) {
    openKeys.value = keys
    return
  }
  const latest = keys.find((k) => !openKeys.value.includes(k))
  openKeys.value = latest ? [latest] : keys
}

function onClick(info: { key: string | number }) {
  const key = String(info.key)
  if (key.startsWith('dir:')) return
  router.push(key)
}
</script>

<template>
  <div class="flex h-full flex-col overflow-hidden">
    <AppLogo
      class="shadow-[0_1px_4px_rgba(0,21,41,0.08)]"
      :collapsed="collapsed"
      :dark="isDark"
    />
    <div class="min-h-0 flex-1 overflow-auto [&::-webkit-scrollbar]:w-1">
      <AMenu
        class="!h-full !border-e-0"
        mode="inline"
        :theme="appConfig.state.menuTheme"
        :inline-collapsed="collapsed"
        :selected-keys="selectedKeys"
        :open-keys="openKeys"
        :items="menuItems"
        @click="onClick"
        @open-change="onOpenChange"
      />
    </div>
  </div>
</template>
