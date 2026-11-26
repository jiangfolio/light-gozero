import { http } from '@/lib/http'

export function pageAuditLogin(params: any) {
  return http.get('/api/system/v1/audit/login_page', { params }) as Promise<any>
}

export function pageAuditOper(params: any) {
  return http.get('/api/system/v1/audit/oper_page', { params }) as Promise<any>
}
