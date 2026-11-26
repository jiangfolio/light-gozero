import { http } from '@/lib/http'

export function getUserResources() {
  return http.get('/api/system/v1/user/resources') as Promise<any>
}

export function createUser(data: any) {
  return http.post('/api/system/v1/user/create', data) as Promise<any>
}

export function updateUser(data: any) {
  return http.post('/api/system/v1/user/update', data) as Promise<any>
}

export function unlockUser(id: any) {
  return http.post('/api/system/v1/user/unlock', { id }) as Promise<any>
}

export function deleteUser(ids: any) {
  return http.post('/api/system/v1/user/delete', { ids }) as Promise<any>
}

export function getUserDetail(id: any) {
  return http.get('/api/system/v1/user/detail', { params: { id } }) as Promise<any>
}

export function pageUser(params: any) {
  return http.get('/api/system/v1/user/page', { params }) as Promise<any>
}

export function bindUserRoles(userId: any, roleIds: any) {
  return http.post('/api/system/v1/user/bind_roles', {
    user_id: userId,
    role_ids: roleIds,
  }) as Promise<any>
}

export function unbindUserRoles(userId: any, roleIds: any) {
  return http.post('/api/system/v1/user/unbind_roles', {
    user_id: userId,
    role_ids: roleIds,
  }) as Promise<any>
}
