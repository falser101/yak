-- 修复 signups 表的 user_id 类型
ALTER TABLE signups ALTER COLUMN user_id TYPE INTEGER USING user_id::INTEGER;
