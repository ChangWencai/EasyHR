# 「易人事」APP 技术架构与详细设计文档

## 一、技术架构设计

### 1.1 总体架构

```
┌──────────────────────────────────────────────────────────────┐
│                         客户端层                             │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────┐ ┌────────┐ │
│  │ Android 原生  │ │  iOS 原生    │ │H5 老板端 │ │微信小程序│ │
│  │  (老板端)     │ │  (老板端)    │ │(管理后台)│ │(员工端) │ │
│  └──────┬───────┘ └──────┬───────┘ └────┬─────┘ └───┬────┘ │
└─────────┼────────────────┼──────────────┼───────────┼──────┘
          │                │              │           │
          └────────────────┼──────────────┼───────────┘
                           │ HTTPS/REST   │           │
                           │    API       │           │
┌──────────────────────────┼──────────────┼───────────┼───────┐
│                      API 网关层                              │
│  ┌─────────────┐  ┌─────────────┐  ┌───────────────────┐   │
│  │  限流/鉴权   │  │  日志/监控   │  │  CORS/SSL         │   │
│  └─────────────┘  └─────────────┘  └───────────────────┘   │
└──────────────────────────┼──────────────────────────────────┘
                           │
┌──────────────────────────┼──────────────────────────────────┐
│                      业务服务层                              │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────────┐    │
│  │ 用户/企业服务 │ │ 员工管理服务  │ │  工资/个税服务    │    │
│  └──────────────┘ └──────────────┘ └──────────────────┘    │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────────┐    │
│  │ 社保管理服务  │ │ 通知/消息服务 │ │ 财务代理记账服务  │    │
│  └──────────────┘ └──────────────┘ └──────────────────┘    │
└──────────────────────────┼──────────────────────────────────┘
                           │
┌──────────────────────────┼──────────────────────────────────┐
│                      数据层                                  │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────────┐    │
│  │  PostgreSQL   │ │    Redis     │ │   阿里云 OSS     │    │
│  │  (核心数据库)  │ │ (缓存/会话)   │ │  (文件存储)      │    │
│  └──────────────┘ └──────────────┘ └──────────────────┘    │
└─────────────────────────────────────────────────────────────┘
```

### 1.2 技术栈选型

| 层级 | 技术选择 | 选择理由 |
|------|----------|----------|
| **Android 客户端（老板端）** | Kotlin + Jetpack Compose | 官方推荐、声明式 UI、与 iOS KMP 可复用逻辑 |
| **iOS 客户端（老板端）** | Swift + SwiftUI | 原生性能、生态成熟、适配 iOS 12+ |
| **H5 管理后台（老板端）** | Vue 3 + Element Plus | PC 端管理后台、复杂数据表格、批量操作 |
| **微信小程序（员工端）** | 原生微信小程序 + WeUI | 员工查看工资条/合同/社保/报销、推送触达率高、原生开发性能最优 |
| **后端框架** | Go 1.23+ + Gin v1.12.0 | 高性能、编译为单二进制、部署简单、并发模型优秀、中国Go生态成熟 |
| **数据库** | PostgreSQL 16+ | ACID 事务保障、JSONB 灵活存储、中国时区支持、行级安全控制 |
| **ORM** | GORM v1.31.1 | Go生态最成熟、自动迁移、中文文档完善、Auto Migration适合快速迭代 |
| **缓存** | Redis 7+ | 会话管理、验证码存储、热点数据缓存、分布式锁 |
| **对象存储** | 阿里云 OSS | 合同/凭证文件存储、CDN 加速、生命周期管理 |
| **消息队列** | Redis Stream（轻量起步） | V1.0 消息通知解耦，V2.0 可替换为 RabbitMQ |
| **定时任务** | go-co-op/gocron + Redis 分布式锁 | 社保/个税到期提醒、工资自动核算触发器 |
| **部署** | Docker + 阿里云 ECS/ACK | 容器化部署、便于扩缩容、CI/CD 友好 |
| **监控** | Prometheus + Grafana + Sentry | 性能监控、错误追踪、告警通知 |
| **CI/CD** | GitHub Actions | 构建、测试、部署自动化 |

### 1.3 项目目录结构

```
EasyHR/
├── cmd/                         # Go 服务入口
├── config/                      # 配置文件
├── internal/                    # 业务模块（模块化单体）
│   ├── common/                  # 公共模块（中间件、响应封装、加密工具）
│   ├── user/                    # 用户/企业服务
│   ├── employee/               # 员工管理服务
│   ├── social/                  # 社保管理服务
│   ├── payroll/                # 工资管理服务
│   ├── tax/                    # 个税服务
│   ├── finance/                # 财务代理记账服务
│   └── notification/           # 通知服务
├── pkg/                         # 可复用包（JWT、OSS、短信）
├── frontend/                    # H5 管理后台（Vue 3 + Element Plus）
├── miniprogram/                 # 微信小程序（员工端）
│   ├── pages/
│   │   ├── payslips/           # 我的工资（首页 Tab）
│   │   ├── payslips-detail/   # 工资单详情
│   │   ├── contracts/          # 我的合同（Tab）
│   │   ├── social/             # 社保记录（Tab）
│   │   ├── expense/            # 费用报销（Tab）
│   │   ├── expense-list/       # 报销记录列表
│   │   ├── login/              # 登录页
│   │   └── mine/               # 我的（Tab）
│   ├── assets/icons/           # TabBar 图标资源
│   ├── app.js                  # 应用入口
│   ├── app.json                # 全局配置（页面路由、TabBar、窗口样式）
│   ├── app.wxss                # 全局样式
│   ├── pages.json              # 页面路径配置
│   └── sitemap.json            # SEO  sitemap
├── miniprogram-design/          # 小程序 UI 设计稿（Pencil）
├── migrations/                  # 数据库迁移
├── go.mod / go.sum
├── Dockerfile / docker-compose.yml
└── CLAUDE.md
```

### 1.4 微信小程序（员工端）实现详情

已在 `miniprogram/` 目录下实现员工端小程序，采用**原生微信小程序框架 + WeUI 组件库**，无需跨端框架额外依赖。

**已实现页面（共 8 个）**：

| 页面 | 文件路径 | 功能描述 |
|------|----------|----------|
| 我的工资（首页） | `pages/payslips/payslips` | Tab 首页，展示工资单列表 |
| 工资单详情 | `pages/payslips-detail/payslips-detail` | 单条工资单明细查看 |
| 我的合同 | `pages/contracts/contracts` | 劳动合同查看 |
| 社保记录 | `pages/social/social` | 社保缴纳记录查询 |
| 费用报销 | `pages/expense/expense` | 费用报销单提交 |
| 报销记录 | `pages/expense-list/expense-list` | 报销历史记录 |
| 我的 | `pages/mine/mine` | 个人中心 |
| 登录 | `pages/login/login` | 微信授权登录 |

**TabBar 配置**（底部导航，共 5 个 Tab）：

| Tab | 页面 | 图标 |
|-----|------|------|
| 我的工资 | payslips | salary.png / salary-active.png |
| 我的合同 | contracts | contract.png / contract-active.png |
| 社保记录 | social | social.png / social-active.png |
| 费用报销 | expense | expense.png / expense-active.png |
| 我的 | mine | mine.png / mine-active.png |

**技术特点**：
- 全局导航栏主题色：`#1677FF`（蓝色）
- 全局样式版本：v2（微信小程序基础库 v2 样式）
- 网络请求：原生 `wx.request` 封装
- 无需引入额外网络库，保持包体积最小

### 1.5 数据库设计原则

1. **行级安全（RLS）**：PostgreSQL 原生支持，确保每个企业只能访问自己的数据
2. **软删除**：所有业务表使用 `deleted_at` 字段，支持数据恢复与合规审计；涉及唯一约束的表使用部分唯一索引（`WHERE deleted_at IS NULL`）避免软删除后冲突
3. **审计追踪**：`created_by`, `created_at`, `updated_by`, `updated_at` 统一审计字段
4. **数据隔离**：每个企业使用 `org_id` 作为租户标识，实现逻辑多租户
5. **敏感字段加密**：身份证号、手机号等使用 AES-256 加密存储
6. **外键约束**：核心关联表使用 FK 约束保障数据一致性（详见 §2.5 外键关系表）；ORM 层使用 Ent/GORM 的关系定义，数据库层根据性能评估决定是否启用物理 FK

---

## 二、数据模型设计

### 2.1 核心实体关系图

