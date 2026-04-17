# Stack Research -- v1.3 新功能技术栈补充

**Domain:** 人事管理系统 v1.3 -- 考勤管理、审批流引擎、薪资增强、组织架构可视化、数据看板
**Researched:** 2026-04-17
**Confidence:** HIGH
**Scope:** 仅覆盖 v1.3 新功能所需的栈补充，不重复已有技术

## 现有栈确认（不需变更）

以下技术已在 v1.2 验证可用，v1.3 继续使用，无需重新评估：

- Go 1.25+ / Gin / GORM / PostgreSQL / go-redis / Viper / Zap / asynq / gocron
- Vue 3 / Element Plus / Vite / TypeScript / Pinia / Axios / dayjs / ECharts（已安装但未充分使用）
- shopspring/decimal v1.4.0（已在 go.mod 中，薪资精确计算）

---

## v1.3 新增后端依赖

### 审批流引擎

| Library | Version | Purpose | Why | Confidence |
|---------|---------|---------|-----|------------|
| qmuntal/stateless | v1.8.0 | 有限状态机（FSM） | 审批流核心引擎。支持层级状态（SubstateOf）-- 可将"待审批"分为"待经理审批""待老板审批"子状态；支持守卫条件（Guard）-- 可校验审批人权限；支持 Entry/Exit 动作 -- 审批状态变更时自动触发通知。比 looplab/fsm 更适合：后者不支持层级状态，无法表达"请假审批中"包含"待经理审批"的嵌套关系。比 Temporal/Cadence 更轻量：小微企业的审批流是同步状态转换，不需要分布式工作流引擎的运维负担 | HIGH |

**集成方式：** 封装为 `internal/approval/` 模块，FSM 实例与 GORM 模型绑定。每次审批操作时从 DB 加载当前状态，通过 FSM 执行转换，持久化新状态。不引入额外基础设施。

**为什么不用其他方案：**

| 排除方案 | 原因 |
|----------|------|
| looplab/fsm | 不支持层级状态（SubstateOf），审批流需要表达"审批中"包含多个子阶段的嵌套关系；event/callback 风格在复杂审批场景下不如声明式 API 清晰 |
| Temporal/Cadence | 分布式工作流引擎，需要独立 Server 集群，运维复杂度远超小微企业需求。审批流是秒级同步操作，不需要持久化工作流 |
| 自研状态机 | 审批流状态转换逻辑复杂（7种假类型 x 多级审批），自研容易遗漏边界条件。qmuntal/stateless 的声明式 API 和层级状态是核心价值 |
| go-workflow / Watermill | 面向事件驱动架构，审批流是同步状态转换，用消息流是杀鸡用牛刀 |

### 薪资算法增强

| Library | Version | Purpose | Why | Confidence |
|---------|---------|---------|-----|------------|
| shopspring/decimal | v1.4.0 (已安装) | 高精度十进制运算 | **已在 go.mod 中**，无需新增。个税累进计算、绩效系数乘法、社保公积金扣减全部使用 decimal 而非 float64，避免 0.1+0.2!=0.3 精度问题。v1.3 将深度使用此库 | HIGH |
| excelize | v2.10.1 (已安装) | 个税批量上传解析 | **已在 go.mod 中**。个税上传功能需要解析 Excel 文件中的员工个税数据，excelize 已支持。无需新增 | HIGH |

**关键决策：不引入个税计算库**

没有找到成熟的中国个税 Go 计算库（搜索了 GitHub、npm、Web）。原因合理：中国个税政策每年可能微调（专项附加扣除标准、起征点），任何开源库都可能滞后。因此：

- **策略：自建 `internal/tax/calculator.go` 模块**，基于七级超额累进税率表硬编码
- 税率表和速算扣除数作为 JSON 配置文件（`configs/tax_brackets.json`），随政策更新修改配置而非代码
- 专项附加扣除项作为可配置的枚举（子女教育、住房贷款、赡养老人等）
- 累计预扣预缴算法用 shopspring/decimal 保证精度

### 考勤管理

**不需要新增后端库。** 考勤管理的核心技术挑战是业务逻辑而非技术栈：

- **打卡时间计算：** dayjs（前端）+ Go time 包（后端），均为标准库级别
- **排班模式（固定班次/排班制/自由工时）：** 纯业务规则，用 Go struct + GORM 模型表达
- **出勤月报统计：** PostgreSQL 聚合查询 + excelize 导出（已安装）
- **打卡实况 WebSocket 推送：** 不做。v1.3 考勤是管理后台操作，不是实时打卡。今日打卡实况用定时刷新（5秒轮询）即可，50人规模下完全够用

---

## v1.3 新增前端依赖

### 组织架构可视化

| Library | Version | Purpose | Why | Confidence |
|---------|---------|---------|-----|------------|
| ECharts Tree Chart | 6.0.0 (已安装) | 组织架构树形图 | **ECharts 内置 tree 系列类型**，支持垂直/水平方向、节点自定义渲染、折叠展开。不需要额外安装任何包。ECharts 是项目已有依赖，tree 图是其 20+ 内置图表类型之一。中国开发者社区大量使用 ECharts tree 做组织架构图的实战案例（掘金、阿里云开发者社区均有教程） | HIGH |

