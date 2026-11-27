import type { Router } from 'vue-router'

import { usePermissionStore } from '@/stores/permission'

/**
 * 资源/角色变更后重拉当前用户菜单与权限，并重建动态路由。
 */
export async function refreshMenus(router: Router) {
  const permissionStore = usePermissionStore()
  // 1. 卸掉旧动态路由，避免残留无权或已删菜单
  for (const name of [...permissionStore.dynamicRouteNames]) {
    if (name && router.hasRoute(name)) {
      router.removeRoute(name)
    }
  }
  permissionStore.dynamicRouteNames = []
  // 2. 重新拉取 resources/perms 并注入
  await permissionStore.loadResources()
  const routes = permissionStore.buildRoutes()
  for (const route of routes) {
    router.addRoute('RootLayout', route)
  }
}
