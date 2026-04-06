# Architecture Patterns

**Domain:** 小微企业人事管理系统（EasyHR / 易人事）
**Researched:** 2026-04-06
**Confidence:** HIGH

## Recommended Architecture

**Modular Monolith** -- 单进程、单数据库、按业务边界划分模块的 Go 应用。

```
                    +---------------------------------------------+
                    |            Clients (多端)                    |
                    | Android(Kotlin) iOS(Swift) H5(Vue3) MiniProg |
                    +------+------------------+-------------------+
                           |                  |
                    HTTPS / REST API (JSON)   |
                           |                  |
                    +------v------------------v-------------------+
                    |              API Gateway Layer              |
                    | Middleware: Auth / RateLimit / CORS / Audit |
                    +------+-------------------------------------+
                           |
          +----------------+----------------+
          |                |                |
   +------v------+  +------v------+  +------v------+
   | user module |  |employee mod |  | social mod  |  ...
   | Handler     |  | Handler     |  | Handler     |
   | Service     |  | Service     |  | Service     |
   | Repository  |  | Repository  |  | Repository  |
   +------+------+  +------+------+  +------+------+
          |                |                |
          +--------+-------+--------+-------+
                   |                |
            +------v------+  +------v------+
            | PostgreSQL  |  |   Redis     |
            | (shared DB) |  | (cache/sess)|
            +-------------+  +-------------+
```

### Why Modular Monolith (Not Microservices)

| Factor | Modular Monolith | Microservices |
|--------|-----------------|---------------|
| 运维复杂度 | 单二进制部署 | 6+ 服务需要 K8s/服务网格 |
| 团队规模 | 1-3 人可维护 | 每服务至少 2 人 |
| 数据一致性 | 单库事务，强一致 | 分布式事务，最终一致 |
| 开发速度 | 模块间直接调用 | 需要定义 API/Proto |
| V1.0 用户量 | <1000 在线 | 大材小用 |
| 未来拆分 | 按模块边界拆分即可 | 已经拆好 |

**关键决策：模块边界即未来微服务边界。** 模块间通过明确的接口通信，禁止直接跨模块访问 Repository。

---

## Component Boundaries

### Module Dependency Graph

```
                    +-----------+
                    |   user    |  用户/组织/权限
                    | (user/org)|
                    +-----+-----+
                          |
               +----------+----------+
               |                     |
        +------v------+       +------v------+
        |  employee   |       |notification |
        | (员工管理)   |       | (通知/消息)  |
        +------+------+       +-------------+
               |
       +-------+-------+
       |       |       |
  +----v--+ +--v---+ +-v--------+
  |social | |payroll| |  tax     |
  |(社保) | |(工资) | | (个税)   |
  +----+--+ +--+---+ +----+-----+
       |       |           |
       +-------+-----------+
               |
        +------v------+
        |   finance   |
        | (财务/记账)  |
        +-------------+
```

### Component Definitions

| Component | Responsibility | Owns Tables | Communicates With |
|-----------|---------------|-------------|-------------------|
| **user** | 认证、用户管理、组织管理、RBAC 权限 | users, organizations | 所有模块（提供当前用户/组织上下文） |
| **employee** | 员工入职、离职、档案、合同、银行账户 | employees, contracts, employee_bank_accounts, resignation_checklists, attachments | user (获取组织信息), social (离职触发社保停缴), payroll (薪资结构) |
| **social** | 社保参保、停缴、变更、政策匹配、缴费提醒 | social_insurance, social_history, social_policies | employee (获取员工信息), payroll (提供社保扣款数据), notification (到期提醒) |
| **payroll** | 工资表创建、核算、发放、工资单推送 | salary_tables, salary_records, salary_payments | employee (员工列表), social (社保扣款), tax (个税扣款), finance (发放生成凭证) |
| **tax** | 个税计算、专项附加扣除、申报提醒、增值税/企业所得税 | tax_records, special_deductions, special_deduction_details | employee (员工信息), payroll (工资数据), notification (申报提醒) |
| **finance** | 会计科目、凭证、账簿、发票、费用报销、会计期间、报表 | accounting_accounts, accounting_vouchers, voucher_entries, accounting_periods, invoices, expense_claims | payroll (工资发放凭证), social (社保缴费凭证), tax (税费凭证), employee (报销人信息) |
| **notification** | 站内消息、短信、微信推送 | notifications, notification_templates | 所有模块（接收通知事件） |
| **common** | 中间件、加密、统一响应、工具函数 | audit_logs | 所有模块（横切关注点） |

