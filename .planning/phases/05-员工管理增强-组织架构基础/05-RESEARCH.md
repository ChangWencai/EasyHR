# Phase 5: 员工管理增强 + 组织架构基础 - Research

**Researched:** 2026-04-17
**Domain:** 员工管理增强（数据看板、组织架构、信息登记、离职审批、花名册、Excel导出）
**Confidence:** HIGH

## Summary

Phase 5 在现有员工管理基础上进行六大增强：员工数据看板（4张纯数字卡片）、组织架构可视化（ECharts tree + 邻接表部门模型）、员工信息登记（Token链接机制 + 短信/二维码转发）、办离职优化（行内审批 + 社保减员联动）、花名册增强（6列默认显示 + Drawer详情 + 搜索）、Excel导出增强（扩展更多列）。

核心复用点：Invitation Token 机制可直接复用于员工信息登记链接（D-06）；Dashboard errgroup 并发聚合模式用于员工数据看板；excelize ExportExcel 已有基础可扩展；Contract 模型已存在可直接查询合同到期天数。关键新增：Department 模型（邻接表 parent_id）、ECharts tree 图表（需新装依赖）、阿里云 SMS 对接、离职驳回状态。

**Primary recommendation:** 新建 `internal/department/` 模块承载组织架构，Employee 模型新增 `department_id` 字段，前端安装 echarts + vue-echarts，员工信息登记复用 Invitation Token 模式但使用独立的 Registration 模型。

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- **D-01:** 纯数字卡片风格，4 张卡片（在职人数/当月新入职/当月离职/当月离职率），不引入图表组件
- **D-02:** 离职率仅展示当月数字（离职人数/(离职人数+期末人数)x100%），不带环比趋势
- **D-03:** 层级深度最多 3 层（部门->岗位->员工），使用邻接表模型（parent_id）存储部门层级
- **D-04:** 使用 ECharts tree 图表渲染组织架构，顶部搜索框输入关键字后树图自动定位并高亮匹配节点
- **D-05:** Department 模型包含 id/name/parent_id/org_id，Employee 模型新增 department_id 字段
- **D-06:** 独立 H5 页面 + Token 链接机制（复用现有 Invitation 模型的 Token 生成逻辑），员工无需登录即可填写
- **D-07:** 转发方式同时支持两种：二维码 + 复制链接（微信转发）和短信发送（对接阿里云 SMS）
- **D-08:** 员工提交后直接更新员工档案（提交即更新），管理员可后续编辑修正。以最新提交版本为准覆盖
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

