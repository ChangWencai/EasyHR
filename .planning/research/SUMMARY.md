# Project Research Summary

**Project:** EasyHR v1.3 (易人事 -- 小微企业人事管理系统)
**Domain:** HR SaaS -- 考勤管理、审批流、薪资增强、社保增强、待办中心、员工管理增强
**Researched:** 2026-04-17
**Confidence:** HIGH

## Executive Summary

EasyHR v1.3 是一个面向 10-50 人小微企业 HR SaaS 系统的功能增强版本。核心挑战不是新建独立模块，而是**跨模块数据聚合**（待办中心需聚合 6+ 模块的待办事项）和**跨模块深度联动**（考勤审批数据驱动薪资核算、离职流程触发社保减员）。现有代码库采用模块化单体架构，通过 adapter 接口解耦跨模块依赖（salary -> tax/si/employee），v1.3 需延续此模式并扩展 `AttendanceProvider`、`TodoCreator`、`DepartmentProvider` 三个新接口。

建议采用 5 阶段递进式构建顺序：先建部门基础数据（Phase A），再建考勤模块作为薪资上游数据源（Phase B），然后并行推进薪资增强和社保增强（Phase C/D），最后构建待办中心聚合层（Phase E）。新增依赖极少：后端仅需 `qmuntal/stateless` 状态机库，前端新增 `vue-echarts` 和 `@vueuse/core` 两个包。v1.3 的核心风险集中在**薪资计算公式变更覆盖历史数据**和**考勤排班跨天班次归属错误**两个陷阱上，恢复代价分别为 CRITICAL 和 HIGH，必须在设计阶段预防。

## Key Findings

### Recommended Stack

v1.3 在已有技术栈（Go/Gin/GORM/PostgreSQL/Vue 3/Element Plus）基础上，新增极少依赖。核心技术决策：(1) 审批流使用 `qmuntal/stateless` 有限状态机而非分布式工作流引擎；(2) 个税计算自建引擎，税率表用 JSON 配置文件；(3) 组织架构可视化复用已有 ECharts 的 tree 图表类型，不引入专用组件。

**新增后端依赖：**
- `qmuntal/stateless` v1.8.0：审批流状态机引擎 -- 支持层级状态和守卫条件，比 looplab/fsm 功能更完整，比 Temporal 运维负担低
- `shopspring/decimal` v1.4.0（已安装，深度使用）：薪资精确计算，避免浮点精度丢失
- `excelize` v2.10.1（已安装，复用）：个税 Excel 上传解析

**新增前端依赖：**
- `vue-echarts` v8.0.1：ECharts 的 Vue 3 封装组件，10+ 图表节省约 400 行样板代码
- `@vueuse/core` v14.2.x：组合式工具函数（useIntervalFn/useDebounceFn 等）
- `echarts` v6.0.0（待安装）：vue-echarts 的 peer dependency

**明确不引入：** Temporal/Cadence（过重）、D3.js 全量（500KB+ 只为树图）、vue3-tree-org（3年未更新）、float64 薪资计算（精度问题）、WebSocket（5秒轮询够用）。

### Expected Features

**Must have（P1 -- Phase 1）：**
- 待办事项汇总列表 + 快捷办事入口 -- Core Value "3步完成操作"的直接体现
- 限时任务引擎（合同签约/个税申报/社保缴费等 7 种任务的定时生成和超时管理）
- 完成率环形图（全部/限时任务两个维度）
- 员工数据看板（在职/新入职/离职/离职率 4 指标）
- 花名册增强（岗位薪资/在职年限/合同到期天数等列扩展）

**Must have（P2 -- Phase 2）：**
- 考勤打卡设置（固定时间/按排班/自由工时 3 种模式）
- 审批流引擎（请假 7 种类型 + 补卡/出差/外出/调班，共 11 种审批类型）
- 今日打卡实况 + 出勤月报
- 薪资数据看板（应发/实发/社保/个税环比）
- 单人调薪 + 按部门普调
- 绩效系数设置（0%-100%，月度维度）