### Module Communication Rules

**严格分层通信，禁止跨层访问：**

```
Module A Handler  --> Module A Service  --> Module A Repository
                                          |
                              Module B Service (通过接口调用)
                                          |
                                    Module B Repository
```

**规则：**

1. **Handler 只调用本模块 Service** -- Handler 不直接调用其他模块的 Service 或 Repository
2. **Service 可调用其他模块 Service（通过接口）** -- 跨模块调用走 Service 层接口，不直接访问 Repository
3. **Repository 只操作本模块表** -- 禁止跨模块 JOIN（除特定读场景外）
4. **事件通知走 notification 模块** -- 模块间异步通知通过 notification.Service 统一调度

**跨模块调用示例：**

```go
// payroll/service.go -- 工资核算时需要社保数据
type PayrollService struct {
    socialService   SocialServiceInterface  // 接口，不是具体实现
    taxService      TaxServiceInterface
    employeeService EmployeeServiceInterface
}

func (s *PayrollService) CalculateSalary(ctx context.Context, tableID int64) error {
    // 通过接口调用 social 模块
    socialRecords, err := s.socialService.GetByEmployeeIDs(ctx, employeeIDs)
    // ...
}
```

**模块间事件通知模式：**

```go
// 定义事件接口（common 层）
type Event struct {
    Type    string
    OrgID   int64
    Payload json.RawMessage
}

type EventBus interface {
    Publish(ctx context.Context, event Event) error
    Subscribe(eventType string, handler EventHandler)
}

// 发布端：employee 模块
func (s *EmployeeService) Resign(ctx context.Context, id int64) error {
    // ... 业务逻辑
    s.eventBus.Publish(ctx, Event{
        Type:    "employee.resigned",
        OrgID:   orgID,
        Payload: payload,
    })
}

// 订阅端：social 模块
eventBus.Subscribe("employee.resigned", func(ctx context.Context, e Event) {
    // 触发社保停缴提醒
})
```

V1.0 EventBus 使用**进程内同步调用**（直接函数调用），不引入消息队列。好处是简单可靠，坏处是耦合略高。当 V2.0 需要拆分微服务时，替换为 Redis Stream 或 RabbitMQ 实现。

---

## Data Flow

### 1. 核心入职流程

```
老板 APP             Handler              Service              Repository          外部服务
   |                    |                    |                    |                    |
   |-- POST /employees->|                    |                    |                    |
   |                    |-- CreateEmployee()->|                    |                    |
   |                    |                    |-- 生成邀请码         |                    |
   |                    |                    |-- Insert(employee)-->|                    |
   |                    |                    |                    |<-- id ----         |
   |                    |                    |-- 创建合同草稿       |                    |
   |                    |                    |-- Insert(contract)-->|                    |
   |                    |                    |-- 发送通知---------->|--------------------|-->短信/推送
   |                    |                    |-- 发布事件(resigned) |                    |
   |<-- 200 OK ---------|                    |                    |                    |
```

### 2. 工资核算流程（跨模块数据聚合）

```
payroll.Service.CalculateSalary(orgID, period)
     |
     |-- 1. employee.Service.ListActive(orgID)         --> 获取在职员工列表
     |-- 2. social.Service.GetDeductions(orgID, emps)  --> 获取社保扣款数据
     |-- 3. tax.Service.Calculate(orgID, emps, period) --> 计算个税
     |-- 4. 汇总: 基本工资 + 绩效 + 奖金 - 社保 - 个税 - 其他扣款 = 实发
     |-- 5. payroll.Repository.BatchInsert(records)    --> 批量写入工资明细
     |-- 6. notification.Service.Notify payslip        --> 推送工资单
```

### 3. 费用报销审批到自动凭证（跨模块联动）

