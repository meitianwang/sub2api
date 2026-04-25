-- 092_drop_subscription_tables.sql
-- 移除"用户订阅"功能：删除 user_subscriptions / subscription_plans 表，
-- 以及相关引用列（payment_orders 的 order_type/plan_id/subscription_group_id/subscription_days、
-- usage_logs.subscription_id、billing_usage_entries.subscription_id、groups.subscription_type）。

-- ============================================================
-- 1. 移除引用了 user_subscriptions(id) 的外键列
-- ============================================================
ALTER TABLE usage_logs DROP COLUMN IF EXISTS subscription_id;
ALTER TABLE billing_usage_entries DROP COLUMN IF EXISTS subscription_id;

-- ============================================================
-- 2. 移除 payment_orders 中所有订阅相关列
--    （plan_id 引用 subscription_plans，必须先删才能 DROP TABLE）
-- ============================================================
ALTER TABLE payment_orders DROP COLUMN IF EXISTS plan_id;
ALTER TABLE payment_orders DROP COLUMN IF EXISTS subscription_group_id;
ALTER TABLE payment_orders DROP COLUMN IF EXISTS subscription_days;
ALTER TABLE payment_orders DROP COLUMN IF EXISTS order_type;
DROP INDEX IF EXISTS idx_payment_orders_order_type;

-- ============================================================
-- 3. 移除 groups.subscription_type（订阅/标准分类已无意义）
-- ============================================================
DROP INDEX IF EXISTS idx_groups_subscription_type;
ALTER TABLE groups DROP COLUMN IF EXISTS subscription_type;

-- ============================================================
-- 4. 删除两张订阅业务表
--    CASCADE 会自动清理任何残留的依赖（索引、视图等）
-- ============================================================
DROP TABLE IF EXISTS user_subscriptions CASCADE;
DROP TABLE IF EXISTS subscription_plans CASCADE;
