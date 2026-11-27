import { http } from '@/lib/http'

export function createNotice(data: any) {
  return http.post('/api/system/v1/notice/create', data) as Promise<any>
}

export function updateNotice(data: any) {
  return http.post('/api/system/v1/notice/update', data) as Promise<any>
}

export function deleteNotice(ids: any) {
  return http.post('/api/system/v1/notice/delete', { ids }) as Promise<any>
}

export function getNoticeDetail(id: any) {
  return http.get('/api/system/v1/notice/detail', { params: { id } }) as Promise<any>
}

export function pageNotice(params: any) {
  return http.get('/api/system/v1/notice/page', { params }) as Promise<any>
}

/** 登录用户可读：仅已发布 */
export function pagePublishedNotice(params: any) {
  return http.get('/api/system/v1/notice/published_page', { params }) as Promise<any>
}

export function getPublishedNoticeDetail(id: any) {
  return http.get('/api/system/v1/notice/published_detail', { params: { id } }) as Promise<any>
}
