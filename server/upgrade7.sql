-- =====================================================
-- upgrade7.sql: 租车装备模块 + 支付模块
-- =====================================================

-- 1. 给 brand_models 添加租车相关字段
ALTER TABLE brand_models ADD COLUMN IF NOT EXISTS deposit DECIMAL(10, 2) DEFAULT 0;
COMMENT ON COLUMN brand_models.deposit IS '押金（元）';

ALTER TABLE brand_models ADD COLUMN IF NOT EXISTS specs JSONB DEFAULT '{}';
COMMENT ON COLUMN brand_models.specs IS '规格参数 JSON: frame/derailleur/brake/wheel/height';

ALTER TABLE brand_models ADD COLUMN IF NOT EXISTS notes TEXT DEFAULT '';
COMMENT ON COLUMN brand_models.notes IS '租车说明/注意事项';

-- 价格体系：day/hour/team/distance (覆盖原来的 price 字段语义)
-- 注意：price 字段保留作为默认日租价格

-- 2. 租车车辆表（平台提供的可租赁车辆）
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
    status INTEGER DEFAULT 1 CHECK (status IN (1, 2)),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_rental_bikes_brand ON rental_bikes(brand_id);
CREATE INDEX IF NOT EXISTS idx_rental_bikes_type ON rental_bikes(bike_type);
CREATE INDEX IF NOT EXISTS idx_rental_bikes_status ON rental_bikes(status);

-- 3. 租车订单表
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
    status INTEGER DEFAULT 0 CHECK (status IN (0, 1, 2, 3, 4)),
    contact_name VARCHAR(50) DEFAULT '',
    contact_phone VARCHAR(20) DEFAULT '',
    remark TEXT DEFAULT '',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    pay_time TIMESTAMP,
    pay_method INTEGER DEFAULT 0 CHECK (pay_method IN (0, 1, 2, 3))
);

CREATE INDEX IF NOT EXISTS idx_rental_orders_user ON rental_orders(user_id);
CREATE INDEX IF NOT EXISTS idx_rental_orders_status ON rental_orders(status);
CREATE INDEX IF NOT EXISTS idx_rental_orders_no ON rental_orders(order_no);

COMMENT ON COLUMN rental_orders.status IS '0=待支付, 1=已支付, 2=已取消, 3=已完成, 4=已退款';
COMMENT ON COLUMN rental_orders.pay_method IS '0=未选择, 1=微信支付, 2=支付宝, 3=到店支付';

-- 4. 支付记录表
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

COMMENT ON COLUMN payment_records.status IS '0=待支付, 1=已支付, 2=已取消, 3=已退款';

-- 5. 插入测试租车数据
INSERT INTO rental_bikes (brand_id, model_id, name, cover, bike_type, tag, price_day, price_hour, price_team, deposit, specs, notes, status) VALUES
(1, 1, 'TCR Advanced SL 公路车', '', '公路车', '热门', 299.00, 50.00, 800.00, 2000.00,
 '{"frame":"M/L","derailleur":"Shimano Dura-Ace","brake":"碟刹","wheel":"700C","height":"170-185cm"}',
 '含头盔、手套。押金2000元，还车退还。', 1),
(1, 3, 'Defy Advanced 平把公路', '', '平把公路', '推荐', 199.00, 35.00, 600.00, 1500.00,
 '{"frame":"S/M","derailleur":"Shimano 105","brake":"圈刹","wheel":"700C","height":"160-175cm"}',
 '适合长途骑行，含车灯。押金1500元，还车退还。', 1),
(2, 5, 'ATX830 山地车', '', '山地车', '热门', 159.00, 30.00, 500.00, 1000.00,
 '{"frame":"26寸","derailleur":"Shimano Deore","brake":"碟刹","wheel":"26寸","height":"165-180cm"}',
 '适合山地路况，含护具。押金1000元，还车退还。', 1),
(6, 9, 'Allez 公路车', '', '公路车', '新品', 249.00, 45.00, 700.00, 1800.00,
 '{"frame":"52/54/56","derailleur":"Shimano Tiagra","brake":"碟刹","wheel":"700C","height":"168-183cm"}',
 '入门级碳纤维公路车，舒适度高。押金1800元，还车退还。', 1),
(8, 7, 'Triban RC500 公路车', '', '公路车', '推荐', 99.00, 20.00, 300.00, 800.00,
 '{"frame":"XS/S/M/L","derailleur":"Shimano Claris","brake":"圈刹","wheel":"700C","height":"160-180cm"}',
 '性价比首选，适合新手。押金800元，还车退还。', 1)
ON CONFLICT DO NOTHING;
