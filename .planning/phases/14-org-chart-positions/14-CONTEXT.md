# Phase 14: 组织架构图（部门+岗位管理） - Context

**Gathered:** 2026-04-21
**Status:** Ready for planning

<domain>
## Phase Boundary

完善组织架构体系，新增独立岗位管理，增强组织架构图可视化交互，让老板直观管理和调整团队结构。

具体包含：
- 岗位管理（创建/编辑/删除岗位，名称/所属部门/排序，岗位可跨部门复用）
- 组织架构图增强（可视化展示部门→岗位→员工三级树，支持搜索高亮、节点展开/折叠、拖拽调整部门层级）
- 部门管理完善（架构图内联编辑部门名称/排序/删除，含员工转移提示）
- 员工关联（创建/编辑时通过下拉选择部门和岗位，替代自由文本输入）

**Scope:** H5管理后台 + 后端API，前端为主
**Depends on:** Phase 10（UX基础）

</domain>

<decisions>
## Implementation Decisions

### 岗位建模（ORG-01）
- **D-14-01:** 新建 `Position` 表（独立岗位管理）
  - 字段：`id` / `name`（varchar(100)）/ `department_id`（int，可NULL）/ `sort_order`（int，默认0）/ `org_id`
  - `Employee` 新增 `position_id`（int，可NULL）字段，保留 `position`（varchar）字段用于冗余存储显示名
  - 唯一性约束：`(org_id, department_id, name)` 联合唯一，避免同一部门内岗位重名
- **D-14-02:** 迁移策略：读取现有 `Employee.position` 唯一值，自动创建对应 Position 记录，关联现有员工

### 岗位归属（ORG-01 跨部门复用）
- **D-14-03:** 通用岗位模式：`Position.department_id = NULL` 表示通用岗位（任何部门可用）
- **D-14-04:** 员工选择岗位时，下拉列表根据员工的 `department_id` 过滤：
  - 先显示「部门专属岗位」（该部门创建的岗位）
  - 再显示「通用岗位」（`department_id=NULL` 的岗位）
- **D-14-05:** 新建岗位时默认「通用岗位」（`department_id=NULL`），可选切换为部门专属

### 架构图交互（ORG-02）
- **D-14-06:** 支持拖拽调整部门层级：ECharts tree 节点拖拽后自动更新 `Department.parent_id`
- **D-14-07:** 保留搜索高亮和节点展开/折叠功能（Phase 05 已实现）
- **D-14-08:** 架构图顶部搜索框支持按部门名/岗位名/员工姓名搜索，匹配节点高亮蓝色，未匹配节点降低透明度

### 部门管理（ORG-03）
- **D-14-09:** 删除有员工的部门时，引导用户选择目标部门进行员工转移，再执行删除
  - 删除流程：点击删除 → 弹窗列出该部门员工列表 → 选择目标部门 → 确认转移 → 删除部门
  - 员工转移后 `department_id` 更新，`position_id` 保留（岗位可能已不存在，需兼容处理）
- **D-14-10:** 内联编辑：点击部门节点名称 → 进入编辑模式（类似 Excel 行内编辑），blur 或回车保存
- **D-14-11:** 部门排序：通过 `sort_order` 字段控制，同级部门按 `sort_order` 升序排列

### 员工表单岗位选择（ORG-04）
- **D-14-12:** 员工创建/编辑表单中，将岗位字段从 `el-input` 改为 `el-select` 下拉选择
- **D-14-13:** 下拉选项列表根据当前选择的部门动态筛选（部门专属岗位 + 通用岗位）
- **D-14-14:** 无岗位情况：员工可选择「未分配岗位」（即 `position_id = NULL`），保留 `position` 字段存储自由文本作为备选显示
- **D-14-15:** 部门为空时，下拉列表显示所有通用岗位（`department_id=NULL`）

### Claude's Discretion
- 架构图节点样式（颜色/字体/图标）的具体配置细节
- 拖拽时的视觉反馈（节点高亮/阴影/吸附效果）
- 部门删除转移弹窗的具体 UI 布局
- 岗位下拉选择器的分组展示方式（专属/通用两组的视觉区分）
- 迁移脚本的具体实现（后台静默执行还是提供进度提示）

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Requirements
- `.planning/REQUIREMENTS.md` — ORG-01~ORG-04 需求定义
- `.planning/ROADMAP.md` §Phase 14 — 阶段目标、成功标准、依赖关系

### Existing Code Patterns
- `internal/department/model.go` — Department 模型（id/name/parent_id/sort_order，邻接表）
- `internal/department/service.go` — BuildTree 方法（按部门→岗位→员工三层构建，搜索高亮逻辑）
- `internal/department/handler.go` — 部门 API 路由
- `internal/department/repository.go` — 部门 Repository
- `internal/employee/model.go` — Employee 模型（含 Position 字段），需新增 position_id
- `internal/employee/repository.go` — Employee Repository（CountByDepartment 方法已存在）
- `frontend/src/views/employee/EmployeeCreate.vue` — 员工创建表单，岗位字段需改造为下拉选择
- `frontend/src/views/employee/OrgChart.vue` — 现有组织架构图（ECharts tree），需增强拖拽
- `frontend/src/api/department.ts` — 部门 API 客户端

### Project Decisions
- `.planning/PROJECT.md` — 核心价值：3步内完成核心操作，零学习成本
- `.planning/PROJECT.md` — 组织架构复用 ECharts tree 图表（Key Decisions 表）
- `.planning/phases/05-员工管理增强-组织架构基础/05-CONTEXT.md` — Phase 05 决策（ECharts tree/邻接表/3层深度）

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `department/service.go BuildTree`: 已实现按岗位名分组员工的逻辑，改用 Position 表后需修改分组逻辑（按 `position_id` 而非文本值）
- `department/service.go markMatches`: 搜索高亮已实现，可直接复用
- `Employee.Position`: 自由文本字段，新方案需保留但作为冗余显示字段
- `el-select`: Element Plus 下拉组件，EmployeeCreate.vue 中可复用
- `asynq`: 后端批量迁移任务可复用 Phase 08 的批量操作框架模式

### Established Patterns
- Handler → Service → Repository 三层架构（所有模块统一）
- org_id 逻辑多租户隔离，GORM Scope 自动注入
- ECharts tree 图表配置（Phase 05 已验证）
- Soft delete 模式（已用于其他模块）

### Integration Points
- Position 表新建后需注册到 GORM AutoMigrate
- Employee 新增 `position_id` 字段，需迁移脚本
- 员工下拉选择器需要部门列表 + 岗位列表的联动 API
- 架构图拖拽后端需要 `UpdateDepartment` 支持 `parent_id` 变更

</code_context>

<specifics>
## Specific Ideas

- Position 表字段：`id`, `org_id`, `name`, `department_id`（可NULL）, `sort_order`, `created_by`, `updated_by`
- Employee 新增字段：`position_id`（可NULL，FK → positions.id）
- 唯一性约束：同一部门内岗位名唯一（通用岗位视为 department_id=NULL 的特殊组）
- 迁移：批量查询现有 Employee.position 唯一值，创建 Position 记录，关联 Employee.position_id
- 员工岗位下拉：API 返回 `{ dept_positions: [...], common_positions: [...] }`，前端按组渲染

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 14-组织架构图（部门+岗位管理）*
*Context gathered: 2026-04-21*