### Deferred Ideas (OUT OF SCOPE)
None
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| EMP-01 | 员工数据看板展示在职人数、当月新入职人数、当月离职人数 | 复用 Dashboard GetEmployeeStats（active/joined/left），新建前端纯数字卡片组件 |
| EMP-02 | 员工数据看板展示离职率（离职人数/(离职人数+期末人数)x100%） | 在 GetEmployeeStats 基础上增加离职率计算，D-02 锁定仅当月数字 |
| EMP-03 | 管理员可查看组织架构可视化图表（部门名称->岗位名称->员工名称层级树） | 新建 Department 模型（邻接表），ECharts tree 渲染，D-03/D-04/D-05 已锁定方案 |
| EMP-04 | 组织架构支持按部门/岗位/员工检索定位 | 后端树形数据 API + 前端 ECharts search + 高亮节点 |
| EMP-05 | 管理员可创建员工信息登记表（必填：姓名/部门/岗位/入职日期） | 新建 Registration 模型，关联 Employee 和 Department |
| EMP-06 | 管理员填写或转发员工填写：手机号码/住址/身份证正反面/银行卡正反面/学历证书/紧急联系人 | Token 链接 H5 页面，D-07 短信+二维码两种转发 |
| EMP-07 | 员工信息登记支持重新创建已入库员工的信息（以手机号或身份证号匹配） | 以最新版为准覆盖（D-08），通过 phone_hash/id_card_hash 匹配已存在员工 |
| EMP-08 | 员工信息登记提交后，数据更新到员工个人信息 | 复用 Invitation SubmitInvitation 事务模式，提交即更新 Employee 记录 |
| EMP-09 | 离职待办列表展示事项/发起人/时间/状态/排序 | 扩展 OffboardingList 现有组件，新增 rejected 状态筛选 |
| EMP-10 | 管理员可审批离职申请（同意/驳回） | 新增 RejectResign 方法 + 行内按钮（D-09） |
| EMP-11 | 离职审批通过后可立即减员（跳转社保公积金减员） | approved 状态行内显示"去减员"按钮，router.push 携带 employee_id+name 参数 |
| EMP-12 | 减员完成后离职状态自动更新 | SocialInsuranceEventHandler 已有 OnEmployeeResigned 接口可扩展 |
| EMP-13 | 花名册显示员工/状态/岗位薪资/在职年限/合同到期天数/手机号码 | 关联 salary_items（岗位薪资）、计算 hire_date 差值（在职年限）、关联 contracts.end_date（到期天数） |
| EMP-14 | 花名册点击"更多"跳转员工个人信息窗口（基本信息+员工档案） | 新建 EmployeeDrawer 组件（D-10 Drawer 抽屉），调用 GetSensitiveInfo API |
| EMP-15 | 花名册支持按关键字搜索 | 扩展 ListEmployees 搜索参数（已支持 search），增加 department_id 筛选 |
| EMP-16 | 员工列表支持 Excel 格式下载导出 | 扩展现有 ExportExcel，增加列：岗位薪资/在职年限/合同到期天数/部门 |
</phase_requirements>

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| 员工数据看板（4卡片数字） | Frontend (Vue) | API (Go) | 前端负责卡片渲染，后端仅提供聚合数据 |
| 离职率计算 | API (Go) | -- | 后端计算离职率公式，前端仅展示 |
| 组织架构树数据构建 | API (Go) | Database (PostgreSQL) | 后端递归查询部门树 + 员工，构建 ECharts tree JSON |
| 组织架构树渲染 | Frontend (Vue + ECharts) | -- | 前端负责 ECharts tree 配置、搜索定位、节点高亮 |
| 部门 CRUD | API (Go) | Database (PostgreSQL) | 标准 Handler->Service->Repository 三层 |
| 员工信息登记表创建 | API (Go) | Database (PostgreSQL) | 管理员创建登记表，生成 Token |
| 员工信息登记 H5 页面 | Frontend (Vue) | -- | 独立路由 /register/:token，无需登录 |
| 短信发送 | API (Go) | Aliyun SMS | 后端调用阿里云 SMS API |
| 离职审批（同意/驳回） | API (Go) | Database (PostgreSQL) | 状态转换 + 事件触发 |
| 社保减员联动 | Frontend (路由跳转) | API (Go - 状态回调) | 前端跳转传参，后端接收减员完成回调 |
| 花名册多列展示 | Frontend (Vue + Element Plus) | API (Go) | 前端表格列配置，后端返回扩展字段 |
| 花名册 Drawer 详情 | Frontend (Vue) | API (Go) | 前端 Drawer 组件，后端 GetSensitiveInfo |
| Excel 导出增强 | API (Go) | -- | excelize 服务端生成，扩展列数据 |

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| ECharts | 6.0.0 | 组织架构树可视化 | D-04 锁定选择，tree 图表类型成熟 [VERIFIED: npm registry] |
| vue-echarts | 8.0.1 | Vue 3 ECharts 组件封装 | 官方推荐 Vue 集成方案，提供 v-chart 指令 [VERIFIED: npm registry] |
| excelize | v2.10.1 | Excel 导出增强 | 项目已安装，扩展更多列即可 [VERIFIED: go.mod] |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| resty | v2.17.2 | 阿里云 SMS HTTP 调用 | 发送员工信息登记短信时使用 [VERIFIED: go.mod] |
| golang-jwt | v5.3.1 | Token 生成 | 复用 Invitation Token 生成模式（crypto/rand 32-byte hex） [VERIFIED: go.mod] |
| dayjs | 1.11.20 | 前端日期计算 | 在职年限、合同到期天数前端展示用 [VERIFIED: package.json] |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| ECharts tree | D3.js tree | D3 更灵活但学习曲线陡峭，ECharts tree 开箱即用、中国生态成熟 |
| 邻接表 (parent_id) | 物化路径 (ltree) | 物化路径查询更快但 PostgreSQL 特有复杂度高，3 层深度邻接表够用 |
| 阿里云 SMS | 腾讯云 SMS | 项目已确定阿里云生态（ECS/OSS），SMS 保持一致 |

**Installation:**

Wave 0 前端依赖安装：
```bash
cd frontend && npm install echarts@6.0.0 vue-echarts@8.0.1
```

后端无需新增依赖（resty 已在 go.mod 中）。

## Architecture Patterns

### System Architecture Diagram

