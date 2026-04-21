# Phase 14: 组织架构图（部门+岗位管理） - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-21
**Phase:** 14-组织架构图（部门+岗位管理）
**Areas discussed:** 岗位建模, 岗位归属, 架构图交互, 部门删除, 员工表单岗位选择

---

## 岗位建模

| Option | Description | Selected |
|--------|-------------|----------|
| 独立岗位表 | 新建 Position 表（id/name/department_id/sort_order），Employee 新增 position_id，保留 position 字段冗余存储 | ✓ |
| 保留自由文本 | Position 表只存岗位名称列表用于下拉选择，Employee.position 仍是自由文本 | |
| Claude决定 | 由 Claude 决定实现方案 | |

**User's choice:** 独立岗位表
**Notes:** 用户认可独立岗位表可以避免拼写不一致、支持岗位管理和排序，是长期更优方案。

---

## 岗位归属

| Option | Description | Selected |
|--------|-------------|----------|
| 通用岗位（department_id=NULL，可跨部门复用） | Position.department_id=NULL 表示通用岗位（任何部门可用），员工下拉根据 department_id 过滤（专属+通用） | ✓ |
| 部门专属（必须绑定部门） | 每个岗位必须属于某部门，跨部门复用需手动复制 | |

**User's choice:** 通用岗位
**Notes:** 通用岗位模式更灵活，跨部门复用场景（如同名"销售"在多个部门）直接支持，无需手动操作。

---

## 架构图交互

| Option | Description | Selected |
|--------|-------------|----------|
| 支持拖拽（推荐） | ECharts tree 节点拖拽后自动更新 parent_id，@chang callback 实现 | ✓ |
| 仅按钮操作 | 通过编辑弹窗手动调整 parent_id | |

**User's choice:** 支持拖拽
**Notes:** 组织架构图的标配交互，小微企业老板会期望这个功能。ECharts tree 原生支持，实现成本可控。

---

## 部门删除

| Option | Description | Selected |
|--------|-------------|----------|
| 引导转移（推荐） | 删除前引导选择目标部门进行员工转移，再执行删除，一气呵成 | ✓ |
| 强制阻止 | 有员工的部门禁止删除，需手动转移员工后再操作 | |

**User's choice:** 引导转移
**Notes:** 老板期望一气呵成，不想跑多个页面操作。引导转移弹窗需列出员工并选择目标部门。

---

## 员工表单岗位选择

（作为 ORG-04 的自然延伸，用户确认改造为下拉选择）

| Decision | Description |
|----------|-------------|
| 下拉选择 | 岗位字段从 el-input 改为 el-select |
| 部门联动 | 下拉选项根据当前部门动态筛选（专属+通用岗位） |
| 未分配支持 | 员工可选择「未分配岗位」（position_id=NULL） |
| 冗余字段保留 | Employee.position 保留用于显示名备份 |

**Notes:** 与 EmployeeCreate.vue 编辑模式的表单样式保持一致（glass-card/section-header 样式）。

---

## Claude's Discretion

- 架构图节点样式（颜色/字体/图标）的具体配置细节
- 拖拽时的视觉反馈（节点高亮/阴影/吸附效果）
- 部门删除转移弹窗的具体 UI 布局
- 岗位下拉选择器的分组展示方式（专属/通用两组的视觉区分）
- 迁移脚本的具体实现（后台静默执行还是提供进度提示）

## Deferred Ideas

None — discussion stayed within phase scope.
