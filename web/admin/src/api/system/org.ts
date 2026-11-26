import { http } from '@/lib/http'

export function createOrg(data: any) {
  return http.post('/api/system/v1/org/create', data) as Promise<any>
}

export function updateOrg(data: any) {
  return http.post('/api/system/v1/org/update', data) as Promise<any>
}

export function deleteOrg(ids: any) {
  return http.post('/api/system/v1/org/delete', { ids }) as Promise<any>
}

export function getOrgDetail(id: any) {
  return http.get('/api/system/v1/org/detail', { params: { id } }) as Promise<any>
}

export function pageOrg(params: any) {
  return http.get('/api/system/v1/org/page', { params }) as Promise<any>
}

export function treeOrg(params?: any) {
  return http.get('/api/system/v1/org/tree', { params }) as Promise<any>
}
