import type { AxiosInstance, AxiosResponse, InternalAxiosRequestConfig } from 'axios'

const TOKEN_KEY = 'token'

/** 与 common/response.Body 对齐 */
type ApiBody = {
  code: number
  message: string
  data?: unknown
}

export function setupInterceptors(client: AxiosInstance) {
  client.interceptors.request.use((config: InternalAxiosRequestConfig) => {
    const token = localStorage.getItem(TOKEN_KEY)
    if (token) {
      // 后端 Auth 拒绝 Bearer 前缀，直接传原始 token
      config.headers.Authorization = token
    }
    return config
  })

  client.interceptors.response.use(
    (response: AxiosResponse) => {
      const body = response.data as ApiBody
      // 非统一包络（如文件流）原样返回
      if (body == null || typeof body !== 'object' || !('code' in body)) {
        return response.data
      }
      if (body.code === 0) {
        return body.data
      }
      if (body.code === 401) {
        localStorage.removeItem(TOKEN_KEY)
        if (typeof window !== 'undefined' && !window.location.pathname.startsWith('/login')) {
          window.location.href = `/login?redirect=${encodeURIComponent(window.location.pathname + window.location.search)}`
        }
      }
      return Promise.reject({ code: body.code, message: body.message })
    },
    (error) => {
      const message = error?.message || '网络异常'
      return Promise.reject({ code: -1, message })
    },
  )
}
