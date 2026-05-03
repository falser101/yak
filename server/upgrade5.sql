-- =====================================================
-- upgrade5.sql: 活动模块扩展 + 报名字段补全
-- =====================================================

-- 1. 给 activities 表添加新字段
ALTER TABLE activities ADD COLUMN IF NOT EXISTS category VARCHAR(20) DEFAULT 'activity' CHECK (category IN ('activity', 'race', 'charity', 'club'));
COMMENT ON COLUMN activities.category IS '活动分类: activity=赛事, race=竞赛, charity=公益, club=俱乐部';

ALTER TABLE activities ADD COLUMN IF NOT EXISTS rules TEXT DEFAULT '';
COMMENT ON COLUMN activities.rules IS '活动规程/规则';

ALTER TABLE activities ADD COLUMN IF NOT EXISTS route TEXT DEFAULT '';
COMMENT ON COLUMN activities.route IS '路线轨迹信息';

ALTER TABLE activities ADD COLUMN IF NOT EXISTS awards TEXT DEFAULT '';
COMMENT ON COLUMN activities.awards IS '奖项设置';

-- 2. 给 signups 表添加身份证号字段
ALTER TABLE signups ADD COLUMN IF NOT EXISTS id_number VARCHAR(20) DEFAULT '';
COMMENT ON COLUMN signups.id_number IS '身份证号';

-- 3. 更新 activities status 语义（原来 0=报名中 1=已满 2=已结束，改为更清晰的语义）
-- 原来: 0=报名中 1=已满 2=已结束
-- 新语义: 0=草稿 1=报名中 2=进行中 3=已结束
-- 注意：这个改动需要数据迁移，先不改旧数据，只给新字段说明

-- 4. 添加测试数据（带分类的活动）
INSERT INTO activities (title, cover, date, location, max_participants, price, description, category, rules, route, awards, created_by, status) VALUES
('2026高原骑行挑战赛', '', '2026-06-01 08:00:00', '青海湖', 100, 299.00, '环青海湖高原骑行挑战，4天3夜', 'race', '1. 必须年满18周岁\n2. 需提供近3个月体检证明\n3. 必须佩戴头盔', '全程约360公里，沿青海湖环湖公路', '冠军: 5000元+奖杯\n亚军: 3000元\n季军: 1000元', 1, 1),
('为爱骑行·公益募捐', '', '2026-05-20 09:00:00', '拉萨市区', 200, 0.00, '为西藏山区儿童募集学习物资', 'charity', '1. 公益活动，免费参与\n2. 建议捐款捐物', '拉萨市区约30公里', '所有完成者获得公益证书', 1, 1),
('亲子骑行日', '', '2026-05-01 10:00:00', '东湖绿道', 50, 0.00, '亲子骑行活动，培养孩子骑行兴趣', 'club', '1. 儿童需家长陪同\n2. 必须佩戴护具', '东湖绿道约10公里', '参与奖状', 1, 1),
('周末俱乐部例骑', '', '2026-05-10 08:00:00', '江滩公园', 30, 0.00, '俱乐部每周例行骑行活动', 'club', '1. 俱乐部成员优先\n2. 必须服从领队安排', '江滩至东湖约25公里', '无', 1, 1)
ON CONFLICT DO NOTHING;

-- 5. 添加 signups 的测试数据（带上 id_number）
INSERT INTO signups (activity_id, user_id, name, phone, emergency_contact, emergency_phone, remark, id_number, status) VALUES
(1, 1, '张三', '13800138000', '张妻', '13800138001', '有高原骑行经验', '420106199001011234', 1),
(2, 1, '测试用户', '13800138000', '家人', '13800138001', '', '420106199001011234', 1)
ON CONFLICT (activity_id, user_id) DO NOTHING;
