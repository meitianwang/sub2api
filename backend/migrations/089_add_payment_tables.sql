-- 089_add_payment_tables.sql
-- 支付系统表：订单、审计日志、渠道、订阅套餐、支付服务商实例

-- ============================================================
-- 1. payment_provider_instances（先建，被 payment_orders 引用）
-- ============================================================
CREATE TABLE IF NOT EXISTS payment_provider_instances (
    id              BIGSERIAL PRIMARY KEY,
    provider_key    VARCHAR(30)  NOT NULL,
    name            VARCHAR(100) NOT NULL,
    config          TEXT         NOT NULL,
    supported_types VARCHAR(255) NOT NULL DEFAULT '',
    enabled         BOOLEAN      NOT NULL DEFAULT true,
    sort_order      INT          NOT NULL DEFAULT 0,
    limits          TEXT,
    refund_enabled  BOOLEAN      NOT NULL DEFAULT false,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payment_provider_instances_provider_key
    ON payment_provider_instances(provider_key);
CREATE INDEX IF NOT EXISTS idx_payment_provider_instances_provider_key_enabled
    ON payment_provider_instances(provider_key, enabled);

-- ============================================================
-- 2. subscription_plans（先建，被 payment_orders 引用）
-- ============================================================
CREATE TABLE IF NOT EXISTS subscription_plans (
    id              BIGSERIAL PRIMARY KEY,
    group_id        BIGINT REFERENCES groups(id) ON DELETE SET NULL,
    name            VARCHAR(100) NOT NULL,
    description     TEXT,
    price           DECIMAL(10,2) NOT NULL,
    original_price  DECIMAL(10,2),
    validity_days   INT          NOT NULL DEFAULT 30,
    validity_unit   VARCHAR(10)  NOT NULL DEFAULT 'day',
    features        TEXT,
    product_name    VARCHAR(255),
    for_sale        BOOLEAN      NOT NULL DEFAULT false,
    sort_order      INT          NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_subscription_plans_for_sale_sort_order
    ON subscription_plans(for_sale, sort_order);

-- ============================================================
-- 3. payment_orders
-- ============================================================
CREATE TABLE IF NOT EXISTS payment_orders (
    id                      BIGSERIAL PRIMARY KEY,
    user_id                 BIGINT       NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    user_email              VARCHAR(255),
    user_name               VARCHAR(100),
    user_notes              TEXT,
    amount                  DECIMAL(10,2) NOT NULL,
    pay_amount              DECIMAL(10,2),
    fee_rate                DECIMAL(5,4),
    recharge_code           VARCHAR(64)  NOT NULL,
    status                  VARCHAR(20)  NOT NULL DEFAULT 'pending',
    payment_type            VARCHAR(30)  NOT NULL,
    payment_trade_no        VARCHAR(255),
    pay_url                 TEXT,
    qr_code                 TEXT,
    qr_code_img             TEXT,
    refund_amount           DECIMAL(10,2),
    refund_reason           TEXT,
    refund_at               TIMESTAMPTZ,
    force_refund            BOOLEAN      NOT NULL DEFAULT false,
    refund_requested_at     TIMESTAMPTZ,
    refund_request_reason   TEXT,
    refund_requested_by     BIGINT,
    expires_at              TIMESTAMPTZ  NOT NULL,
    paid_at                 TIMESTAMPTZ,
    completed_at            TIMESTAMPTZ,
    failed_at               TIMESTAMPTZ,
    failed_reason           TEXT,
    client_ip               VARCHAR(45),
    src_host                VARCHAR(255),
    src_url                 TEXT,
    order_type              VARCHAR(20)  NOT NULL DEFAULT 'balance',
    plan_id                 BIGINT REFERENCES subscription_plans(id) ON DELETE SET NULL,
    subscription_group_id   BIGINT,
    subscription_days       INT,
    provider_instance_id    BIGINT REFERENCES payment_provider_instances(id) ON DELETE SET NULL,
    created_at              TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Unique constraint
CREATE UNIQUE INDEX IF NOT EXISTS idx_payment_orders_recharge_code
    ON payment_orders(recharge_code);

-- Query indexes
CREATE INDEX IF NOT EXISTS idx_payment_orders_user_id
    ON payment_orders(user_id);
CREATE INDEX IF NOT EXISTS idx_payment_orders_status
    ON payment_orders(status);
CREATE INDEX IF NOT EXISTS idx_payment_orders_expires_at
    ON payment_orders(expires_at);
CREATE INDEX IF NOT EXISTS idx_payment_orders_created_at
    ON payment_orders(created_at);
CREATE INDEX IF NOT EXISTS idx_payment_orders_paid_at
    ON payment_orders(paid_at);
CREATE INDEX IF NOT EXISTS idx_payment_orders_payment_type_paid_at
    ON payment_orders(payment_type, paid_at);
CREATE INDEX IF NOT EXISTS idx_payment_orders_order_type
    ON payment_orders(order_type);

-- ============================================================
-- 4. payment_audit_logs
-- ============================================================
CREATE TABLE IF NOT EXISTS payment_audit_logs (
    id          BIGSERIAL PRIMARY KEY,
    order_id    BIGINT       NOT NULL REFERENCES payment_orders(id) ON DELETE CASCADE,
    action      VARCHAR(50)  NOT NULL,
    detail      TEXT,
    operator    VARCHAR(100),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payment_audit_logs_order_id
    ON payment_audit_logs(order_id);
CREATE INDEX IF NOT EXISTS idx_payment_audit_logs_created_at
    ON payment_audit_logs(created_at);

-- ============================================================
-- 5. payment_channels
-- ============================================================
CREATE TABLE IF NOT EXISTS payment_channels (
    id              BIGSERIAL PRIMARY KEY,
    group_id        BIGINT UNIQUE REFERENCES groups(id) ON DELETE SET NULL,
    name            VARCHAR(100)  NOT NULL,
    platform        VARCHAR(50)   NOT NULL DEFAULT '',
    rate_multiplier DECIMAL(10,4) NOT NULL,
    description     TEXT,
    models          TEXT,
    features        TEXT,
    sort_order      INT           NOT NULL DEFAULT 0,
    enabled         BOOLEAN       NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payment_channels_sort_order
    ON payment_channels(sort_order);
