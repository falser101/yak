-- 自行车/品牌功能

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

-- 索引
CREATE INDEX IF NOT EXISTS idx_brand_models_brand ON brand_models(brand_id);
CREATE INDEX IF NOT EXISTS idx_bikes_user ON bikes(user_id);
CREATE INDEX IF NOT EXISTS idx_bikes_brand ON bikes(brand_id);

-- 预置品牌数据
INSERT INTO brands (name, logo, description) VALUES
('Giant', 'G', 'GIANT 是全球自行车产业的领导者，致力于为骑行者提供高品质、高性能的产品。GIANT 结合先进的技术与创新的设计理念，为各类骑行者提供最佳的骑行体验。'),
('捷安特', '捷', '捷安特是台湾巨大集团旗下的自行车品牌，以卓越的品质和创新的技术享誉全球。作为入门级到专业级的全面选择，捷安特为每一位骑行者提供最适合的产品。'),
('UCC', 'U', 'UCC 是来自美国的知名自行车品牌，专注于高性能公路车和山地车的研发与制造。'),
('Merida', 'M', 'Merida（美利达）是台湾自行车品牌，以其卓越的工艺和性价比著称，产品涵盖公路车、山地车等多个领域。'),
('Trek', '崔', 'Trek（崔克）是美国顶级自行车品牌，以创新技术和专业品质闻名，是众多职业车队和业余爱好者的首选。'),
('Specialized', '闪', 'Specialized（闪电）是美国顶级自行车品牌，以先进的气动技术和人体工学设计著称，为骑行者提供极致的性能体验。'),
('喜德盛', '喜', '喜德盛是中国本土知名品牌，以性价比高的产品赢得广大骑行者的喜爱，产品线丰富覆盖各类车型。'),
('Decathlon', '迪', 'Decathlon（迪卡侬）提供从入门到专业的全系列自行车，以亲民的价格和可靠的品质著称，是很多新手骑友的第一选择。'),
('Cannondale', '佳', 'Cannondale（佳能戴尔）是美国知名自行车品牌，以创新的铝合金技术和独特的左右不对称设计闻名。'),
('瑞豹', '瑞', '瑞豹（Pardus）是国内专业自行车品牌，专注于碳纤维公路车的研发与生产，是多项国际赛事的官方供应商。')
ON CONFLICT DO NOTHING;

-- 预置车型数据
INSERT INTO brand_models (brand_id, name, price, bike_type) VALUES
-- Giant (id=1)
(1, 'TCR Advanced SL', 29999.00, '公路'),
(1, 'Propel Advanced', 25999.00, '公路'),
(1, 'Defy Advanced', 18999.00, '公路'),
(1, 'Contend AR', 8999.00, '平把公路'),
-- 捷安特 (id=2)
(2, 'OCR5500', 4298.00, '公路'),
(2, 'ATX830', 2998.00, '山地'),
(2, 'Escape 3', 1998.00, '平把公路'),
-- UCC (id=3)
(3, 'Ultimate', 22999.00, '公路'),
(3, 'Ares', 15999.00, '公路'),
-- Merida (id=4)
(4, 'Reacto', 27999.00, '公路'),
(4, 'Scultura', 19999.00, '公路'),
(4, 'Big.Nine', 8999.00, '山地'),
-- Trek (id=5)
(5, 'Madone', 39999.00, '公路'),
(5, 'Emonda', 29999.00, '公路'),
(5, 'Domane', 24999.00, '公路'),
-- Specialized (id=6)
(6, 'Venge', 45999.00, '公路'),
(6, 'Tarmac', 35999.00, '公路'),
(6, 'Allez', 12999.00, '公路'),
-- 喜德盛 (id=7)
(7, 'RS350', 3999.00, '公路'),
(7, '侠客600', 2999.00, '山地'),
-- Decathlon (id=8)
(8, 'RC100', 999.00, '平把公路'),
(8, 'Triban RC500', 3999.00, '公路'),
-- Cannondale (id=9)
(9, 'CAAD13', 18999.00, '公路'),
(9, 'SuperSix', 29999.00, '公路'),
-- 瑞豹 (id=10)
(10, 'Robin', 15999.00, '公路'),
(10, 'Sprint', 9999.00, '公路');
