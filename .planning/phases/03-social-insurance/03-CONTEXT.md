# Phase 3: 社保管理 - Context

**Gathered:** 2026-04-07
**Status:** Ready for planning

<domain>
## Phase Boundary

根据员工城市+岗位自动匹配社保参保基数（自建30+城市五险一金政策库），老板一键办理参保/停缴（支持批量），缴费到期前3天自动提醒，记录缴费明细和变更历史。覆盖 SOCL-01 ~ SOCL-07 全部需求。

</domain>

<decisions>
## Implementation Decisions

### 社保政策库数据结构
- **D-01:** 单表+JSONB存储。一个城市一条记录，字段包括：城市ID、生效年份、五险一金配置（JSONB）。JSONB结构包含每个险种的企业缴费比例、个人缴费比例、基数下限、基数上限。利用PostgreSQL JSONB查询能力。
- **D-02:** 管理员在H5后台手动录入/编辑政策数据。提供政策编辑页面，按城市录入各险种比例和基数上下限。初始30+城市数据通过管理后台逐个录入。
- **D-03:** 五险一金全覆盖：养老保险、医疗保险、失业保险、工伤保险、生育保险、住房公积金。JSONB中每个险种一个key。
- **D-04:** 政策按年度生效（effective_year字段）。每年7月左右各地调整政策，管理员更新后新建下一年度记录。查询时取 effective_year <= 当前年份 的最新记录。

### 参保/停缴操作流程
- **D-05:** 参保流程3步：老板选择员工（支持多选）→ 系统根据员工城市自动匹配五险一金基数，显示各险种企业和个人缴费金额预览 → 老板确认 → 批量生效。符合"3步内完成核心操作"原则。
- **D-06:** 支持批量参保/停缴。老板可勾选多个员工一次性操作。前端传 employee_ids 数组，后端批量创建社保记录。
- **D-07:** 三种停缴触发方式：
  1. 老板手动停缴：在社保模块选择员工手动停缴
  2. 离职自动触发提醒：复用 Phase 2 `onEmployeeResigned` 接口，离职后生成"需办理社保停缴"待办提醒（仅提醒，不自动执行停缴）
  3. 转正自动检查：员工从试用期转正时，系统检查社保状态，若未参保则生成提醒
- **D-08:** 社保记录状态：待参保（pending）→ 参保中（active）→ 停缴（stopped）。参保记录记录参保月份、基数、各险种明细。

### 缴费提醒机制
- **D-09:** 使用 gocron v2.19.1 实现定时检查。每日定时扫描所有企业的社保缴费到期情况，到期前3天生成提醒记录。
- **D-10:** 提醒方式：APP内消息 + 首页待办卡片（Phase 7 工作台消费）。不使用短信或微信模板消息（V1.0 成本控制）。
- **D-11:** 缴费截止日为每月固定日期（系统默认值，如每月15日或25日）。全局配置，非企业自定义。

### 社保与工资联动
- **D-12:** 社保模块提供查询接口 `GetSocialInsuranceDeduction(orgID, employeeID, month)` 返回指定月份各险种个人扣款金额。Phase 5 工资核算模块调用此接口获取社保扣款数据。单向依赖，社保模块不依赖工资模块。
- **D-13:** 薪资变动时只提醒不自动调整社保基数（SOCL-06）。检测方式：工资核算完成后，对比员工当前薪资与社保基数，偏差超过阈值时生成"建议调整社保基数"提醒。具体实现留 Phase 5 联动。

### RBAC 权限（社保模块）
- **D-14:** 社保模块权限分配：
  - OWNER：全部操作（参保、停缴、变更、查看明细、导出凭证、管理政策库）
  - ADMIN：同 OWNER（参保、停缴、变更、查看明细、导出凭证）
  - MEMBER：仅查看自己社保记录（通过 user_id → employee_id 关联查询）

