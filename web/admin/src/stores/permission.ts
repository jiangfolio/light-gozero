import type { RouteRecordRaw } from 'vue-router'
import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import { getUserResources, type ResourceNode } from '@/api/system/user'

const viewModules = import.meta.glob('@/views/**/*.vue')

function joinPath(parent: string, path: string) {
  if (!path) return parent || '/'
  if (path.startsWith('/')) return path
  const base = parent.endsWith('/') ? parent.slice(0, -1) : parent
  if (!base || base === '/') return `/${path}`
  return `${base}/${path}`.replace(/\/{2,}/g, '/')
}

function isEnabled(status: string) {
  return String(status) === '1'
}

function isTruthyFlag(v: string) {
  return String(v) === '1' || String(v).toLowerCase() === 'true'
}

function resolveView(componentPath: string) {
  const normalized = componentPath.replace(/^\//, '').replace(/\.vue$/, '')
  const suffix = `/views/${normalized}.vue`
  const key = Object.keys(viewModules).find((k) => k.endsWith(suffix))
  if (key) return viewModules[key]
  return () => import('@/views/exception/not-found-component.vue')
}

/**
 * 资源树、权限码、动态路由构建。
 */
export const usePermissionStore = defineStore('permission', () => {
  const resources = ref<ResourceNode[]>([])
  const perms = ref<string[]>([])
  const loaded = ref(false)
  const dynamicRouteNames = ref<string[]>([])

  const menuTree = computed(() => filterMenuTree(resources.value))

  function hasPerm(code: string | string[]) {
    if (perms.value.includes('*')) return true
    const codes = Array.isArray(code) ? code : [code]
    if (!codes.length) return true
    return codes.some((c) => perms.value.includes(c))
  }

  async function loadResources() {
    const data = await getUserResources()
    resources.value = data.resources || []
    perms.value = data.perms || []
    loaded.value = true
  }

  /** 将 resources 中的 menu 扁平挂到 Layout children。 */
  function buildRoutes(): RouteRecordRaw[] {
    const routes: RouteRecordRaw[] = []
    const names: string[] = []

    const walk = (nodes: ResourceNode[], parentPath: string) => {
      for (const node of nodes) {
        if (!isEnabled(node.status)) continue
        if (node.type === 'button') continue

        const fullPath = joinPath(parentPath, node.path)

        if (node.type === 'menu' && node.component_path) {
          const name = node.route_name || fullPath.replace(/\//g, '_')
          routes.push({
            path: fullPath.replace(/^\//, ''),
            name,
            component: resolveView(node.component_path),
            meta: {
              title: node.name,
              icon: node.icon,
              keepAlive: isTruthyFlag(node.keep_alive),
              affix: isTruthyFlag(node.affix),
              hidden: isTruthyFlag(node.hidden),
            },
          })
          names.push(name)
        }

        if (node.children?.length) {
          walk(node.children, fullPath)
        }
      }
    }

    walk(resources.value, '')
    dynamicRouteNames.value = names
    return routes
  }

  function reset() {
    resources.value = []
    perms.value = []
    loaded.value = false
    dynamicRouteNames.value = []
  }

  return {
    resources,
    perms,
    loaded,
    dynamicRouteNames,
    menuTree,
    hasPerm,
    loadResources,
    buildRoutes,
    reset,
  }
})

/** 侧栏/顶栏菜单：directory + menu，过滤 hidden / 禁用。 */
function filterMenuTree(nodes: ResourceNode[]): ResourceNode[] {
  const result: ResourceNode[] = []
  for (const node of nodes) {
    if (!isEnabled(node.status)) continue
    if (node.type !== 'directory' && node.type !== 'menu') continue
    if (isTruthyFlag(node.hidden)) continue
    const children = node.children?.length ? filterMenuTree(node.children) : []
    result.push({ ...node, children })
  }
  return result
}

export function resolveMenuFullPath(node: ResourceNode, parentPath = ''): string {
  return joinPath(parentPath, node.path)
}
