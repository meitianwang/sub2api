-- 090_add_payment_setting_keys.sql
-- 插入支付系统相关的默认配置项

INSERT INTO settings (key, value, updated_at) VALUES
    ('pay_order_timeout_minutes',    '5',                       NOW()),
    ('pay_min_recharge_amount',      '1',                       NOW()),
    ('pay_max_recharge_amount',      '1000',                    NOW()),
    ('pay_max_daily_recharge_amount','10000',                   NOW()),
    ('pay_product_name',             'Sub2API Balance Recharge', NOW()),
    ('pay_providers',                '',                        NOW()),
    ('pay_help_image_url',           '',                        NOW()),
    ('pay_help_text',                '',                        NOW()),
    ('pay_max_daily_amount_alipay',  '0',                       NOW()),
    ('pay_max_daily_amount_wxpay',   '0',                       NOW()),
    ('pay_max_daily_amount_stripe',  '0',                       NOW())
ON CONFLICT (key) DO NOTHING;
