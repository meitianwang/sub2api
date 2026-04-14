-- 091_remove_purchase_subscription_settings.sql
-- Remove deprecated external payment iframe settings (payment is now built-in)
DELETE FROM settings WHERE key IN ('purchase_subscription_enabled', 'purchase_subscription_url');
