/// <reference types="vite/client" />

declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<object, object, any>
  export default component
}

declare module 'sm-crypto' {
  export const sm2: {
    doEncrypt: (msg: string, publicKey: string, cipherMode?: number) => string
    doDecrypt: (cipher: string, privateKey: string, cipherMode?: number) => string
  }
}
