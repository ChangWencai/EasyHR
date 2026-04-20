# Phase 11: 合同合规 - Context

**Gathered:** 2026-04-20
**Status:** Ready for planning

<domain>
## Phase Boundary

劳动合同全生命周期管理——从标准模板选择、员工签署到存档查询。

具体包含：
- 固定标准模板管理（劳动合同/实习协议/兼职合同，无自定义模板）
- 中文 PDF 生成（内嵌字体，固定模板自动分页）
- 签署验证流程（手机号+短信验证码，7天有效期，3天未签自动提醒老板）
- 合同存档与查询（员工详情抽屉内，短信链接→H5签署页，员工无需注册）

**Scope:** H5管理后台（老板端合同管理+员工端H5签署页），后端 API 配合。
**Depends on:** Phase 10

</domain>

<decisions>
## Implementation Decisions

### 模板管理界面（COMP-01）
- **D-11-01:** 提供 3 个固定标准模板：劳动合同（固定期限）、实习协议、兼职合同；不允许用户自定义模板条款
- **D-11-02:** 模板变量仅含基本信息：员工姓名、身份证号、合同期限（起止日期）；薪资/岗位等敏感字段不写入合同正文，避免频繁重签
- **D-11-03:** 合同管理入口放在员工详情抽屉内（EmployeeDrawer Tab），不新增独立菜单；符合员工模块内管理的交互习惯

### 中文 PDF 生成（COMP-02）
- **D-11-04:** PDF 中文字体：内嵌开源中文字体文件（如思源黑体/阿里巴巴普惠体，约 2-4MB），不依赖网络，确保离线签发可用
- **D-11-05:** PDF 布局：固定模板 + 自动分页；字体大小/行间距微调，确保常用场景 2-3 页 A4，超出范围自动分页

### 签署验证流程（COMP-03）
- **D-11-06:** 员工签署方式：手机号 + 短信 6 位验证码；员工无需提前注册微信小程序，收到短信链接即可签署
- **D-11-07:** 签署链接有效期：7 天；超期后老板重新发起签署，短信重新发送
- **D-11-08:** 未签署自动提醒：合同发起后第 3 天，若员工未签署，系统通过待办中心给老板发提醒
- **D-11-09:** 老板发起签署流程（≤ 3 步）：
  - Step 1：选择员工 + 选择合同类型（劳动合同/实习协议/兼职合同）
  - Step 2：预览生成的 PDF，确认内容无误
  - Step 3：点击「发送签署链接」→ 系统上传 OSS → 发送短信给员工

### 合同列表与查询（COMP-04）
- **D-11-10:** 老板在员工详情抽屉内查看该员工的历史合同列表（含类型/期限/状态/到期倒计时）；无需跨员工全局汇总视图
- **D-11-11:** 员工签署方式：短信链接 → 打开 H5 签署页 → 输入手机号+验证码 → 确认签署 → 显示已签 PDF；签署后合同状态变为"已签/生效"
- **D-11-12:** 签署完成后合同 PDF 存档（OSS），老板和员工均可查看；老板可下载，员工可在签署页查看

### Claude's Discretion
- 固定模板的具体条款内容（法律条款文本由标准合同规范决定）
- PDF 字体具体选型（思源黑体/阿里巴巴普惠体/其他开源字体）
- 签署 H5 页面的具体 UI 样式
- 待办提醒的具体文案内容

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Requirements
- `.planning/REQUIREMENTS.md` §v1.4 — COMP-01~COMP-04 需求定义（4个合同合规需求）
- `.planning/ROADMAP.md` §Phase 11 — 阶段目标、成功标准、依赖关系

### Prior Phase Context
- `.planning/phases/10-UX基础-流程简化与引导体系/10-CONTEXT.md` — Phase 10 决策（3步向导/Tour/useMessage统一错误处理）
- `.planning/PROJECT.md` — 核心约束：3步内完成核心操作，Vue 3 + Element Plus + go-pdf/fpdf
- `.planning/PROJECT.md` §Blockers — 合同 PDF 生成依赖模板格式（go-pdf/fpdf已选型）；员工签署短信验证码依赖阿里云 SMS（需模板ID配置）