**Must have（P3 -- Phase 3）：**
- 薪资算法增强（考勤联动：基本工资/加班费/病假工资自动计算）
- 个税上传自动抓取（Excel 解析 -> 关键字匹配 -> 更新工资表）
- 社保数据看板 + 增减员弹窗优化 + 缴费渠道选择 + 欠缴状态管理
- 组织架构可视化（ECharts Tree）
- 员工信息登记（转发员工填写）
- 办离职优化（审批 + 社保减员联动）

**Defer（v2+）：**
- GPS 定位打卡、人脸识别打卡 -- 需要移动端 APP 原生支持
- 复杂工作流引擎（多级审批）-- 小微企业不需要
- 排班自动优化算法 -- NP-hard 问题，V1.3 不值得投入
- 考勤硬件对接 -- 超出 SaaS 产品范畴

### Architecture Approach

v1.3 延续现有模块化单体架构，新增 `internal/todo`（待办中心）和 `internal/attendance`（考勤管理）两个独立模块，在现有 `internal/salary`、`internal/socialinsurance`、`internal/employee` 模块内扩展增强。跨模块通信通过 adapter 接口解耦，现有 6 个 adapter 接口新增 3 个（AttendanceProvider、TodoCreator、DepartmentProvider）。

**Major components:**
1. `internal/todo` -- 全新模块，聚合展示层。物化 `todo_items` 表 + 事件驱动写入（各模块状态变更时主动推送），避免跨表 UNION ALL 实时聚合
2. `internal/attendance` -- 全新模块，打卡规则引擎（3 种模式）+ 审批流状态机（qmuntal/stateless）+ 出勤月报预计算
3. `internal/salary` 扩展 -- 调薪/普调（INSERT ONLY 策略）、绩效系数（独立月度表）、个税上传解析、薪资算法增强（集成 AttendanceProvider）
4. `internal/socialinsurance` 扩展 -- 缴费状态从 3 种扩展为 5 种（新增 arrears/transferred），定时任务驱动状态流转
5. `internal/employee` 扩展 -- 新增部门模型（邻接表 + path 字段预留）、员工信息登记（Token 邀请复用现有 Invitation 模式）

### Critical Pitfalls

1. **调薪生效期历史追溯破坏已有工资数据** -- 恢复代价 CRITICAL。调薪操作必须 INSERT 新记录，禁止 UPDATE 历史 salary_items；已确认月份强制只读保护
2. **工资核算公式变更覆盖历史数据** -- 恢复代价 CRITICAL。v1.3 新增病假系数/加班费等复杂逻辑后，已 confirmed/paid 的 PayrollRecord 禁止重新核算
3. **考勤排班跨天班次导致打卡归属日期错乱** -- 恢复代价 HIGH。班次模型必须包含 workDateOffset 字段，打卡记录表设计 work_date 与 punch_time 分离
4. **离职流程多模块联动的事务一致性** -- 恢复代价 HIGH。离职审批通过必须同一数据库事务（Offboarding + Employee 状态），联动操作用 asynq 事件队列异步处理
5. **审批流状态机遗漏转换路径导致死锁** -- 恢复代价 MEDIUM。11 种审批类型需统一状态机（draft/pending/approved/rejected/cancelled/timeout），超时机制由 gocron 驱动
6. **病假系数计算不符合地方法规** -- 恢复代价 MEDIUM。参照社保政策库模式建 sick_leave_policies 配置表，计算结果必须校验 >= 当地最低工资 80%
7. **待办中心事项聚合的实时性与一致性** -- 恢复代价 LOW。采用物化 todo_items 表 + 事件驱动写入，避免跨 5+ 表 UNION ALL 实时聚合

## Implications for Roadmap

基于依赖关系分析、功能耦合度和陷阱预防需求，建议以下阶段结构：

### Phase 1: 基础数据与展示层

