package schema

import (
	"github.com/Wei-Shaw/sub2api/ent/schema/mixins"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// PaymentAuditLog holds the schema definition for the PaymentAuditLog entity.
//
// 删除策略：不删除（append-only）
// 支付审计日志是不可修改的记录，只追加不删除。
type PaymentAuditLog struct {
	ent.Schema
}

func (PaymentAuditLog) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "payment_audit_logs"},
	}
}

func (PaymentAuditLog) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.TimeMixin{},
	}
}

func (PaymentAuditLog) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("order_id"),
		field.String("action").
			MaxLen(50).
			NotEmpty(),
		field.String("detail").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.String("operator").
			MaxLen(100).
			Optional().
			Nillable(),
	}
}

func (PaymentAuditLog) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("order", PaymentOrder.Type).
			Ref("audit_logs").
			Field("order_id").
			Unique().
			Required(),
	}
}

func (PaymentAuditLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("order_id"),
		index.Fields("created_at"),
	}
}
