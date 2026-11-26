import { sm2 } from 'sm-crypto'

/**
 * 使用后端公钥按 C1C3C2 加密明文密码，返回 hex 密文。
 * sm-crypto 要求公钥带 04 前缀；密文需补 04 才能被 gmsm 解密。
 */
export function encryptPassword(plain: string, publicKeyHex: string) {
  let key = publicKeyHex.trim().toLowerCase()
  if (!key.startsWith('04')) {
    key = `04${key}`
  }
  // cipherMode 1 = C1C3C2，与后端 gmsm 一致
  const cipher = sm2.doEncrypt(plain, key, 1)
  return cipher.startsWith('04') ? cipher : `04${cipher}`
}