```
employee(MiniProg)      expense_claim         finance
      |                     |                    |
      |-- 提交报销单-------->|                    |
      |                     |-- 状态=PENDING     |
      |                     |                    |
boss(APP)                   |                    |
      |-- 审批通过---------->|                    |
      |                     |-- 更新状态=APPROVED|
      |                     |-- 触发联动事件----->|
      |                     |                    |-- 生成费用凭证
      |                     |                    |-- 凭证：借 管理费用
      |                     |                    |-- 凭证：贷 银行存款/现金
      |                     |                    |-- 更新 expense_claim.voucher_id
```

### 4. 多租户数据隔离流

```
HTTP Request
     |
     v
Auth Middleware
     |-- JWT 解析 --> user_id, org_id, role
     |-- 注入到 Context: ctx = context.WithValue(ctx, "org_id", orgID)
     |
     v
Handler
     |-- 从 Context 获取 orgID
     |-- 传递给 Service
     |
     v
Service
     |-- 业务逻辑（orgID 作为参数传递）
     |
     v
Repository
     |-- WHERE org_id = ? （所有查询自动追加租户条件）
     |-- 使用 Scope/Hook 自动注入 org_id
```

**关键设计：org_id 透传链路。** 从中间件解析 JWT 获取 org_id，通过 Context 传递到 Handler -> Service -> Repository 每一层。Repository 层使用 Scope（GORM）或 Hook（Ent）自动追加 `WHERE org_id = ?` 条件，防止越权访问。

---

## Patterns to Follow

### Pattern 1: Module Interface Contract（模块接口契约）

**What:** 每个模块对外暴露 Service 接口（interface），内部实现（struct）不导出。

**When:** 所有模块间调用。

**Example:**

```go
// social/service.go -- 定义接口
type SocialService interface {
    GetByEmployeeIDs(ctx context.Context, employeeIDs []int64) ([]SocialInsurance, error)
    GetDeductions(ctx context.Context, orgID int64, period string) ([]Deduction, error)
    SuspendByEmployee(ctx context.Context, employeeID int64) error
}

// social/service_impl.go -- 内部实现
type socialService struct {
    repo repository.SocialRepository
}

func NewSocialService(repo repository.SocialRepository) SocialService {
    return &socialService{repo: repo}
}
```

**好处：** 可测试（mock 接口）、可替换（换实现不影响调用方）、为未来拆微服务预留。

### Pattern 2: Transaction Script（事务脚本）

**What:** 复杂业务操作封装在 Service 层的单个方法中，使用数据库事务保障一致性。

**When:** 跨表写操作（如入职 = 创建员工 + 创建合同 + 创建社保记录）。

**Example:**

```go
func (s *employeeService) Onboard(ctx context.Context, req OnboardRequest) (*Employee, error) {
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return nil, err
    }
    defer tx.Rollback()

    // 1. 创建员工
    emp, err := s.repo.CreateWithTx(tx, employee)
    // 2. 创建合同
    contract, err := s.contractRepo.CreateWithTx(tx, contract)
    // 3. 初始化社保记录（如需）
    if req.NeedSocial {
        _, err = s.socialRepo.CreateWithTx(tx, socialRecord)
    }

    if err := tx.Commit(); err != nil {
        return nil, err
    }
    return emp, nil
}
```

### Pattern 3: Encrypted Field（加密字段双列模式）

**What:** 敏感字段存储两列：加密值（用于展示）+ 哈希值（用于查找）。

**When:** 手机号、身份证号、银行账号等 PII 数据。

**Example:**

```go
// common/crypto/field.go
type EncryptedField struct {
    Encrypted string // AES-256-GCM 加密值，用于展示时解密
    Hash      string // SHA-256 哈希值，用于 WHERE 精确查找
}

func EncryptField(plaintext string, key []byte) (*EncryptedField, error) {
    encrypted, err := aes256gcm.Encrypt(plaintext, key)
    hash := sha256.Sum256([]byte(plaintext))
    return &EncryptedField{
        Encrypted: encrypted,
        Hash:      hex.EncodeToString(hash[:]),
    }, err
}

// Repository 层查找
func (r *employeeRepo) FindByPhone(ctx context.Context, phone string) (*Employee, error) {
    hash := crypto.HashSHA256(phone)
    return r.db.Where("phone_hash = ? AND org_id = ?", hash, getOrgID(ctx)).First(&Employee{}).Error
}
```

