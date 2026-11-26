import { http } from '@/lib/http'

export function pageSecurityConfig(params: any) {
  return http.get('/api/system/v1/config/security/page', { params }) as Promise<any>
}

export function updateSecurityConfig(items: any) {
  return http.post('/api/system/v1/config/security/update', { items }) as Promise<any>
}
