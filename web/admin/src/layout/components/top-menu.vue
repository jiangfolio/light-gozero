<script setup lang="ts">
import { computed, h } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import type { ResourceNode } from '@/api/system/user'
import { resolveMenuFullPath, usePermissionStore } from '@/stores/permission'

import MenuIcon from './menu-icon.vue'

const route = useRoute()
const router = useRouter()
const permissionStore = usePermissionStore()

function toItems(nodes: ResourceNode[], parentPath = ''): unknown[] {
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

const menuItems = computed(() => [
  {
    key: '/home',
    label: '首页',
    icon: () => h(MenuIcon, { name: 'Home' }),
  },
  ...toItems(permissionStore.menuTree),
])

const selectedKeys = computed(() => [route.path])

function onClick(info: { key: string | number }) {
  const key = String(info.key)
  if (key.startsWith('dir:')) return
  router.push(key)
}
</script>

<template>
  <AMenu
    class="top-menu-bar"
    mode="horizontal"
    theme="dark"
    :selected-keys="selectedKeys"
    :items="menuItems as any"
    @click="onClick"
  />
</template>

<style scoped lang="less">
@import '../styles/layout-variables.less';

.top-menu-bar {
  min-width: 0;
  flex: 1;
  height: @top-menu-height !important;
  border-bottom: 0 !important;
  background: transparent !important;
  line-height: @top-menu-height !important;

  :deep(.ant-menu-overflow),
  :deep(.ant-menu-item),
  :deep(.ant-menu-submenu),
  :deep(.ant-menu-submenu-title) {
    display: inline-flex !important;
    height: @top-menu-height !important;
    align-items: center !important;
    line-height: @top-menu-height !important;
    margin-block: 0 !important;
    padding-block: 0 !important;
  }

  :deep(.ant-menu-item .ant-menu-item-icon),
  :deep(.ant-menu-item .anticon),
  :deep(.ant-menu-title-content) {
    display: inline-flex;
    align-items: center;
    line-height: 1;
  }

  :deep(.ant-menu-item-selected) {
    background: #1677ff !important;
  }
}
</style>
