ALTER TABLE groups
ADD COLUMN IF NOT EXISTS model_pricing JSONB DEFAULT '{}';

COMMENT ON COLUMN groups.model_pricing IS '按模型定价配置: {"model_name": {"sell_input_price": X, "sell_output_price": X, "cost_input_price": X, "cost_output_price": X}} 价格单位: USD/M tokens';