### Existing Contract Backend Code
- `backend/internal/employee/contract_model.go` — Contract GORM 模型，status 常量（draft/pending_sign/signed/active/terminated/expired），type 常量（fixed_term/indefinite/intern）
- `backend/internal/employee/contract_repository.go` — 数据访问层
- `backend/internal/employee/contract_service.go` — 业务逻辑（CreateContract/GeneratePDF/UploadSigned/TerminateContract/CheckContractRenewalReminders）
- `backend/internal/employee/contract_handler.go` — HTTP endpoints（已有 POST/GET/PUT generate-pdf upload-signed）
- `backend/internal/employee/contract_dto.go` — Request/Response DTOs
- `backend/internal/employee/pdf.go` — 现有 PDF 生成（Helvetica 英文占位符，需改为中文字体内嵌）
- `frontend/src/views/employee/EmployeeDrawer.vue` — 现有员工抽屉，含合同显示区域（已嵌入，无独立合同 Tab）
- `backend/internal/sms/` — 现有短信服务模块（复用 Phase 05 的 generateToken 模式）

### Tech Stack
- `go-pdf/fpdf`（已在 go.mod）— PDF 生成，需扩展中文字体支持
- `github.com/xuri/excelize/v2`（已在 go.mod）— 参考 Excel 导出模式
- 阿里云 OSS — 合同 PDF 存储（复用现有 oss 模块）
- 阿里云 SMS — 签署链接短信发送（复用现有 sms 模块）

### Project Patterns
- Phase 08 asynq 批量操作框架 — 后端定时任务参考模式
- Phase 10 StepWizard 向导 — 前端步骤向导组件参考
- Phase 10 useMessage — 错误处理复用统一模式
- `backend/internal/employee/` 目录结构 — 新合同签署相关代码放此处

### Integration Points
- `EmployeeDrawer.vue` → 新增「合同」Tab，含合同列表+新建合同入口
- `contract_service.go` → OSS（上传签署后 PDF）
- `contract_service.go` → SMS（发送签署链接）
- `contract_service.go` → Todo（未签提醒触发待办）
- `contract_handler.go` → 新签署验证 endpoint（校验验证码，更新合同状态）
- 员工 H5 签署页（新建）→ 员工无需登录，输入手机号+验证码即完成签署

</canonical_refs>

<codebase_context>
## Existing Code Insights

### Reusable Assets
- `EmployeeDrawer.vue`: 现有员工抽屉，可扩展 Tab（合同Tab）；复用现有 glass-card/section-header 样式
- `contract_service.go`: 已有 CreateContract/GeneratePDF/UploadSigned；新增签署验证逻辑（短信验证码校验）
- `pdf.go`: 已有 PDF 生成框架（AddPage/SetFont/CellFormat）；需替换字体为中文字体
- `sms/` 模块: 复用 generateToken 模式生成6位签署验证码
- `oss/` 模块: 复用 GenerateUploadURL 模式上传合同 PDF

### Established Patterns
- EmployeeDrawer Tab 扩展: 现有抽屉可扩展 Tab 切换（参考员工信息各字段Tab布局）
- asynq/gocron 定时任务: CheckContractRenewalReminders（到期提醒）已存在；新增"3天未签提醒"可复用同一模式
- StepWizard 向导（Phase 10）: 员工入职向导组件可参考用于合同签署流程

</codebase_context>

<specifics>
## Specific Ideas

- 合同类型选项：劳动合同（固定期限）/ 实习协议 / 兼职合同
- 老板签署流程3步：选员工+类型 → PDF预览确认 → 发送短信签署
- 签署 H5 页面：员工打开短信链接 → 输入手机号 → 收到验证码 → 输入验证码 → 确认签署 → 显示已签 PDF
- PDF 包含字段：甲方（公司名）、乙方（员工姓名+身份证号）、合同期限（起始日期、终止日期）、签订日期
- 签署状态：draft → pending_sign → signed/active → terminated/expired
- 合同到期提醒：已签合同到期前30天自动给老板发待办（CheckContractRenewalReminders 已有）

</specifics>

<deferred>
## Deferred Ideas

### Reviewed Todos (not folded)
- **员工绩效管理** — 超出 Phase 11 范围，属于 v1.4 之后的功能
- **多城市社保分别计算** — V1.0 明确排除
- **微信电子签章集成** — 明确 V2.0 范围，V1.0 降级为 PDF 模板+短信签署

</deferred>

---

*Phase: 11-合同合规*
*Context gathered: 2026-04-20*
