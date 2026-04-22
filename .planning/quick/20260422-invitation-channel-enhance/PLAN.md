# Plan: 入职邀请渠道增强

增强入职邀请功能，支持两种推送渠道：微信小程序（需保存手机号+岗位）和邮箱（需关联模板）。

## 1. 数据库迁移

新增 `invitations` 表字段：

```sql
ALTER TABLE invitations ADD COLUMN channel VARCHAR(20) NOT NULL DEFAULT 'wechat' COMMENT '推送渠道: wechat=微信小程序, email=邮箱';
ALTER TABLE invitations ADD COLUMN phone VARCHAR(20) COMMENT '手机号(用于微信渠道)';
ALTER TABLE invitations ADD COLUMN email_template_id BIGINT COMMENT '邮箱模板ID(用于邮箱渠道)';
```

## 2. 后端 DTO（invitation_dto.go）

`CreateInvitationRequest` 新增字段：
- `channel`（必填，wechat | email）
- `position`（微信渠道必填）
- `phone`（微信渠道必填）
- `email_template_id`（邮箱渠道必填）

`InvitationListItem` 新增字段：
- `name`
- `phone`
- `channel`

`CreateInvitationResponse` 新增字段：
- `channel`

## 3. 后端 Model（invitation_model.go）

`Invitation` struct 新增字段：
- `Channel`（string, default: wechat）
- `Phone`（string）
- `EmailTemplateID`（int64）

## 4. 后端 Service（invitation_service.go）

`CreateInvitation` 方法：
- 根据 `channel` 校验必填字段
- 微信渠道：存储手机号+岗位
- 邮箱渠道：存储模板ID

`ListInvitations` 方法：
- 返回 `name`、`phone`、`channel` 字段

## 5. 前端表单（InvitationList.vue）

弹窗表单新增字段：
- 推送渠道选择（el-radio-group: 微信小程序 / 邮箱）
- 微信小程序渠道：姓名、手机号、岗位
- 邮箱渠道：姓名、邮箱模板下拉

## 6. 邮箱模板（EmailTemplate）

确认是否已有邮箱模板相关接口和表。如无，需新建。
