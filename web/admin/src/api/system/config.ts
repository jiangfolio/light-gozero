import { http } from '@/lib/http'

export function pageConfig(params: any) {
  return http.get('/api/system/v1/config/page', { params }) as Promise<any>
}

export function getConfigDetail(params: any) {
  return http.get('/api/system/v1/config/detail', { params }) as Promise<any>
}

export function updateConfig(items: any) {
  return http.post('/api/system/v1/config/update', { items }) as Promise<any>
}
