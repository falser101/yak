-- 修复 users 表结构
DROP TABLE IF EXISTS users CASCADE;

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    openid VARCHAR(100) UNIQUE NOT NULL,
    nickname VARCHAR(100),
    avatar VARCHAR(500),
    phone VARCHAR(20),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 重建索引
CREATE INDEX IF NOT EXISTS idx_signups_activity ON signups(activity_id);
CREATE INDEX IF NOT EXISTS idx_signups_user ON signups(user_id);
CREATE INDEX IF NOT EXISTS idx_activities_status ON activities(status, date);

-- 插入测试用户
INSERT INTO users (openid, nickname, phone) VALUES
    ('dev_test_user', '测试用户', '13800138000');
