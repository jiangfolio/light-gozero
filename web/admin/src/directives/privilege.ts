import type { App, Directive } from 'vue'

import { usePermissionStore } from '@/stores/permission'

/**
 * v-privilege：无权限时从 DOM 移除节点。
 * 用法：v-privilege="'system:org:create'" 或 v-privilege="['a','b']"
 */
export const privilegeDirective: Directive<HTMLElement, string | string[]> = {
  mounted(el, binding) {
    const store = usePermissionStore()
    const value = binding.value
    if (!store.hasPerm(value)) {
      el.parentNode?.removeChild(el)
    }
  },
}

export function setupPrivilegeDirective(app: App) {
  app.directive('privilege', privilegeDirective)
}
