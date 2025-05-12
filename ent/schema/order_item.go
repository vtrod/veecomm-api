package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"time"
)

// OrderItem define o schema da entidade Item do Pedido
type OrderItem struct {
	ent.Schema
}

// Fields define os campos da entidade Item do Pedido
func (OrderItem) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Unique().
			StorageKey("id").
			Immutable(),
		field.String("order_id").
			Optional(),
		field.String("product_id").
			Optional(),
		field.String("name").
			NotEmpty(),
		field.Float("price").
			Positive(),
		field.String("image").
			NotEmpty(),
		field.Int("quantity").
			Positive(),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges define as relações desta entidade com outras entidades
func (OrderItem) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("order", Order.Type).
			Ref("order_items").
			Field("order_id").
			Unique(),
		edge.From("product", Product.Type).
			Ref("order_items").
			Field("product_id").
			Unique(),
	}
} 