import { http } from '@/lib/http'

export function fetchCaptcha() {
  return http.get('/api/system/v1/auth/captcha') as Promise<any>
}

export function fetchPublicKey() {
  return http.get('/api/system/v1/auth/public_key') as Promise<any>
}

export function login(data: any) {
  return http.post('/api/system/v1/auth/login', data) as Promise<any>
}

export function logout() {
  return http.post('/api/system/v1/auth/logout') as Promise<any>
}

/** 自助改密：old_password/new_password 需先经公钥 SM2 加密 */
export function changePassword(data: any) {
  return http.post('/api/system/v1/auth/change_password', data) as Promise<any>
}