```
[管理员浏览器]
    |
    |--- GET /api/v1/employee/dashboard -----> [Handler.GetDashboard]
    |                                          |-> errgroup 并发查询
    |                                          |   -> GetEmployeeStats (active/joined/left)
    |                                          |   -> 计算离职率
    |                                          |-> 返回 DashboardResult
    |
    |--- GET /api/v1/departments/tree --------> [DepartmentHandler.GetTree]
    |                                          |-> 递归查询 departments + employees
    |                                          |-> 构建 ECharts tree JSON
    |
    |--- POST /api/v1/registrations ----------> [RegistrationHandler.Create]
    |                                          |-> 生成 Token (crypto/rand)
    |                                          |-> 存储 Registration 记录
    |
    |--- POST /api/v1/offboardings/:id/reject > [OffboardingHandler.Reject]
    |                                          |-> 更新状态为 rejected
    |
[员工手机浏览器 (H5 /register/:token)]
    |
    |--- GET /api/v1/public/register/:token --> [RegistrationHandler.GetDetail] (无需认证)
    |                                          |-> 验证 Token 有效性
    |
    |--- POST /api/v1/public/register/:token -> [RegistrationHandler.Submit] (无需认证)
    |                                          |-> 验证 Token + 过期时间
    |                                          |-> 匹配已有员工 (phone_hash/id_card_hash)
    |                                          |-> 加密敏感字段
    |                                          |-> 事务更新 Employee 记录
    |                                          |-> 更新 Registration 状态为 used
    |
[阿里云 SMS]
    |
    |<-- POST https://dysmsapi.aliyuncs.com/ -- [SMSService.SendRegistrationLink]
         (resty HTTP 调用)                      |-> 短信模板 + Token 链接
```

### Recommended Project Structure

```
internal/
├── department/           # 新增：部门/组织架构模块
│   ├── model.go          # Department 模型 (id/name/parent_id/org_id/sort_order)
│   ├── dto.go            # 请求/响应 DTO
│   ├── repository.go     # 数据访问层
│   ├── service.go        # 业务逻辑 (CRUD + 树构建)
│   └── handler.go        # HTTP 端点
├── employee/
│   ├── model.go          # 新增 department_id 字段
│   ├── service.go        # 扩展 ExportExcel、ListEmployees
│   ├── offboarding_model.go   # 新增 rejected 状态常量
│   ├── offboarding_service.go # 新增 RejectResign 方法
│   ├── registration_model.go  # 新增：员工信息登记模型
│   ├── registration_service.go # 新增：登记业务逻辑（复用 Token 模式）
│   ├── registration_handler.go # 新增：登记 HTTP 端点（含公开路由）
│   └── contract_model.go # 已存在：Contract 模型（用于合同到期天数查询）
├── dashboard/
│   └── service.go        # 扩展 GetEmployeeStats 增加离职率
└── sms/                  # 新增：短信服务模块（或放在 internal/common/sms/）
    ├── service.go        # 阿里云 SMS 封装
    └── templates.go      # 短信模板常量

frontend/src/
├── views/employee/
│   ├── EmployeeDashboard.vue  # 新增：员工数据看板（4卡片）
│   ├── OrgChart.vue           # 新增：组织架构可视化（ECharts tree）
│   ├── EmployeeList.vue       # 扩展：新增列 + Drawer + 搜索
│   ├── EmployeeDrawer.vue     # 新增：员工详情抽屉
│   ├── RegistrationList.vue   # 新增：信息登记管理列表
│   ├── RegistrationCreate.vue # 新增：创建信息登记表
│   ├── RegisterPage.vue       # 新增：员工填写信息 H5（独立路由，无登录）
│   └── OffboardingList.vue    # 扩展：同意/驳回/去减员
├── api/
│   ├── employee.ts       # 扩展：新增 dashboard/department/registration API
│   └── department.ts     # 新增：部门 CRUD API
└── router/
    └── index.ts          # 扩展：新增路由（org-chart, register/:token）
```

### Pattern 1: Department Tree 构建（邻接表递归）

**What:** 从扁平的部门记录（parent_id）构建嵌套的树形结构
**When to use:** 组织架构树 API、ECharts tree 数据源
**Example:**

```go
// Source: 基于 D-03/D-05 决策 + GORM 邻接表标准模式
type Department struct {
    model.BaseModel
    Name     string  `gorm:"column:name;type:varchar(100);not null" json:"name"`
    ParentID *int64  `gorm:"column:parent_id;index" json:"parent_id"`
    SortOrder int    `gorm:"column:sort_order;not null;default:0" json:"sort_order"`
}

type TreeNode struct {
    ID       int64       `json:"id"`
    Name     string      `json:"name"`
    Type     string      `json:"type"` // "department" | "position" | "employee"
    Children []*TreeNode `json:"children,omitempty"`
}

// BuildTree 从扁平记录构建树（单次查询 + 内存递归，3层深度足够）
func BuildTree(depts []Department, employees []Employee) []*TreeNode {
    // 1. 按 parent_id 分组
    // 2. 递归构建部门树
    // 3. 员工挂载到岗位节点下
}
```

### Pattern 2: Registration Token 复用 Invitation 模式

**What:** 独立 Registration 模型，复用 crypto/rand Token 生成
**When to use:** 员工信息登记链接生成和验证
**Example:**

