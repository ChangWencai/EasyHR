# Phase 2: 员工管理 - Context

**Gathered:** 2026-04-06
**Status:** Ready for planning

<domain>
## Phase Boundary

老板可在3步内完成员工入职，集中管理员工档案（搜索+导出Excel），办理离职并自动生成交接清单、触发社保停缴提醒。合同管理（V1.0降级：PDF模板生成+手动签署上传）。覆盖 EMPL-01 ~ EMPL-08 全部需求。

</domain>

<decisions>
## Implementation Decisions

### 入职邀请机制
- **D-01:** 老板创建入职邀请时，后端生成唯一 invite_token，返回可分享的链接和二维码。链接格式：`/invite/{token}`。
- **D-02:** 邀请链接打开 H5 页面（无需下载APP），员工直接在网页填写基本信息（姓名、手机号、身份证号、岗位、入职日期）。
- **D-03:** 员工提交信息后，创建 Employee 记录，状态为"待入职"。老板在员工列表中确认入职后，状态转为"在职"（或"试用期"）。
- **D-04:** 邀请链接有效期 7 天。过期后不可提交，需老板重新生成。invite_token 存储于数据库（非 Redis），含过期时间字段。
- **D-05:** 邀请填写阶段不创建 User 账号。Employee 独立于 User，后续如需开通登录权限再关联。

### 员工数据模型
- **D-06:** Employee 独立模型，与 User 分离。可选 user_id 外键（int64, nullable）关联 User 模型。逻辑：不是所有员工都需要登录系统。
- **D-07:** 员工状态生命周期：待入职（pending）→ 试用期（probation）→ 在职（active）→ 离职（resigned）。使用枚举字段 `status`（varchar(20)）。
- **D-08:** 员工档案字段分为基础信息和扩展信息：
  - 基础信息（必填）：姓名、手机号、身份证号、岗位、入职日期
  - 扩展信息（选填）：性别（从身份证自动提取）、出生日期（从身份证提取）、银行卡号、紧急联系人+电话、住址、备注
- **D-09:** 敏感字段沿用 Phase 1 双列模式：
  - 身份证号：id_card_encrypted（AES-256-GCM）+ id_card_hash（SHA-256，唯一索引 WHERE deleted_at IS NULL）
  - 手机号：phone_encrypted + phone_hash（复用 Phase 1 的 crypto 包）
  - 银行卡号：bank_account_encrypted + bank_account_hash
- **D-10:** API 响应中敏感字段返回脱敏数据。复用 `crypto.MaskPhone()`、`crypto.MaskIDCard()`。银行卡号脱敏规则：保留后4位（`****5678`）。

### 员工档案管理
- **D-11:** 员工列表支持分页 + 搜索。搜索维度：姓名（模糊）、岗位（模糊）、手机号（精确，通过 hash 查询）、状态（精确筛选）。
- **D-12:** 导出 Excel 使用 excelize 库。导出字段与列表一致，敏感字段导出脱敏数据（与API响应一致，符合数据安全原则）。仅 OWNER 和 ADMIN 可导出。
- **D-13:** 员工详情页展示全部信息（脱敏），OWNER/ADMIN 可查看完整敏感信息（点击"查看"按钮后临时解密返回）。

### 离职流程
- **D-14:** 双方均可发起离职：
  - 老板直接办理：选择员工 → 填写离职日期+原因 → 确认 → 自动执行离职流程
  - 员工申请（V1.0 为 H5 页面）：员工提交离职申请（日期+原因）→ 老板审批 → 审批通过后执行离职流程
- **D-15:** 交接清单为模板化生成，包含固定分类：资产归还（电脑、门禁卡等）、工作交接（项目/任务清单）、权限回收（系统账号、钥匙等）。老板可编辑补充具体内容。
- **D-16:** 离职完成后：员工状态改为"离职"，记录离职日期（resignation_date）和离职原因（resignation_reason）。数据归档不删除，仍可搜索查看。
- **D-17:** 离职操作通过事件机制触发后续流程。V1.0 实现方式：在 service 层直接调用（同步），预留事件接口供 Phase 3（社保停缴提醒）接入。具体：`employee.OnResigned(employeeID)` 方法，Phase 3 实现具体逻辑。

### 合同管理
- **D-18:** V1.0 合同流程：老板创建合同 → 系统生成 PDF 模板（预填充员工和企业信息）→ 老板下载打印 → 线下双方签署 → 上传签署后的扫描件至 OSS → 合同状态更新为"已签署"。
- **D-19:** 合同状态：草稿（draft）→ 待签署（pending_sign）→ 已签署（signed）→ 履行中（active）→ 已终止（terminated）/ 已到期（expired）。
- **D-20:** 一个员工可有多份合同（如续签）。合同模型包含：employee_id（外键）、合同类型（固定期限/无固定期限/实习）、起始日期、终止日期、PDF 文件 OSS URL、签署扫描件 OSS URL、状态。
- **D-21:** PDF 模板使用 go-pdf/fpdf 库生成。模板内容为标准劳动合同格式，自动填充：企业名称、统一社会信用代码、员工姓名、身份证号、岗位、薪资、合同期限、签署日期。模板固定，不支持自定义。

