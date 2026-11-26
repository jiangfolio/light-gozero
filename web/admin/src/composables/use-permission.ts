import { usePermissionStore } from '@/stores/permission'

/** 脚本侧权限判断。 */
export function usePermission() {
  const store = usePermissionStore()
  return {
    hasPerm: (code: string | string[]) => store.hasPerm(code),
    perms: store.perms,
  }
}
