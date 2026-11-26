<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { resolveMenuFullPath, usePermissionStore } from '@/stores/permission'

const route = useRoute()
const router = useRouter()
const permissionStore = usePermissionStore()

type Crumb = { title: string; path?: string }

const crumbs = computed<Crumb[]>(() => {
  const result: Crumb[] = [{ title: '首页', path: '/home' }]
  if (route.path === '/home') return result

  const matched: Crumb[] = []
  const walk = (nodes: any[], parentPath: string) => {
    for (const n of nodes) {
      const full = resolveMenuFullPath(n, parentPath)
      if (route.path === full || route.path.startsWith(`${full}/`)) {
        matched.push({
          title: n.name,
          path: n.type === 'menu' ? full : undefined,
        })
        if (n.children?.length) walk(n.children, full)
        break
      }
    }
  }
  walk(permissionStore.menuTree, '')
  return [...result, ...matched]
})

function go(path?: string) {
  if (path) router.push(path)
}
</script>

<template>
  <!-- 小屏（<1024px）隐藏面包屑，把顶栏空间留给页签与图标；大屏恢复 -->
  <ABreadcrumb class="hidden text-sm lg:block">
    <ABreadcrumbItem v-for="(c, i) in crumbs" :key="i">
      <a v-if="c.path && i < crumbs.length - 1" @click.prevent="go(c.path)">{{ c.title }}</a>
      <span v-else>{{ c.title }}</span>
    </ABreadcrumbItem>
  </ABreadcrumb>
</template>

<style scoped lang="less">
/*
 * 窄屏保护：antdv 面包屑 ol 默认 flex-wrap:wrap，
 * 空间不足时折行会把顶栏撑高（实测 66px）并压到页签行，看似「页签换行」。
 * 强制单行，超出各级省略；分隔符不压缩。
 */
/* 注意：组件根节点就是 .ant-breadcrumb，:deep 选择器不能再以它开头 */
:deep(ol) {
  flex-wrap: nowrap !important;
}

:deep(.ant-breadcrumb-item) {
  min-width: 0;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

:deep(.ant-breadcrumb-separator) {
  flex-shrink: 0;
}
</style>
