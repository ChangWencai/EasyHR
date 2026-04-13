# EasyHR 员工端小程序 — 功能规格说明

> 本文档描述每个页面的功能规格、字段说明与数据接口。

---

## 页面 1：首页（Dashboard）

### 路径
`pages/index/index`

### 功能描述
员工打开小程序的第一屏，展示当月工资/社保状态、最新公告、快捷入口。

### 数据来源
| 区块 | 接口 | 方法 |
|------|------|------|
| 用户信息 | `/api/v1/employee/profile` | GET |
| 当月工资摘要 | `/api/v1/payslip/current/summary` | GET |
| 当月社保状态 | `/api/v1/social/current/status` | GET |
| 公告列表（最新3条）| `/api/v1/announcements?limit=3` | GET |

### 字段说明

**用户信息**
```typescript
interface EmployeeProfile {
  id: string;
  name: string;           // "张三"
  department: string;    // "产品部"
  position: string;      // "产品经理"
  employeeNo: string;    // "E0012"
  entryDate: string;      // "2024-06-01"
  avatarUrl?: string;    // 头像 URL
}
```

**工资摘要**
```typescript
interface PayslipSummary {
  month: string;          // "2026-03"
  grossSalary: number;    // 9500.00（税前）
  netSalary: number;     // 8534.75（实发）
  payDate: string;        // "2026-03-25"
  status: 'paid' | 'pending' | 'processing';
}
```

**社保状态**
```typescript
interface SocialStatus {
  month: string;
  pension: { personal: number; company: number; status: 'paid' | 'pending' };
  medical: { personal: number; company: number; status: 'paid' | 'pending' };
  unemployment: { personal: number; company: number; status: 'paid' | 'pending' };
  housing: { personal: number; company: number; status: 'paid' | 'pending' };
  allPaid: boolean;      // true = 全部已缴
}
```

### 交互规则
- 页面加载时显示骨架屏，数据返回后渐隐过渡
- 工资/社保卡片点击 → 跳对应 Tab 并高亮当前项
- 公告点击 → 公告详情页（`/pages/announcement/detail?id=xxx`）
- 快捷入口：合同 → 合同列表， 报销 → 报销表单， 我的 → 我的页面

---

## 页面 2：工资条列表

### 路径
`pages/payslip/list/index`

### 功能描述
展示员工所有月份的工资记录，支持折叠/展开、按月筛选。

### 接口
| 操作 | 接口 | 方法 |
|------|------|------|
| 获取列表 | `/api/v1/payslip/list?page=1&limit=12` | GET |
| 导出工资条 | `/api/v1/payslip/{id}/export?format=pdf` | GET |

### 列表项字段
```typescript
interface PayslipListItem {
  id: string;
  month: string;          // "2026-03"
  yearMonth: string;      // "2026年3月"
  grossSalary: number;    // 9500.00
  socialDeduction: number; // 1472.50
  taxDeduction: number;   // 492.75
  netSalary: number;      // 8534.75
  payDate?: string;      // "2026-03-25"（paid 时有）
  status: 'paid' | 'pending' | 'processing';
  summary: {
    // 折叠时显示：基础工资/岗位工资/绩效奖金各1项
    baseSalary: number;
    positionSalary: number;
    bonus: number;
  };
}
```

### 交互规则
- 首次加载显示骨架屏（12条占位）
- 下拉刷新：重新请求第一页
- 上拉加载：page+1，追加到列表（无限滚动）
- 卡片折叠/展开动画：200ms height 过渡
- 导出按钮：显示底部操作面板（PDF / 截图 / 分享）
- 空数据：显示空状态插图 + 引导文案

---

## 页面 3：工资明细

### 路径
`pages/payslip/detail/index`

### 参数
```
?id=payslip_id
```

### 接口
| 操作 | 接口 | 方法 |
|------|------|------|
| 获取明细 | `/api/v1/payslip/{id}` | GET |