### Pattern 4: Soft Delete with Partial Unique Index

**What:** 软删除通过 `deleted_at` 字段实现，唯一约束使用 `WHERE deleted_at IS NULL` 条件。

**When:** 所有业务实体（员工、合同、工资表等）。

**Example:**

```sql
-- 唯一约束排除已软删除记录
CREATE UNIQUE INDEX idx_employees_org_invite
    ON employees(org_id, invite_code)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX idx_org_credit_code
    ON organizations(credit_code)
    WHERE deleted_at IS NULL;
```

```go
// Repository 层统一 Scope
func NotDeleted(db *gorm.DB) *gorm.DB {
    return db.Where("deleted_at IS NULL")
}

func (r *employeeRepo) List(ctx context.Context, orgID int64) ([]Employee, error) {
    var employees []Employee
    err := r.db.Scopes(NotDeleted).
        Where("org_id = ?", orgID).
        Find(&employees).Error
    return employees, err
}
```

### Pattern 5: Tenant-Scoped Query（租户隔离 Scope）

**What:** Repository 层所有查询自动追加 `org_id` 条件。

**When:** 所有业务模块的数据访问。

**Example:**

```go
func TenantScope(orgID int64) func(db *gorm.DB) *gorm.DB {
    return func(db *gorm.DB) *gorm.DB {
        return db.Where("org_id = ?", orgID)
    }
}

// 使用
db.Scopes(TenantScope(orgID), NotDeleted).Find(&employees)
```

### Pattern 6: Audit Log Middleware（审计日志中间件）

**What:** 通过 Gin 中间件自动记录所有写操作的审计日志。

**When:** 所有 POST/PUT/DELETE 请求。

```go
func AuditLog(logger *AuditLogger) gin.HandlerFunc {
    return func(c *gin.Context) {
        if c.Request.Method == "GET" {
            c.Next()
            return
        }
        start := time.Now()
        c.Next()
        logger.Record(AuditEntry{
            UserID:    GetUserID(c),
            OrgID:     GetOrgID(c),
            Module:    getModuleFromPath(c.Request.URL.Path),
            Action:    c.Request.Method,
            TargetID:  c.Param("id"),
            Detail:    fmt.Sprintf("status=%d duration=%v", c.Writer.Status(), time.Since(start)),
            IPAddress: c.ClientIP(),
        })
    }
}
```

### Pattern 7: Domain Event for Cross-Module Triggers

**What:** 模块间通过事件解耦，避免循环依赖。

**When:** 一个模块的业务操作需要触发其他模块的副作用。

**V1.0 事件清单：**

| 事件 | 发布模块 | 订阅模块 | 触发动作 |
|------|---------|---------|---------|
| `employee.onboarded` | employee | social | 提醒办理社保参保 |
| `employee.resigned` | employee | social | 触发社保停缴提醒 |
| `employee.resigned` | employee | notification | 通知老板完成交接 |
| `payroll.confirmed` | payroll | finance | 生成工资发放凭证 |
| `payroll.confirmed` | payroll | notification | 推送工资单给员工 |
| `social.payment_confirmed` | social | finance | 生成社保缴费凭证 |
| `social.due_soon` | social | notification | 社保到期提醒 |
| `tax.calculated` | tax | payroll | 返回个税扣款数据 |
| `tax.due_soon` | tax | notification | 个税申报提醒 |
| `expense.approved` | finance | finance | 自动生成费用凭证 |
| `period.closed` | finance | finance | 生成财务报表快照 |

---

## Anti-Patterns to Avoid

### Anti-Pattern 1: Shared Repository Access（跨模块直接访问 Repository）

**What:** 模块 A 直接 import 模块 B 的 Repository。
**Why bad:** 破坏模块边界，耦合数据模型，拆分时无法独立。
**Instead:** 通过模块 B 的 Service 接口调用。如果需要跨模块读数据且性能敏感，可在 Service 层提供专门的查询方法。

### Anti-Pattern 2: God Service（巨型 Service）

