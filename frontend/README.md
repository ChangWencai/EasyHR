# EasyHR H5 管理后台

易人事（EasyHR）老板端 H5 管理后台，采用 Vue 3 + TypeScript + Vite 构建。

## 技术栈

- **框架**: Vue 3.5 + Composition API
- **构建**: Vite 8 + TypeScript 6
- **UI 组件**: Element Plus 2
- **状态管理**: Pinia 3
- **路由**: Vue Router 5
- **HTTP**: Axios 1
- **工具库**: @vueuse/core 14, dayjs 1

## 项目结构

```
frontend/
├── src/
│   ├── api/          # API 接口封装
│   ├── components/   # 公共组件
│   ├── stores/       # Pinia 状态管理
│   ├── utils/        # 工具函数
│   ├── views/        # 页面组件
│   │   ├── dashboard/   # Dashboard
│   │   ├── employee/    # 员工管理
│   │   ├── finance/     # 财务记账
│   │   ├── home/        # 首页
│   │   ├── layout/      # 布局组件
│   │   ├── mine/        # 我的
│   │   ├── onboarding/  # 企业入驻
│   │   ├── salary/      # 工资管理
│   │   ├── tool/        # 工具（社保、个税）
│   │   └── wxmployees/  # 员工管理
│   ├── App.vue
│   └── main.ts
├── index.html
├── vite.config.ts
└── tsconfig.json
```

## 开发

```bash
# 安装依赖
pnpm install

# 开发模式
pnpm dev

# 类型检查
pnpm type-check

# 构建生产版本
pnpm build
```

## 页面路由

| 路径 | 页面 | 说明 |
|------|------|------|
| `/login` | 登录页 | 手机号登录/注册 |
| `/onboarding/org-setup` | 企业入驻 | 首次登录填写企业信息 |
| `/home` | 首页 | Dashboard + 待办事项 |
| `/employee` | 员工列表 | 员工管理 |
| `/employee/create` | 新增员工 | 入职登记 |
| `/invitation` | 邀请管理 | 邀请员工入职 |
| `/offboarding` | 离职管理 | 员工离职 |
| `/finance` | 财务记账 | 凭证、发票、报销 |
| `/salary` | 工资管理 | 工资条、考勤 |
| `/social` | 社保管理 | 参保、缴费 |
| `/tax` | 个税申报 | 个税计算、扣除项 |
| `/mine` | 我的 | 个人中心、企业设置 |