```go
// Source: 复用 internal/employee/invitation_service.go generateToken() 模式
type Registration struct {
    ID            int64      `gorm:"primaryKey;autoIncrement"`
    OrgID         int64      `gorm:"column:org_id;index;not null"`
    EmployeeID    *int64     `gorm:"column:employee_id;index"`
    Token         string     `gorm:"column:token;type:varchar(64);uniqueIndex;not null"`
    Status        string     `gorm:"column:status;type:varchar(20);default:pending"`
    ExpiresAt     time.Time  `gorm:"column:expires_at;not null"`
    UsedAt        *time.Time `gorm:"column:used_at"`
    CreatedBy     int64      `gorm:"column:created_by;not null"`
    CreatedAt     time.Time  `gorm:"column:created_at;autoCreateTime"`
}
// Token 生成复用: generateToken() -> crypto/rand 32-byte -> hex 64 chars
```

### Pattern 3: ECharts Tree 配置（Vue 组件）

**What:** vue-echarts 组件渲染组织架构树
**When to use:** OrgChart.vue 页面
**Example:**

```vue
<!-- Source: ECharts 6 tree 配置 + vue-echarts 8 标准用法 -->
<template>
  <v-chart :option="chartOption" autoresize style="height: 600px" />
</template>

<script setup lang="ts">
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import { TreeChart } from 'echarts/charts'
import { TooltipComponent, TitleComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'

use([TreeChart, TooltipComponent, TitleComponent, CanvasRenderer])

// chartOption.data 格式: { name, children: [{ name, children: [...] }] }
// 搜索高亮: 遍历 data 设置 itemStyle.color / label.color
</script>
```

### Pattern 4: Offboarding 状态扩展（rejected）

**What:** 在现有 pending/approved/completed 基础上新增 rejected 状态
**When to use:** 离职审批驳回功能
**Example:**

```go
// Source: 扩展 internal/employee/offboarding_model.go
const OffboardingStatusRejected = "rejected" // 新增

// RejectResign 驳回离职申请
func (s *OffboardingService) RejectResign(orgID, rejectorID, offboardingID int64, reason string) error {
    ob, err := s.obRepo.FindByID(orgID, offboardingID)
    if err != nil {
        return fmt.Errorf("离职记录不存在")
    }
    if ob.Status != OffboardingStatusPending {
        return fmt.Errorf("当前状态不可驳回（状态: %s）", ob.Status)
    }
    updates := map[string]interface{}{
        "status":     OffboardingStatusRejected,
        "updated_by": rejectorID,
    }
    return s.obRepo.Update(orgID, offboardingID, updates)
}
```

### Anti-Patterns to Avoid

- **N+1 查询花名册列表:** 花名册新增"岗位薪资"和"合同到期天数"列时，不要在循环中逐条查询 salary_items 和 contracts，应使用 JOIN 或批量预加载
- **递归 SQL 查询组织架构树:** 3 层深度的邻接表不应使用 WITH RECURSIVE，应一次性查询全部记录后内存构建树
- **Employee 模型直接存储薪资:** 不要在 Employee 表新增 salary 字段，应关联查询 salary_items，保持数据来源单一
- **前端 ECharts 全量引入:** 不要 `import * as echarts from 'echarts'`，应按需引入 TreeChart + TooltipComponent + CanvasRenderer

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Token 安全生成 | 自定义随机字符串 | crypto/rand 32-byte hex (复用 generateToken) | 密码学安全随机数，已有成熟实现 |
| 敏感字段加密 | 自定义加密 | crypto.Encrypt/HashSHA256 (已有) | AES-256-GCM + SHA-256 双列模式已验证 |
| 手机号/身份证脱敏 | 自定义截取 | crypto.MaskPhone/MaskIDCard (已有) | 考虑了长度边界和格式规范 |
| Excel 文件生成 | 手动拼接 CSV | excelize v2.10.1 (已有) | 支持 xlsx 格式、样式、冻结首行 |
| 组织架构树渲染 | 自定义 DOM 树组件 | ECharts tree 图表 | 拖拽/缩放/搜索定位开箱即用 |
| 阿里云 SMS 签名 | 自行实现 API 签名 | resty + 阿里云 SMS SDK 签名算法 | HMAC-SHA1 签名复杂度高，有现成库 |
| 前端状态管理 | 组件内 ref 存储全局数据 | Pinia store | 跨组件共享部门列表、员工数据 |

**Key insight:** 本 Phase 大部分功能可复用现有模式。真正"新"的部分只有：Department 模型（新建）、ECharts tree（新依赖）、阿里云 SMS（新对接）、rejected 状态（扩展）。

## Common Pitfalls

### Pitfall 1: ECharts Tree 大数据量性能

**What goes wrong:** 企业 50 人 + 3 层树可能有 100+ 节点，ECharts 默认配置下初始渲染和搜索高亮变慢
**Why it happens:** ECharts tree 默认开启全部动画和阴影效果
**How to avoid:** 关闭初始动画 (`animation: false`)、使用 `series.roam: true` 支持缩放平移、搜索时仅更新受影响节点的 `itemStyle`
**Warning signs:** 树图首次渲染超过 2 秒、搜索输入卡顿

