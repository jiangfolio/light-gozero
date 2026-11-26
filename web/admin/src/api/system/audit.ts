import { http } from '@/lib/http'

export function pageAuditLogin(params: any) {
  return http.get('/api/system/v1/audit/login_page', { params }) as Promise<any>
}

export function pageAuditOper(params: any) {
  return http.get('/api/system/v1/audit/oper_page', { params }) as Promise<any>
}

/** 我的登录日志：用户名由后端取自登录态，无需传参 */
export function pageMyAuditLogin(params: any) {
  return http.get('/api/system/v1/audit/my_login_page', { params }) as Promise<any>
}

/** 我的操作日志：用户名由后端取自登录态，无需传参 */
export function pageMyAuditOper(params: any) {
  return http.get('/api/system/v1/audit/my_oper_page', { params }) as Promise<any>
}
