package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"time"
)

// Category define o schema da entidade Categoria
type Category struct {
	ent.Schema
}

// Fields define os campos da entidade Categoria
func (Category) Fields() []ent.Field {
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
		field.String("image").
			Optional(),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges define as relações desta entidade com outras entidades
func (Category) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("products", Product.Type),
	}
} 