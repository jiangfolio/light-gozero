<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import type { ResourceNode } from '@/api/system/user'
import { useAppConfigStore } from '@/stores/app-config'
import { resolveMenuFullPath, usePermissionStore } from '@/stores/permission'

import AppLogo from '../app-logo.vue'
import MenuIcon from '../menu-icon.vue'
import { selectedTopMenu } from './shared'

const route = useRoute()
const router = useRouter()
const appConfig = useAppConfigStore()
const permissionStore = usePermissionStore()

const menuTree = computed(() => permissionStore.menuTree)
const theme = computed(() => appConfig.state.menuTheme)
const selectedKeys = ref<string[]>([])

function menuNameAdapter(name: string) {
  return name.slice(0, 2)
}

function findFirstMenu(node: ResourceNode, basePath: string): string | null {
  // 1. 优先叶子菜单；目录则递归其子节点
  for (const child of node.children || []) {
    const childFull = resolveMenuFullPath(child, basePath)
    if (child.type === 'menu') return childFull
    const nested = findFirstMenu(child, childFull)
    if (nested) return nested
  }
  return null
}

function onSelectMenu(item: ResourceNode) {
  // 1. 高亮一级；无子节点的菜单直接跳转；有子节点交给二级栏
  selectedKeys.value = [item.id]
  selectedTopMenu.value = item
  const full = resolveMenuFullPath(item, '')
  if (item.type === 'menu' && !(item.children || []).some((c) => c.type === 'menu' || c.type === 'directory')) {
    router.push(full)
    return
  }
  if (item.type === 'menu') {
    router.push(full)
    return
  }
  const first = findFirstMenu(item, full)
  if (first) router.push(first)
}

function syncFromRoute() {
  // 1. 按当前路由命中一级目录/菜单
  const hit = menuTree.value.find((r) => {
    const full = resolveMenuFullPath(r, '')
    return route.path === full || route.path.startsWith(`${full}/`)
  })
  if (hit) {
    selectedKeys.value = [hit.id]
    selectedTopMenu.value = hit
  } else if (route.path === '/home') {
    selectedKeys.value = []
    selectedTopMenu.value = null
  }
}

watch(() => route.path, syncFromRoute, { immediate: true })
watch(menuTree, syncFromRoute)
</script>

<template>
  <div class="top-menu-container">
    <AppLogo :dark="theme === 'dark'" class="side-expand-logo" />
    <AMenu :selected-keys="selectedKeys" mode="inline" :theme="theme">
      <AMenuItem
        v-for="item in menuTree"
        :key="item.id"
        @click="onSelectMenu(item)"
      >
        <template #icon>
          <MenuIcon :name="item.icon" />
        </template>
        {{ menuNameAdapter(item.name) }}
      </AMenuItem>
    </AMenu>
  </div>
</template>

<style scoped lang="less">
@import '../../styles/layout-variables.less';

.top-menu-container {
  height: 100%;
}

.side-expand-logo {
  height: @header-user-height !important;
  min-height: @header-user-height;
}

:deep(.ant-menu-item) {
  padding-inline: 12px !important;
}
</style>
