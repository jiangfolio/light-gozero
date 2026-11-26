const TOKEN_KEY = 'token'
const USER_INFO_KEY = 'user_info'
const REMEMBER_USER_KEY = 'remember_username'

export function getToken() {
  return localStorage.getItem(TOKEN_KEY) || ''
}

export function setToken(token: string) {
  localStorage.setItem(TOKEN_KEY, token)
}

export function clearToken() {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(USER_INFO_KEY)
}

export function getUserInfo<T = unknown>() {
  try {
    const raw = localStorage.getItem(USER_INFO_KEY)
    return raw ? (JSON.parse(raw) as T) : null
  } catch {
    return null
  }
}

export function setUserInfo(info: unknown) {
  if (info == null) localStorage.removeItem(USER_INFO_KEY)
  else localStorage.setItem(USER_INFO_KEY, JSON.stringify(info))
}

export function getRememberUsername() {
  return localStorage.getItem(REMEMBER_USER_KEY) || ''
}

export function setRememberUsername(username: string) {
  if (username) localStorage.setItem(REMEMBER_USER_KEY, username)
  else localStorage.removeItem(REMEMBER_USER_KEY)
}
