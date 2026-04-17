# Phase 5: 员工管理增强 + 组织架构基础 - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-17
**Phase:** 05-员工管理增强-组织架构基础
**Areas discussed:** 员工数据看板, 组织架构可视化, 员工信息登记流程, 办离职审批+花名册增强

---

## 员工数据看板

### 展示风格

| Option | Description | Selected |
|--------|-------------|----------|
| 纯数字卡片 | 4张卡片，简洁直观，符合"简单好用"原则 | ✓ |
| 数字卡片 + 趋势图 | 加折线图，开发量稍大 | |
| 数字卡片 + 环形图 | 加离职率环形图 | |

**User's choice:** 纯数字卡片
**Notes:** 与首页仪表盘风格保持一致

### 离职率展示

| Option | Description | Selected |
|--------|-------------|----------|
| 当月离职率数字 | 单月数字如"5.2%"，简单明了 | ✓ |
| 当月 + 环比上月 | 带环比箭头如 ↑0.8% | |
| 近3个月趋势数字 | 展示3个数字 | |

**User's choice:** 当月离职率数字

---

## 组织架构可视化

### 层级深度

| Option | Description | Selected |
|--------|-------------|----------|
| 最多3层 | 部门→岗位→员工，小微企业足够 | ✓ |
| 不限层级 | 支持多级嵌套，实现复杂 | |
| 仅2层 | 不单独展示岗位 | |

**User's choice:** 最多3层（部门→岗位→员工）

### 交互方式

| Option | Description | Selected |
|--------|-------------|----------|
| 搜索框 + 高亮定位 | ECharts tree + 搜索框自动定位高亮 | ✓ |
| 左右分栏布局 | 左侧树形列表 + 右侧详情面板 | |
| 纯树图 + 点击弹窗 | 全屏图表，点击弹出信息卡片 | |

**User's choice:** 搜索框 + 高亮定位（ECharts tree 图表）

---

## 员工信息登记流程

### 登记表形态

| Option | Description | Selected |
|--------|-------------|----------|
| 独立 H5 页面 + 链接 | 复用 Invitation Token 机制，员工微信打开填写 | ✓ |
| 后台弹窗表单 | 管理员直接填写，无法收集照片 | |

**User's choice:** 独立 H5 页面 + Token 链接

### 转发方式

| Option | Description | Selected |
|--------|-------------|----------|
| 二维码 + 复制链接 | 无需对接短信，微信转发 | |
| 短信发送链接 | 需对接阿里云 SMS，有额外成本 | |
| 两种方式都支持 | 管理员可选择 | ✓ |

**User's choice:** 两种方式都支持（二维码+复制链接 + 短信发送）

### 数据同步

| Option | Description | Selected |
|--------|-------------|----------|
| 提交即更新 | 员工提交后直接更新档案，管理员可后续编辑 | ✓ |
| 管理员审核后更新 | 提交后进入"待确认"状态，管理员审核后更新 | |

**User's choice:** 提交即更新

---

## 办离职审批 + 花名册增强

### 离职审批交互

| Option | Description | Selected |
|--------|-------------|----------|
| 列表审批 + 减员按钮跳转 | 列表中同意/驳回，同意后显示"去减员"按钮 | ✓ |
| 详情弹窗一站式操作 | 一个弹窗内完成审批+减员 | |

**User's choice:** 列表审批 + "去减员"按钮跳转社保减员页面

### 花名册详情展示

| Option | Description | Selected |
|--------|-------------|----------|
| 右侧 Drawer 抽屉 | 点击"更多"弹出侧边抽屉展示完整信息 | ✓ |
| 跳转详情页 | 跳转到 /employee/:id 页面 | |

**User's choice:** 右侧 Drawer 抽屉

### 花名册列展示

| Option | Description | Selected |
|--------|-------------|----------|
| 全部显示 | 所有新增列默认显示 | ✓ |
| 基础列 + 展开详情 | 仅显示基础列，其他点击展开 | |

**User's choice:** 全部显示（姓名/状态/岗位薪资/在职年限/合同到期天数/手机号）

---

## Claude's Discretion

- 数据看板卡片排列顺序和具体样式细节
- 组织架构树 ECharts 配置（布局方向、节点样式、动画效果）
- 员工信息登记 H5 页面的具体布局和字段分组
- 离职审批列表的列排序和筛选器
- 花名册列宽分配和默认排序
- Department 模型的 sort_order 字段设计
- 短信模板内容

## Deferred Ideas

None — discussion stayed within phase scope
