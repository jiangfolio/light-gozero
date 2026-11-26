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