**为什么不用专门的 org-chart 组件：**

| 排除方案 | 原因 |
|----------|------|
| vue3-tree-org (v4.2.2) | 最后更新约 3 年前，与 Vue 3.5.x 兼容性未经验证；无正式 GitHub releases；功能与 ECharts tree 重叠，增加一个维护停滞的依赖不值得 |
| vue3-org-chart | 维护者少，npm 下载量低，社区验证不足 |
| D3.js 组织架构图 | 引入整个 D3.js（500KB+）只为树形图是严重过度；ECharts 已满足需求 |
| 自研 SVG 树组件 | ECharts tree 支持节点自定义渲染（rich text）、折叠展开、缩放拖拽，自研没有优势 |

### 数据看板图表增强

| Library | Version | Purpose | Why | Confidence |
|---------|---------|---------|-----|------------|
| vue-echarts | v8.0.1 | ECharts Vue 3 封装组件 | v1.3 需要大量图表（薪资看板、社保看板、考勤统计、待办完成率环形图）。vue-echarts 提供声明式 Vue 组件写法（`<v-chart :option="chartOption">`），比直接操作 ECharts 实例代码量减少 60%+。v8.0.1 要求 echarts ^6.0.0 + vue ^3.3.0，与现有栈完全匹配。支持按需引入图表类型，减小打包体积 | HIGH |

**为什么需要 vue-echarts 而非直接用 ECharts：**

v1.3 有 5 个数据看板页面，每个页面有 2-4 个图表。直接用 ECharts 需要手动管理实例生命周期（init、resize、dispose），每个图表组件约 40 行样板代码。vue-echarts 封装后只需 `<v-chart :option="option" />` 一行，自动处理 resize/dispose。10+ 个图表节省约 400 行代码。

### 待办中心环形进度图

| Library | Version | Purpose | Why | Confidence |
|---------|---------|---------|-----|------------|
| ECharts Pie/Bar (已安装) + vue-echarts | 见上 | 待办完成率环形图 | ECharts pie 系列的 `radius: ['60%', '80%']` 配置即为环形图（donut chart），无需任何额外库。vue-echarts 提供更好的 Vue 组件封装 | HIGH |

### 前端工具补充

| Library | Version | Purpose | Why | Confidence |
|---------|---------|---------|-----|------------|
| @vueuse/core | v14.2.x (待安装) | 组合式工具函数 | v1.3 大量使用 `useIntervalFn`（考勤实况定时刷新）、`useLocalStorage`（看板布局偏好）、`useDebounceFn`（搜索防抖）。虽然 CLAUDE.md 已规划但 package.json 尚未安装 | HIGH |

---

## 安装命令

```bash
# 后端 -- 仅新增一个依赖
cd /Users/wencai/github/EasyHR
go get github.com/qmuntal/stateless@v1.8.0

# 前端 -- 新增两个依赖
cd /Users/wencai/github/EasyHR/frontend
npm install vue-echarts@8.0.1
npm install @vueuse/core@14.2.1
```

**注意：** vue-echarts 要求 echarts 作为 peer dependency。项目 CLAUDE.md 规划了 echarts 6.0.0 但 package.json 尚未安装。需要同时安装：

```bash
npm install echarts@6.0.0 vue-echarts@8.0.1 @vueuse/core@14.2.1
```

---

## Alternatives Considered

### 审批流引擎

| Recommended | Alternative | When to Use Alternative |
|-------------|-------------|-------------------------|
| qmuntal/stateless | looplab/fsm | 如果审批流极简（只有 Draft -> Approved/Rejected 两步），looplab/fsm 更轻量。但 v1.3 有 7 种假类型 + 多级审批，stateless 的层级状态更合适 |
| qmuntal/stateless | Temporal | 如果未来（V3.0+）需要跨服务分布式审批流（如对接外部电子签 API），考虑迁移到 Temporal |
| qmuntal/stateless | 自研 DB 状态字段 | 如果只有 1-2 种审批类型，直接用 status 字段 + switch 就够了。但 7 种假类型 + 撤回/催办/转交，FSM 的声明式配置比散落的 if/else 更可维护 |

### 组织架构可视化

| Recommended | Alternative | When to Use Alternative |
|-------------|-------------|-------------------------|
| ECharts tree | vue3-tree-org | 如果需要拖拽节点重构组织架构（ECharts tree 不支持节点拖拽重排）。但 v1.3 PRD 未要求拖拽重排，只需要可视化展示 |
| ECharts tree | D3.js tree | 如果需要极度定制化的节点渲染（如节点内嵌入头像、进度条、按钮等复杂 HTML）。ECharts tree 支持 rich text 但不支持嵌入 HTML。v1.3 只需展示姓名+职位，ECharts 足够 |

---

