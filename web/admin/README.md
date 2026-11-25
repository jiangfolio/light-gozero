# light-gozero admin

Vue 3 + TypeScript + Vite + Tailwind + Pinia + Axios + antdv-next + IconPark。

```bash
# 仓库根目录
make web-install
make web-dev   # http://127.0.0.1:8081
```

- 开发代理：`/api/system` → `:8888`，`/api/demo` → `:8889`
- HTTP 默认无 baseURL；完整路径写在 `src/api/**` 请求文件
- 响应解包对齐后端 `code` / `message` / `data`
- 自动导入：`unplugin-vue-components` + `@antdv-next/auto-import-resolver`（模板用 `AButton` / `a-button`）；图标 `unplugin-icons` + `@iconify-json/icon-park*`（离线），模板内 `<IconParkHome />`