### Claude's Discretion
- 社保政策库 JSONB 内部具体字段命名
- 参保记录的具体数据模型（是否按险种拆分行 vs 一条记录存所有险种）
- 缴费明细的记录粒度（按月汇总 vs 按次记录）
- 导出凭证的 Excel/PDF 格式细节
- 社保模块内部目录结构

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### 项目规范
- `prd.md` — 产品需求文档，V1.0功能范围、验收标准
- `ui-ux.md` — UI/UX设计原型，社保管理页面布局、交互流程
- `tech-architecture.md` — 技术架构设计、数据模型、模块结构、API设计

### Phase 1-2 上下文
- `.planning/phases/01-foundation-auth/1-CONTEXT.md` — Phase 1 全部决策（三层架构、多租户、加密、RBAC 等）
- `.planning/phases/02-employee-management/02-CONTEXT.md` — Phase 2 决策，关键：D-17 离职事件接口定义

### 研究报告
- `.planning/research/STACK.md` — 技术栈选型（gocron v2.19.1, asynq v0.26.0, excelize 等）
- `.planning/research/ARCHITECTURE.md` — 架构设计建议、模块边界
- `.planning/research/PITFALLS.md` — 常见陷阱

### 需求追踪
- `.planning/REQUIREMENTS.md` — SOCL-01 ~ SOCL-07 需求定义
- `.planning/ROADMAP.md` — Phase 3 定义和成功标准

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/common/model/base.go` — BaseModel（ID、OrgID、审计字段、软删除），社保所有模型嵌入此基类
- `internal/common/crypto/` — AES-256-GCM 加密/解密、SHA-256 哈希（社保不需要额外加密，基数非敏感数据）
- `internal/common/response/` — 统一响应封装（Success、Error、PageSuccess）
- `internal/common/middleware/` — 认证、RBAC（RequireRole）、多租户（TenantScope）
- `internal/city/` — 城市列表模块（37城市），社保政策库通过 city_id 关联
- `internal/employee/` — 员工模块，关键集成点：
  - `model.go` Employee 结构体（城市、岗位、状态）
  - `offboarding_service.go:282` `onEmployeeResigned` — Phase 3 需实现此回调
- `pkg/oss/client.go` — OSS 签名URL生成，凭证文件存储可复用
- `internal/employee/pdf.go` — PDF生成参考实现（go-pdf/fpdf），参保材料PDF复用此模式

### Established Patterns
- 三层架构：handler → service → repository（社保模块遵循相同模式）
- DTO 模式：独立请求/响应结构体
- 路由注册：RegisterRoutes 方法在 main.go 统一注册
- 审计日志：GORM Hook 自动记录，Module="social_insurance"
- 多租户：Repository 层自动注入 org_id

### Integration Points
- `cmd/server/main.go` AutoMigrate — 新增社保相关模型
- `cmd/server/main.go` 路由注册 — 新增 socialInsuranceHandler.RegisterRoutes(v1, authMiddleware)
- `cmd/server/main.go` gocron 初始化 — 注册社保缴费提醒定时任务
- `internal/employee/offboarding_service.go:282` — 实现 onEmployeeResigned 回调
- Phase 5 工资核算 → 调用社保查询接口获取扣款数据

### 新增依赖
- `gocron v2.19.1` — 定时任务（已在技术栈中选定）
- `go-pdf/fpdf v0.9.0` — 参保材料PDF生成（已在 Phase 2 引入）
- `excelize v2.10.1` — 缴费凭证Excel导出（已在 Phase 2 引入）

</code_context>

<specifics>
## Specific Ideas

- 社保政策库按城市+年度存储，管理员通过H5后台维护，不需要Excel导入
- 参保操作支持批量多选员工，一键确认生效
- 离职后只生成停缴提醒待办，不自动执行停缴（老板确认后操作）
- 缴费提醒通过APP内消息+待办卡片触达，不用短信降低成本
- 薪资变动时生成调整建议提醒，不自动修改社保基数

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---
*Phase: 03-social-insurance*
*Context gathered: 2026-04-07*