## What NOT to Use

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| Temporal/Cadence | 需要独立 Server 集群、学习曲线陡峭、运维成本高。小微企业的审批流是同步状态转换，不需要分布式工作流引擎 | qmuntal/stateless + PostgreSQL |
| D3.js（全量引入） | 打包体积 500KB+，只为树形图引入整个库严重浪费。ECharts 已满足所有图表需求 | ECharts 6.0 内置 tree/pie/bar |
| vue3-tree-org | 3 年未更新、无正式 release、与 Vue 3.5.x 兼容性未经验证 | ECharts tree 系列（已安装） |
| float64 做薪资计算 | 浮点精度丢失（0.1 + 0.2 = 0.30000000000000004），薪资场景不可接受 | shopspring/decimal（已安装） |
| WebSocket 实时推送 | v1.3 考勤管理后台只需定时刷新（5秒轮询），50 人规模下完全够用。WebSocket 增加连接管理复杂度 | HTTP 轮询 + asynq 后台任务 |
| 自研个税计算库 | 中国个税政策每年微调，硬编码在代码中不可维护 | 配置文件（JSON）+ 通用计算引擎 |
| RXjs / Observable | v1.3 的异步场景简单（定时刷新、表单提交），Vue 3 Composition API + async/await 完全够用 | async/await + Vue 3 reactive |

---

## Version Compatibility

| Package | Compatible With | Notes |
|---------|-----------------|-------|
| vue-echarts@8.0.1 | echarts@^6.0.0, vue@^3.3.0 | 项目 vue@3.5.x 满足；echarts 6.0 需要安装 |
| qmuntal/stateless@v1.8.0 | Go 1.21+ | 项目 Go 1.25 满足 |
| @vueuse/core@14.2.x | vue@^3.5.0 | 项目 vue@3.5.x 满足 |
| ECharts 6.0 tree | vue-echarts@8.0+ | tree 系列是 ECharts 内置类型，无需额外引入 |

---

## 后端模块结构建议

```
internal/
  approval/          # 新增 -- 审批流引擎
    engine.go        # FSM 初始化和配置
    types.go         # 审批状态、事件类型定义
    service.go       # 审批业务逻辑
    handlers.go      # Gin HTTP handlers
  attendance/        # 新增 -- 考勤管理
    model.go         # 打卡记录、排班模型
    service.go       # 考勤业务逻辑
    calculator.go    # 工时计算
    report.go        # 月报生成
    handlers.go      # Gin HTTP handlers
  todo/              # 新增 -- 待办中心
    model.go         # 待办事项模型
    service.go       # 待办聚合逻辑
    handlers.go      # Gin HTTP handlers
  salary/            # 增强 -- 薪资模块扩展
    dashboard.go     # 数据看板查询
    adjustment.go    # 调薪/普调逻辑
    tax_upload.go    # 个税上传解析
    performance.go   # 绩效系数计算
    slip.go          # 工资条发送
```

---

## 前端目录结构建议

```
frontend/src/
  views/
    attendance/      # 新增 -- 考勤管理页面
    approval/        # 新增 -- 审批管理页面
    todo/            # 新增 -- 待办中心页面
  components/
    charts/          # 新增 -- 图表封装组件
      DonutChart.vue      # 环形进度图（待办完成率）
      BarChart.vue        # 柱状图（薪资分布、考勤统计）
      LineChart.vue       # 折线图（趋势图）
      OrgTreeChart.vue    # 组织架构树形图
      StatCard.vue        # 统计卡片组件（通用）
```

---

## Sources

- [qmuntal/stateless GitHub](https://github.com/qmuntal/stateless) -- v1.8.0, 2026-02-10 发布，支持层级状态、守卫条件、Entry/Exit 动作 -- HIGH
- [vue-echarts npm](https://www.npmjs.com/package/vue-echarts) -- v8.0.1, 要求 echarts ^6.0.0 + vue ^3.3.0 -- HIGH
- [shopspring/decimal GitHub](https://github.com/shopspring/decimal) -- v1.4.0, 已在 go.mod 中，高精度十进制运算 -- HIGH
- [ECharts Tree 官方文档](https://echarts.apache.org/en/option.html#series-tree) -- 内置 tree 系列类型，支持垂直/水平/径向布局 -- HIGH
- [掘金：ECharts Tree 组织结构图实战](https://juejin.cn/post/7041048687933390884) -- 中国开发者使用 ECharts tree 做组织架构图的实践案例 -- MEDIUM
- [阿里云开发者社区：Vue3 使用 ECharts 树图](https://developer.aliyun.com/article/1600528) -- Vue3 + ECharts tree 组件化封装 -- MEDIUM
- [Medium：Simple Workflow Engine in Go Using Stateless](https://medium.com/@jhberges/simple-workflow-engine-in-go-using-stateless-9db4464b93ec) -- 用 stateless 构建工作流引擎的实践 -- MEDIUM
- [vue3-tree-org npm](https://www.npmjs.com/package/vue3-tree-org) -- v4.2.2，约 3 年前发布，维护状态存疑 -- LOW（排除依据）

---
*Stack research for: EasyHR v1.3 新功能*
*Researched: 2026-04-17*
