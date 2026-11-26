import { http } from '@/lib/http'

export function fetchCaptcha() {
  return http.get('/api/system/v1/auth/captcha') as Promise<{
    captcha_id: string
    captcha_image: string
  }>
}

export function fetchPublicKey() {
  return http.get('/api/system/v1/auth/public_key') as Promise<{ public_key: string }>
}

export function login(data: {
  username: string
  password: string
  captcha_id: string
  captcha_code: string
}) {
  return http.post('/api/system/v1/auth/login', data) as Promise<{
    token: string
    must_change_password: string
    user: { id: string; username: string; nickname: string; org_id: string }
  }>
}

export function logout() {
  return http.post('/api/system/v1/auth/logout') as Promise<unknown>
}
