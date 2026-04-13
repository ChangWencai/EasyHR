# EasyHR 小程序图标规范

## 概述

微信小程序员工端使用 **WeUI 图标库** 作为基础图标，辅以少量自定义 SVG 图标补充。

## 基础图标库：WeUI Icons

WeUI 是微信官方提供的基础样式库，包含 86 个常用操作图标，完全兼容小程序 WXML。

### 引入方式
```json
// app.json 或页面的 json 配置
{
  "usingComponents": {
    "van-icon": "@vant/weapp/icon/index"
  }
}
```

> **推荐方案**：使用 [@vant/weapp](https://vant-contrib.gitee.io/vant-weapp/#/home) 的 Icon 组件，它扩展了 WeUI 图标集，覆盖更全面。

### 核心图标对照表

| 功能 | WeUI 图标名 | 代码 |
|------|------------|------|
| 首页 | home | `icon-home` |
| 钱包/工资 | wallet | `icon-wallet` |
| 盾牌/社保 | shield | `icon-shield` 或用 `icon-safe` |
| 用户/我的 | user | `icon-contact` |
| 合同/文档 | document | `icon-description` |
| 报销/钱 | money | `icon-pending` 或 `icon-card` |
| 设置 | setting | `icon-setting` |
| 帮助 | question | `icon-question` |
| 关于 | info | `icon-info` |
| 公告 | announ | `icon-speaker` 或 `icon-add-dot` |
| 箭头右 | 右箭头 | `icon-arrow` |
| 返回 | 返回 | `icon-arrow-left` |
| 关闭 | 关闭 | `icon-cross` |
| 通知铃铛 | 铃铛 | `icon-bell` |
| 复制 | 复制 | `icon-copy` |
| 筛选 | 筛选 | `icon-filter` |
| 导出 | 下载 | `icon-download` |
| 分享 | 分享 | `icon-share` |
| 上传 | 加号 | `icon-plus` |
| 删除 | 删除 | `icon-delete` |
| 勾选成功 | 勾选 | `icon-checked` |
| 警告 | 警告 | `icon-warning` |
| 错误 | 错误 | `icon-fail` |
| 加载 | 加载 | `icon-loading` |
| 编辑 | 编辑 | `icon-edit` |
| 查看 | 眼睛 | `icon-see` |
| 请假 | 日历 | `icon-calendar` |
| 电话 | 电话 | `icon-phone` |

## 自定义图标（补充）

以下图标需要自定义 SVG 或使用 iconfont：

### 社保险种图标（4个）
```
social-pension.svg     # 养老保险（小人+盾牌）
social-medical.svg      # 医疗保险（+字）
social-unemployment.svg # 失业保险（向下箭头）
social-housing.svg      # 住房公积金（房子）
```

### 快捷入口图标
```
quick-contract.svg       # 合同
quick-expense.svg        # 报销
quick-profile.svg        # 我的
quick-help.svg           # 帮助
```

### 插图（空状态用）
```
empty-payslip.svg       # 无工资记录
empty-social.svg         # 无社保记录
empty-contract.svg       # 无合同
empty-expense.svg        # 无报销记录
empty-network.svg        # 网络错误
```

## 图标设计规范

### 风格
- 统一使用 **线性图标**（Line icons），非填充图标
- 线宽：1.5px（2x 分辨率下为 3px）
- 圆角：图标本身的圆角统一 2px
- 颜色：当前色 `--text-muted`，选中/激活色 `--primary`

### 尺寸
| 用途 | 尺寸 |
|------|------|
| TabBar 图标 | 24×24px |
| 列表项图标 | 20×20px |
| 快捷入口图标 | 32×32px |
| 空状态插图 | 80×80px |
| 按钮图标 | 18×18px |

### 不使用 Emoji
所有图标必须为 SVG矢量，**禁止使用 Emoji** 作为功能图标。

## 图标色值

```css
/* 未选中状态 */
.icon-default {
  color: #64748B;  /* --text-muted */
}

/* 选中/激活状态 */
.icon-active {
  color: #4F46E5;  /* --primary */
}

/* 功能色图标 */
.icon-success { color: #10B981; }  /* 成功 */
.icon-warning { color: #F59E0B; }  /* 警告 */
.icon-danger  { color: #EF4444; }  /* 错误 */
.icon-info    { color: #0EA5E9; }  /* 信息 */
```

---

*图标规范版本：v1.0 | 最后更新：2026-04-13*