```
┌──────────────┐       ┌──────────────┐       1:N      ┌──────────────┐
│   User 用户   │──1:1──│ Organization │◄┌─────────────┐│   Employee    │
│  (登录/权限)  │       │   企业        │ │  ┌─────────┤◄──  员工       │
└──────────────┘       └──────────────┘ │  │         │  └──────┬───────┘
                                         │  │         │         │
┌──────────────┐       ┌──────────────┐  │  1:N     1:N    ┌────┴────┐
│   Contract   │       │  SalaryTable │  │         │    Contract│Social│
│   电子合同    │◄──────│   工资表      │───┼─────────┘    │记录  │记录  │
└──────────────┘  N:1  └──────────────┘  │                  └────┬────┘
                                         │                       │
┌──────────────┐       ┌──────────────┐  │                  ┌────┴────┐
│ TaxDeclaration│◄─────│ TaxRecord   │  │                  │Social  │Tax   │
│  个税申报     │  1:1  │  个税记录    │───┘                  │变更历史 │申报记录│
└──────────────┘       └──────────────┘                       └────────┘
```

### 2.2 表结构设计

#### 2.2.1 用户与组织表

**`users` 用户表**
| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGSERIAL PRIMARY KEY | 主键 |
| phone | VARCHAR(20) UNIQUE NOT NULL | 手机号（加密存储） |
| phone_hash | VARCHAR(64) | 手机号哈希值（用于精确查找，SHA-256） |
| password_hash | VARCHAR(255) | 登录密码哈希 |
| nickname | VARCHAR(50) | 用户昵称 |
| avatar_url | VARCHAR(500) | 头像 URL |
| org_id | BIGINT NOT NULL | 所属企业ID |
| role | VARCHAR(20) NOT NULL | 角色：OWNER/ADMIN/MEMBER |
| wechat_openid | VARCHAR(64) | 微信 OpenID（登录绑定） |
| wechat_unionid | VARCHAR(64) | 微信 UnionID（跨应用识别） |
| last_login_at | TIMESTAMP | 最后登录时间 |
| last_login_ip | VARCHAR(50) | 最后登录 IP |
| device_id | VARCHAR(64) | 设备标识（登录保护） |
| status | VARCHAR(20) DEFAULT 'ACTIVE' | 状态：ACTIVE/INACTIVE |
| created_at | TIMESTAMP DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMP DEFAULT NOW() | 更新时间 |
| deleted_at | TIMESTAMP | 软删除时间 |

**`organizations` 企业表**
| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGSERIAL PRIMARY KEY | 主键 |
| name | VARCHAR(200) NOT NULL | 企业名称 |
| credit_code | VARCHAR(18) UNIQUE NOT NULL | 统一社会信用代码 |
| province | VARCHAR(50) | 省份 |
| city | VARCHAR(50) NOT NULL | 所在城市 |
| district | VARCHAR(50) | 区县 |
| industry | VARCHAR(50) | 行业分类（影响工伤保险费率） |
| employee_count | INT | 企业规模 |
| contact_name | VARCHAR(50) | 联系人 |
| contact_phone | VARCHAR(20) | 联系电话（加密存储） |
| logo_url | VARCHAR(500) | 企业 LOGO |
| social_account_no | VARCHAR(50) | 社保单位编号 |
| tax_account_no | VARCHAR(50) | 税务登记号 |
| subscription_plan | VARCHAR(20) DEFAULT 'FREE' | 订阅计划：FREE/PRO/ENTERPRISE（V3.0 付费预留） |
| status | VARCHAR(20) DEFAULT 'ACTIVE' | 状态：ACTIVE/FREEZE/CLOSED |
| created_at | TIMESTAMP DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMP DEFAULT NOW() | 更新时间 |
| deleted_at | TIMESTAMP | 软删除时间 |

#### 2.2.2 员工管理表

**`employees` 员工表**
| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGSERIAL PRIMARY KEY | 主键 |
| org_id | BIGINT NOT NULL | 所属企业ID |
| name | VARCHAR(50) NOT NULL | 姓名（加密存储） |
| phone | VARCHAR(20) | 手机号（加密存储） |
| phone_hash | VARCHAR(64) | 手机号哈希值（精确查找） |
| id_card | VARCHAR(18) | 身份证号（加密存储） |
| id_card_hash | VARCHAR(64) | 身份证号哈希值（精确查找） |
| gender | VARCHAR(5) | 性别（影响退休年龄、生育保险） |
| birth_date | DATE | 出生日期 |
| education | VARCHAR(20) | 学历（PRD 入职登记要求） |
| position | VARCHAR(100) | 岗位 |
| department | VARCHAR(100) | 部门 |
| email | VARCHAR(100) | 邮箱（接收工资条、合同） |
| address | VARCHAR(200) | 居住地址 |
| emergency_contact | VARCHAR(50) | 紧急联系人 |
| emergency_phone | VARCHAR(20) | 紧急联系人电话 |
| avatar_url | VARCHAR(500) | 头像 |
| entry_date | DATE | 入职日期 |
| leave_date | DATE | 离职日期（可为空） |
| probation_end_date | DATE | 试用期截止日 |
| status | VARCHAR(20) DEFAULT 'ACTIVE' | 状态：ACTIVE/ONBOARDING/RESIGNED |
| invite_code | VARCHAR(64) UNIQUE | 入职邀请码 |
| tax_id | VARCHAR(30) | 纳税人识别号 |
| created_by | BIGINT NOT NULL | 创建人ID |
| created_at | TIMESTAMP DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMP DEFAULT NOW() | 更新时间 |
| deleted_at | TIMESTAMP | 软删除时间 |

**`contracts` 合同表**
| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGSERIAL PRIMARY KEY | 主键 |
| employee_id | BIGINT NOT NULL | 员工ID |
| org_id | BIGINT NOT NULL | 企业ID |
| contract_type | VARCHAR(20) | 类型：FULL_TIME/PART_TIME/INTERNSHIP |
| start_date | DATE NOT NULL | 合同开始日期 |
| end_date | DATE | 合同结束日期（无固定为 NULL） |
| probation_months | INT | 试用期月数 |
| salary | DECIMAL(10,2) | 合同约定薪资 |
| work_location | VARCHAR(100) | 工作地点 |
| template_id | VARCHAR(50) | 电子签模板ID |
| sign_url | VARCHAR(500) | 签署链接 |
| status | VARCHAR(20) DEFAULT 'DRAFT' | 状态：DRAFT/SIGNED/EXPIRED/TERMINATED |
| file_url | VARCHAR(500) | 电子合同文件 OSS URL |
| employee_sign_time | TIMESTAMP | 员工签署时间 |
| org_sign_time | TIMESTAMP | 企业签署时间 |
| created_at | TIMESTAMP DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMP DEFAULT NOW() | 更新时间 |
| deleted_at | TIMESTAMP | 软删除时间 |

#### 2.2.3 社保管理表

**`social_insurance` 社保记录表**
| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGSERIAL PRIMARY KEY | 主键 |
| employee_id | BIGINT NOT NULL | 员工ID |
| org_id | BIGINT NOT NULL | 企业ID |
| social_no | VARCHAR(30) | 社保编号 |
| insurance_city | VARCHAR(50) NOT NULL | 参保城市 |
| base_salary | DECIMAL(10,2) | 社保缴费基数 |
| pension_company_rate | DECIMAL(5,2) | 养老保险企业比例(%) |
| pension_personal_rate | DECIMAL(5,2) | 养老保险个人比例(%) |
| medical_company_rate | DECIMAL(5,2) | 医疗保险企业比例(%) |
| medical_personal_rate | DECIMAL(5,2) | 医疗保险个人比例(%) |
| unemployment_company_rate | DECIMAL(5,2) | 失业保险企业比例(%) |
| unemployment_personal_rate | DECIMAL(5,2) | 失业保险个人比例(%) |
| injury_rate | DECIMAL(5,2) | 工伤保险比例(%) |
| maternity_rate | DECIMAL(5,2) | 生育保险比例(%) |
| housing_fund_base | DECIMAL(10,2) | 公积金基数 |
| housing_fund_rate | DECIMAL(5,2) | 公积金比例(%) |
| housing_fund_amount | DECIMAL(10,2) | 公积金月缴额 |
| personal_total | DECIMAL(10,2) | 个人缴纳合计 |
| company_total | DECIMAL(10,2) | 企业缴纳合计 |
| total_amount | DECIMAL(10,2) | 月缴费总额 |
| status | VARCHAR(20) DEFAULT 'ACTIVE' | 状态：ACTIVE/SUSPENDED |
| start_date | DATE NOT NULL | 参保开始日期 |
| end_date | DATE | 参保结束日期 |
| created_at | TIMESTAMP DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMP DEFAULT NOW() | 更新时间 |

