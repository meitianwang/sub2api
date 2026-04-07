ALTER TABLE usage_logs
ADD COLUMN IF NOT EXISTS upstream_cost DECIMAL(20,10) DEFAULT 0;

COMMENT ON COLUMN usage_logs.upstream_cost IS '上游成本（平台支付给上游的费用）单位: USD';
