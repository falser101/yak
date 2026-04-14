-- 实名认证字段
ALTER TABLE users ADD COLUMN IF NOT EXISTS rz_status INTEGER DEFAULT 0;  -- 0:未认证, 1:认证中, 2:已认证
ALTER TABLE users ADD COLUMN IF NOT EXISTS rz_real_name VARCHAR(100);     -- 真实姓名
ALTER TABLE users ADD COLUMN IF NOT EXISTS rz_id_card VARCHAR(20);        -- 证件号
ALTER TABLE users ADD COLUMN IF NOT EXISTS rz_gender INTEGER;             -- 性别: 0:未知, 1:男, 2:女
ALTER TABLE users ADD COLUMN IF NOT EXISTS rz_emergency_name VARCHAR(100);  -- 紧急联系人姓名
ALTER TABLE users ADD COLUMN IF NOT EXISTS rz_emergency_phone VARCHAR(20); -- 紧急联系人电话
ALTER TABLE users ADD COLUMN IF NOT EXISTS rz_verified_at TIMESTAMP;      -- 认证时间
