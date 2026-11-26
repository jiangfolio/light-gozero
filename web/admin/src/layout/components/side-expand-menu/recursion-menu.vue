<script setup lang="ts">
import { computed, h, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { resolveMenuFullPath } from '@/stores/permission'

import MenuIcon from '../menu-icon.vue'
import { selectedTopMenu } from './shared'

const route = useRoute()
const router = useRouter()

const topMenu = selectedTopMenu
const openKeys = ref<string[]>([])

const showSecond = computed(
  () => Boolean(topMenu.value?.children?.some((c: any) => c.type === 'menu' || c.type === 'directory')),
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
    <div class="top-menu">
      <span>{{ topMenu?.name }}</span>
    </div>
    <div class="bottom-menu">
      <AMenu
        mode="inline"
        :selected-keys="selectedKeys"
        :open-keys="openKeys"
        :items="menuItems"
        @click="onClick"
        @open-change="(keys: string[]) => (openKeys = keys)"
      />
    </div>
  </div>
</template>

<style scoped lang="less">
@import '../../styles/layout-variables.less';

.recursion-container {
  height: 100vh;
  background-color: #fff;
  border-right: 1px solid #f0f0f0;
}

.recursion-container ::-webkit-scrollbar {
  width: 0 !important;
}

.bottom-menu {
  overflow: auto;
  display: flex;
  height: 90%;
  color: #515a6e;

  :deep(.ant-menu) {
    width: 100%;
    border-inline-end: none !important;
  }
}

.top-menu {
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  height: @header-user-height;
  font-size: 16px;
  border-bottom: 1px solid #f0f0f0;
  border-right: 1px solid #f0f0f0;
}
</style>
