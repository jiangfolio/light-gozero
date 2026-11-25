import { createHttpClient } from './createClient'

/** 默认 HTTP 客户端：无服务前缀，路径写在各 src/api/** 请求文件中 */
export const http = createHttpClient()

export { createHttpClient } from './createClient'
