<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { useAppConfigStore } from '@/stores/app-config'
import { resolveMenuFullPath, usePermissionStore } from '@/stores/permission'

import MenuIcon from '../menu-icon.vue'
import { hasSideMenu, selectedTopMenu } from './shared'

const props = defineProps<{
  collapsed?: boolean
}>()

const route = useRoute()
const router = useRouter()
const appConfig = useAppConfigStore()
const permissionStore = usePermissionStore()

const menuTree = computed(() => permissionStore.menuTree)
const selectedKeys = ref<string[]>([])

function findFirstMenu(node: any, basePath: string): string | null {
  for (const child of node.children || []) {
    const childFull = resolveMenuFullPath(child, basePath)
    if (child.type === 'menu') return childFull
    const nested = findFirstMenu(child, childFull)
    if (nested) return nested
  }
  return null
}

function onSelectMenu(item: any) {
  selectedKeys.value = [item.id]
  selectedTopMenu.value = item
  const full = resolveMenuFullPath(item, '')
  if (item.type === 'menu') {
    router.push(full)
    return
  }
  const first = findFirstMenu(item, full)
  if (first) router.push(first)
}

function syncFromRoute() {
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

const menuInfo = computed(() => {
  let width = '100vw'
  let right = '-100vw'
  if (hasSideMenu.value) {
    if (props.collapsed) {
      width = 'calc(100vw - 80px)'
      right = 'calc(-100vw + 80px)'
    } else {
      const w = appConfig.state.menuWidth
      width = `calc(100vw - ${w}px)`
      right = `calc(-100vw + ${w}px)`
    }
  }
  return { width, right }
})
</script>

<template>
  <div class="top-menu" :style="{ width: menuInfo.width, right: menuInfo.right }">
    <div class="top-menu-container">
      <AMenu :selected-keys="selectedKeys" mode="horizontal" theme="dark">
        <AMenuItem
          v-for="item in menuTree"
          :key="item.id"
          @click="onSelectMenu(item)"
        >
          <template #icon>
            <MenuIcon :name="item.icon" />
          </template>
          {{ item.name }}
        </AMenuItem>
      </AMenu>
    </div>
  </div>
</template>

<style scoped lang="less">
.top-menu {
  position: absolute;
  transition:
    all 0.2s,
    background 0s;
  top: 0;
  flex-shrink: 0;
  z-index: 20;
  background: #001529;
}

:deep(.ant-menu-item-selected) {
  background: #0958d9 !important;
  color: #fff !important;
}

.top-menu-container {
  height: 100%;
}
</style>
