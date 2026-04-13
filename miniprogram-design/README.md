# EasyHR 员工端小程序设计

> 易人事（EasyHR）员工端微信小程序界面设计与规范

## 目录结构

```
miniprogram-design/
├── README.md           # 本文件，项目概述
├── DESIGN.md           # 完整设计规范（色彩、字体、组件、交互）
├── WIREFRAMES.md       # 页面线框图与交互说明
├── SPEC.md             # 页面功能规格说明
└── assets/             # 设计资源（图标、插图）
    └── icons/
```

## 设计规范速查

| 维度 | 规范值 |
|------|--------|
| 主色 | `#4F46E5` (Indigo) |
| 背景 | `#F8FAFC` |
| 字体 | Noto Sans SC |
| 圆角 | 卡片 12px / 按钮 999px |
| 触摸目标 | ≥ 44×44px |
| TabBar | 4项（首页/工资条/社保/我的） |

## 页面清单

| 页面 | 路径 | 说明 |
|------|------|------|
| 首页 | `pages/index/index` | Dashboard，展示工资/社保摘要+公告 |
| 工资条列表 | `pages/payslip/list/index` | 月度工资列表，支持折叠展开 |
| 工资明细 | `pages/payslip/detail/index` | 单月工资完整明细 |
| 社保主页 | `pages/social/index/index` | 社保公积金缴纳状态与记录 |
| 我的 | `pages/profile/index/index` | 用户信息+功能入口 |
| 合同列表 | `pages/profile/contract/list` | 劳动合同/保密协议列表 |
| 合同详情 | `pages/profile/contract/detail` | 单份合同 PDF 预览 |
| 报销列表 | `pages/profile/expense/list` | 报销历史记录 |
| 报销申请 | `pages/profile/expense/form` | 提交新报销（表单） |
| 设置 | `pages/profile/settings/index` | 消息通知/隐私设置 |
| 帮助反馈 | `pages/profile/help/index` | 帮助中心+提交反馈 |

## 技术栈

- **框架**：微信小程序原生框架（无跨端框架）
- **UI 库**：WeUI（微信官方基础样式库）
- **图标**：WeUI Icons + 自定义 SVG 图标
- **字体**：Noto Sans SC（Google Fonts CDN）
- **状态管理**：微信小程序 `Behavior` + `Component`
- **数据**：wx.request 封装，JWT 鉴权

## 设计文件说明

- **DESIGN.md**：色彩令牌、字体系统、组件规范、交互动效、可访问性要求
- **WIREFRAMES.md**：每个页面的 ASCII 线框图、页面跳转关系、全局样式类
- **SPEC.md**：每个页面的功能规格、字段说明、数据接口描述

## 设计原则

1. **3步完成**：核心操作不超过3步
2. **专业可信**：Indigo 色系 + 清晰数据展示，传递信任感
3. **零学习成本**：界面符合微信小程序规范，用户无需指导
4. **触摸友好**：最小触摸目标 44×44px，间距合理
5. **加载体验**：骨架屏替代 spinner，感知速度更快

## 设计关键词

```
企业 SaaS · 员工自助 · HR 工具 · 工资条 · 社保公积金
微信小程序 · WeUI · Indigo #4F46E5 · 简洁专业 · 敏感数据安全
```

## 相关文档

- [易人事产品需求文档](../docs/)
- [后端 API 文档](../backend/docs/)
- [H5管理后台设计](../h5-design/)

---

*最后更新：2026-04-13*