**`social_history` 社保变更历史表**
| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGSERIAL PRIMARY KEY | 主键 |
| social_id | BIGINT NOT NULL | 社保记录ID |
| change_type | VARCHAR(20) NOT NULL | 类型：ENROLL/CHANGE/SUSPEND |
| old_value | JSONB | 变更前（存储 JSON 快照） |
| new_value | JSONB | 变更后（存储 JSON 快照） |
| reason | VARCHAR(200) | 变更原因 |
| created_at | TIMESTAMP DEFAULT NOW() | 变更时间 |

#### 2.2.4 工资管理表

**`salary_tables` 工资表**
| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGSERIAL PRIMARY KEY | 主键 |
| org_id | BIGINT NOT NULL | 企业ID |
| period | VARCHAR(7) NOT NULL | 工资月份（如 2026-04） |
| status | VARCHAR(20) DEFAULT 'DRAFT' | 状态：DRAFT/CONFIRMED/SENT/PAID |
| total_count | INT DEFAULT 0 | 员工数 |
| total_amount | DECIMAL(12,2) DEFAULT 0 | 工资总额 |
| created_by | BIGINT NOT NULL | 创建人ID |
| created_at | TIMESTAMP DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMP DEFAULT NOW() | 更新时间 |

**`salary_records` 工资明细表**
| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGSERIAL PRIMARY KEY | 主键 |
| table_id | BIGINT NOT NULL | 工资表ID |
| employee_id | BIGINT NOT NULL | 员工ID |
| org_id | BIGINT NOT NULL | 企业ID（租户标识，避免 JOIN 查询） |
| base_salary | DECIMAL(10,2) DEFAULT 0 | 基本工资 |
| performance | DECIMAL(10,2) DEFAULT 0 | 绩效工资 |
| bonus | DECIMAL(10,2) DEFAULT 0 | 奖金 |
| overtime_pay | DECIMAL(10,2) DEFAULT 0 | 加班费 |
| allowance | DECIMAL(10,2) DEFAULT 0 | 补贴 |
| deduction | DECIMAL(10,2) DEFAULT 0 | 扣款 |
| leave_deduction | DECIMAL(10,2) DEFAULT 0 | 请假扣款 |
| social_deduction | DECIMAL(10,2) DEFAULT 0 | 社保个人扣款 |
| housing_fund_deduction | DECIMAL(10,2) DEFAULT 0 | 公积金个人扣款 |
| tax_deduction | DECIMAL(10,2) DEFAULT 0 | 个税扣款 |
| gross_salary | DECIMAL(10,2) DEFAULT 0 | 应发工资 |
| net_salary | DECIMAL(10,2) DEFAULT 0 | 实发工资 |
| remark | VARCHAR(500) | 备注说明 |
| confirmed | BOOLEAN DEFAULT FALSE | 员工是否确认 |
| confirm_time | TIMESTAMP | 确认时间 |
| created_at | TIMESTAMP DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMP DEFAULT NOW() | 更新时间 |

#### 2.2.5 个税管理表

**`tax_records` 个税记录表**
| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGSERIAL PRIMARY KEY | 主键 |
| employee_id | BIGINT NOT NULL | 员工ID |
| org_id | BIGINT NOT NULL | 企业ID |
| period | VARCHAR(7) NOT NULL | 所属月份 |
| gross_salary | DECIMAL(10,2) | 应发工资 |
| special_deductions | DECIMAL(10,2) DEFAULT 0 | 专项附加扣除总额 |
| taxable_income | DECIMAL(10,2) | 应纳税所得额 |
| tax_amount | DECIMAL(10,2) | 应纳税额 |
| declared | BOOLEAN DEFAULT FALSE | 是否已申报 |
| declared_at | TIMESTAMP | 申报时间 |
| created_at | TIMESTAMP DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMP DEFAULT NOW() | 更新时间 |

**`special_deductions` 专项附加扣除表**
| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGSERIAL PRIMARY KEY | 主键 |
| employee_id | BIGINT NOT NULL | 员工ID |
| type | VARCHAR(30) NOT NULL | 类型：CHILD_EDUCATION/CONTINUING/DISEASE/LOAN/RENT/SUPPORT |
| amount | DECIMAL(10,2) | 月扣除金额 |
| start_month | VARCHAR(7) NOT NULL | 开始月份 |
| end_month | VARCHAR(7) | 结束月份 |
| status | VARCHAR(20) DEFAULT 'ACTIVE' | 状态：ACTIVE/EXPIRED |
| created_at | TIMESTAMP DEFAULT NOW() | 创建时间 |

#### 2.2.6 通知与审计表

**`notifications` 通知表**
| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGSERIAL PRIMARY KEY | 主键 |
| user_id | BIGINT NOT NULL | 接收用户ID |
| org_id | BIGINT NOT NULL | 企业ID |
| type | VARCHAR(30) | 类型：SOCIAL_DUE/TAX_DUE/CONTRACT_EXPIRE/PAYROLL |
| target_type | VARCHAR(30) | 关联对象类型（EMPLOYEE/CONTRACT/SALARY 等） |
| target_id | BIGINT | 关联对象ID |
| title | VARCHAR(200) | 通知标题 |
| content | TEXT | 通知内容 |
| channel | VARCHAR(20) DEFAULT 'APP' | 发送渠道：APP/SMS/WECHAT |
| send_status | VARCHAR(20) DEFAULT 'PENDING' | 发送状态：PENDING/SENT/FAILED |
| send_time | TIMESTAMP | 实际发送时间 |
| error_msg | VARCHAR(500) | 发送失败原因 |
| read | BOOLEAN DEFAULT FALSE | 是否已读 |
| read_at | TIMESTAMP | 已读时间 |
| created_at | TIMESTAMP DEFAULT NOW() | 创建时间 |

**`audit_logs` 审计日志表**
| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGSERIAL PRIMARY KEY | 主键 |
| org_id | BIGINT NOT NULL | 企业ID |
| user_id | BIGINT | 操作用户ID |
| module | VARCHAR(30) NOT NULL | 操作模块 |
| action | VARCHAR(50) NOT NULL | 操作类型 |
| target_id | BIGINT | 目标对象ID |
| target_type | VARCHAR(30) | 目标类型 |
| detail | JSONB | 操作详情 |
| ip_address | VARCHAR(50) | 操作IP |
| created_at | TIMESTAMP DEFAULT NOW() | 操作时间 |

### 2.3 补充表结构

#### 2.3.1 员工银行账户表 `employee_bank_accounts`

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGSERIAL PRIMARY KEY | 主键 |
| employee_id | BIGINT NOT NULL | 员工ID |
| org_id | BIGINT NOT NULL | 企业ID |
| bank_name | VARCHAR(50) | 开户行名称 |
| bank_code | VARCHAR(20) | 银行编码 |
| account_number | VARCHAR(30) | 银行账号（加密存储） |
| account_name | VARCHAR(50) | 账户名 |
| is_default | BOOLEAN DEFAULT FALSE | 是否默认账户 |
| created_at | TIMESTAMP DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMP DEFAULT NOW() | 更新时间 |

#### 2.3.2 工资发放记录表 `salary_payments`

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGSERIAL PRIMARY KEY | 主键 |
| table_id | BIGINT NOT NULL | 工资表ID |
| employee_id | BIGINT NOT NULL | 员工ID |
| org_id | BIGINT NOT NULL | 企业ID |
| amount | DECIMAL(10,2) | 实发金额 |
| payment_method | VARCHAR(20) | 发放方式：BANK_TRANSFER/CASH/OTHER |
| payment_date | DATE | 实际发放日期 |
| transaction_no | VARCHAR(64) | 银行流水号 |
| status | VARCHAR(20) DEFAULT 'PENDING' | 状态：PENDING/PROCESSING/SUCCESS/FAILED |
| failure_reason | VARCHAR(200) | 失败原因 |
| created_at | TIMESTAMP DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMP DEFAULT NOW() | 更新时间 |

