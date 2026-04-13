# Phase 2: 员工管理 - Research

**Date:** 2026-04-06
**Phase:** 02-employee-management
**Status:** Research complete

---

## 1. 技术选型验证

### excelize v2.10.1 — Excel 导出

**可靠性：** HIGH — Go 生态最成熟的 Excel 库，GitHub 17k+ stars

**关键模式：**
- `excelize.NewFile()` 创建文件
- `SetCellValue()` 设置单元格值
- `SetColWidth()` 设置列宽
- `SetCellStyle()` 设置样式（字体、对齐、边框）
- `WriteToBuffer()` 输出到 buffer，无需临时文件
- 中文支持：默认支持 UTF-8，无需特殊配置

**导出员工档案建议流程：**
1. 创建 workbook + 设置 header 样式
2. 写入表头行（姓名、手机号、身份证号、岗位、入职日期等）
3. 遍历员工数据写入行（敏感字段使用脱敏值）
4. 自适应列宽 + 冻结首行
5. 返回 `[]byte` 给 HTTP handler 设置 `Content-Disposition` 下载

**性能考量：** 1000行员工数据导出 < 100ms，无需流式处理。

### go-pdf/fpdf v0.9.0 — 合同 PDF 生成

**可靠性：** MEDIUM — 纯 Go 实现，无 CGO 依赖，轻量

**中文支持关键：**
- fpdf 默认字体不含中文，**必须** 注册中文字体文件
- 推荐使用开源中文字体（如思源黑体 SourceHanSansCN 或文泉驿）
- 字体文件可嵌入二进制或放 OSS 按需下载
- `AddUTF8Font()` 注册字体，`SetFont()` 使用

**合同 PDF 模板设计：**
```
┌──────────────────────────────┐
│       劳动合同书               │
│                               │
│  甲方：{企业名称}              │
│  统一社会信用代码：{信用代码}   │
│                               │
│  乙方：{员工姓名}              │
│  身份证号：{身份证号}          │
│                               │
│  合同期限：{起始日}至{终止日}  │
│  工作岗位：{岗位}              │
│  工作地点：{城市}              │
│  薪资待遇：{薪资}元/月         │
│                               │
│  甲方签章：______ 日期：______ │
│  乙方签章：______ 日期：______ │
└──────────────────────────────┘
```

**替代方案：** gofpdi 可导入现有 PDF 模板填充字段，但增加依赖。V1.0 用 fpdf 直接生成更简单。

### 邀请 Token 安全

**推荐方案：** `crypto/rand` 生成 32 字节随机 token + hex 编码（64字符）

```go
token := make([]byte, 32)
crypto_rand.Read(token)
inviteToken := hex.EncodeToString(token)
```

**不用 UUID 的理由：** UUID v4 包含时间信息和固定格式，可预测性略高。纯随机 hex 更适合短期邀请链接。

**安全措施：**
- Token 一次性使用（提交后标记已使用）
- 7天过期（数据库字段 `expires_at`）
- 每个邀请限制提交次数为1
- Rate limiting：同一 token 每分钟最多 10 次查看

---

## 2. 数据模型设计

### Employee 模型

```
employees 表:
├── BaseModel (id, org_id, created_by/at, updated_by/at, deleted_at)
├── name              varchar(50)     NOT NULL  姓名
├── phone_encrypted   varchar(200)    NOT NULL  手机号(加密)
├── phone_hash        varchar(64)     UNIQUE(partial)  手机号哈希索引
├── id_card_encrypted varchar(200)              身份证号(加密)
├── id_card_hash      varchar(64)     UNIQUE(partial)  身份证哈希索引
├── gender            varchar(10)               性别(从身份证提取)
├── birth_date        date                      出生日期(从身份证提取)
├── position          varchar(100)    NOT NULL  岗位
├── hire_date         date            NOT NULL  入职日期
├── status            varchar(20)     NOT NULL  状态(pending/probation/active/resigned)
├── user_id           int64                     关联用户ID(nullable)
├── bank_name         varchar(100)              开户银行
├── bank_account_encrypted varchar(200)         银行卡号(加密)
├── bank_account_hash varchar(64)                银行卡哈希索引
├── emergency_contact varchar(50)               紧急联系人
├── emergency_phone_encrypted varchar(200)      紧急联系人电话(加密)
├── emergency_phone_hash varchar(64)            紧急联系人电话哈希
├── address           varchar(500)              住址
├── remark            text                      备注
├── resignation_date  date                      离职日期
├── resignation_reason varchar(500)             离职原因
```

**唯一约束（部分索引）：**
- `phone_hash WHERE deleted_at IS NULL` — 同一企业内手机号唯一
- `id_card_hash WHERE deleted_at IS NULL` — 同一企业内身份证号唯一