### 字段说明
```typescript
interface PayslipDetail {
  id: string;
  month: string;           // "2026-03"
  payDate: string;
  status: 'paid' | 'pending' | 'processing';

  income: {
    baseSalary: number;    // 5000.00
    positionSalary: number; // 2000.00
    bonus: number;        // 2500.00
    overtimePay?: number;  // 0.00
    allowances?: number;   // 0.00
    grossTotal: number;   // 9500.00
  };

  deduction: {
    pension: number;       // 760.00
    medical: number;       // 190.00
    unemployment: number; // 47.50
    housingFund: number;   // 475.00
    personalTotal: number; // 1472.50
    companyPension: number;    // 1200.00
    companyMedical: number;    // 480.00
    companyUnemployment: number; // 120.00
    companyHousing: number;    // 475.00
    companyTotal: number;     // 2275.00
  };

  tax: {
    taxableAmount: number; // 7027.50
    taxRate: number;       // 0.10
    quickDeduction: number; // 210.00
    taxAmount: number;     // 492.75
  };

  netSalary: number;      // 8534.75
}
```

### 交互规则
- 页面加载时显示骨架屏
- 金额字段右对齐，数字使用等宽字体（tabular）
- 分享按钮：调用 `wx.showShareMenu`，分享当前页面卡片截图
- 保存到相册：使用 `wx.canvasToTempFilePath` 生成图片
- 底部隐私提示文字，灰色小字

---

## 页面 4：社保公积金

### 路径
`pages/social/index/index`

### 接口
| 操作 | 接口 | 方法 |
|------|------|------|
| 获取社保信息 | `/api/v1/social/current` | GET |
| 获取缴纳记录 | `/api/v1/social/records?page=1&limit=12` | GET |

### 字段说明
```typescript
interface SocialInsurance {
  employeeName: string;
  socialAccountNo: string;    // 脱敏：310********1234
  housingAccountNo: string;   // 脱敏：310********5678
  insuranceType: string;      // "城镇职工社保"
  baseAmount: number;         // 6000.00（缴费基数）

  currentMonth: string;       // "2026-04"
  contributions: {
    pension: {
      personal: number;  // 760.00
      company: number;  // 1200.00
      rate: string;     // "8%"
      status: 'paid' | 'pending';
    };
    medical: {
      personal: number;  // 190.00
      company: number;  // 480.00
      rate: string;     // "2%"
      status: 'paid' | 'pending';
    };
    unemployment: {
      personal: number;  // 47.50
      company: number;  // 120.00
      rate: string;     // "0.5%"
      status: 'paid' | 'pending';
    };
    housingFund: {
      personal: number;  // 475.00
      company: number;  // 475.00
      rate: string;     // "12%"
      status: 'paid' | 'pending';
    };
  };
}
```

### 交互规则
- 账号信息提供"复制"功能（`wx.setClipboardData`）
- 各险种 2×2 网格卡片布局，状态徽章标识
- 缴纳记录分月展示，tap 展开查看当月详细
- "查看历年缴纳明细" → `pages/social/history/index`

---

## 页面 5：我的

### 路径
`pages/profile/index/index`

### 功能
- 显示个人信息摘要
- 导航到各子功能页面

### 子功能列表

| 入口 | 目标页面 | 说明 |
|------|----------|------|
| 我的合同 | `/pages/profile/contract/list` | 查看所有合同文件 |
| 报销申请 | `/pages/profile/expense/form` | 新建报销 |
| 报销历史 | `/pages/profile/expense/list` | 已有报销记录 |
| 请假记录 | `/pages/profile/leave/list` | 请假申请与状态 |
| 紧急联系人 | `/pages/profile/emergency/index` | 紧急联系人管理 |
| 设置 | `/pages/profile/settings/index` | 通知/隐私设置 |
| 帮助与反馈 | `/pages/profile/help/index` | 帮助中心+问题反馈 |
| 关于我们 | `/pages/profile/about/index` | App 版本/公司信息 |

---

## 页面 6：报销申请表单

### 路径
`pages/profile/expense/form`

### 表单字段

| 字段 | 类型 | 必填 | 校验规则 |
|------|------|------|----------|
| 报销类型 | picker（下拉） | 是 | 不能为空 |
| 发票金额 | number | 是 | > 0，最多2位小数 |
| 发票日期 | date picker | 是 | 距今 ≤ 90天 |
| 费用说明 | textarea | 否 | 最多200字 |
| 发票图片 | image[] | 是 | 1-5张，单张 ≤ 5MB |