#### 2.3.3 离职交接清单表 `resignation_checklists`

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGSERIAL PRIMARY KEY | 主键 |
| employee_id | BIGINT NOT NULL | 员工ID |
| org_id | BIGINT NOT NULL | 企业ID |
| resignation_date | DATE | 离职日期 |
| reason | VARCHAR(200) | 离职原因 |
| approver_id | BIGINT | 审批人ID |
| approval_status | VARCHAR(20) DEFAULT 'PENDING' | 审批状态：PENDING/APPROVED/REJECTED |
| approval_time | TIMESTAMP | 审批时间 |
| items | JSONB | 交接事项清单（动态结构，含资产、工作内容等） |
| status | VARCHAR(20) DEFAULT 'PENDING' | 交接状态：PENDING/IN_PROGRESS/COMPLETED |
| created_at | TIMESTAMP DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMP DEFAULT NOW() | 更新时间 |
| deleted_at | TIMESTAMP | 软删除时间 |

#### 2.3.4 附件管理表 `attachments`

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGSERIAL PRIMARY KEY | 主键 |
| org_id | BIGINT NOT NULL | 企业ID |
| target_type | VARCHAR(30) | 关联类型：CONTRACT/SOCIAL_VOUCHER/SALARY_SLIP/TAX_VOUCHER |
| target_id | BIGINT NOT NULL | 关联对象ID |
| file_name | VARCHAR(200) | 原始文件名 |
| file_type | VARCHAR(20) | 文件类型：PDF/EXCEL/IMAGE |
| file_size | BIGINT | 文件大小（bytes） |
| file_url | VARCHAR(500) | OSS URL |
| uploaded_by | BIGINT | 上传人ID |
| created_at | TIMESTAMP DEFAULT NOW() | 创建时间 |

#### 2.3.5 通知模板表 `notification_templates`

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGSERIAL PRIMARY KEY | 主键 |
| type | VARCHAR(30) NOT NULL | 类型：SOCIAL_DUE/TAX_DUE/CONTRACT_EXPIRE/PAYROLL |
| channel | VARCHAR(20) NOT NULL | 渠道：APP/SMS/WECHAT |
| title_template | VARCHAR(200) | 标题模板 |
| content_template | TEXT | 内容模板 |
| is_active | BOOLEAN DEFAULT TRUE | 是否启用 |
| created_at | TIMESTAMP DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMP DEFAULT NOW() | 更新时间 |

#### 2.3.6 个税扣除详情表 `special_deduction_details`

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGSERIAL PRIMARY KEY | 主键 |
| deduction_id | BIGINT NOT NULL | 关联 special_deductions.id |
| detail_type | VARCHAR(30) | 具体细项 |
| cert_no | VARCHAR(50) | 证书/证明编号 |
| cert_file_url | VARCHAR(500) | 证明文件 |
| spouse_name | VARCHAR(50) | 配偶姓名（子女教育/房贷需要） |
| spouse_id_card | VARCHAR(18) | 配偶身份证号 |
| created_at | TIMESTAMP DEFAULT NOW() | 创建时间 |

#### 2.3.7 财务代理记账模块表

##### `accounting_accounts` 会计科目表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGSERIAL PRIMARY KEY | 主键 |
| org_id | BIGINT NOT NULL | 企业ID |
| code | VARCHAR(20) NOT NULL | 科目编码（如 1001、2001、1002.01） |
| name | VARCHAR(100) NOT NULL | 科目名称 |
| parent_id | BIGINT | 父级科目ID（树形结构） |
| level | INT NOT NULL | 科目层级（1-5级） |
| category | VARCHAR(20) NOT NULL | 科目类别：ASSET/LIABILITY/EQUITY/COST/REVENUE |
| direction | VARCHAR(10) NOT NULL | 余额方向：DEBIT/CREDIT |
| is_leaf | BOOLEAN DEFAULT TRUE | 是否末级科目 |
| status | VARCHAR(20) DEFAULT 'ACTIVE' | 状态：ACTIVE/DISABLED |
| created_at | TIMESTAMP DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMP DEFAULT NOW() | 更新时间 |

> 常用科目预设：1001 库存、1002 银行存款、1002.01 库存-工行、2001 应收账款→2002 应付账款→3001 管理费用→3002 销售费用→4001 实收资本→4002 未分配利润

##### `accounting_vouchers` 会计凭证表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGSERIAL PRIMARY KEY | 主键 |
| org_id | BIGINT NOT NULL | 企业ID |
| voucher_no | VARCHAR(20) NOT NULL | 凭证编号（如 PZ20260400001） |
| voucher_date | DATE NOT NULL | 凭证日期 |
| period | VARCHAR(7) NOT NULL | 会计期间（如 2026-04） |
| description | VARCHAR(500) | 凭证摘要 |
| total_debit | DECIMAL(12,2) DEFAULT 0 | 借方合计 |
| total_credit | DECIMAL(12,2) DEFAULT 0 | 借方合计 |
| source_type | VARCHAR(20) | 来源：MANUAL/SALARY/SOCIAL/TAX/INVOICE/REIMBURSE |
| source_id | BIGINT | 关联来源对象ID |
| status | VARCHAR(20) DEFAULT 'DRAFT' | 状态：DRAFT/APPROVED/REVERSED |
| created_by | BIGINT NOT NULL | 制单人ID |
| approved_by | BIGINT | 审核人ID |
| approved_at | TIMESTAMP | 审核时间 |
| created_at | TIMESTAMP DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMP DEFAULT NOW() | 更新时间 |

##### `voucher_entries` 凭证明细表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGSERIAL PRIMARY KEY | 主键 |
| voucher_id | BIGINT NOT NULL | 凭证ID |
| account_id | BIGINT NOT NULL | 会计科目ID |
| description | VARCHAR(200) | 明细摘要 |
| debit_amount | DECIMAL(12,2) DEFAULT 0 | 借方金额 |
| credit_amount | DECIMAL(12,2) DEFAULT 0 | 贷方金额 |
| sort_order | INT NOT NULL | 排序序号 |

##### `accounting_periods` 会计期间表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGSERIAL PRIMARY KEY | 主键 |
| org_id | BIGINT NOT NULL | 企业ID |
| period | VARCHAR(7) NOT NULL | 会计期间（如 2026-04） |
| start_date | DATE NOT NULL | 期间开始日期 |
| end_date | DATE NOT NULL | 期间结束日期 |
| is_closed | BOOLEAN DEFAULT FALSE | 是否已结账 |
| closed_by | BIGINT | 结账人ID |
| closed_at | TIMESTAMP | 结账时间 |
| created_at | TIMESTAMP DEFAULT NOW() | 创建时间 |

##### `invoices` 发票管理表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGSERIAL PRIMARY KEY | 主键 |
| org_id | BIGINT NOT NULL | 企业ID |
| invoice_type | VARCHAR(10) NOT NULL | 类型：INPUT（进项）/OUTPUT（销项） |
| invoice_code | VARCHAR(50) | 发票代码 |
| invoice_no | VARCHAR(50) NOT NULL | 发票号码 |
| invoice_date | DATE NOT NULL | 开票日期 |
| seller_name | VARCHAR(200) | 销方名称 |
| seller_tax_no | VARCHAR(30) | 销方税号 |
| buyer_name | VARCHAR(200) | 购方名称 |
| buyer_tax_no | VARCHAR(30) | 购方税号 |
| amount_without_tax | DECIMAL(12,2) NOT NULL | 不含税金额 |
| tax_rate | DECIMAL(5,2) | 税率(%) |
| tax_amount | DECIMAL(12,2) | 税额 |
| total_amount | DECIMAL(12,2) NOT NULL | 价税合计 |
| file_url | VARCHAR(500) | 发票文件 OSS URL |
| voucher_id | BIGINT | 关联凭证ID |
| status | VARCHAR(20) DEFAULT 'PENDING' | 状态：PENDING/VERIFIED/REJECTED |
| verified_by | BIGINT | 审核人ID |
| verified_at | TIMESTAMP | 审核时间 |
| remark | VARCHAR(500) | 备注 |
| created_at | TIMESTAMP DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMP DEFAULT NOW() | 更新时间 |

##### `expense_claims` 费用报销单表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGSERIAL PRIMARY KEY | 主键 |
| org_id | BIGINT NOT NULL | 企业ID |
| claim_no | VARCHAR(20) NOT NULL | 报销单号 |
| employee_id | BIGINT NOT NULL | 报销人ID |
| expense_type | VARCHAR(30) NOT NULL | 费用类型：OFFICE/TRAVEL/MEAL/OTHER |
| amount | DECIMAL(10,2) NOT NULL | 报销金额 |
| description | VARCHAR(500) | 费用说明 |
| file_urls | JSONB | 附件文件URL列表 |
| status | VARCHAR(20) DEFAULT 'PENDING' | 状态：PENDING/APPROVED/REJECTED/PAID |
| approved_by | BIGINT | 审批人ID |
| approved_at | TIMESTAMP | 审批时间 |
| paid_at | TIMESTAMP | 支付时间 |
| voucher_id | BIGINT | 关联凭证ID |
| created_at | TIMESTAMP DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMP DEFAULT NOW() | 更新时间 |