### Pitfall 2: 员工信息登记 Token 安全

**What goes wrong:** Token 被猜测或泄露，恶意提交员工信息
**Why it happens:** Token 生成不够随机、有效期过长、无频率限制
**How to avoid:** 使用 crypto/rand 32-byte（已有）、7 天有效期（与 Invitation 一致）、公开接口添加 IP 频率限制
**Warning signs:** 短时间内大量提交请求

### Pitfall 3: 花名册多列数据聚合性能

**What goes wrong:** 花名册每行需要关联 salary_items（薪资）、contracts（合同）、employees（在职年限），100 条数据触发 300+ 查询
**Why it happens:** ListEmployees 方法目前只查 employees 表
**How to avoid:** 使用单条 SQL JOIN 查询或预加载 (`Preload`)，在 Repository 层做一次聚合查询返回 DTO
**Warning signs:** 花名册加载时间超过 1 秒

### Pitfall 4: Department 删除时的级联影响

**What goes wrong:** 删除部门后，该部门下员工的 department_id 成为悬空引用
**Why it happens:** 未检查部门下是否有员工或子部门
**How to avoid:** 删除前校验：部门下无员工、无子部门。否则返回错误提示
**Warning signs:** 员工列表中 department_name 显示为空

### Pitfall 5: 离职率计算精度

**What goes wrong:** 期末人数为 0 时除零错误，或离职率为负数
**Why it happens:** 公式 `离职人数/(离职人数+期末人数)` 在极端情况下边界条件未处理
**How to avoid:** 分母为 0 时返回 0%，结果 clamp 到 [0, 100]
**Warning signs:** 离职率显示 NaN 或负值

### Pitfall 6: 阿里云 SMS 模板审核延迟

**What goes wrong:** 短信模板需要审核（通常 2 小时，偶尔 1-2 天），审核期间功能不可用
**Why it happens:** 阿里云 SMS 短信签名和模板必须审核通过后才能发送
**How to avoid:** Wave 0 即创建短信签名和模板、同时支持二维码+链接复制作为备用转发方式（D-07）
**Warning signs:** 短信发送返回 `isv.SMS_TEMPLATE_ILLEGAL`

## Code Examples

### 员工数据看板 - 后端扩展

```go
// Source: 扩展 internal/dashboard/service.go GetDashboard 模式
type EmployeeDashboardResult struct {
    ActiveCount      int     `json:"active_count"`
    JoinedThisMonth  int     `json:"joined_this_month"`
    LeftThisMonth    int     `json:"left_this_month"`
    TurnoverRate     float64 `json:"turnover_rate"` // 离职率百分比
}

func (s *Service) GetEmployeeDashboard(ctx context.Context, orgID int64) (*EmployeeDashboardResult, error) {
    active, joined, left, err := s.repo.GetEmployeeStats(ctx, orgID)
    if err != nil {
        return nil, err
    }

    // 离职率 = 离职人数 / (离职人数 + 期末人数) x 100%  (D-02)
    turnoverRate := 0.0
    denominator := float64(left + active)
    if denominator > 0 {
        turnoverRate = float64(left) / denominator * 100
    }

    return &EmployeeDashboardResult{
        ActiveCount:     active,
        JoinedThisMonth: joined,
        LeftThisMonth:   left,
        TurnoverRate:    math.Round(turnoverRate*100) / 100, // 保留2位小数
    }, nil
}
```

### Department 模型定义

```go
// Source: D-05 决策 + BaseModel 模式
package department

import "github.com/wencai/easyhr/internal/common/model"

type Department struct {
    model.BaseModel
    Name      string `gorm:"column:name;type:varchar(100);not null;comment:部门名称" json:"name"`
    ParentID  *int64 `gorm:"column:parent_id;index;comment:父部门ID（顶级为空）" json:"parent_id"`
    SortOrder int    `gorm:"column:sort_order;not null;default:0;comment:排序" json:"sort_order"`
}

func (Department) TableName() string { return "departments" }
```

### Registration 提交 - 更新员工档案