### Invitation 模型

```
invitations 表:
├── id                int64 PK
├── org_id            int64 NOT NULL
├── token             varchar(64) UNIQUE NOT NULL  邀请token
├── position          varchar(100)                 预设岗位
├── status            varchar(20) NOT NULL         pending/used/expired/cancelled
├── created_by        int64 NOT NULL
├── created_at        timestamp NOT NULL
├── expires_at        timestamp NOT NULL           过期时间
├── used_at           timestamp                    使用时间
├── employee_id       int64                        关联的员工ID(提交后填充)
```

### Contract 模型

```
contracts 表:
├── BaseModel (id, org_id, created_by/at, updated_by/at, deleted_at)
├── employee_id       int64 NOT NULL       关联员工
├── contract_type     varchar(20) NOT NULL fixed_term/indefinite/intern
├── start_date        date NOT NULL        合同起始日期
├── end_date          date                 合同终止日期(无固定期限为空)
├── salary            decimal(10,2)        合同薪资
├── status            varchar(20) NOT NULL draft/pending_sign/signed/active/terminated/expired
├── pdf_url           varchar(500)         生成的PDF的OSS URL
├── signed_pdf_url    varchar(500)         签署后上传的扫描件OSS URL
├── sign_date         date                 签署日期
├── terminate_date    date                 终止日期
├── terminate_reason  varchar(500)         终止原因
```

### Offboarding 模型

```
offboardings 表:
├── BaseModel (id, org_id, created_by/at, updated_by/at, deleted_at)
├── employee_id       int64 NOT NULL       关联员工
├── type              varchar(20) NOT NULL voluntary/involuntary  主动/被动
├── resignation_date  date NOT NULL        离职日期
├── reason            varchar(500)         离职原因
├── status            varchar(20) NOT NULL pending/approved/completed
├── checklist_items   jsonb NOT NULL       交接清单[{category, items:[{name, completed}]}]
├── completed_at      timestamp            完成时间
├── approved_by       int64                审批人
├── approved_at       timestamp            审批时间
```

---

## 3. API 端点设计

### 员工管理
```
POST   /api/v1/employees                    手动创建员工
GET    /api/v1/employees                    员工列表(分页+搜索)
GET    /api/v1/employees/:id                员工详情
PUT    /api/v1/employees/:id                更新员工信息
DELETE /api/v1/employees/:id                删除员工(软删除)
GET    /api/v1/employees/export             导出Excel

POST   /api/v1/employees/:id/sensitive      查看敏感信息(临时解密)
```

### 入职邀请
```
POST   /api/v1/invitations                  创建入职邀请
GET    /api/v1/invitations                  邀请列表
DELETE /api/v1/invitations/:token           取消邀请

POST   /api/v1/invitations/:token/submit    员工提交信息(公开接口,无需auth)
GET    /api/v1/invitations/:token           查看邀请详情(公开接口)
POST   /api/v1/employees/:id/confirm        老板确认入职
```

### 离职管理
```
POST   /api/v1/employees/:id/resign         老板办理离职
POST   /api/v1/employees/:id/resign/apply   员工申请离职(H5)
PUT    /api/v1/offboardings/:id/approve     审批离职申请
PUT    /api/v1/offboardings/:id/complete    完成交接
GET    /api/v1/offboardings/:id             离职详情+交接清单
PUT    /api/v1/offboardings/:id/checklist   更新交接清单
```

### 合同管理
```
POST   /api/v1/employees/:id/contracts      创建合同
GET    /api/v1/employees/:id/contracts      合同列表
GET    /api/v1/contracts/:id                合同详情
PUT    /api/v1/contracts/:id                更新合同
POST   /api/v1/contracts/:id/generate-pdf   生成PDF
POST   /api/v1/contracts/:id/upload-signed  上传签署扫描件
PUT    /api/v1/contracts/:id/terminate      终止合同
```

---

## 4. 搜索策略

### 模糊搜索（姓名、岗位）
```sql
-- PostgreSQL ILIKE 模糊匹配，GORM 实现
WHERE name ILIKE '%keyword%' AND org_id = ?
```

**性能：** 10-50人规模，无需全文索引。GORM `db.Where("name ILIKE ?", "%"+keyword+"%")` 足够。

### 精确搜索（手机号、身份证号）
```sql
-- 通过哈希索引精确查找
WHERE phone_hash = SHA256(phone) AND org_id = ?
```

**理由：** 加密字段无法直接 LIKE，通过 SHA-256 哈希索引实现精确匹配。

