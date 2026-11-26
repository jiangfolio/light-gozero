import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import * as authApi from '@/api/system/auth'
import {
  clearToken,
  getToken,
  getUserInfo,
  setToken,
  setUserInfo,
} from '@/lib/auth-token'
import { encryptPassword } from '@/lib/crypto/sm2'

import { usePermissionStore } from './permission'
import { useTabsStore } from './tabs'

type LoginUser = {
  id: string
  username: string
  nickname: string
  org_id: string
}

/**
 * 登录态：token、用户信息、登录/退出流程。
 */
export const useUserStore = defineStore('user', () => {
  const token = ref(getToken())
  const userInfo = ref<LoginUser | null>(getUserInfo<LoginUser>())

  const isLogin = computed(() => Boolean(token.value))

  /** 拉取公钥后 SM2 加密密码并登录，成功后写入 token。 */
  async function login(payload: {
    username: string
    password: string
    captcha_id: string
    captcha_code: string
  }) {
    // 1. 取公钥并加密密码
    const { public_key } = await authApi.fetchPublicKey()
    const cipher = encryptPassword(payload.password, public_key)
    // 2. 提交登录
    const data = await authApi.login({
      username: payload.username,
      password: cipher,
      captcha_id: payload.captcha_id,
      captcha_code: payload.captcha_code,
    })
    // 3. 持久化 token 与用户信息
    setToken(data.token)
    setUserInfo(data.user)
    token.value = data.token
    userInfo.value = data.user
    return data
  }

  /** 退出：尽量调后端注销，并清空本地路由/权限/页签。 */
  async function logout(callApi = true) {
    // 1. 有 token 时尝试通知后端（失败忽略）
    if (callApi && token.value) {
      try {
        await authApi.logout()
      } catch {
        // 忽略网络失败，本地仍要清会话
      }
    }
    // 2. 清本地会话与权限缓存
    clearToken()
    token.value = ''
    userInfo.value = null
    usePermissionStore().reset()
    useTabsStore().reset()
  }

  function hydrateFromStorage() {
    token.value = getToken()
    userInfo.value = getUserInfo<LoginUser>()
  }

  return { token, userInfo, isLogin, login, logout, hydrateFromStorage }
})
