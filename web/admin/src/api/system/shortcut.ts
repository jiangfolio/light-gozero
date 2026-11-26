import { http } from '@/lib/http'

/** 当前用户快捷入口列表（非 page） */
export function listShortcut() {
  return http.get('/api/system/v1/shortcut/list') as Promise<any>
}

export function createShortcut(data: any) {
  return http.post('/api/system/v1/shortcut/create', data) as Promise<any>
}

export function updateShortcut(data: any) {
  return http.post('/api/system/v1/shortcut/update', data) as Promise<any>
}

export function deleteShortcut(ids: any) {
  return http.post('/api/system/v1/shortcut/delete', { ids }) as Promise<any>
}
