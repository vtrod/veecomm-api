package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"time"
)

// Product define o schema da entidade Produto
type Product struct {
	ent.Schema
}

// Fields define os campos da entidade Produto
func (Product) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Unique().
			StorageKey("id").
			Immutable(),
		field.String("name").
			NotEmpty(),
		field.String("slug").
			Unique().
			NotEmpty(),
		field.Text("description").
			NotEmpty(),
		field.Float("price").
			Positive(),
		field.Float("sale_price").
			Optional(),
		field.Bool("on_sale").
			Default(false),
		field.Int("stock").
			Default(0),
		field.String("sku").
			NotEmpty(),
		field.String("category_id").
			Optional(),
		field.JSON("images", []string{}).
			Default([]string{}),
		field.Float("rating").
			Default(0),
		field.Int("review_count").
			Default(0),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges define as relações desta entidade com outras entidades
func (Product) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("avaliations", Avaliation.Type),
		edge.From("category", Category.Type).
			Ref("products").
			Field("category_id").
			Unique(),
		edge.To("order_items", OrderItem.Type),
		edge.To("cart_items", CartItem.Type),
	}
} 