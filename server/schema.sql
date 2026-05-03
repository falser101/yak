-- =====================================================
-- 骑行平台数据库 Schema
-- 合并所有升级脚本，发版时重建数据库使用
-- =====================================================

-- =====================================================
-- init.sql: 基础表结构
-- =====================================================

-- 用户表（微信登录）
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    openid VARCHAR(100) UNIQUE NOT NULL,
    nickname VARCHAR(100),
    avatar VARCHAR(500),
    phone VARCHAR(20),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 活动表
CREATE TABLE IF NOT EXISTS activities (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    cover VARCHAR(500) DEFAULT '',
    date TIMESTAMP NOT NULL,
    location VARCHAR(255) NOT NULL,
    max_participants INTEGER DEFAULT 50,
    price DECIMAL(10, 2) DEFAULT 0,
    description TEXT,
    status INTEGER DEFAULT 0 CHECK (status IN (0, 1, 2, 3)),  -- 0=草稿, 1=报名中, 2=进行中, 3=已结束
    signup_end_time TIMESTAMP,
    created_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 报名表
CREATE TABLE IF NOT EXISTS signups (
    id SERIAL PRIMARY KEY,
    activity_id INTEGER REFERENCES activities(id),
    user_id INTEGER REFERENCES users(id),
    name VARCHAR(100) NOT NULL,
    phone VARCHAR(20) NOT NULL,
    emergency_contact VARCHAR(100),
    emergency_phone VARCHAR(20),
    remark TEXT,
    id_number VARCHAR(20) DEFAULT '',
    status INTEGER DEFAULT 1 CHECK (status IN (1, 2, 3)),  -- 1=已报名, 2=已取消, 3=已完成
    pay_method INTEGER DEFAULT 0 CHECK (pay_method IN (0, 1, 2, 3)),  -- 0=未支付, 1=微信, 2=支付宝, 3=到店
    pay_time TIMESTAMP,
    amount DECIMAL(10, 2) DEFAULT 0,
    addons TEXT DEFAULT '',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(activity_id, user_id)
);

-- =====================================================
-- upgrade3.sql: 自行车/品牌功能
-- =====================================================

-- 品牌表
CREATE TABLE IF NOT EXISTS brands (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    logo VARCHAR(500) DEFAULT '',
    description TEXT DEFAULT '',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 车型表
CREATE TABLE IF NOT EXISTS brand_models (
    id SERIAL PRIMARY KEY,
    brand_id INTEGER REFERENCES brands(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    price DECIMAL(10, 2) DEFAULT 0,
    cover VARCHAR(500) DEFAULT '',
    bike_type VARCHAR(50) DEFAULT '',
    deposit DECIMAL(10, 2) DEFAULT 0,
    specs JSONB DEFAULT '{}',
    notes TEXT DEFAULT '',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 用户自行车表
CREATE TABLE IF NOT EXISTS bikes (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    brand_id INTEGER REFERENCES brands(id) ON DELETE SET NULL,
    model_id INTEGER REFERENCES brand_models(id) ON DELETE SET NULL,
    name VARCHAR(100) NOT NULL,
    cover VARCHAR(500) DEFAULT '',
    bike_type VARCHAR(50) DEFAULT '',
    purchase_date DATE,
    cost DECIMAL(10, 2) DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- =====================================================
-- upgrade7.sql: 租车装备模块 + 支付模块
-- =====================================================

-- 租车车辆表
CREATE TABLE IF NOT EXISTS rental_bikes (
    id SERIAL PRIMARY KEY,
    brand_id INTEGER REFERENCES brands(id) ON DELETE SET NULL,
    model_id INTEGER REFERENCES brand_models(id) ON DELETE SET NULL,
    name VARCHAR(100) NOT NULL,
    cover VARCHAR(500) DEFAULT '',
    bike_type VARCHAR(50) DEFAULT '',
    tag VARCHAR(20) DEFAULT '' CHECK (tag IN ('热门', '推荐', '新品', '')),
    price_day DECIMAL(10, 2) DEFAULT 0,
    price_hour DECIMAL(10, 2) DEFAULT 0,
    price_team DECIMAL(10, 2) DEFAULT 0,
    price_distance DECIMAL(10, 2) DEFAULT 0,
    deposit DECIMAL(10, 2) DEFAULT 0,
    specs JSONB DEFAULT '{}',
    notes TEXT DEFAULT '',
    status INTEGER DEFAULT 1 CHECK (status IN (1, 2)),  -- 1=上架, 2=下架
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 租车订单表
CREATE TABLE IF NOT EXISTS rental_orders (
    id SERIAL PRIMARY KEY,
    order_no VARCHAR(32) UNIQUE NOT NULL,
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    bike_id INTEGER REFERENCES rental_bikes(id) ON DELETE SET NULL,
    bike_name VARCHAR(100) DEFAULT '',
    bike_cover VARCHAR(500) DEFAULT '',
    bike_color VARCHAR(50) DEFAULT '',
    package VARCHAR(20) DEFAULT 'day' CHECK (package IN ('day', 'hour', 'team', 'distance')),
    quantity INTEGER DEFAULT 1,
    rental_date DATE,
    amount DECIMAL(10, 2) DEFAULT 0,
    deposit DECIMAL(10, 2) DEFAULT 0,
    status INTEGER DEFAULT 0 CHECK (status IN (0, 1, 2, 3, 4)),  -- 0=待支付, 1=已支付, 2=已取消, 3=已完成, 4=已退款
    contact_name VARCHAR(50) DEFAULT '',
    contact_phone VARCHAR(20) DEFAULT '',
    remark TEXT DEFAULT '',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    pay_time TIMESTAMP,
    pay_method INTEGER DEFAULT 0 CHECK (pay_method IN (0, 1, 2, 3))
);

-- 支付记录表
CREATE TABLE IF NOT EXISTS payment_records (
    id SERIAL PRIMARY KEY,
    order_type VARCHAR(20) DEFAULT '' CHECK (order_type IN ('signup', 'rental', '')),
    order_id INTEGER DEFAULT 0,
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    amount DECIMAL(10, 2) DEFAULT 0,
    pay_method INTEGER DEFAULT 0 CHECK (pay_method IN (0, 1, 2, 3)),
    pay_time TIMESTAMP,
    transaction_id VARCHAR(64) DEFAULT '',
    status INTEGER DEFAULT 0 CHECK (status IN (0, 1, 2, 3)),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- =====================================================
-- 所有索引
-- =====================================================
CREATE INDEX IF NOT EXISTS idx_signups_activity ON signups(activity_id);
CREATE INDEX IF NOT EXISTS idx_signups_user ON signups(user_id);
CREATE INDEX IF NOT EXISTS idx_activities_status ON activities(status, date);
CREATE INDEX IF NOT EXISTS idx_brand_models_brand ON brand_models(brand_id);
CREATE INDEX IF NOT EXISTS idx_bikes_user ON bikes(user_id);
CREATE INDEX IF NOT EXISTS idx_bikes_brand ON bikes(brand_id);
CREATE INDEX IF NOT EXISTS idx_rental_bikes_brand ON rental_bikes(brand_id);
CREATE INDEX IF NOT EXISTS idx_rental_bikes_type ON rental_bikes(bike_type);
CREATE INDEX IF NOT EXISTS idx_rental_bikes_status ON rental_bikes(status);
CREATE INDEX IF NOT EXISTS idx_rental_orders_user ON rental_orders(user_id);
CREATE INDEX IF NOT EXISTS idx_rental_orders_status ON rental_orders(status);
CREATE INDEX IF NOT EXISTS idx_rental_orders_no ON rental_orders(order_no);

