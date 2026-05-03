-- =====================================================
-- upgrade6.sql: 用户模块扩展
-- =====================================================

-- 1. 给 users 表添加会员相关字段
ALTER TABLE users ADD COLUMN IF NOT EXISTS membership_level INTEGER DEFAULT 0 CHECK (membership_level >= 0 AND membership_level <= 3);
COMMENT ON COLUMN users.membership_level IS '会员等级: 0=普通会员, 1=银卡会员, 2=金卡会员, 3=钻石会员';

ALTER TABLE users ADD COLUMN IF NOT EXISTS total_rides INTEGER DEFAULT 0;
COMMENT ON COLUMN users.total_rides IS '累计骑行里程（公里）';

ALTER TABLE users ADD COLUMN IF NOT EXISTS status INTEGER DEFAULT 1 CHECK (status IN (1, 2));
COMMENT ON COLUMN users.status IS '用户状态: 1=正常, 2=禁用';

-- 2. 给 signups 表添加支付相关字段（为支付模块做准备）
ALTER TABLE signups ADD COLUMN IF NOT EXISTS pay_method INTEGER DEFAULT 0 CHECK (pay_method >= 0 AND pay_method <= 3);
COMMENT ON COLUMN signups.pay_method IS '支付方式: 0=未支付, 1=微信支付, 2=支付宝, 3=到店支付';

ALTER TABLE signups ADD COLUMN IF NOT EXISTS pay_time TIMESTAMP;
COMMENT ON COLUMN signups.pay_time IS '支付时间';

ALTER TABLE signups ADD COLUMN IF NOT EXISTS amount DECIMAL(10, 2) DEFAULT 0;
COMMENT ON COLUMN signups.amount IS '缴费金额';

ALTER TABLE signups ADD COLUMN IF NOT EXISTS addons TEXT DEFAULT '';
COMMENT ON COLUMN signups.addons IS '附加选项JSON数组';

-- 3. 更新测试用户的会员信息
UPDATE users SET membership_level = 2, total_rides = 520 WHERE openid = 'dev_test_user';
