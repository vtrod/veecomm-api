package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"time"
)

// Avaliation define o schema da entidade Avaliação
type Avaliation struct {
	ent.Schema
}

// Fields define os campos da entidade Avaliação
func (Avaliation) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Unique().
			StorageKey("id").
			Immutable(),
		field.String("product_id").
			Optional(),
		field.String("user_id").
			Optional(),
		field.String("user_name").
			NotEmpty(),
		field.Int("rating").
			Range(1, 5),
		field.Text("comment").
			NotEmpty(),
		field.Time("date").
			Default(time.Now),
		field.JSON("images", []string{}).
			Optional(),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges define as relações desta entidade com outras entidades
func (Avaliation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("product", Product.Type).
			Ref("avaliations").
			Field("product_id").
			Unique(),
		edge.From("user", User.Type).
			Ref("avaliations").
			Field("user_id").
			Unique(),
	}
} 