### 报销类型枚举
```typescript
type ExpenseType =
  | 'travel'     // 差旅费
  | 'transport'  // 交通费
  | 'comm'       // 通讯费
  | 'meal'       // 餐饮费
  | 'entertain'  // 业务招待费
  | 'office'     // 办公用品
  | 'other';     // 其他
```

### 接口
| 操作 | 接口 | 方法 |
|------|------|------|
| 提交报销 | `/api/v1/expense/submit` | POST |
| 上传发票 | `/api/v1/expense/upload` | POST（formData）|

### 校验规则
- 发票金额：正数，最多2位小数
- 发票日期：不超过今天，不早于90天前
- 图片：JPG/PNG，单张 ≤ 5MB，总数 1-5张
- 提交时显示 loading，禁止重复提交

### 交互规则
- 必填字段未填时提交按钮禁用（disabled + 灰色）
- 上传图片：最多5张，超过提示"最多上传5张"
- 图片预览：tap 放大，tap X 删除
- 提交成功：显示成功 toast → 跳转报销列表
- 提交失败：显示错误 toast，不跳转

---

## 页面 7：合同列表

### 路径
`pages/profile/contract/list`

### 接口
| 操作 | 接口 | 方法 |
|------|------|------|
| 获取合同列表 | `/api/v1/contract/list` | GET |

### 字段说明
```typescript
interface Contract {
  id: string;
  type: 'labor' | 'confidentiality' | 'non-compete' | 'other';
  typeName: string;      // "劳动合同（全职）"
  title: string;        // "张三劳动合同"
  signDate: string;      // "2024-06-01"
  startDate?: string;    // 合同起始日期
  endDate?: string;      // 合同结束日期
  status: 'active' | 'expired' | 'terminated';
  position?: string;     // "产品经理"
  salary?: string;       // "¥9,500/月"
  compensation?: string;  // 竞业限制补偿金
  duration?: string;      // 协议期限
}
```

### 筛选标签
```
全部 | 劳动合同 | 保密协议 | 竞业限制 | 其他
```

---

## 通用交互规范

### 骨架屏时机
- 页面初始加载，数据未返回时（> 200ms 时长）

### 空状态
- 无数据时显示：插图 + 说明文案 + 主操作按钮
- 每个列表页都需要空状态设计

### 错误状态
- 网络错误：显示断网插图 + "重新加载"按钮
- 接口错误：显示 toast 提示错误原因，3s 自动消失

### 下拉刷新
- 所有列表页支持下拉刷新（`onPullDownRefresh`）
- 刷新时顶部显示加载动画

### 上拉加载
- 列表达到一定长度后，底部显示"上拉加载更多"
- 无更多数据时显示"— 已加载全部 —"

---

## 路由配置（app.json pages 字段）

```json
{
  "pages": [
    "pages/index/index",
    "pages/payslip/list/index",
    "pages/payslip/detail/index",
    "pages/social/index/index",
    "pages/profile/index/index",
    "pages/profile/contract/list/index",
    "pages/profile/contract/detail/index",
    "pages/profile/expense/list/index",
    "pages/profile/expense/form/index",
    "pages/profile/leave/list/index",
    "pages/profile/emergency/index/index",
    "pages/profile/settings/index/index",
    "pages/profile/help/index/index",
    "pages/profile/about/index/index"
  ],
  "window": {
    "navigationBarTitleText": "易人事",
    "navigationBarBackgroundColor": "#FFFFFF",
    "navigationBarTextStyle": "black",
    "backgroundColor": "#F8FAFC"
  },
  "tabBar": {
    "color": "#64748B",
    "selectedColor": "#4F46E5",
    "backgroundColor": "#FFFFFF",
    "borderStyle": "white",
    "list": [
      { "pagePath": "pages/index/index", "text": "首页" },
      { "pagePath": "pages/payslip/list/index", "text": "工资条" },
      { "pagePath": "pages/social/index/index", "text": "社保" },
      { "pagePath": "pages/profile/index/index", "text": "我的" }
    ]
  }
}
```

---

*功能规格版本：v1.0 | 最后更新：2026-04-13*