```go
// Source: 复用 invitation_service.go SubmitInvitation 事务模式 + D-08 提交即更新
func (s *RegistrationService) SubmitRegistration(token string, req *SubmitRegistrationRequest) error {
    reg, err := s.regRepo.FindByToken(token)
    // ... Token 验证 + 过期检查 ...

    aesKey := s.aesKey()

    // 查找已存在员工（通过 phone_hash 或 id_card_hash）
    var emp *Employee
    if req.Phone != "" {
        phoneHash := crypto.HashSHA256(req.Phone)
        emp, _ = s.empRepo.FindByPhoneHash(reg.OrgID, phoneHash)
    }
    if emp == nil && req.IDCard != "" {
        idCardHash := crypto.HashSHA256(req.IDCard)
        emp, _ = s.empRepo.FindByIDCardHash(reg.OrgID, idCardHash)
    }

    return s.regRepo.DB().Transaction(func(tx *gorm.DB) error {
        if emp != nil {
            // D-08: 以最新版本覆盖更新 Employee
            updates := buildEmployeeUpdates(req, aesKey)
            if err := tx.Model(&Employee{}).Where("id = ?", emp.ID).Updates(updates).Error; err != nil {
                return fmt.Errorf("更新员工信息失败: %w", err)
            }
        } else {
            // 新员工：创建 Employee 记录
            newEmp := buildNewEmployee(reg, req, aesKey)
            if err := tx.Create(newEmp).Error; err != nil {
                return fmt.Errorf("创建员工失败: %w", err)
            }
        }

        // 更新 Registration 状态
        now := time.Now()
        return tx.Model(&Registration{}).Where("token = ?", token).Updates(map[string]interface{}{
            "status": "used",
            "used_at": &now,
        }).Error
    })
}
```

### 花名册扩展查询 - 多表关联

```go
// Source: 避免 N+1 的批量聚合查询模式
type EmployeeRosterItem struct {
    ID                  int64   `json:"id"`
    Name                string  `json:"name"`
    Status              string  `json:"status"`
    Position            string  `json:"position"`
    DepartmentName      string  `json:"department_name"`
    Phone               string  `json:"phone"` // 脱敏后
    SalaryAmount        float64 `json:"salary_amount"`
    YearsOfService      float64 `json:"years_of_service"` // 在职年限（年）
    ContractExpiryDays  *int    `json:"contract_expiry_days"` // 合同到期天数（nil=无合同/无固定期限）
}

// Repository 层一条 SQL 完成（或分批预加载）
func (r *Repository) ListRoster(orgID int64, params SearchParams, page, pageSize int) ([]EmployeeRosterItem, int64, error) {
    // 方案：先查 employees 分页 -> 批量查 salary_items -> 批量查 contracts
    // 内存中组装 DTO，避免复杂 JOIN 影响分页
}
```

### 离职审批行内操作 - 前端

