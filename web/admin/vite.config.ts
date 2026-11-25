import { AntdvNextResolver } from '@antdv-next/auto-import-resolver'
import tailwindcss from '@tailwindcss/vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'
import AutoImport from 'unplugin-auto-import/vite'
import IconsResolver from 'unplugin-icons/resolver'
import Icons from 'unplugin-icons/vite'
import Components from 'unplugin-vue-components/vite'
import { defineConfig } from 'vite'

export default defineConfig({
  plugins: [
    vue(),
    tailwindcss(),
    // Vue / Router / Pinia API 自动导入
    AutoImport({
      imports: ['vue', 'vue-router', 'pinia'],
      dts: 'src/auto-imports.d.ts',
      vueTemplate: true,
    }),
    // antdv-next 组件 + IconPark（Iconify JSON，npm 离线）按需自动导入
    Components({
      dts: 'src/components.d.ts',
      resolvers: [
        AntdvNextResolver(),
        IconsResolver({
          prefix: 'Icon',
          // IconParkHome → @iconify-json/icon-park-outline
          alias: {
            park: 'icon-park-outline',
            parkSolid: 'icon-park-solid',
            parkFilled: 'icon-park',
          },
          enabledCollections: ['icon-park-outline', 'icon-park-solid', 'icon-park'],
        }),
      ],
    }),
    // 禁止运行时联网拉 icon；仅用已安装的 @iconify-json/*
    Icons({
      compiler: 'vue3',
      autoInstall: false,
    }),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    port: 8081,
    proxy: {
      // /api/system/v1/... → system :8888 /api/v1/...
      '/api/system': {
        target: 'http://127.0.0.1:8888',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api\/system/, '/api'),
      },
      // /api/demo/v1/... → demoservice :8889 /api/v1/...
      '/api/demo': {
        target: 'http://127.0.0.1:8889',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api\/demo/, '/api'),
      },
    },
  },
})