**What:** 一个 Service 文件超过 800 行，包含所有业务逻辑。
**Why bad:** 难以维护和测试，违反单一职责。
**Instead:** 按子领域拆分 Service 文件。例如 payroll 模块拆分为 `salary_service.go`（工资表管理）、`payslip_service.go`（工资单生成/推送）、`payment_service.go`（发放记录管理）。

### Anti-Pattern 3: Business Logic in Handler

**What:** Handler 层包含业务判断逻辑（if/else、状态转换）。
**Why bad:** 无法复用，无法单元测试。
**Instead:** Handler 只做参数解析/校验 + 调用 Service + 返回响应。

### Anti-Pattern 4: Distributed Transaction Prematurely

**What:** 在 V1.0 引入分布式事务（Saga、TCC 等）。
**Why bad:** 单库事务足够，分布式事务增加 10x 复杂度。
**Instead:** V1.0 使用 PostgreSQL 本地事务。跨模块写操作在同一事务中完成。

### Anti-Pattern 5: org_id in JWT Only, Not Verified at DB Level

**What:** 只依赖中间件从 JWT 解析 org_id，不在数据库查询层强制校验。
**Why bad:** 任何一个遗漏 org_id 过滤的查询就是数据泄露。
**Instead:** Repository 层使用 TenantScope 自动注入，同时考虑 PostgreSQL RLS（Row Level Security）作为防御层。

### Anti-Pattern 6: Over-Abstracted Event System

**What:** V1.0 就引入 Kafka/RabbitMQ 做事件驱动。
**Why bad:** 运维成本激增，对 <1000 用户无意义。
**Instead:** V1.0 使用进程内同步事件分发（函数调用），接口抽象好，V2.0 按需替换。

---

## Scalability Considerations

| Concern | At 100 users (V1.0) | At 10K users (V2.0) | At 100K+ users (V3.0) |
|---------|---------------------|---------------------|----------------------|
| **部署** | 单 ECS 实例 + Docker | 多实例 + 负载均衡 | K8s + HPA 自动伸缩 |
| **数据库** | 单 RDS PostgreSQL | 读写分离 + 连接池优化 | 分库（按模块拆）+ 读写分离 |
| **缓存** | Redis 单实例 | Redis Sentinel | Redis Cluster |
| **文件存储** | 阿里云 OSS | OSS + CDN | OSS + CDN + 图片处理 |
| **多租户** | 逻辑隔离（org_id） | 逻辑隔离 + RLS | 评估 Schema 隔离 |
| **模块通信** | 进程内函数调用 | Redis Stream | 消息队列（RabbitMQ/Kafka） |
| **搜索** | SQL LIKE | PostgreSQL Full-Text | Elasticsearch |
| **监控** | 日志 + Sentry | Prometheus + Grafana | 全链路 APM |

---

## Build Order (Recommended Implementation Sequence)

基于模块依赖关系和业务优先级，建议以下构建顺序：

### Phase 0: Foundation (Week 1-2)

```
搭建基础框架，所有后续模块的基石。
```

| 组件 | 内容 | 原因 |
|------|------|------|
| 项目脚手架 | cmd/server/main.go, internal/common/, pkg/ | 所有模块的容器 |
| common/middleware | Auth(JWT), RateLimit, CORS, AuditLog, TenantScope | 横切关注点，每个 API 都需要 |
| common/response | 统一响应格式 (success/error/meta) | API 契约 |
| common/crypto | AES-256-GCM 加解密, SHA-256 哈希, 脱敏 | 敏感数据保护 |
| pkg/jwt | JWT 生成/验证/刷新 | 认证基础 |
| pkg/sms | 阿里云短信客户端 | 登录验证码 |
| pkg/oss | 阿里云 OSS 客户端 | 文件上传/下载 |
| PostgreSQL | 初始迁移、连接池、GORM/Ent 配置 | 数据层 |
| Redis | 连接配置、Session 存储 | 缓存层 |
| CI/CD | GitHub Actions 构建流水线 | 自动化 |

**交付物：** 可运行的空项目框架 + 健康检查 API。

### Phase 1: User & Organization Module (Week 2-3)

