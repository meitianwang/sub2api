package schema

import (
	"github.com/Wei-Shaw/sub2api/ent/schema/mixins"
	"github.com/Wei-Shaw/sub2api/internal/domain"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/shopspring/decimal"
)

// PaymentOrder holds the schema definition for the PaymentOrder entity.
//
// 删除策略：硬删除（实际不删除，通过 status 管理生命周期）
// 支付订单通过 status 字段追踪完整生命周期，不需要软删除。
type PaymentOrder struct {
	ent.Schema
}

func (PaymentOrder) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "payment_orders"},
	}
}

func (PaymentOrder) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.TimeMixin{},
	}
}

func (PaymentOrder) Fields() []ent.Field {
	return []ent.Field{
		// 用户信息
		field.Int64("user_id"),
		field.String("user_email").
			MaxLen(255).
			Optional().
			Nillable(),
		field.String("user_name").
			MaxLen(100).
			Optional().
			Nillable(),
		field.String("user_notes").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),

		// 金额（使用 shopspring/decimal 避免浮点精度问题）
		field.Other("amount", decimal.Decimal{}).
			SchemaType(map[string]string{dialect.Postgres: "decimal(10,2)"}),
		field.Other("pay_amount", decimal.Decimal{}).
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "decimal(10,2)"}),
		field.Other("fee_rate", decimal.Decimal{}).
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "decimal(5,4)"}),

		// 订单标识
		field.String("recharge_code").
			MaxLen(64).
			NotEmpty().
			Unique(),
		field.String("status").
			MaxLen(20).
			Default(domain.PaymentOrderStatusPending),
		field.String("payment_type").
			MaxLen(30).
			NotEmpty(),
		field.String("payment_trade_no").
			MaxLen(255).
			Optional().
			Nillable(),

		// 支付链接
		field.String("pay_url").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.String("qr_code").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.String("qr_code_img").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),

		// 退款信息
		field.Other("refund_amount", decimal.Decimal{}).
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "decimal(10,2)"}),
		field.String("refund_reason").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.Time("refund_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Bool("force_refund").
			Default(false),
		field.Time("refund_requested_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.String("refund_request_reason").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.Int64("refund_requested_by").
			Optional().
			Nillable(),

		// 时间线
		field.Time("expires_at").
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("paid_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("completed_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("failed_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.String("failed_reason").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),

		// 来源追踪
		field.String("client_ip").
			MaxLen(45).
			Optional().
			Nillable(),
		field.String("src_host").
			MaxLen(255).
			Optional().
			Nillable(),
		field.String("src_url").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),

		// 订单类型与订阅关联
		field.String("order_type").
			MaxLen(20).
			Default(domain.PaymentOrderTypeBalance),
		field.Int64("plan_id").
			Optional().
			Nillable(),
		field.Int64("subscription_group_id").
			Optional().
			Nillable(),
		field.Int("subscription_days").
			Optional().
			Nillable(),

		// 支付服务商实例
		field.Int64("provider_instance_id").
			Optional().
			Nillable(),
	}
}

func (PaymentOrder) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("payment_orders").
			Field("user_id").
			Unique().
			Required(),
		edge.From("plan", SubscriptionPlan.Type).
			Ref("orders").
			Field("plan_id").
			Unique(),
		edge.From("provider_instance", PaymentProviderInstance.Type).
			Ref("orders").
			Field("provider_instance_id").
			Unique(),
		edge.To("audit_logs", PaymentAuditLog.Type),
	}
}

func (PaymentOrder) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("status"),
		index.Fields("expires_at"),
		index.Fields("created_at"),
		index.Fields("paid_at"),
		index.Fields("payment_type", "paid_at"),
		index.Fields("order_type"),
		index.Fields("provider_instance_id", "paid_at"),
	}
}
