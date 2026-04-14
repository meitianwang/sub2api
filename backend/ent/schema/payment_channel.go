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
	"github.com/shopspring/decimal"
)

// PaymentChannel holds the schema definition for the PaymentChannel entity.
//
// 支付渠道展示配置，关联 Group 以展示对应分组的充值/订阅入口。
type PaymentChannel struct {
	ent.Schema
}

func (PaymentChannel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "payment_channels"},
	}
}

func (PaymentChannel) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.TimeMixin{},
	}
}

func (PaymentChannel) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("group_id").
			Optional().
			Nillable().
			Unique(),
		field.String("name").
			MaxLen(100).
			NotEmpty(),
		field.String("platform").
			MaxLen(50).
			Default("default"),
		field.Other("rate_multiplier", decimal.Decimal{}).
			SchemaType(map[string]string{dialect.Postgres: "decimal(10,4)"}),
		field.String("description").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.String("models").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.String("features").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.Int("sort_order").
			Default(0),
		field.Bool("enabled").
			Default(true),
	}
}

func (PaymentChannel) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("group", Group.Type).
			Ref("payment_channels").
			Field("group_id").
			Unique(),
	}
}

func (PaymentChannel) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("sort_order"),
	}
}