```
user 模块是所有模块的依赖基础，必须最先完成。
```

| 组件 | 内容 | 依赖 |
|------|------|------|
| 用户注册/登录 | 手机号验证码登录、JWT 签发 | pkg/sms, pkg/jwt, Redis |
| 企业信息管理 | 企业创建、信息编辑、行业/城市选择 | PostgreSQL |
| RBAC 权限 | OWNER/ADMIN/MEMBER 三级权限中间件 | common/middleware |
| Token 管理 | Access/Refresh Token 轮换、Redis 黑名单 | Redis |
| 审计日志 | 所有写操作自动记录 | common/middleware |

**交付物：** 用户可通过手机号登录、创建/管理企业、分配角色。

**依赖关系：** 无外部模块依赖（仅依赖 common/pkg）。

### Phase 2: Employee Module (Week 3-5)

```
员工管理是核心实体，social/payroll/tax 都依赖员工数据。
```

| 组件 | 内容 | 依赖 |
|------|------|------|
| 员工入职 | 创建员工档案、生成邀请码 | user (org上下文) |
| 员工信息编辑 | 个人信息、岗位、薪资结构 | common/crypto (加密) |
| 合同管理 | 合同创建、PDF 模板生成、签署状态 | pkg/oss (文件存储) |
| 员工离职 | 离职审批、交接清单生成 | notification (提醒) |
| 员工搜索 | 按姓名/岗位搜索、导出 Excel | PostgreSQL (索引优化) |

**交付物：** 老板可完成员工入职到离职的全生命周期管理。

**依赖关系：** 依赖 user 模块（组织信息、权限校验）。

### Phase 3: Social Insurance Module (Week 5-7)

```
社保管理依赖员工数据，也是工资核算的上游（提供社保扣款数据）。
```

| 组件 | 内容 | 依赖 |
|------|------|------|
| 社保政策库 | 30+ 城市基数/比例管理、政策匹配 | PostgreSQL |
| 参保/停缴 | 员工社保登记、变更、停缴 | employee (员工信息) |
| 社保核算 | 自动计算企业和个人缴费金额 | employee (岗位/薪资) |
| 缴费提醒 | 到期前自动提醒 | notification |
| 参保材料 | 生成 PDF 参保/停缴材料 | pkg/oss |

**交付物：** 老板可为员工办理社保参保/停缴，自动计算缴费金额。

**依赖关系：** 依赖 employee（员工信息、离职事件）。

### Phase 4: Payroll Module (Week 7-9)

```
工资核算需要聚合员工、社保、个税三个模块的数据。
```

| 组件 | 内容 | 依赖 |
|------|------|------|
| 薪资结构 | 自定义薪资项目（基本/绩效/奖金/补贴/扣款） | employee |
| 工资表 | 创建月度工资表、复制上月、一键核算 | employee, social, tax |
| 工资核算引擎 | 应发 = 基本+绩效+奖金+补贴-扣款-社保-个税 | social (社保扣款), tax (个税) |
| 工资单推送 | 生成电子工资单、推送至员工 | notification, pkg/oss |
| 工资发放 | 记录发放状态/金额/方式 | finance (凭证联动) |
| Excel 导入导出 | 考勤表导入、工资表导出 | - |

**交付物：** 老板可一键核算工资，生成工资单推送给员工。

**依赖关系：** 依赖 employee + social + tax，是最复杂的跨模块聚合点。

### Phase 5: Tax Module (Week 6-8, 可与 Payroll 并行开发)

```
个税模块与工资模块紧密耦合，但核心计算逻辑独立。
```

| 组件 | 内容 | 依赖 |
|------|------|------|
| 个税计算引擎 | 累计预扣法计算、税率表匹配 | employee |
| 专项附加扣除 | 六项扣除登记/管理 | employee |
| 申报提醒 | 月度申报截止提醒 | notification |
| 申报表生成 | 生成个税申报表 PDF | pkg/oss |
| 税务计算 | 增值税/企业所得税计算 | finance (发票数据) |

**交付物：** 基于工资数据自动计算个税，生成申报辅助材料。

**依赖关系：** 依赖 employee。与 payroll 双向依赖（payroll 需要个税扣款数据，tax 需要应发工资数据）。

