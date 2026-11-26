import { http } from '@/lib/http'

export function createResource(data: any) {
  return http.post('/api/system/v1/resource/create', data) as Promise<any>
}

export function updateResource(data: any) {
  return http.post('/api/system/v1/resource/update', data) as Promise<any>
}

export function deleteResource(ids: any) {
  return http.post('/api/system/v1/resource/delete', { ids }) as Promise<any>
}

export function getResourceDetail(id: any) {
  return http.get('/api/system/v1/resource/detail', { params: { id } }) as Promise<any>
}

export function pageResource(params: any) {
  return http.get('/api/system/v1/resource/page', { params }) as Promise<any>
}

/** 按 parent_id 一次性拉子资源（按钮权限等） */
export function listResource(params: any) {
  return http.get('/api/system/v1/resource/list', { params }) as Promise<any>
}

export function treeResource(params?: any) {
  return http.get('/api/system/v1/resource/tree', { params }) as Promise<any>
}
