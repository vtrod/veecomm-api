package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"time"
)

// Order define o schema da entidade Pedido
type Order struct {
	ent.Schema
}

// Fields define os campos da entidade Pedido
func (Order) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Unique().
			StorageKey("id").
			Immutable(),
		field.String("user_id").
			Optional(),
		field.Time("date").
			Default(time.Now),
		field.Float("total").
			Positive(),
		field.Float("shipping").
			Default(0),
		field.Float("discount").
			Default(0),
		field.Enum("delivery_type").
			Values("pickup", "delivery"),
		field.Enum("status").
			Values(
				"pending", 
				"processing", 
				"shipped", 
				"delivered", 
				"cancelled",
			),
		field.String("address_id").
			Optional(),
		field.String("payment_method").
			NotEmpty(),
		field.String("payment_status").
			NotEmpty(),
		field.String("coupon_code").
			Optional(),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges define as relações desta entidade com outras entidades
func (Order) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("orders").
			Field("user_id").
			Unique(),
		edge.From("address", Address.Type).
			Ref("orders").
			Field("address_id").
			Unique(),
		edge.To("order_items", OrderItem.Type),
	}
} 