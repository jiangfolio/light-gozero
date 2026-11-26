import { http } from '@/lib/http'

export function createRole(data: any) {
  return http.post('/api/system/v1/role/create', data) as Promise<any>
}

export function updateRole(data: any) {
  return http.post('/api/system/v1/role/update', data) as Promise<any>
}

export function deleteRole(ids: any) {
  return http.post('/api/system/v1/role/delete', { ids }) as Promise<any>
}

export function getRoleDetail(id: any) {
  return http.get('/api/system/v1/role/detail', { params: { id } }) as Promise<any>
}

export function pageRole(params: any) {
  return http.get('/api/system/v1/role/page', { params }) as Promise<any>
}

export function bindRoleResources(roleId: any, resourceIds: any) {
  return http.post('/api/system/v1/role/bind_resources', {
    role_id: roleId,
    resource_ids: resourceIds,
  }) as Promise<any>
}

export function unbindRoleResources(roleId: any, resourceIds: any) {
  return http.post('/api/system/v1/role/unbind_resources', {
    role_id: roleId,
    resource_ids: resourceIds,
  }) as Promise<any>
}
