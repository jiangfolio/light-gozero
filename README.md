<div align="center">

# light-gozero

**基于 go-zero 的国密友好快速开发脚手架**

多服务 Monorepo · Spec-First（`.api` + goctl）· Redis 不透明会话 · RBAC / 数据范围 · **满足基础等保要求**

[![Go](https://img.shields.io/badge/Go-1.27+-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![go-zero](https://img.shields.io/badge/go--zero-1.9.2-blue)](https://go-zero.dev/)
[![goctl](https://img.shields.io/badge/goctl-1.9.2-5C6BC0)](https://go-zero.dev/)
[![GORM](https://img.shields.io/badge/GORM-MySQL-red)](https://gorm.io/)
[![Redis](https://img.shields.io/badge/Redis-Session%20%7C%20Cache-DC382D?logo=redis&logoColor=white)](https://redis.io/)
[![ai-context](https://img.shields.io/badge/ai--context-zeromicro-black)](https://github.com/zeromicro/ai-context)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](./LICENSE)

</div>

> 工程约定与 AI 协作流程对齐 [zeromicro/ai-context](https://github.com/zeromicro/ai-context)：先写 `.api`，再用 goctl 生成，logic 填业务；密码传输走国密 SM2（对外不暴露算法细节），落库 bcrypt。内置身份鉴别、访问控制、安全审计、入侵防范等**基础等保**常见技术控制能力（面向内网 / 二级场景基线，正式测评还需结合部署与管理制度）。

---

## 目录

- [特性](#特性)
- [基础等保能力](#基础等保能力)
- [技术栈](#技术栈)
- [目录结构](#目录结构)
- [环境依赖](#环境依赖)
- [快速开始](#快速开始)
- [配置说明](#配置说明)
- [API 约定](#api-约定)
- [AI 协作（ai-context）](#ai-协作ai-context)
- [冒烟与安全脚本](#冒烟与安全脚本)
- [说明](#说明)
- [License](#license)

## 特性

| 模块 | 说明 |
|------|------|
| 认证会话 | 图形验证码登录；Redis **不透明 Token**（禁止 `Bearer`）；失败锁定、强制改密、会话踢出 |
| 国密传输 | `GET /auth/public_key` + 登录/建用户/改密密码密文；内部 SM2，对外仅称「密码密文」 |
| RBAC | 资源树 directory/menu/button；权限码 `module:resource:action`；logic 内 `RbacGuard.Require` |
| 数据范围 | 角色 `data_scope` 1～5 + 指定部门；用户/组织读写按范围校验 |
| 系统管理 | 组织 / 用户 / 角色 / 资源 / **单表字典** / 动态配置 `sys_config` |
| 审计 | 登录日志、操作日志；logic 可写中文摘要；敏感字段脱敏 |
| 运行时防护 | go-zero MaxBytes / Timeout / Shedding / Breaker；登录·验证码·公钥周期限流 |
| 脚手架形态 | `system` 为底座服务；`demoservice` 为占位示例；可按同模式横向加业务服务 |

## 基础等保能力

面向常见**二级等保**应用侧技术要求做了开箱能力（策略多可在 `sys_config` 热调）。TLS 终止、物理/网络分区、管理制度等仍需部署与运维侧配合。

| 控制点 | 本脚手架落地 |
|--------|----------------|
| 身份鉴别 | 验证码防机器人；密码复杂度 / 最小长度 / 历史密码 / 过期强制改密；连续失败锁定 |
| 访问控制 | RBAC 菜单按钮权限；角色数据范围；越权 IDOR 在用户/组织读写侧校验 |
| 安全审计 | 登录日志、操作日志；可读中文摘要；密码等字段脱敏 |
| 通信完整性 / 保密性（应用层） | 登录与改密等敏感字段国密密文传输；会话 Token 存 Redis、不落 JWT 明文声明 |
| 入侵防范 | 公开接口周期限流；请求体大小 / 超时 / 自适应卸载；批量 ID 与深分页上限 |
| 数据备份与恢复（应用） | 标准 MySQL / Redis，便于与现有备份体系对接（脚本与策略由运维配置） |

## 技术栈

| 层 | 选型 |
|----|------|
| 框架 | Go 1.27+ · [go-zero](https://github.com/zeromicro/go-zero) 1.9.2 · goctl 1.9.2 |
| 持久化 | MySQL · GORM |
| 缓存 / 会话 | Redis（go-zero redis） |
| 密码 | 传输 SM2（gmsm）· 存储 bcrypt |
| ID | UUIDv7（可选 auto_increment DDL） |
| AI 上下文 | [ai-context](https://github.com/zeromicro/ai-context)（本仓库 submodule：`.cursorrules`） |

## 目录结构

```text
.
├── .cursorrules/              # zeromicro/ai-context（workflows / patterns / goctl 约定）
├── common/                    # 跨服务公共库（authctx / errorx / rbac / pwdcipher / pageutil …）
├── deploy/sql/                # uuidv7 / auto_increment 初始化脚本
├── scripts/                   # HTTP 冒烟、安全用例、pwdenc
└── service/
    ├── system/api/            # 系统底座（鉴权 / RBAC / 字典 / 配置 / 审计）
    │   ├── desc/              # .api 契约（auth org user role resource dict config audit）
    │   ├── etc/system.yaml    # 本地配置（生产用环境变量覆盖密钥）
    │   └── internal/          # handler · logic · model · middleware · svc
    └── demoservice/api/       # 最小示例服务
```

## 环境依赖

| 依赖 | 说明 |
|------|------|
| Go 1.27+ | 后端；`goctl@v1.9.2` |
| MySQL 8+ | 库名默认 `light_gozero` |
| Redis | 会话、验证码、RBAC 缓存、限流 |

默认约定本机已有（可与 Docker 容器同名）：

| 服务 | 端口 | 示例凭证 |
|------|------|----------|
| MySQL | 3306 | root / `infra123!`，库 `light_gozero` |
| Redis | 6379 | 密码 `infra123!` |

## 快速开始

```bash
git clone <本仓库> && cd light-gozero   # 或当前工程根目录
git submodule update --init --recursive

# 1. 初始化库表（UUIDv7 模式含种子数据）
docker exec -i mysql mysql -uroot -p'infra123!' < deploy/sql/light_gozero_init_uuidv7.sql
# 或本机：mysql -uroot -p'infra123!' < deploy/sql/light_gozero_init_uuidv7.sql

# 2. 启动 system
cd service/system/api
go run . -f etc/system.yaml
```

| 项 | 地址 |
|----|------|
| API | http://127.0.0.1:8888 |
| Prometheus | http://127.0.0.1:9101/metrics |

| 账号 | 密码 | 说明 |
|------|------|------|
| `superadmin` | `Admin@123456` | 超级管理员（种子数据）；上线后请立刻修改 |

登录需先取验证码与公钥，密码字段传 **密文**（可用 `scripts/pwdenc` 本地加密）：

```bash
# 验证码
curl -s 'http://127.0.0.1:8888/api/v1/auth/captcha'
# 公钥
curl -s 'http://127.0.0.1:8888/api/v1/auth/public_key'
# 登录（password 为密文 hex）
curl -s -X POST 'http://127.0.0.1:8888/api/v1/auth/login' \
  -H 'Content-Type: application/json' \
  -d '{"username":"superadmin","password":"<cipher>","captcha_id":"...","captcha_code":"..."}'
```

鉴权头为原始 Token：`Authorization: <token>`（不要加 `Bearer `）。

## 配置说明

主配置：[`service/system/api/etc/system.yaml`](service/system/api/etc/system.yaml)。

| 项 | 作用 |
|----|------|
| `IdGen.Mode` | `uuidv7`（默认）或 `auto_increment`（配套另一份 DDL） |
| `PwdCipher.PrivateKey` | 传输解密私钥；可用 `SYSTEM_PWD_PRIVATE_KEY` 覆盖 |
| `Mysql.DataSource` | 可用 `SYSTEM_MYSQL_DATASOURCE` 覆盖 |
| `TrustProxy` | 默认 `false`；仅在可信反代后才信 `X-Forwarded-For` |
| `RateLimit.*` | 登录 / 验证码 / 公钥周期配额 |
| `MaxBytes` / `Timeout` / `CpuThreshold` | go-zero Rest 防护 |

运行时策略（Token TTL、锁定阈值、密码复杂度等）在表 `sys_config`，管理端可热更新；`GET /config/public` 仅返回白名单键。

本地覆盖文件可放 `etc/system-local.yaml`（已 gitignore），勿把生产密钥提交进库。

## API 约定

- 前缀：`/api/v1`
- 方法：查询 **GET**，写操作 **POST**；参数优先 query / body，少用 path 变量
- 数值字段对外一律 **string**（防 JS 大整数精度问题）
- 统一包络：`{ "code": 0, "message": "成功", "data": ... }`；业务错误仍 HTTP 200，看 `code` / `message`
- 分页：必传 `page` / `page_size`，`page_size ≤ 100`
- 字典：管理 CRUD 需 `system:dict:*`；`GET /dict/by_code` 仅需登录（弱鉴权只读）

主要分组：`/auth` · `/org` · `/user` · `/role` · `/resource` · `/dict` · `/config` · `/audit`。

新增接口：改 `service/system/api/desc/**/*.api` → `goctl api go` → 补全 logic / model。

## AI 协作（ai-context）

本仓库通过 submodule 引入 [zeromicro/ai-context](https://github.com/zeromicro/ai-context)（路径 `.cursorrules`），开发时按其中约定：

1. **Spec-First**：契约先改 `.api`，再 `goctl api validate` / `goctl api go`
2. **完整实现**：logic 写满业务，不要留 stub
3. **验证**：`go build ./...`；需要时跑冒烟脚本
4. 详细流程见 [`.cursorrules/00-instructions.md`](.cursorrules/00-instructions.md)、[`workflows.md`](.cursorrules/workflows.md)、[`patterns.md`](.cursorrules/patterns.md)

克隆后请执行：

```bash
git submodule update --init --recursive
```

## 冒烟与安全脚本

依赖本机已启动的 system API、MySQL、Redis（脚本默认通过 `docker exec` 读验证码）：

```bash
# 功能冒烟（覆盖主要成功路径）
bash scripts/smoke_system_http.sh

# 安全用例（未登录、坏 Token、分页/批量上限、登录限流等）
bash scripts/sec_system_http.sh
```

密码加密小工具：

```bash
go run ./scripts/pwdenc <private_key_hex> '<plain_password>'
```

## 说明

- 「满足基础等保要求」指应用侧已覆盖身份鉴别、访问控制、审计、传输密文、限流等常见技术基线；**不构成**等保测评结论，正式过保需按等保对象与测评机构要求补齐网络、主机、管理制度与证据。
- TLS / CORS 建议由网关终止；本进程默认明文监听，便于内网联调。
- `demoservice` 仅作 go-zero 服务脚手架示例，无业务鉴权。
- 仓库内 yaml 默认密钥仅供本地演示；生产必须用环境变量或私有配置注入并轮换。

## License

本项目基于 [Apache License 2.0](LICENSE) 开源。完整条款见 [LICENSE](LICENSE)，版权声明见 [NOTICE](NOTICE)。

---

基于 [go-zero](https://github.com/zeromicro/go-zero) · 工程约定参考 [ai-context](https://github.com/zeromicro/ai-context)
