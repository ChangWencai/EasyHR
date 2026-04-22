-- 添加邀请渠道字段，支持微信小程序和邮箱两种推送方式
-- 2026-04-22

ALTER TABLE invitations ADD COLUMN channel VARCHAR(20) NOT NULL DEFAULT 'wechat' COMMENT '推送渠道: wechat=微信小程序, email=邮箱';
ALTER TABLE invitations ADD COLUMN phone VARCHAR(20) COMMENT '手机号(用于微信渠道)';
ALTER TABLE invitations ADD COLUMN email_template_id BIGINT COMMENT '邮箱模板ID(用于邮箱渠道)';
