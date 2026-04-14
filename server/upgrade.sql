-- 为 activities 表添加报名截止时间字段
ALTER TABLE activities ADD COLUMN IF NOT EXISTS signup_end_time TIMESTAMP;