**Rationale:** 部门是薪资普调和组织架构的基础维度；数据看板是纯展示逻辑，无复杂业务规则，可快速交付用户价值；待办中心框架先行，为后续模块提供推送接口。
**Delivers:** 员工/薪资/社保 3 个数据看板、花名册增强、部门管理、待办中心框架（汇总列表 + 快捷入口 + 完成率环形图）、限时任务引擎
**Addresses:** P1 全部功能（FEATURES.md 待办中心标配 + 员工管理看板/花名册）
**Avoids:** Pitfall #8（待办聚合 -- 第一时间建立物化 todo_items 表和事件写入模式）
**Stack:** vue-echarts + @vueuse/core + ECharts（图表组件）、gocron（限时任务调度）

### Phase 2: 考勤管理与审批流

**Rationale:** 考勤是薪资计算增强的上游数据源，必须在薪资增强之前完成。审批流产生的请假/加班数据直接影响薪资核算。打卡规则引擎独立性强，可与审批流并行开发。
**Delivers:** 3 种打卡模式（固定/排班/自由）、审批流引擎（11 种审批类型）、今日打卡实况、出勤月报、薪资基础增强（看板 + 调薪/普调 + 绩效系数）
**Addresses:** P2 全部功能（FEATURES.md 考勤标配 + 薪资基础增强）
**Avoids:** Pitfall #1（跨天班次 -- 班次模型第一天就支持 workDateOffset）、Pitfall #2（状态机死锁 -- 先画状态转换图再写代码）、Pitfall #3（调薪历史覆盖 -- INSERT ONLY）、Pitfall #5（绩效系数审计 -- 独立月度表）
**Stack:** qmuntal/stateless（审批流状态机）、asynq（审批状态变更异步通知考勤）

### Phase 3: 数据联动与模块增强

**Rationale:** 考勤模块完成后，薪资算法增强（考勤联动）才具备数据基础。社保增强相对独立，可与薪资增强并行。员工信息登记和组织架构可视化作为最后一批功能。
**Delivers:** 薪资算法增强（考勤联动 -- 基本工资/加班费/病假工资自动计算）、个税上传自动抓取、社保增减员优化 + 缴费渠道 + 欠缴状态管理、组织架构可视化、员工信息登记、办离职优化（审批 + 社保减员联动）、待办中心完善（邀请协办/轮播图）
**Addresses:** P3 全部功能
**Avoids:** Pitfall #4（病假系数合规 -- 配置化政策库）、Pitfall #6（社保欠缴流转 -- 完整状态机 + 定时任务）、Pitfall #7（离职联动 -- 事件驱动 + 幂等消费）、Pitfall #9（组织架构性能 -- 邻接表 + path 字段 + 缓存）、Pitfall #10（工资公式变更 -- 已确认月份只读保护）
**Stack:** excelize（个税上传）、shopspring/decimal（深度使用于薪资精确计算）

### Phase Ordering Rationale

- **Phase 1 先行**：部门是跨模块基础数据（薪资普调按部门、组织架构按部门展示）；数据看板是低风险高价值功能，快速交付；待办中心框架先行可让后续模块逐步接入
- **Phase 2 承上启下**：考勤数据是薪资算法增强的必要输入（出勤天数、请假时长、加班时长），必须在薪资联动前完成；审批流状态机是 v1.3 最复杂的业务逻辑，需专注攻克
- **Phase 3 收尾联动**：跨模块深度集成（考勤->薪资、离职->社保）放在最后，此时各模块核心逻辑已稳定，联调风险最低
- **避免的重构风险**：调薪 INSERT ONLY 策略（Pitfall #3）在 Phase 2 建立基础时就落地，确保 Phase 3 薪资增强时历史数据已受保护

### Research Flags

需要深入研究的阶段：
- **Phase 2（审批流）：** 11 种审批类型的状态机设计需详细建模，每种类型的薪资影响规则不同（尤其是病假系数的地方法规差异），建议执行 `/gsd-research-phase` 细化状态转换矩阵
- **Phase 3（薪资考勤联动）：** 薪资算法增强是 v1.3 最重要的跨模块集成点，病假系数按城市配置、加班费三档费率、基本工资/应出勤/计薪天数的计算公式需精确到小数位，建议执行 `/gsd-research-phase` 细化计算引擎

