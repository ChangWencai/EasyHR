# Phase 11 Plan 01: 后端签署验证与中文PDF

## 执行摘要

中文PDF劳动合同生成 + 手机验证码电子签署流程后端实现。老板发起签署 → 员工手机号+6位验证码完成签署 → 3天未签自动提醒。

## 元数据

| 字段 | 值 |
|------|-----|
| Phase | 11 |
| Plan | 01 |
| Subsystem | employee (合同模块) |
| 标签 | contract, pdf, signing, sms |
| 依赖 | SMS/OSS 客户端（已有） |
| 影响 | 新增 contract_sign_codes 表 |
| 技术栈 | Go 1.25 + go-pdf/fpdf v0.9.0 + embed.FS |
| 完成时间 | 2026-04-20T03:15:43Z |
| 执行时长 | 857s (~14min) |

## 任务完成情况

| # | 任务 | Commit | 文件 |
|---|------|--------|------|
| 1 | 下载中文字体文件 | 78efacd | assets/fonts/*.ttf, internal/employee/fonts/*.ttf |
| 2 | 重写 pdf.go 实现中文PDF生成 | 78efacd | internal/employee/pdf.go |
| 3 | 实现签署验证码模型与DTO | de5137b | contract_model.go, contract_dto.go, contract_repository.go |
| 4 | 实现签署验证服务逻辑 | 66ce6ea | contract_service.go, repository.go, main.go |
| 5 | 注册签署验证HTTP端点 | 787a1f8 | contract_handler.go, main.go |
| 6 | 注册3天未签提醒定时任务 | 709da9a | todo/scheduler.go |

## Commits

```
78efacd feat(11-01): 添加中文字体并重写PDF生成支持3种合同模板
de5137b feat(11-01): 添加签署验证码模型、DTO和Repository方法
66ce6ea feat(11-01): 实现签署验证码服务逻辑与依赖注入
787a1f8 feat(11-01): 注册签署验证HTTP端点
709da9a feat(11-01): 注册3天未签提醒定时任务
```

## 关键决策

| 决策 | 理由 |
|------|------|
| 字体文件同时存于 assets/fonts/ 和 internal/employee/fonts/ | assets/fonts/ 为项目资源目录，internal/employee/fonts/ 供 Go embed 指令使用 |
| 使用 AddUTF8FontFromBytes 动态加载字体 | go-pdf/fpdf v0.9.0 的 AddUTF8FontFromBytes 直接接收 TTF 字节，无需预编译 JSON/z 文件 |
| 新增 FindByPhoneHashGlobal 跨租户查询 | 签署端点无认证，通过手机号哈希跨租户查找员工（phone_hash 全局唯一索引） |
| FindByPhoneHashGlobal 替代方案 | 不使用 TenantScope(orgID=0)，避免多租户隔离逻辑干扰 |
| ContractSignCode 独立表 | 验证码数据与合同主表分离，避免泄露合同信息 |
| OSS 客户端 nil-safe 处理 | SMS/OSS 客户端可能未配置，方法内做空检查不影响核心流程 |

## 功能实现

### 中文PDF生成（3种合同模板）

- **劳动合同（fixed_term）**：7条条款，含合同期限/工作内容/薪资/社保/劳动保护/生效
- **实习协议（intern）**：6条条款，含实习期限/内容/补贴/保险/协议解除
- **兼职合同（indefinite）**：7条条款，非全日制用工，4小时/日限制，工伤险

字体：NotoSansSC Regular/Bold，嵌入 Go 二进制（~16MB）

### 签署验证流程

```
老板发起 → POST /contracts/:id/send-sign-link
           ↓ (生成PDF → 上传OSS → 发短信)
员工收到短信 → POST /contracts/sign/send-code (手机号+合同ID)
           ↓ (发送6位验证码)
员工输入验证码 → POST /contracts/sign/verify-code (手机号+合同ID+验证码)
           ↓ (返回 sign_token, 有效期30分钟)
确认签署 → POST /contracts/sign/confirm (合同ID+sign_token)
           ↓ (更新合同状态: pending_sign → signed/active)
```

### 签署端点（无认证，phone隔离租户）

| 端点 | 方法 | 说明 |
|------|------|------|
| /contracts/sign/send-code | POST | 发送验证码到员工手机 |
| /contracts/sign/verify-code | POST | 校验验证码，返回 sign_token |
| /contracts/sign/confirm | POST | 用 sign_token 确认签署 |
| /contracts/:id/signed-pdf | GET | 获取已签PDF URL |
| /contracts/:id/send-sign-link | POST | 老板发起签署（需认证） |

### 3天未签提醒

- 每日 09:00 CST 扫描所有 `status=pending_sign AND created_at <= 3天前` 的合同
- 为每份合同创建 `contract_pending_sign` 待办（通过 ExistsBySource 防止重复）
- 待办标题：`员工 {name} 的合同已发起签署 {n} 天，员工尚未签署，请跟进`

## 威胁缓解（STRIDE）

| Threat | 缓解措施 |
|--------|----------|
| T-11-01 伪造签署请求 | phone→emp→contract 链验证；3次/小时 频率限制（待前端） |
| T-11-02 验证码暴力破解 | 6位数字+5分钟有效期+锁定；sign_token 64-char hex |
| T-11-03 签署否认 | SignToken 绑定 contract_id+phone，30分钟有效期 |
| T-11-04 信息泄露 | 验证码错误不返回具体原因；合同信息仅在 VerifySignCode 成功后返回 |
| T-11-05 DoS 攻击 | 频率限制（待实现）；无认证但 phone→contract 验证 |
| T-11-06 PDF 未授权访问 | OSS 签名 URL（1小时有效期）；URL 非可猜测 |

## 偏差记录

### Rule 3 - 阻塞问题修复

**[Rule 3 - 阻塞修复] go:embed 路径修正**
- 发现：原始 `../../assets/fonts/` 相对路径导致 `invalid pattern syntax` 错误
- 修复：字体文件同时复制到 `internal/employee/fonts/`，使用 `fonts/NotoSansSC-*.ttf` 路径
- 根因：go:embed 不支持向上多级相对路径

**[Rule 3 - 阻塞修复] fpdf.AddUTF8FontFromBytes 无返回值**
- 发现：Go 文档显示 AddUTF8FontFromBytes 返回 error，但源码为 void 方法
- 修复：移除 `if err := ...` 包装，直接调用 `pdf.AddUTF8FontFromBytes(...)`
- 验证：`go doc` 显示旧签名，`go mod cache` 源码确认实际为 void

**[Rule 3 - 阻塞修复] main.go 缺失 OSS 客户端**
- 发现：NewContractService 新增 smsClient/ossClient 参数，但 main.go 未创建 ossClient
- 修复：新增 `oss.NewClient` 调用，传入 cfg.OSS 配置
- 验证：`go build ./...` 编译通过

**[Rule 3 - 阻塞修复] generateToken 重声明**
- 发现：invitation_service.go 已定义 `generateToken`，直接复用导致 redeclared
- 修复：移除 contract_service.go 中的重复定义，直接使用 invitation_service.go 的版本

**[Rule 3 - 阻塞修复] FindByPhoneHash 租户隔离问题**
- 发现：FindByPhoneHash 使用 TenantScope(orgID=0)，在无认证场景下无法跨租户查找
- 修复：新增 `FindByPhoneHashGlobal` 方法（无 TenantScope），用于签署等无认证流程

## 已知限制

| 限制 | 说明 |
|------|------|
| OSS 上传未完整实现 | uploadPdfToOss 返回 objectKey，实际 OSS 上传由前端直传（需 Phase 11 Plan 02 前端实现） |
| 频率限制未实现 | 验证码发送频率限制（3次/小时/手机号）待后续迭代 |
| 签署链接 URL 不含 token | SendSignCode 生成的签署链接未将 token 存入数据库，仅发短信时生成一次 |

## 待配置

| 配置项 | 来源 | 用途 |
|--------|------|------|
| `ALIYUN_SMS_CONTRACT_TEMPLATE_CODE` | 阿里云短信控制台 | 合同签署短信模板（含 link 占位符） |
| `APP_BASE_URL` | 环境变量 | 生成签署链接域名 |

## 自检

- [x] `go build ./...` 编译通过
- [x] 5个新 commit，每个任务独立
- [x] PDF 使用中文字体（NotoSansSC），3种模板全中文
- [x] 签署验证码：6位数字，5分钟有效期，存储在 contract_sign_codes 表
- [x] 签署流程：send-code → verify-code（返回 sign_token） → confirm（完成签署）
- [x] 合同状态：pending_sign → signed/active
- [x] 3天未签提醒：每日 09:00 CST
- [x] generateToken 在 invitation_service.go 中定义（复用已有）
- [x] ContractSignCode 所有字段有 JSON tag
- [x] TodoCreator.ExistsBySource 已存在（Phase 09 已实现）
- [x] Contract/Employee OrgID 通过 BaseModel 继承已有 JSON tag