### 2.4 外键关系表

> **说明**：V1.0 使用 ORM 层关系定义（Ent/GORM），数据库层暂不启用物理 FK 约束以保障写入性能。通过应用层事务和校验保障引用完整性。V2.0 评估后决定是否启用物理 FK。

| 子表 | 字段 | 父表 | 字段 | 删除策略 |
|------|------|------|------|---------|
| users | org_id | organizations | id | RESTRICT（企业不可删除） |
| employees | org_id | organizations | id | CASCADE |
| employees | created_by | users | id | SET NULL |
| contracts | employee_id | employees | id | CASCADE |
| contracts | org_id | organizations | id | CASCADE |
| social_insurance | employee_id | employees | id | CASCADE |
| social_insurance | org_id | organizations | id | CASCADE |
| social_history | social_id | social_insurance | id | CASCADE |
| salary_tables | org_id | organizations | id | CASCADE |
| salary_tables | created_by | users | id | SET NULL |
| salary_records | table_id | salary_tables | id | CASCADE |
| salary_records | employee_id | employees | id | CASCADE |
| salary_records | org_id | organizations | id | CASCADE |
| tax_records | employee_id | employees | id | CASCADE |
| tax_records | org_id | organizations | id | CASCADE |
| special_deductions | employee_id | employees | id | CASCADE |
| notifications | user_id | users | id | CASCADE |
| notifications | org_id | organizations | id | CASCADE |
| audit_logs | org_id | organizations | id | RESTRICT |
| audit_logs | user_id | users | id | SET NULL |
| attachments | org_id | organizations | id | CASCADE |
| employee_bank_accounts | employee_id | employees | id | CASCADE |
| employee_bank_accounts | org_id | organizations | id | CASCADE |
| salary_payments | table_id | salary_tables | id | CASCADE |
| salary_payments | employee_id | employees | id | CASCADE |
| resignation_checklists | employee_id | employees | id | CASCADE |
| accounting_accounts | org_id | organizations | id | CASCADE |
| accounting_accounts | parent_id | accounting_accounts | id | RESTRICT |
| accounting_vouchers | org_id | organizations | id | CASCADE |
| accounting_vouchers | created_by | users | id | SET NULL |
| voucher_entries | voucher_id | accounting_vouchers | id | CASCADE |
| voucher_entries | account_id | accounting_accounts | id | RESTRICT |
| accounting_periods | org_id | organizations | id | CASCADE |
| invoices | org_id | organizations | id | CASCADE |
| invoices | voucher_id | accounting_vouchers | id | SET NULL |
| expense_claims | org_id | organizations | id | CASCADE |
| expense_claims | employee_id | employees | id | CASCADE |
| expense_claims | approved_by | users | id | SET NULL |
| expense_claims | voucher_id | accounting_vouchers | id | SET NULL |

### 2.5 索引策略

```sql
-- 高频查询索引
CREATE INDEX idx_employees_org_status ON employees(org_id, status) WHERE deleted_at IS NULL;
CREATE INDEX idx_salary_tables_org_period ON salary_tables(org_id, period);
CREATE INDEX idx_social_insurance_org_status ON social_insurance(org_id, status);
CREATE INDEX idx_tax_records_org_period_declared ON tax_records(org_id, period, declared);
CREATE INDEX idx_audit_logs_org_module_time ON audit_logs(org_id, module, created_at);

-- 财务模块索引
CREATE INDEX idx_vouchers_org_period_date ON accounting_vouchers(org_id, period, voucher_date);
CREATE INDEX idx_voucher_entries_voucher_id ON voucher_entries(voucher_id);
CREATE INDEX idx_invoices_org_type_date ON invoices(org_id, invoice_type, invoice_date);
CREATE INDEX idx_accounting_periods_org ON accounting_periods(org_id, period);
CREATE INDEX idx_expense_claims_org_status ON expense_claims(org_id, status);

-- 唯一约束索引（软删除表使用部分唯一索引，WHERE deleted_at IS NULL）
CREATE UNIQUE INDEX idx_employees_org_invite ON employees(org_id, invite_code) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_org_credit_code ON organizations(credit_code) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_voucher_no_org ON accounting_vouchers(org_id, voucher_no);
CREATE UNIQUE INDEX idx_invoice_no_org ON invoices(org_id, invoice_no);
CREATE UNIQUE INDEX idx_accounting_period_org ON accounting_periods(org_id, period);
```

---

## 三、第三方集成方案

### 3.1 集成总览

| 集成方 | 用途 | 集成方式 | 状态 |
|--------|------|----------|------|
| 电子签服务 | 电子劳动合同签署 | V1.0 降级：PDF 模板 + 手动签署；V2.0 对接 e签宝/上上签 API | V1.0 降级方案 |
| 社保政策接口 | 各城市社保基数/比例 | 自建政策库（30+ 城市），管理员手动更新；V2.0 对接第三方数据服务商 | V1.0 自建库 |
| 短信服务 | 验证码、提醒通知 | 阿里云短信 SDK | V1.0 需要 |
| 微信开放平台 | 微信小程序推送/登录 | 微信服务端 API | V1.0 需要 |
| 个税申报系统 | 个税数据对接 | V1.0 自动计算 + 手动提交至自然人电子税务局；V2.0 对接 API | V2.0 规划 |
| 社保在线办理 | 社保参保自动对接 | V1.0 生成参保材料 PDF 供线下办理；V2.0 对接政务接口 | V2.0 规划 |
| 发票查验接口 | 发票真伪查验 | V1.0 手动登记；V2.0 对接国家税务总局查验平台 | V2.0 规划 |

### 3.2 电子签服务集成

**推荐方案**：对接「上上签」或「e签宝」（国内合规电子签主流厂商）

```
用户发起合同 → 后端调用电子签 API → 生成签署链接
     ↓
员工手机端签署 → 电子签回调通知后端 → 更新合同状态 + 保存签署文件 OSS URL
     ↓
合同归档 → 触发通知提醒老板
```

**关键配置**：
| 配置项 | 说明 |
|--------|------|
| 认证方式 | OAuth 2.0 + AppKey/AppSecret |
| 模板预置 | 标准劳动合同模板预存至电子签平台 |
| 签署流程 | 企业方授权 → 员工方填写 → 双方签署 → 自动归档 |
| 回调处理 | 签署结果回调 + 主动轮询兜底 |
| 文件存储 | 签署完成后将 PDF 下载至公司 OSS |

**V1.0 降级方案**：若电子签对接周期长，先支持 PDF 模板生成 + 手动签署上传。

**风险与应对**：

| 风险 | 应对措施 |
|------|----------|
| 合规风险：CA 认证与可信时间戳差异 | 选用前确认厂商通过工信部 CA 认证，支持可信第三方时间戳 |
| 成本风险：按量计费线性增长 | V1.0 采用降级方案避免第三方费用，预留接口抽象层（Strategy 模式）便于后续切换厂商 |
| 依赖风险：API 变更或服务中断 | 回调 + 主动轮询兜底，核心业务不依赖电子签可用性 |
| 模板维护：各地劳动合同条款差异 | 初期只维护「标准版」模板，城市定制版延后 |

### 3.3 社保政策接口

**实现策略**：聚合第三方社保政策数据 + 本地缓存

```
数据源方案：
├── 方案A：对接「51社保」等人力资源 SaaS 的开放 API
├── 方案B：爬虫定期采集各城市社保官网数据（需要持续维护）
└── 方案C（推荐）：自建社保政策库，初期人工维护 + 后续对接官方接口
```

**V1.0 推荐方案（方案 C 改良版）**：

1. **初始数据**：导入主流城市（北上广深杭等 30+ 城市）社保基数和比例
2. **更新机制**：管理员后台手动更新 + 系统推送政策变动通知
3. **本地缓存**：PostgreSQL 存储 + Redis 缓存热数据

