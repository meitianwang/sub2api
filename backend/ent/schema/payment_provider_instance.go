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

// PaymentProviderInstance holds the schema definition for the PaymentProviderInstance entity.
//
// 支付服务商多实例配置，支持同一渠道配置多个商户号并负载均衡。
// config 字段存储加密的 JSON 凭据，不同 provider_key 有不同结构。
type PaymentProviderInstance struct {
	ent.Schema
}

func (PaymentProviderInstance) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "payment_provider_instances"},
	}
}

func (PaymentProviderInstance) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.TimeMixin{},
	}
}

func (PaymentProviderInstance) Fields() []ent.Field {
	return []ent.Field{
		field.String("provider_key").
			MaxLen(30).
			NotEmpty(),
		field.String("name").
			MaxLen(100).
			NotEmpty(),
		field.String("config").
			NotEmpty().
			SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.String("supported_types").
			MaxLen(255).
			Default(""),
		field.Bool("enabled").
			Default(true),
		field.Int("sort_order").
			Default(0),
		field.String("limits").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.Bool("refund_enabled").
			Default(false),
	}
}

func (PaymentProviderInstance) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("orders", PaymentOrder.Type),
	}
}

func (PaymentProviderInstance) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("provider_key"),
		index.Fields("provider_key", "enabled"),
	}
}