### RBAC 权限（员工模块）
- **D-22:** 员工模块权限分配：
  - OWNER：全部操作（入职、编辑、离职、合同管理、导出、查看敏感信息）
  - ADMIN：同 OWNER（入职、编辑、离职、合同管理、导出）
  - MEMBER：仅查看员工列表（脱敏数据），不可导出，不可查看敏感信息详情

### Claude's Discretion
- 员工模块内部目录结构（可拆分为 employee/、contract/ 子包或合为一个包）
- 邀请 H5 页面是否复用现有 H5 管理后台框架
- 交接清单的编辑交互细节
- 搜索性能优化（是否需要全文索引）
- 合同 PDF 模板的具体排版和样式
- 员工头像上传（可选功能，是否在 V1.0 实现）

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### 项目规范
- `prd.md` — 产品需求文档，V1.0功能范围、验收标准
- `ui-ux.md` — UI/UX设计原型，员工管理页面布局、交互流程
- `tech-architecture.md` — 技术架构设计、数据模型、模块结构、API设计

### Phase 1 上下文
- `.planning/phases/01-foundation-auth/1-CONTEXT.md` — Phase 1 全部决策（三层架构、多租户、加密、RBAC 等）

### 研究报告
- `.planning/research/STACK.md` — 技术栈选型（excelize v2.10.1, go-pdf/fpdf v0.9.0 等）
- `.planning/research/ARCHITECTURE.md` — 架构设计建议、模块边界
- `.planning/research/PITFALLS.md` — 常见陷阱（敏感数据加密、界面复杂度）

### 需求追踪
- `.planning/REQUIREMENTS.md` — EMPL-01 ~ EMPL-08 需求定义
- `.planning/ROADMAP.md` — Phase 2 定义和成功标准

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/common/crypto/` — AES-256-GCM 加密/解密、SHA-256 哈希、手机号/身份证脱敏（MaskPhone、MaskIDCard）
- `internal/common/response/` — 统一响应封装（Success、Error、PageSuccess）
- `internal/common/middleware/` — 认证、RBAC（RequireRole）、多租户（TenantScope）、限流
- `internal/common/model/base.go` — BaseModel（ID、OrgID、审计字段、软删除），所有 Employee 模型嵌入此基类
- `pkg/oss/client.go` — OSS 签名URL生成，按 bizType/org_id/日期 组织，合同上传复用此组件
- `pkg/jwt/jwt.go` — JWT 工具（Employee 无需，但员工关联 User 时可能用到）
- `internal/user/` — 完整的 handler→service→repository 三层架构参考实现
- `test/testutil/` — 测试工具（SQLite 内存测试、CreateTestOrg、CreateTestUser）

### Established Patterns
- 三层架构：handler（HTTP+路由注册）→ service（业务逻辑+加密）→ repository（GORM+多租户scope）
- 敏感字段双列模式：加密值列 + SHA-256 哈希索引列，API 响应脱敏
- DTO 模式：独立请求/响应结构体，binding tag 做参数校验
- 路由注册：模块的 RegisterRoutes 方法在 main.go 中统一注册
- 审计日志：GORM Hook 自动记录，Module="employee" 区分模块

### Integration Points
- `cmd/server/main.go` AutoMigrate — 新增 Employee、Contract、Invitation 模型
- `cmd/server/main.go` 路由注册 — 新增 employeeHandler.RegisterRoutes(v1, authMiddleware)
- `cmd/server/main.go` 依赖注入 — employeeRepo → employeeSvc → employeeHandler
- OSS 合同文件上传 — bizType="contracts"，复用现有签名URL模式
- 审计日志 — Module="employee"，Action="create"/"update"/"resign"等
- 离职事件接口 — `OnResigned(employeeID)` 方法，Phase 3 社保模块消费

### 新增依赖
- `excelize v2.10.1` — Excel 导出（go get 引入）
- `go-pdf/fpdf v0.9.0` — PDF 合同模板生成（go get 引入）

</code_context>

<specifics>
## Specific Ideas

- 入职邀请链接应同时支持生成二维码，方便面对面扫码填写
- 员工身份证号可自动提取性别和出生日期，减少手动输入
- 交接清单模板固定，但允许老板补充自定义条目
- 离职事件机制预留接口，Phase 3 实现社保停缴具体逻辑

</specifics>

<deferred>
## Deferred Ideas

None — analysis stayed within phase scope

</deferred>

---
*Phase: 02-employee-management*
*Context gathered: 2026-04-06*