标准模式可跳过研究的阶段：
- **Phase 1（数据看板/待办框架）：** 数据看板是标准 CRUD + 聚合查询，待办中心物化表模式在 PITFALLS.md 中已充分论证，无需额外研究
- **Phase 1（部门管理）：** 邻接表模式是经典方案，代码库中已有类似层级数据结构

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | 新增仅 3 个依赖（qmuntal/stateless、vue-echarts、@vueuse/core），均有 GitHub/npm 官方源验证版本兼容性 |
| Features | HIGH | 基于 PRD 1.1 详细需求 + 现有代码库分析 + 中国 HR SaaS 竞品（钉钉/飞书/企微/i人事）功能对比 |
| Architecture | HIGH | 基于现有代码库架构审查（adapter 接口模式、模块化单体、GORM 模型），v1.3 延续而非重构 |
| Pitfalls | HIGH | 基于代码审查（salary/service.go、socialinsurance/service.go、offboarding_service.go）+ 领域知识 + 搜索验证 |

**Overall confidence:** HIGH

### Gaps to Address

- **病假系数政策库初始数据：** 需要确定 v1.3 支持的城市范围。建议初期只支持一线城市（北上广深），通过 `sick_leave_policies` JSONB 配置录入，后续按需扩展
- **个税 Excel 模板标准化：** 个税上传自动抓取依赖用户上传的 Excel 格式。需要定义标准模板格式，非标准格式的容错处理策略需在 Phase 3 细化
- **排班模式复杂度控制：** 排班管理是 v1.3 单项复杂度最高的功能（班次定义 + 排班周期 + 员工分配 + 特殊日期 + 调班审批）。建议 Phase 2 先实现最简版（固定班次 + 手动排班表），不搞智能排班
- **缴费渠道对接：** 代理缴费的 API 对接在 v1.3 先做 UI 和流程，实际支付集成可后续迭代
- **审批流 7 种请假类型的硬编码 vs 配置化权衡：** v1.3 MVP 阶段硬编码 7 种假类型可接受，V2.0 需配置化。架构上通过 approval_type + subtype 区分，预留配置化扩展空间

## Sources

### Primary (HIGH confidence)
- PRD 1.1 (`prd1.1.md`) -- v1.3 详细产品需求定义
- 现有代码库审查 -- `internal/salary/`, `internal/socialinsurance/`, `internal/employee/`, `internal/dashboard/`, `cmd/server/main.go`
- qmuntal/stateless GitHub -- v1.8.0, 2026-02-10 发布
- vue-echarts npm -- v8.0.1, peer dependency echarts ^6.0.0 + vue ^3.3.0
- shopspring/decimal GitHub -- v1.4.0, 已在 go.mod 中
- ECharts 官方文档 -- tree 系列内置类型
- 上海市人社局 -- 病假工资计算法定标准
- 《企业职工患病或非因工负伤医疗期规定》 -- 医疗期计算规则

### Secondary (MEDIUM confidence)
- HR 系统选型实战：2026 年主流产品深度对比（mokahr.com） -- 竞品功能对比
- 薪资管理系统对比：2026 年主流产品深度评测（mokahr.com） -- 薪资管理行业标准
- 2025 年新版：员工月工资计算标准（sina.com.cn） -- 假期工资计算法规
- Go 设计模式实战 -- 用状态模式实现系统工作流（cnblogs.com） -- 审批流状态机参考
- 腾讯云开发者社区 -- 人事 OA 考勤管理架构
- 知乎 -- 考勤系统复杂班次工时计算
- 博客园 -- 考勤排班规则详解
- Medium -- Simple Workflow Engine in Go Using Stateless

### Tertiary (LOW confidence)
- 掘金/阿里云开发者社区 -- ECharts Tree 组织结构图实战案例
- vue3-tree-org npm -- 排除依据（3 年未更新）
- easy-workflow GitHub -- Go 审批流参考实现

---
*Research completed: 2026-04-17*
*Ready for roadmap: yes*
