import type { Router } from 'vue-router'

import { getToken } from '@/lib/auth-token'
import { usePermissionStore } from '@/stores/permission'
import { useUserStore } from '@/stores/user'

/**
 * 全局路由守卫：鉴权 + 首次进入时拉 resources 并注入动态路由。
 */
export function setupRouterGuard(router: Router) {
  // 使用返回值风格（vue-router 4.5+ 弃用 next 回调）：true 放行、对象/路径重定向
  router.beforeEach(async (to) => {
    const token = getToken()
    const isPublic = Boolean(to.meta.public)

    // 1. 登录页：已登录则进首页，未登录放行
    if (to.path === '/login') {
      return token ? { path: '/' } : true
    }

    // 2. 无 token：公开页放行，否则跳登录并带回跳地址
    if (!token) {
      return isPublic ? true : { path: '/login', query: { redirect: to.fullPath } }
    }

    // 3. 有 token 但资源未加载：拉 resources → addRoute → replace 重进当前路由；失败则登出回登录页
    const permissionStore = usePermissionStore()
    if (!permissionStore.loaded) {
      try {
        useUserStore().hydrateFromStorage()
        await permissionStore.loadResources()
        const routes = permissionStore.buildRoutes()
        for (const route of routes) {
          router.addRoute('RootLayout', route)
        }
        // 按路径重进，让新注入的动态路由参与匹配。两个坑：
        // 1. 不能展开 to（此时 to 命中 404 兜底路由，会带上兜底 name，按名字导航又跳回 404）
        // 2. vue-router 5 对象形式的 path 串内嵌 query/hash 会被丢弃，必须拆字段显式传递
        return { path: to.path, query: to.query, hash: to.hash, replace: true }
      } catch {
        // 资源加载失败按未登录处理：先移除可能残留的动态路由，再清会话（顺序同上，不能反）
        resetDynamicRoutes(router)
        await useUserStore().logout(false)
        return { path: '/login', query: { redirect: to.fullPath } }
      }
    }

    // 4. 常规放行
    return true
  })
}

/**
 * 退出时移除动态路由，避免残留权限。
 * 必须在 userStore.logout() 之前调用：logout 会清空 permission store 里的路由名列表。
 */
export function resetDynamicRoutes(router: Router) {
  const permissionStore = usePermissionStore()
  for (const name of permissionStore.dynamicRouteNames) {
    if (name && router.hasRoute(name)) {
      router.removeRoute(name)
    }
  }
  permissionStore.dynamicRouteNames = []
}
