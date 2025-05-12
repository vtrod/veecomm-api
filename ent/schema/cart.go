package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"time"
)

// Cart define o schema da entidade Carrinho
type Cart struct {
	ent.Schema
}

// Fields define os campos da entidade Carrinho
func (Cart) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Unique().
			StorageKey("id").
			Immutable(),
		field.String("user_id").
			Optional(),
		field.Float("subtotal").
			Default(0),
		field.Float("shipping").
			Optional(),
		field.Float("discount").
			Default(0),
		field.Float("total").
			Default(0),
		field.Bool("applied_coupon").
			Default(false),
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
func (Cart) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("cart").
			Field("user_id").
			Unique(),
		edge.To("cart_items", CartItem.Type),
	}
} 