# Phase 5: 员工管理增强 + 组织架构基础 - Context

**Gathered:** 2026-04-17
**Status:** Ready for planning

<domain>
## Phase Boundary

管理员获得完整的员工数据洞察和组织管理能力，部门模型为后续薪资普调（Phase 7）和考勤排班（Phase 6）提供基础维度。

具体包含：
- 员工数据看板（在职/新入职/离职/离职率）
- 组织架构可视化（部门→岗位→员工层级树 + 检索定位）
- 员工信息登记（创建登记表 + 转发给员工填写 + 自动更新档案）
- 办离职优化（审批 + 社保减员联动）
- 花名册增强（完整信息列 + 详情查看 + 搜索）
- Excel 导出增强

</domain>

<decisions>
## Implementation Decisions

### 员工数据看板
- **D-01:** 纯数字卡片风格，4 张卡片（在职人数/当月新入职/当月离职/当月离职率），不引入图表组件
- **D-02:** 离职率仅展示当月数字（离职人数/(离职人数+期末人数)×100%），不带环比趋势

### 组织架构可视化
- **D-03:** 层级深度最多 3 层（部门→岗位→员工），使用邻接表模型（parent_id）存储部门层级
- **D-04:** 使用 ECharts tree 图表渲染组织架构，顶部搜索框输入关键字后树图自动定位并高亮匹配节点
- **D-05:** Department 模型包含 id/name/parent_id/org_id，Employee 模型新增 department_id 字段

### 员工信息登记流程
- **D-06:** 独立 H5 页面 + Token 链接机制（复用现有 Invitation 模型的 Token 生成逻辑），员工无需登录即可填写
- **D-07:** 转发方式同时支持两种：二维码 + 复制链接（微信转发）和短信发送（对接阿里云 SMS）
- **D-08:** 员工提交后直接更新员工档案（提交即更新），管理员可后续编辑修正。以最新提交版本为准覆盖

### 办离职审批 + 花名册增强
- **D-09:** 离职审批在列表中直接操作（同意/驳回按钮），审批通过后行内显示"去减员"按钮跳转社保减员页面，减员完成后自动更新离职状态
- **D-10:** 花名册点击"更多"弹出右侧 Drawer 抽屉展示员工完整信息（基本信息 + 身份证/合同/银行卡/简历）
- **D-11:** 花名册全部新增列默认显示：姓名、状态、岗位薪资、在职年限、合同到期天数、手机号。不使用折叠展开

### Claude's Discretion
- 数据看板卡片排列顺序和具体样式细节
- 组织架构树 ECharts 配置（布局方向、节点样式、动画效果）
- 员工信息登记 H5 页面的具体布局和字段分组
- 离职审批列表的列排序和筛选器
- 花名册列宽分配和默认排序
- Department 模型的 sort_order 字段设计
- 短信模板内容

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Requirements
- `.planning/REQUIREMENTS.md` §5 — Phase 5 需求定义（EMP-01~EMP-16）
- `.planning/ROADMAP.md` §Phase 5 — 阶段目标、成功标准、依赖关系

### Research
- `.planning/research/FEATURES.md` — v1.3 功能研究（组织架构、员工管理增强的竞品分析和技术方案）
- `.planning/research/SUMMARY.md` — v1.3 研究综合报告

### Existing Code Patterns
- `internal/employee/model.go` — Employee 模型（敏感字段双列模式：encrypted + hash）
- `internal/employee/service.go` — Service 层模式（CRUD + 加密/解密/脱敏）
- `internal/employee/invitation_model.go` — Invitation Token 机制（可复用于员工信息登记链接）
- `internal/employee/offboarding_dto.go` — 离职 DTO 结构
- `internal/dashboard/service.go` — Dashboard 聚合模式（errgroup 并发）
- `frontend/src/views/employee/EmployeeList.vue` — 员工列表页组件
- `frontend/src/api/employee.ts` — 前端 API 层

### Project Decisions
- `.planning/PROJECT.md` — Key Decisions 表（多租户 org_id、RBAC、加密策略）
- `.planning/STATE.md` — 项目状态（已锁定决策汇总）

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/dashboard/service.go`: GetEmployeeStats 已实现 active/joined/left 统计，可直接复用并扩展离职率计算
- `internal/employee/invitation_model.go`: Invitation Token 生成和验证机制，可直接复用于员工信息登记链接
- `internal/employee/service.go`: ExportExcel 已实现 excelize 导出模式，需扩展更多列（岗位薪资/在职年限/合同到期天数）
- `internal/employee/offboarding_dto.go`: OffboardingDetailResponse 结构可扩展审批状态字段
- `frontend/src/views/home/HomeView.vue`: 卡片布局和样式可参考用于员工数据看板

### Established Patterns
- Handler → Service → Repository 三层架构（所有模块统一）
- 敏感字段双列模式（encrypted + hash），所有敏感数据必须加密
- org_id 逻辑多租户隔离，GORM Scope 自动注入
- errgroup 并发聚合数据（Dashboard 模式）
- 前端使用 Element Plus + Composition API + Pinia Store

### Integration Points
- Employee 模型需新增 department_id 字段（关联 Department）
- 新建 Department 模型（id/name/parent_id/org_id/sort_order），为 Phase 6 排班和 Phase 7 普调提供维度
- Offboarding 流程需增加 approved/rejected 状态转换（现有只有 pending_review → approved → completed）
- 社保减员页面需支持接收离职员工参数（employee_id + name），实现一键减员
- 员工信息登记 H5 页面需新建独立路由（/register/:token），无需登录认证
- 阿里云 SMS 对接需新增短信模板（员工信息登记邀请）

</code_context>

<specifics>
## Specific Ideas

- 员工数据看板与首页数据概览风格保持一致（蓝色主题、卡片圆角）
- 组织架构搜索要支持部门名、岗位名、员工姓名三种维度的匹配
- 员工信息登记链接需设置有效期（参考 Invitation 的 7 天过期机制）
- 花名册的"岗位薪资"列显示月薪数字（需关联 salary 模块数据）
- 合同到期天数需要关联 contract 模块计算，负数标红显示

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 05-员工管理增强-组织架构基础*
*Context gathered: 2026-04-17*