```vue
<!-- Source: D-09 决策，行内同意/驳回 + 去减员 -->
<el-table-column label="操作" width="260" fixed="right">
  <template #default="{ row }">
    <el-button v-if="row.status === 'pending'" size="small" type="primary"
      @click="handleApprove(row.id)">同意</el-button>
    <el-button v-if="row.status === 'pending'" size="small" type="danger"
      @click="handleReject(row.id)">驳回</el-button>
    <el-button v-if="row.status === 'approved'" size="small" type="warning"
      @click="goToSIRegister(row.employee_id, row.employee_name)">去减员</el-button>
    <el-button v-if="row.status === 'approved'" size="small" type="success"
      @click="handleComplete(row.id)">完成离职</el-button>
  </template>
</el-table-column>
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| ECharts 5.x import 方式 | ECharts 6.x 按需引入 (TreeChart) | ECharts 6.0 2025 | 包体积更小，tree API 向后兼容 |
| vue-echarts 手动 init | vue-echarts 8 的 `<v-chart>` 组件 | vue-echarts 8.0 2025 | Composition API 原生支持 |
| Invitation Token 单一用途 | Registration Token 多场景 | Phase 5 新增 | Token 模式扩展为通用"无登录填写"机制 |
| Offboarding 3 状态 | Offboarding 4 状态（+rejected） | Phase 5 新增 | 需要确保 rejected 在所有筛选/列表中被处理 |

**Deprecated/outdated:**
- robfig/cron: 已停更 4 年+，项目已选用 gocron v2（与本 Phase 无关但注意避免引入）

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | 阿里云 SMS API 签名算法可用 resty 手动实现，无需专用 SDK | Standard Stack | 可能需要引入 aliyun-go-sdk-sms 增加依赖 |
| A2 | ECharts 6.0 tree 配置与 5.x 完全向后兼容 | Architecture Patterns | 可能有 breaking change 需调整配置 |
| A3 | 花名册"岗位薪资"取 salary_items 最近一条 effective_month 的 amount | Phase Requirements | 可能需要取特定薪资项而非全部 amount 之和 |
| A4 | Contract 模型的 end_date 可直接用于计算到期天数 | Phase Requirements | 已确认模型存在且有 EndDate 字段 |
| A5 | 邻接表 3 层深度不需要 WITH RECURSIVE，单次全量查询+内存构建即可 | Architecture Patterns | 如果层级超过 5 层需要改方案，但 D-03 锁定最多 3 层 |

## Open Questions

1. **阿里云 SMS 签名和模板审核** (RESOLVED)
   - What we know: 阿里云 SMS 需要签名+模板审核才能使用，审核时间 2 小时到 2 天
   - What's unclear: 签名是否已注册、模板内容是否需要法务审核
   - Recommendation: Wave 0 即提交签名和模板申请，同时以二维码+链接复制为降级方案
   - Resolution: Plan 05-03 user_setup 已声明阿里云 SMS 配置需求。D-07 已锁定同时支持二维码+链接复制作为降级方案，SMS 审核期间不影响核心功能。执行时由 Plan 05-03 Task 1 在 pkg/sms/client.go 中新增 SendTemplateMessage 方法，TestMode 降级为空操作。

2. **花名册"岗位薪资"数据来源** (RESOLVED)
   - What we know: salary_items 有 employee_id + amount + effective_month
   - What's unclear: "岗位薪资"是取基本工资单项还是所有 income 项之和
   - Recommendation: 取最近的 effective_month 中 template_item 对应"基本工资"的 amount（最简单理解），Claude's Discretion 范围内决定
   - Resolution: 按 Recommendation 方案执行。Plan 05-05 Task 1 在 GetSalaryAmounts 方法中取 salary_items 最近 effective_month 中 template_item 对应"基本工资"的 amount。此决策属于 Claude's Discretion 范围。

3. **社保减员页面跳转参数传递方式** (RESOLVED)
   - What we know: D-09 要求"去减员"按钮跳转社保减员页面
   - What's unclear: 社保减员页面是否已存在，接口是否支持预填 employee_id
   - Recommendation: router.push({ path: '/tool/socialinsurance', query: { action: 'reduce', employee_id, employee_name } })
   - Resolution: 按 Recommendation 方案执行。Plan 05-04 Task 2 使用 router.push 携带 query 参数跳转。减员完成后自动更新离职状态（EMP-12）通过 Plan 05-04 Task 1 确认 SocialInsuranceEventHandler.OnEmployeeResigned 回调或新增 CompleteOffboardingFromSI 方法实现。

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go | 后端全部 | ✓ | 1.25.0 | -- |
| Node.js | 前端构建 | ✓ | 24.1.0 | -- |
| npm | 前端依赖 | ✓ | 11.3.0 | -- |
| PostgreSQL | 数据存储 | ✓ (dev) | 16+ | -- |
| Redis | 缓存 | ✓ | -- | go-redis 已在 go.mod |
| ECharts | 组织架构树 | ✗ (未安装) | -- | Wave 0 安装 |
| vue-echarts | Vue ECharts 集成 | ✗ (未安装) | -- | Wave 0 安装 |
| 阿里云 SMS | 短信发送 | ✗ (未配置) | -- | 降级为二维码+链接 |

**Missing dependencies with no fallback:**
- ECharts + vue-echarts: Wave 0 必须 `npm install`
- 阿里云 SMS 配置: 需要 SMS_ACCESS_KEY / SMS_ACCESS_SECRET 环境变量

**Missing dependencies with fallback:**
- 阿里云 SMS: 二维码+复制链接方式可替代短信发送（D-07 要求同时支持两种）

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go testing + testify (后端), Vitest 4.1.4 (前端) |
| Config file | 无独立 vitest.config（Vite 内联配置） |
| Quick run command (后端) | `go test -race ./internal/employee/... ./internal/department/... -run TestXxx -v` |
| Quick run command (前端) | `cd frontend && npm run test:unit -- --reporter=verbose` |
| Full suite command (后端) | `go test -race ./... -cover` |
| Full suite command (前端) | `cd frontend && npm run test:unit` |

### Phase Requirements -> Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| EMP-01 | 员工数据看板返回在职/新入职/离职人数 | unit | `go test ./internal/dashboard/... -run TestGetEmployeeDashboard -v` | ❌ Wave 0 |
| EMP-02 | 离职率计算正确（含边界：分母为0） | unit | `go test ./internal/dashboard/... -run TestTurnoverRate -v` | ❌ Wave 0 |
| EMP-03 | 组织架构树构建（3层部门+员工） | unit | `go test ./internal/department/... -run TestBuildTree -v` | ❌ Wave 0 |
| EMP-04 | 组织架构搜索匹配部门/岗位/员工 | unit | `go test ./internal/department/... -run TestSearchTree -v` | ❌ Wave 0 |
| EMP-05 | 创建员工信息登记表（Token生成） | unit | `go test ./internal/employee/... -run TestCreateRegistration -v` | ❌ Wave 0 |
| EMP-06 | Token 有效期验证 + 过期拒绝 | unit | `go test ./internal/employee/... -run TestRegistrationExpiry -v` | ❌ Wave 0 |
| EMP-07 | 已入库员工信息覆盖更新 | unit | `go test ./internal/employee/... -run TestRegistrationOverwrite -v` | ❌ Wave 0 |
| EMP-08 | 提交后员工档案数据一致性 | unit | `go test ./internal/employee/... -run TestSubmitRegistration -v` | ❌ Wave 0 |
| EMP-09 | 离职列表包含 rejected 状态 | unit | `go test ./internal/employee/... -run TestListOffboardingsWithRejected -v` | ❌ Wave 0 |
| EMP-10 | 离职审批同意/驳回状态转换 | unit | `go test ./internal/employee/... -run TestRejectResign -v` | ❌ Wave 0 |
| EMP-11 | 审批通过后跳转减员参数 | integration | 手动验证 | N/A |
| EMP-12 | 减员完成后离职状态更新 | integration | 手动验证 | N/A |
| EMP-13 | 花名册返回薪资/年限/合同到期/手机号 | unit | `go test ./internal/employee/... -run TestListRoster -v` | ❌ Wave 0 |
| EMP-14 | Drawer 展示完整员工信息 | unit (前端) | `cd frontend && npx vitest run --reporter=verbose` | ❌ Wave 0 |
| EMP-15 | 花名册关键字搜索（姓名/部门/岗位） | unit | `go test ./internal/employee/... -run TestRosterSearch -v` | ❌ Wave 0 |
| EMP-16 | Excel 导出包含新增列 | unit | `go test ./internal/employee/... -run TestExportExcelEnhanced -v` | ❌ Wave 0 |

### Sampling Rate

- **Per task commit:** `go test -race ./internal/employee/... ./internal/department/... -v`
- **Per wave merge:** `go test -race ./... -cover`
- **Phase gate:** 全部后端 + 前端测试通过

### Wave 0 Gaps

- [ ] `internal/department/model_test.go` -- covers EMP-03/EMP-04
- [ ] `internal/department/service_test.go` -- covers Department CRUD + Tree 构建
- [ ] `internal/employee/registration_model_test.go` -- covers Registration Token 生成
- [ ] `internal/employee/registration_service_test.go` -- covers EMP-05~EMP-08
- [ ] `internal/employee/offboarding_service_test.go` -- covers EMP-10（RejectResign）
- [ ] `internal/dashboard/service_test.go` -- covers EMP-01/EMP-02
- [ ] 前端: `npm install echarts vue-echarts` -- ECharts 依赖安装
- [ ] 阿里云 SMS 签名+模板申请

## Security Domain

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | yes | 员工信息登记 Token 验证（替代认证），现有 JWT 用于管理员端 |
| V3 Session Management | no | 员工信息登记无 Session，Token 一次性使用 |
| V4 Access Control | yes | 公开接口仅限 Token 验证；管理员接口 RequireRole(owner/admin) |
| V5 Input Validation | yes | go-playground/validator struct tag + 前端表单校验 |
| V6 Cryptography | yes | AES-256-GCM 加密敏感字段 + SHA-256 哈希索引（已有） |

### Known Threat Patterns for Go + Vue + Token Auth

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Token 枚举/暴力破解 | Tampering | crypto/rand 32-byte（2^256 空间）+ 7天过期 + IP 频率限制 |
| XSS 通过员工填写内容 | Tampering | Vue 模板自动转义 + DOMPurify（如需富文本） |
| CSRF 伪造提交 | Spoofing | Token 绑定 org_id + EmployeeID，非 Session 机制 |
| 敏感数据泄露（手机号/身份证） | Information Disclosure | 双列加密模式 + toResponse 脱敏 + GetSensitiveInfo 需 ADMIN 权限 |
| SQL 注入（搜索参数） | Tampering | GORM 参数化查询（已有） |

## Sources

### Primary (HIGH confidence)
- `internal/employee/model.go` -- Employee 模型结构、敏感字段双列模式
- `internal/employee/invitation_model.go` + `invitation_service.go` -- Token 生成和验证机制
- `internal/employee/offboarding_model.go` + `offboarding_service.go` -- 离职状态和审批流程
- `internal/dashboard/service.go` + `repository.go` -- Dashboard errgroup 聚合模式
- `internal/salary/model.go` -- SalaryItem/PayrollRecord 模型
- `internal/employee/contract_model.go` -- Contract 模型（EndDate 字段）
- `frontend/package.json` -- 前端依赖现状
- npm registry: ECharts 6.0.0, vue-echarts 8.0.1 -- 版本验证

### Secondary (MEDIUM confidence)
- CONTEXT.md D-01~D-11 -- 用户锁定决策
- REQUIREMENTS.md EMP-01~EMP-16 -- 需求定义
- STATE.md -- 项目状态和已确认决策

### Tertiary (LOW confidence)
- A1: 阿里云 SMS resty 手动签名可行性（需验证 API 文档）
- A2: ECharts 6.0 tree 完全向后兼容（需实际测试）

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - 所有核心依赖已在项目中使用或版本已验证
- Architecture: HIGH - 复用现有模式（三层架构、Token 机制、errgroup），新模块设计明确
- Pitfalls: MEDIUM - 基于项目经验推断，阿里云 SMS 集成风险需实际验证

**Research date:** 2026-04-17
**Valid until:** 2026-05-17（30 天，稳定技术栈）
