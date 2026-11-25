import axios, { type AxiosInstance, type CreateAxiosDefaults } from 'axios'

import { setupInterceptors } from './interceptors'

/** 创建 Axios 实例；默认无 baseURL，完整路径由业务请求文件自行书写。 */
export function createHttpClient(options: CreateAxiosDefaults = {}): AxiosInstance {
  const client = axios.create({
    timeout: 30000,
    ...options,
  })
  setupInterceptors(client)
  return client
}
