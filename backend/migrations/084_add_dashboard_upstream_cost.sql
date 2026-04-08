ALTER TABLE usage_dashboard_daily
ADD COLUMN IF NOT EXISTS upstream_cost DECIMAL(20,10) DEFAULT 0;

ALTER TABLE usage_dashboard_hourly
ADD COLUMN IF NOT EXISTS upstream_cost DECIMAL(20,10) DEFAULT 0;
