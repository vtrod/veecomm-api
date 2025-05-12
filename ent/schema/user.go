package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"time"
)

// User define o schema da entidade Usuário
type User struct {
	ent.Schema
}

// Fields define os campos da entidade Usuário
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Unique().
			StorageKey("id").
			Immutable(),
		field.String("name").
			NotEmpty(),
		field.String("email").
			Unique().
			NotEmpty(),
		field.String("password").
			Sensitive().
			NotEmpty(),
		field.String("phone").
			Optional(),
		field.String("profile_image").
			Optional(),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges define as relações desta entidade com outras entidades
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("addresses", Address.Type),
		edge.To("orders", Order.Type),
		edge.To("avaliations", Avaliation.Type),
		edge.To("cart", Cart.Type).Unique(),
	}
} 