### 组合搜索
```go
// 动态条件构建
func (r *Repo) Search(orgID int64, params SearchParams) ([]Employee, int64, error) {
    query := r.db.Where("org_id = ?", orgID)
    if params.Name != "" {
        query = query.Where("name ILIKE ?", "%"+params.Name+"%")
    }
    if params.Position != "" {
        query = query.Where("position ILIKE ?", "%"+params.Position+"%")
    }
    if params.Phone != "" {
        hash := crypto.HashSHA256(params.Phone)
        query = query.Where("phone_hash = ?", hash)
    }
    if params.Status != "" {
        query = query.Where("status = ?", params.Status)
    }
    // count + paginate
}
```

---

## 5. 事件触发机制

### 离职触发社保停缴提醒

**V1.0 方案：Service 层同步调用**

```go
// internal/employee/service.go
func (s *Service) CompleteResignation(ctx context.Context, orgID, employeeID int64) error {
    // 1. 更新员工状态为"离职"
    // 2. 保存离职记录
    // 3. 触发后续流程
    s.onEmployeeResigned(orgID, employeeID)
    return nil
}

// onEmployeeResigned 预留接口，Phase 3 实现具体逻辑
func (s *Service) onEmployeeResigned(orgID, employeeID int64) {
    // Phase 3: 创建社保停缴提醒
    // s.socialInsuranceSvc.CreateStopReminder(orgID, employeeID)
    logger.Info("employee resigned event", zap.Int64("org_id", orgID), zap.Int64("employee_id", employeeID))
}
```

**V2.0 升级路径：** 引入 asynq 异步任务队列，发布 `{orgID, employeeID, "resigned"}` 事件，Phase 3 消费者订阅处理。V1.0 同步调用满足需求。

---

## 6. 身份证号信息提取

**18位身份证号结构：**
- 1-6位：地区码
- 7-14位：出生日期（YYYYMMDD）
- 15-17位：顺序码（奇数男，偶数女）
- 18位：校验码

```go
func extractFromIDCard(idCard string) (gender string, birthDate time.Time, err error) {
    if len(idCard) != 18 {
        return "", time.Time{}, fmt.Errorf("invalid id card length")
    }
    // 性别
    seq, _ := strconv.Atoi(idCard[16:17])
    if seq%2 == 1 {
        gender = "男"
    } else {
        gender = "女"
    }
    // 出生日期
    birthDate, err = time.Parse("20060102", idCard[6:14])
    if err != nil {
        return "", time.Time{}, fmt.Errorf("invalid birth date in id card")
    }
    return gender, birthDate, nil
}
```

---

## 7. Chinese Labor Law 合规要点

### 入职必备文档
1. 劳动合同（入职30日内必须签署）
2. 员工信息登记表（姓名、身份证、联系方式、紧急联系人）
3. 岗位说明书（可选，建议有）

### 离职合规
1. 试用期离职：提前3天通知
2. 正式员工主动离职：提前30天书面通知
3. 企业辞退：需支付经济补偿金（N+1）
4. 离职证明：离职后15日内出具
5. 档案/社保转移：15日内办理

**V1.0 实现建议：**
- 合同模板包含标准劳动法条款
- 离职流程记录日期和原因（满足合规追溯）
- 离职证明可通过合同模块生成 PDF

---

## 8. Validation Architecture

### 关键验证点

| 维度 | 验证项 | 方法 |
|------|--------|------|
| D1-完整性 | 所有 EMPL 需求有对应测试 | grep test files |
| D2-正确性 | 身份证号提取性别/生日准确 | unit test |
| D3-安全性 | 敏感字段加密存储 | DB assertion |
| D4-安全性 | API响应脱敏 | response assertion |
| D5-多租户 | 不同企业员工数据隔离 | integration test |
| D6-RBAC | MEMBER无法导出/查看敏感信息 | middleware test |
| D7-性能 | 1000员工搜索<500ms | benchmark |
| D8-合规 | 离职流程记录日期和原因 | DB assertion |

### 测试策略

**单元测试：**
- `employee/service_test.go` — 创建、搜索、离职、合同逻辑
- `employee/repository_test.go` — CRUD、搜索、分页
- `contract/service_test.go` — PDF生成、状态流转
- `invitation/service_test.go` — token生成、过期、提交

**集成测试：**
- `test/integration/employee_test.go` — 完整入职→在职→离职流程
- 多租户隔离验证（两个企业的员工互不可见）
- RBAC 权限验证（MEMBER无法访问敏感接口）

---

## RESEARCH COMPLETE

**Summary:**
- excelize v2.10.1 和 go-pdf/fpdf v0.9.0 满足 Excel 导出和 PDF 生成需求
- 邀请 token 使用 crypto/rand 32字节 hex，7天过期
- 搜索使用 ILIKE + SHA-256 hash 混合策略
- 离职事件 V1.0 同步调用，预留 Phase 3 接口
- 身份证号可自动提取性别和出生日期
- 4个数据模型：Employee、Invitation、Contract、Offboarding
