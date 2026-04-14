-- 用户表（微信登录）
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    openid VARCHAR(100) UNIQUE NOT NULL,
    nickname VARCHAR(100),
    avatar VARCHAR(500),
    phone VARCHAR(20),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 活动表（移除 participants 字段，通过 COUNT signups 计算）
CREATE TABLE IF NOT EXISTS activities (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    cover VARCHAR(500) DEFAULT '',
    date TIMESTAMP NOT NULL,
    location VARCHAR(255) NOT NULL,
    max_participants INTEGER DEFAULT 50,
    price DECIMAL(10, 2) DEFAULT 0,
    description TEXT,
    status INTEGER DEFAULT 0,  -- 0:报名中 1:已满 2:已结束
    signup_end_time TIMESTAMP,  -- 报名截止时间
    created_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 报名表（包含详细信息）
CREATE TABLE IF NOT EXISTS signups (
    id SERIAL PRIMARY KEY,
    activity_id INTEGER REFERENCES activities(id),
    user_id INTEGER REFERENCES users(id),
    name VARCHAR(100) NOT NULL,
    phone VARCHAR(20) NOT NULL,
    emergency_contact VARCHAR(100),
    emergency_phone VARCHAR(20),
    remark TEXT,
    status INTEGER DEFAULT 1,  -- 1:已报名 2:已取消 3:已完成
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(activity_id, user_id)  -- 防止重复报名
);

-- 索引
CREATE INDEX IF NOT EXISTS idx_signups_activity ON signups(activity_id);
CREATE INDEX IF NOT EXISTS idx_signups_user ON signups(user_id);
CREATE INDEX IF NOT EXISTS idx_activities_status ON activities(status, date);

-- 插入测试用户
INSERT INTO users (openid, nickname, phone) VALUES
    ('dev_test_user', '测试用户', '13800138000');

-- 插入测试活动
INSERT INTO activities (title, cover, date, location, max_participants, price, description, created_by, status) VALUES
('周末环湖骑行 50KM', '', '2026-04-12 09:00:00', '东湖绿道', 50, 50.00, '休闲骑行活动，沿途欣赏湖光山色', 1, 0),
('山地车进阶技巧培训', '', '2026-04-13 14:00:00', '磨山山地车公园', 20, 100.00, '专业教练指导山地车技巧', 1, 0),
('夜骑江城 30KM 休闲骑', '', '2026-04-11 19:00:00', '江汉路集合', 50, 30.00, '夜间城市骑行', 1, 1);

-- 插入测试报名数据
INSERT INTO signups (activity_id, user_id, name, phone, status) VALUES
(1, 1, '张三', '13800138000', 1),
(1, 1, '李四', '13900139000', 1);