```sql
-- 社保政策表
CREATE TABLE social_policies (
    id BIGSERIAL PRIMARY KEY,
    city VARCHAR(50) NOT NULL,
    period VARCHAR(7) NOT NULL,  -- 如 2026-04
    min_base DECIMAL(10,2),      -- 最低基数
    max_base DECIMAL(10,2),      -- 最高基数
    pension_company_rate DECIMAL(5,2),  -- 养老公司比例%
    pension_personal_rate DECIMAL(5,2), -- 养老个人比例%
    medical_company_rate DECIMAL(5,2),
    medical_personal_rate DECIMAL(5,2),
    unemployment_company_rate DECIMAL(5,2),
    unemployment_personal_rate DECIMAL(5,2),
    injury_rate DECIMAL(5,2),
    maternity_rate DECIMAL(5,2),
    fund_rate DECIMAL(5,2),
    effective_date DATE,
    source VARCHAR(30),              -- 数据来源：MANUAL/API_THIRD_PARTY/OFFICIAL
    verified_at TIMESTAMP,           -- 最后验证时间
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(city, period)
);
```

**风险与应对**：

| 风险 | 应对措施 |
|------|----------|
| 数据准确性：人工维护易出错 | 增加社保计算结果「人工确认」环节，避免自动计算错误直接生效 |
| 时效性：各城市调整时间不统一 | 政策变更时自动标记受影响员工的社保记录，推送确认提醒 |
| 覆盖范围：初期 30 城市可能不足 | 优先覆盖目标用户集中的城市，支持「其他城市」手动输入模式 |
| 维护成本：300+ 地级市政策各异 | V2.0 对接第三方社保数据服务商（如「51社保」），付费获取数据 |

### 3.4 短信服务集成

**推荐**：阿里云短信服务

| 场景 | 模板变量 | 触发条件 |
|------|----------|----------|
| 登录验证码 | `${code}`, `${expire}` | 用户登录时 |
| 社保缴费提醒 | `${orgName}`, `${deadline}` | 截止前3天 |
| 个税申报提醒 | `${orgName}`, `${deadline}` | 截止前3天 |
| 合同到期提醒 | `${employee}`, `${date}` | 到期前7天/30天 |
| 工资单通知 | `${orgName}`, `${amount}`, `${period}` | 工资单生成后 |

**集成要点**：
- 验证码 TTL：Redis 存储，5 分钟过期
- 短信防刷：同一手机号 60 秒间隔 + 每日 10 次上限 + IP 维度限流（同 IP 每小时最多 N 条）
- 防机器人：发送验证码前增加图形验证码/滑块验证
- 失败重试：3 次指数退避重试 + 降级到 APP 推送

**注意事项**：
- 阿里云短信模板审核通常需 1-2 个工作日，提前报备
- 短信到达率约 95%-98%，高峰期可能更低，需有降级方案

### 3.5 微信推送集成

V1.0 通过**微信小程序**实现消息推送：

```
后端生成通知 → 写入 notifications 表 → 调用微信模板消息 API
     ↓
用户收到微信服务通知 → 打开小程序查看详情
```

| 集成项 | 说明 |
|--------|------|
| 登录方式 | 微信授权登录（与手机号绑定） |
| 推送方式 | 微信小程序订阅消息 |
| 模板类型 | 社保到期提醒、个税申报提醒、待办审批通知 |
| V3.0 扩展 | 微信小程序作为轻量级操作入口 |

**注意事项**：
- 微信小程序订阅消息需用户主动授权，且每次授权只能发一条 → V1.0 以 APP 站内消息为主，微信为辅
- 账号体系：手机号为主键，微信为绑定（一个手机号只能绑定一个微信）
- 涉及工资/社保等敏感功能的小程序可能面临更严格的微信审核 → 提前准备资质文件和隐私政策

### 3.6 财务代理记账模块设计

#### 3.6.1 功能范围

| 子模块 | V1.0 功能 | V2.0 扩展 |
|--------|----------|----------|
| 基础记账 | 手动录入凭证、凭证审核、凭证打印 | OCR 识别发票自动生成凭证 |
| 发票管理 | 进项/销项发票登记、查验、归档 | 发票批量导入、自动验真 |
| 费用报销 | 员工提交报销单 → 老板审批 → 自动生成凭证 | 审批流（多级审批） |
| 账簿查询 | 总账、明细账、余额表实时生成 | 多维度辅助核算 |
| 财务报表 | 资产负债表、利润表自动生成 | 现金流量表、自定义报表 |
| 税务管理 | 增值税/企业所得税自动计算 | 一键申报对接 |
| 会计期间 | 月度结账/反结账 | 年度结账、跨年调整 |

#### 3.6.2 核心业务流程

**凭证录入流程**：
```
老板/兼职会计 选择会计期间
     ↓
录入凭证（选择科目 + 填写借贷金额 + 摘要）
     ↓
保存为草稿 → 提交审核
     ↓
审核通过 → 更新科目余额 → 实时生成账簿
     ↓
月末结账 → 锁定期间 → 生成财务报表
```

**费用报销流程**：
```
员工（微信小程序）提交报销单 + 上传票据照片
     ↓
老板（APP/H5）查看报销详情 → 审批（通过/驳回）
     ↓
审批通过 → 自动生成费用凭证（管理费用科目）
     ↓
标记为「已支付」→ 同步更新工资/往来
```

**发票管理流程**：
```
老板上传/拍照录入发票信息
     ↓
系统自动查验（对接税务局发票查验接口）
     ↓
查验通过 → 标记为有效 → 可关联至凭证
     ↓
月末汇总 → 自动计算增值税进项/销项 → 生成纳税申报数据
```

#### 3.6.3 科目体系预置

V1.0 预置小微企业常用科目（简化版）：

```
一、资产类
  1001 库存现金          1002 银行存款
  1012 应收账款          1013 预付账款
  1101 固定资产          1102 累计折旧

二、负债类
  2001 应付账款          2002 预收账款
  2011 应付职工薪酬      2012 应交税费
  2021 其他应付款

三、所有者权益类
  3001 实收资本          3002 资本公积
  3101 盈余公积          3102 本年利润
  3103 利润分配

四、成本类
  4001 生产成本          4002 制造费用

五、损益类
  5001 主营业务收入      5002 其他业务收入
  5011 管理费用          5012 销售费用
  5013 财务费用          5101 主营业务成本
  5102 税金及附加        5103 其他业务成本
  5111 营业外收入        5112 营业外支出
  5121 所得税费用
```

> 支持企业在标准科目基础上自定义增删科目。

#### 3.6.4 与其他模块的联动

| 触发场景 | 联动操作 |
|----------|----------|
| 工资发放确认 | 自动生成「应付职工薪酬 → 银行存款」凭证 |
| 社保缴费确认 | 自动生成「管理费用-社保 → 银行存款」凭证 |
| 个税代扣 | 自动生成「应付职工薪酬 → 应交税费-个税」凭证 |
| 员工报销审批通过 | 自动生成「管理费用-xx → 银行存款/库存现金」凭证 |
| 发票录入（进项） | 自动生成「应交税费-增值税(进项)」分录 |
| 发票录入（销项） | 自动生成「应交税费-增值税(销项)」分录 |

#### 3.6.5 财务报表生成逻辑

**资产负债表**：基于科目余额表，按资产=负债+权益公式自动生成

**利润表**：基于损益类科目发生额，按 收入-成本-费用=利润 公式自动生成

**报表存储**：每次结账时生成快照存入数据库，后续修改不影响已生成报表

#### 3.6.6 权限控制

| 角色 | 凭证录入 | 凭证审核 | 结账 | 报表查看 | 发票管理 |
|------|---------|---------|------|---------|---------|
| OWNER | ✅ | ✅ | ✅ | ✅ | ✅ |
| ADMIN（兼职会计） | ✅ | ❌ | ❌ | ✅ | ✅ |
| MEMBER | ❌ | ❌ | ❌ | ❌ | 仅查看 |

---

## 四、权限与安全设计

### 4.1 RBAC 权限模型

```
角色层级：
┌─────────────────────────────────┐
│         OWNER（企业主）          │  全部权限，不可被删除
├─────────────────────────────────┤
│     ADMIN（管理员/兼职人事）      │  配置权限，可操作全部功能
├─────────────────────────────────┤
│   MEMBER（普通成员/行政）        │  仅可查看和执行部分操作
└─────────────────────────────────┘
```

**权限矩阵**：

