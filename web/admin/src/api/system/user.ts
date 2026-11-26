import { http } from '@/lib/http'

export type ResourceNode = {
  id: string
  parent_id: string
  name: string
  type: string
  perm_code: string
  path: string
  component_path: string
  redirect: string
  route_name: string
  icon: string
  hidden: string
  keep_alive: string
  affix: string
  sort: string
  status: string
  children?: ResourceNode[]
}

export function getUserResources() {
  return http.get('/api/system/v1/user/resources') as Promise<{
    resources: ResourceNode[]
    perms: string[]
  }>
}
