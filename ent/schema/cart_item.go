package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"time"
)

// CartItem define o schema da entidade Item do Carrinho
type CartItem struct {
	ent.Schema
}

// Fields define os campos da entidade Item do Carrinho
func (CartItem) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Unique().
			StorageKey("id").
			Immutable(),
		field.String("cart_id").
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
func (CartItem) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("cart", Cart.Type).
			Ref("cart_items").
			Field("cart_id").
			Unique(),
		edge.From("product", Product.Type).
			Ref("cart_items").
			Field("product_id").
			Unique(),
	}
} 