| 功能模块 | OWNER | ADMIN（兼职会计） | MEMBER（行政） |
|----------|-------|-------|--------|
| 企业信息设置 | ✅ | ✅ | ❌ |
| 子账号管理（增删改） | ✅ | ❌ | ❌ |
| 员工入职 | ✅ | ✅ | ✅ |
| 员工离职审核 | ✅ | ✅ | ❌ |
| 员工档案查看 | ✅ | ✅ | ✅ |
| 员工档案导出 | ✅ | ✅ | ❌ |
| 社保参保/停缴 | ✅ | ✅ | ✅ |
| 社保记录查看 | ✅ | ✅ | ✅ |
| 工资表创建/编辑 | ✅ | ✅ | ✅ |
| 工资发放/确认 | ✅ | ✅ | ❌ |
| 工资条查看 | ✅ | ✅ | ❌ |
| 个税申报提交 | ✅ | ✅ | ❌ |
| 个税记录查看 | ✅ | ✅ | ✅ |
| 凭证录入 | ✅ | ✅ | ❌ |
| 凭证审核 | ✅ | ❌ | ❌ |
| 凭证查看 | ✅ | ✅ | ❌ |
| 结账/反结账 | ✅ | ❌ | ❌ |
| 财务报表查看 | ✅ | ✅ | ❌ |
| 发票管理 | ✅ | ✅ | ❌ |
| 费用报销审批 | ✅ | ✅ | ❌ |
| 科目管理 | ✅ | ✅ | ❌ |
| 税务计算/导出 | ✅ | ✅ | ❌ |
| 通知设置 | ✅ | ✅ | ✅ |
| 数据导出 | ✅ | ✅ | ❌ |

**权限实现**：
```go
// Gin 中间件实现权限校验
func RequireRole(role string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userRole := c.GetString("user_role")
        if !hasPermission(userRole, role) {
            response.Forbidden(c, "权限不足")
            c.Abort()
            return
        }
        c.Next()
    }
}

// 路由注册示例
employeeGroup := r.Group("/api/v1/employees", middleware.Auth())
employeeGroup.POST("", handler.CreateEmployee)                           // 入职
employeeGroup.PUT("/:id/resign", middleware.RequireRole("ADMIN"), handler.Resign) // 离职审核
```

### 4.2 数据安全策略

#### 4.2.1 敏感数据加密

| 数据类型 | 加密方式 | 说明 |
|----------|----------|------|
| 手机号 | AES-256-GCM | 可逆加密，支持查询展示 |
| 身份证号 | AES-256-GCM | 可逆加密，脱敏显示 |
| 用户密码 | BCrypt（cost=12）| 不可逆哈希 |
| 工资金额 | 传输层加密（HTTPS）| 存储明文，支持计算查询 |
| 银行账号 | AES-256-GCM | 工资发放使用 |

**密钥管理方案**：
- 使用阿里云 KMS（Key Management Service）管理加密密钥，应用层通过 KMS API 获取数据加密密钥
- 密钥版本管理，支持密钥轮换时的平滑过渡
- 加密字段采用「哈希索引 + 加密存储」双字段方案，支持精确查找：
  - `phone_encrypted`（AES 加密值，用于展示）+ `phone_hash`（SHA-256 哈希值，用于查找）
  - `id_card_encrypted` + `id_card_hash` 同理

**数据脱敏规则**：

| 数据类型 | 脱敏规则 | 示例 |
|----------|----------|------|
| 手机号 | 保留前 3 后 4 | 138\*\*\*\*5678 |
| 身份证号 | 保留前 3 后 4 | 110\*\*\*\*\*\*\*\*\*1234 |
| 银行账号 | 保留后 4 | \*\*\*\*\*\*\*\*5678 |
| 姓名 | 2字保留姓，3字保留姓+尾字 | 张\* / 张\*明 |

#### 4.2.2 接口安全

| 安全措施 | 实现方式 |
|----------|----------|
| API 鉴权 | JWT Token（golang-jwt） |
| Token 有效期 | Access Token 30 分钟有效，Refresh Token 7 天有效 |
| Token 安全 | JWT payload 含 `device_id` 和 `jti`（JWT ID），服务端校验设备一致性 |
| Token 刷新 | Refresh Token 一次性使用，每次刷新时轮换，旧的立即失效 |
| Token 失效 | Redis 黑名单（至少 Sentinel 模式保障可用性） |
| 接口限流 | 按接口敏感度分级限流（见下表） |
| 登录保护 | 同一设备5次失败后锁定30分钟 |
| 签名校验 | 请求签名防篡改（HMAC-SHA256） |

**接口分级限流策略**：

| 接口类型 | 限流策略 | 理由 |
|----------|----------|------|
| 登录/验证码 | 5 次/分钟/IP | 防暴力破解 |
| 工资查询/导出 | 10 次/分钟/用户 | 敏感数据 |
| 员工列表查询 | 30 次/分钟/用户 | 常规操作 |
| 其他业务接口 | 60 次/分钟/用户 | 默认 |

**API 签名验证方案**：
- 签名输入：`HTTP Method + URL Path + Timestamp + Nonce + Request Body Hash`
- 时间戳窗口：正负 5 分钟（防重放攻击）
- Nonce 在 Redis 中存储，TTL 10 分钟
- 客户端签名密钥不硬编码，登录后由服务端动态下发并绑定 `device_id`

#### 4.2.3 传输安全

- 全站 HTTPS（Let's Encrypt / 阿里云 SSL 证书）
- 敏感接口（登录、工资、社保）额外 HMAC 签名校验
- 请求体加密（可选 V2.0，针对高安全场景）

### 4.3 操作审计

**完整操作日志记录**：

```go
// 审计日志中间件
func AuditLog(logger *AuditLogger) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        c.Next()

        logger.Record(AuditEntry{
            UserID:    middleware.GetUserID(c),
            OrgID:     middleware.GetOrgID(c),
            Module:    getModuleFromPath(c.Request.URL.Path),
            Action:    c.Request.Method,
            TargetID:  c.Param("id"),
            Detail:    fmt.Sprintf("status=%d duration=%v", c.Writer.Status(), time.Since(start)),
            IPAddress: c.ClientIP(),
        })
    }
}
```

**审计要求**：
- 工资发放记录保留至少 3 年（符合劳动法要求）
- 社保/个税申报记录永久归档
- 日志不可修改、不可删除（仅可软删除标记）
- 支持按时间/模块/操作人查询

### 4.4 合规要求

#### 4.4.1 个人信息保护法合规

| 要求 | 实现方案 |
|------|----------|
| 最小必要原则 | 仅收集完成功能所必需的信息 |
| 用户同意 | 首次使用签署《隐私政策同意书》 |
| 数据访问权 | 提供「我的企业数据导出」功能 |
| 数据删除权 | 支持企业注销后的完整数据删除 |
| 数据存储期限 | 离职员工数据保留 2 年后自动脱敏归档 |
| 安全措施披露 | 在隐私政策中明确加密、备份等安全措施 |

#### 4.4.2 劳动合同法合规

| 要求 | 实现方案 |
|------|----------|
| 书面劳动合同 | 电子签章签署，符合《电子签名法》 |
| 合同条款合规 | 使用经法务审核的标准劳动合同模板 |
| 社保缴纳义务 | 系统自动提醒 + 参保记录可追溯 |
| 工资支付凭证 | 工资条电子存档，支持导出 |

#### 4.4.3 数据备份策略

| 备份类型 | 频率 | 保留时间 | 用途 |
|----------|------|----------|------|
| 全量备份 | 每日凌晨 2:00 | 30 天 | 灾难恢复 |
| 增量备份 | 每 4 小时 | 7 天 | Point-in-Time Recovery |
| 归档备份 | 每月 1 日 | 365 天 | 合规审计 |

---

## 五、API 接口设计规范

### 6.1 RESTful API 规范

**URL 命名**：资源路径风格，统一 `/api/v1/` 前缀

```
GET    /api/v1/employees          # 员工列表
POST   /api/v1/employees          # 创建员工
GET    /api/v1/employees/{id}     # 员工详情
PUT    /api/v1/employees/{id}     # 更新员工
DELETE /api/v1/employees/{id}     # 删除员工（软删除）

GET    /api/v1/salary-tables                  # 工资表列表
POST   /api/v1/salary-tables                  # 创建工资表
GET    /api/v1/salary-tables/{id}/records     # 工资表明细
POST   /api/v1/salary-tables/{id}/confirm     # 确认工资表
```

### 6.2 统一响应格式

```json
{
  "success": true,
  "data": { ... },
  "error": null,
  "meta": {
    "total": 50,
    "page": 1,
    "limit": 20
  }
}
```

### 6.3 错误码体系

| 错误码范围 | 模块 | 示例 |
|-----------|------|------|
| 10001-19999 | 用户/认证 | 10001: 手机号格式错误 |
| 20001-29999 | 员工管理 | 20001: 员工不存在 |
| 30001-39999 | 社保管理 | 30001: 社保基数超出范围 |
| 40001-49999 | 工资管理 | 40001: 工资表已确认不可修改 |
| 50001-59999 | 个税管理 | 50001: 专项扣除信息缺失 |
| 90001-99999 | 系统错误 | 90001: 系统繁忙请重试 |