**注意：** payroll 和 tax 的循环依赖通过接口注入解决 -- payroll 引用 tax.Service 接口，tax 引用 payroll.Service 接口。

### Phase 6: Finance Module (Week 9-12)

```
财务模块是最复杂的模块，依赖前面所有模块的数据。
```

| 组件 | 内容 | 依赖 |
|------|------|------|
| 会计科目 | 预置科目体系 + 自定义增删 | - |
| 凭证管理 | 手动录入、借贷平衡校验、审核 | - |
| 费用报销 | 员工提交 -> 老板审批 -> 自动凭证 | employee, notification |
| 发票管理 | 进项/销项登记、月末汇总 | - |
| 账簿查询 | 总账/明细账/余额表实时生成 | - |
| 财务报表 | 资产负债表/利润表 | - |
| 会计期间 | 月度结账/反结账、期间锁定 | - |
| 税务管理 | 增值税/企业所得税申报辅助 | tax |

**交付物：** 完整的小微企业财务记账系统，从凭证录入到报表生成。

**依赖关系：** 依赖 employee（报销人）、payroll（工资凭证）、social（社保凭证）、tax（税费凭证）。

### Phase 7: Notification Module (贯穿开发)

```
通知模块是横切模块，随各业务模块逐步接入。
```

| 阶段 | 接入的通知类型 |
|------|--------------|
| Phase 1 | 验证码短信 |
| Phase 2 | 入职邀请通知、合同签署提醒 |
| Phase 3 | 社保到期提醒、参保变更通知 |
| Phase 4 | 工资单推送、发放状态通知 |
| Phase 5 | 个税申报提醒 |
| Phase 6 | 报销审批通知、结账提醒 |

### Build Order Visualization

```
Week:  1   2   3   4   5   6   7   8   9   10  11  12
       +---+---+---+---+---+---+---+---+---+---+---+---+
P0:    |█████████|                                   Foundation
P1:        |███████|                                 User/Org
P2:            |█████████|                           Employee
P3:                    |█████████|                   Social
P4:                            |█████████|           Payroll
P5:                        |█████████|               Tax (parallel)
P6:                                    |█████████████| Finance
P7:    |==================================================| Notification
       +---+---+---+---+---+---+---+---+---+---+---+---+
```

---

## Critical Architecture Decisions

### Decision 1: GORM vs Ent

**推荐：GORM v2**

理由：
- 生态成熟，中文文档丰富（对国内团队友好）
- Scope 机制天然适合多租户过滤
- 软删除内建支持
- Hook 机制便于审计日志
- Ent 学习曲线较陡，团队可能不熟悉

**但需注意：** GORM 的性能陷阱（Preload N+1 问题），在工资核算等批量场景需手写 SQL。

### Decision 2: Web Framework -- Gin vs Fiber

**推荐：Gin**

理由：
- 社区最大、中间件生态最丰富
- 性能足够（V1.0 不需要 Fiber 的 fasthttp 优势）
- 项目中已有 Gin 的代码示例

### Decision 3: EventBus V1.0 -- Process-Internal vs Redis Stream

**推荐：进程内同步调用（V1.0）**

理由：
- <1000 用户不需要消息队列的异步/解耦能力
- 避免引入额外基础设施依赖
- 通过接口抽象，V2.0 可无缝替换为 Redis Stream

### Decision 4: Single DB vs Multi-Schema

**推荐：单库共享（V1.0）**

理由：
- 跨模块事务简单（工资核算需要同时写工资表、社保记录、个税记录）
- 单库运维成本最低
- 通过 org_id 逻辑隔离足够安全

---

## Sources

- tech-architecture.md -- 项目完整技术架构文档（1374 行，HIGH confidence）
- .planning/PROJECT.md -- 项目需求定义（HIGH confidence）
- Go 社区模块化单体最佳实践（threedots.tech, gobeyond.dev）-- 基于训练数据（MEDIUM confidence，外部搜索受限未能验证）
- GORM v2 官方文档 -- 基于训练数据（MEDIUM confidence）
- Gin 框架官方文档 -- 基于训练数据（MEDIUM confidence）
