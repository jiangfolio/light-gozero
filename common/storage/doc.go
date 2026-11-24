// Package storage 提供可切换的对象存储抽象：local 分片网盘与 S3 兼容预签名直传。
//
// # 配置切换
//
// 通过 yaml Storage.Provider=local|s3 与 AccessMode=private|public 切换。
// 默认 AccessMode=private。AccessMode=public 时必须配置 https 的 PublicBaseURL
// （可用 AllowInsecurePublicURL 仅供本地调试）。
//
// # 上传语义（两端统一）
//
//  Initiate → PreparePart / PutPart → ListParts → Complete / Abort
//
// local：字节经服务端 PutPart，边写边算 SHA-256，Complete 仅聚合 Merkle，不重读正文。
// s3：PreparePart 返回窄权限 UploadPart 预签名；PutPart 返回 ErrDirectUploadOnly。
//
// # 读访问
//
//  private + local：鉴权后 Open 流式，禁止把 Local.Root 做 Nginx 静态直挂。
//  private + s3：PresignGet 短 TTL。
//  public + local：PublicURL 为稳定 HTTP API 直链（无 TTL/签名），形如
//    {PublicBaseURL}/api/v1/file/public?key=...
//    由后续 file 模块无登录 handler 内部 Open；安全依赖 key 校验与磁盘哈希路径屏蔽，无 PublicRoot。
//  public + s3：PublicURL 指向 CDN/桶公网域，无签名。
//
// # 完整性
//
// 分片强制 SHA-256；整文件 Merkle = SHA256(canonical(chunkSize, count, leaves))。
// FullContentHash 默认关闭；开启时仅 local 合并流顺便计算 content SHA-256。
//
// # 后续 file 模块契约（本包不实现 HTTP）
//
// API：
//   POST /file/upload/init
//   GET  /file/upload/parts
//   POST /file/upload/part_url
//   PUT  /file/upload/part          // 仅 local raw；MaxBytes=ChunkSize+margin
//   POST /file/upload/complete
//   POST /file/upload/abort
//   GET  /file/download_url         // 需登录
//   GET  /file/public               // 稳定公开直链入口，无 TTL 参数
//
// 表草案：file_object（owner, merkle, visibility, ...）、file_upload、file_upload_part。
// 会话须绑定 created_by；Part/Complete/Abort 校验所有者；按 user/IP 限流。
// 对象 visibility 与桶策略前缀一致；private 对象永不返回无签名永久链接。
//
// # 运维检查清单
//
//  - S3 private：Block Public Access；IAM 最小权限。
//  - S3 public：建议仅 public/* 前缀匿名 Get；写仍鉴权。
//  - local：Root 目录 0700/文件 0600；禁止静态 alias 到 Root。
//  - 预签名 URL 与 Secret 禁止打 info 日志。
//  - GCExpired 由调用方定时触发，本包不自启 goroutine。
package storage