---

## 六、数据导入导出方案

### 7.1 导入方案

| 场景 | 格式 | 策略 |
|------|------|------|
| 考勤表导入 | Excel (.xlsx) | 逐行校验 + 全部成功才提交事务 |
| 员工档案导入 | Excel (.xlsx) | 逐行校验 + 错误行标注 + 跳过错误行 |
| 社保基数导入 | Excel (.xlsx) | 逐行校验 + 全部成功才提交 |

**导入流程**：
1. 用户上传文件 → 校验格式（文件大小 ≤5MB，行数 ≤1000）
2. 解析数据 → 逐行校验（类型、必填、格式）
3. 返回校验结果（成功数/失败数/失败明细）
4. 用户确认 → 批量写入数据库

### 7.2 导出方案

| 场景 | 格式 | 安全要求 |
|------|------|----------|
| 员工档案导出 | Excel | 需二次验证（密码/验证码） |
| 工资条导出 | Excel/PDF | 需二次验证，文件水印 |
| 社保凭证导出 | PDF | OSS 临时链接，24 小时过期 |
| 个税申报凭证 | PDF | OSS 临时链接，24 小时过期 |
| 离职交接清单 | PDF | 标准模板生成 |

**大文件导出**：采用异步方案，生成完成后推送通知用户下载。

---

## 七、离线与弱网策略

### 8.1 数据本地缓存

| 数据类型 | 缓存策略 | 过期时间 |
|----------|----------|----------|
| 企业信息 | 本地持久化 | 服务端推送更新 |
| 员工列表 | 本地缓存 | 5 分钟或手动刷新 |
| 社保政策 | 本地缓存 | 24 小时（政策变动不频繁） |
| 工资数据 | 不缓存 | 每次从服务端拉取（敏感数据） |

### 8.2 离线操作支持

- 员工信息录入：支持离线填写表单，联网后自动同步
- 离线队列：Room（Android）/ CoreData（iOS）存储待同步操作
- 冲突解决策略：服务端优先（Last-Write-Wins），冲突时提示用户手动合并

---

## 八、多端数据同步

### 9.1 实时更新机制

V1.0 采用**短轮询**方案（简洁可靠）：

- 老板端 APP：关键页面（首页待办）每 30 秒轮询一次
- 员工端 H5：页面进入时拉取最新数据，无实时性要求
- V2.0 可升级为 WebSocket 实现真正的实时推送

### 9.2 账号体系

- 手机号为唯一主键，支持验证码登录
- 微信账号可绑定手机号（一个手机号绑定一个微信）
- 多设备登录：V1.0 仅支持单设备同时在线（新登录踢掉旧设备）

---

## 九、日志与异常处理

### 10.1 结构化日志规范

```json
{
  "timestamp": "2026-04-06T10:00:00.000Z",
  "level": "INFO",
  "traceId": "abc123",
  "service": "employee",
  "action": "createEmployee",
  "userId": 1001,
  "orgId": 5001,
  "message": "Employee created successfully",
  "duration": 120
}
```

### 10.2 异常分类

| 异常类型 | HTTP 状态码 | 处理策略 |
|----------|-------------|----------|
| 业务异常（参数错误、状态冲突） | 400/409 | 返回明确错误信息，前端友好提示 |
| 认证异常（Token 过期、无权限） | 401/403 | 引导重新登录 |
| 第三方异常（短信发送失败） | 502 | 自动重试 + 降级方案 |
| 系统异常（数据库连接失败） | 500 | 告警通知 + 降级响应 |

---

## 十、数据库迁移管理

使用 **golang-migrate** 管理 Schema 变更：

```
backend/
└── migrations/
    ├── 000001_init_schema.up.sql           # 初始建表
    ├── 000001_init_schema.down.sql         # 回滚
    ├── 000002_add_employee_fields.up.sql   # 员工表补充字段
    ├── 000002_add_employee_fields.down.sql
    └── 000003_add_bank_accounts.up.sql     # 新增银行账户表
```

**迁移管理命令**：
```bash
# 创建新迁移文件
migrate create -ext sql -dir migrations -seq add_housing_fund

# 执行迁移
migrate -path migrations -database "postgres://..." up

# 回滚一步
migrate -path migrations -database "postgres://..." down 1
```

---

## 十一、V1.0 交付节奏建议

### 第一批（MVP，4-6 周）
- 用户注册/登录 + 企业信息录入
- 员工入职/离职管理（核心流程）
- 员工档案管理

### 第二批（4 周）
- 工资核算与工资单
- 个税自动计算

### 第三批（4 周）
- 社保管理与政策匹配
- 通知与提醒系统

### 第四批（4 周）
- 财务代理记账模块（凭证管理、发票登记、费用报销、账簿查询、财务报表）
- 会计科目预置 + 自定义科目
- 会计期间管理（结账/反结账）
- 增值税/企业所得税自动计算

### 前置工作（需提前启动）
1. 阿里云 ICP 备案（1-3 周）
2. 微信小程序注册审核（1-2 周）
3. 阿里云短信签名和模板审核（每个模板 1-2 个工作日）
4. 社保政策数据收集与清洗（1-2 周）
5. 电子签服务评估（即使 V1.0 用降级方案也需提前评估 V2.0 可行性）

---

## 十二、补充说明

### 12.1 V1.0 技术债务管理

| 技术债务项 | V1.0 实现 | V2.0/V3.0 优化 |
|-----------|-----------|----------------|
| 消息队列 | Redis Stream | RabbitMQ/Kafka |
| 社保对接 | 手动录入 + 提醒 | 政务接口自动对接 |
| 个税申报 | 自动计算 + 手动申报 | 系统一键提交 |
| 电子签 | 模板生成 + 手动签署 | 全流程自动签署 |
| 发票查验 | 手动登记 | 对接税务局查验接口 + OCR |
| 监控告警 | 基础日志 | 完整 APM + Prometheus + Grafana |
| 数据隔离 | 逻辑隔离（org_id） | 物理隔离（独立 Schema） |

### 12.2 性能指标

| 指标 | V1.0 目标 | 实现手段 |
|------|-----------|----------|
| API 响应时间 | ≤500ms | Redis 缓存 + SQL 优化 |
| 并发用户 | 1000 | 连接池 + 无状态设计 |
| 页面加载时间 | ≤2s | 接口合并 + 前端懒加载 |
| 社保核算性能 | 50人/企业 ≤3s | 批量查询 + 并行计算 |
| 工资核算性能 | 50人/企业 ≤5s | 并行计算 + 数据库优化 |
| 财务报表生成 | 50人/企业 ≤2s | 科目余额预计算 + 快照缓存 |

### 12.3 第三方服务费用估算（V1.0 月度）

| 服务 | 预估费用 | 说明 |
|------|----------|------|
| 阿里云 ECS（2C4G） | ¥300/月 | 应用服务器 |
| 阿里云 RDS PostgreSQL | ¥150/月 | 数据库 |
| 阿里云 Redis | ¥50/月 | 缓存 |
| 阿里云 OSS | ¥30/月 | 文件存储（含合同/凭证/发票） |
| 阿里云短信 | ¥0.04/条 × ~1000条 = ¥40/月 | 验证码 + 通知 |
| 电子签服务 | ¥0（V1.0 降级方案，无 API 调用） | PDF 模板生成 |
| SSL 证书 | ¥0（Let's Encrypt） | 免费 |
| 合计 | **~¥570/月** | V1.0 轻量运行成本很低 |

### 12.4 项目启动清单

- [ ] 阿里云账号初始化（ECS、RDS、Redis、OSS、Sentry）
- [ ] 短信服务签名审核、模板报备
- [ ] 微信小程序注册审核
- [ ] 电子签服务评估（V2.0 对接可行性）
- [ ] 社保政策数据初始化（30+ 城市）
- [ ] Go 项目脚手架搭建（Gin + Ent/GORM + 基础模块初始化）
- [ ] Android 项目创建 + Kotlin 配置
- [ ] iOS 项目创建 + Swift 配置
- [ ] 微信小程序项目创建
- [ ] H5 管理后台项目创建（Vue 3 + Element Plus）
- [ ] CI/CD 流水线搭建（GitHub Actions）
- [ ] PostgreSQL 数据库建表 + 索引创建
- [ ] Redis 缓存策略配置
- [ ] OSS Bucket 创建 + CDN 配置
- [ ] 会计科目预置数据初始化
