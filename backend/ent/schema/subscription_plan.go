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

// SubscriptionPlan holds the schema definition for the SubscriptionPlan entity.
//
// 订阅套餐定义（价格、有效期、关联分组），用户购买后创建 UserSubscription 实例。
// 同一个 Group 可以有多个 Plan（不同价位/时长）。
type SubscriptionPlan struct {
	ent.Schema
}

func (SubscriptionPlan) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "subscription_plans"},
	}
}

func (SubscriptionPlan) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.TimeMixin{},
	}
}

func (SubscriptionPlan) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("group_id").
			Optional().
			Nillable(),
		field.String("name").
			MaxLen(100).
			NotEmpty(),
		field.String("description").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.Other("price", decimal.Decimal{}).
			SchemaType(map[string]string{dialect.Postgres: "decimal(10,2)"}),
		field.Other("original_price", decimal.Decimal{}).
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "decimal(10,2)"}),
		field.Int("validity_days").
			Default(30),
		field.String("validity_unit").
			MaxLen(10).
			Default("day"),
		field.String("features").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.String("product_name").
			MaxLen(255).
			Optional().
			Nillable(),
		field.Bool("for_sale").
			Default(false),
		field.Int("sort_order").
			Default(0),
	}
}

func (SubscriptionPlan) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("group", Group.Type).
			Ref("subscription_plans").
			Field("group_id").
			Unique(),
		edge.To("orders", PaymentOrder.Type),
	}
}

func (SubscriptionPlan) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("for_sale", "sort_order"),
	}
}
