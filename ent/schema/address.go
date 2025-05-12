package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"time"
)

// Address define o schema da entidade Endereço
type Address struct {
	ent.Schema
}

// Fields define os campos da entidade Endereço
func (Address) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Unique().
			StorageKey("id").
			Immutable(),
		field.String("user_id").
			Optional(),
		field.String("cep").
			NotEmpty(),
		field.String("logradouro").
			NotEmpty(),
		field.String("numero").
			NotEmpty(),
		field.String("complemento").
			Optional(),
		field.String("bairro").
			NotEmpty(),
		field.String("cidade").
			NotEmpty(),
		field.String("estado").
			NotEmpty(),
		field.Bool("is_default").
			Default(false),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges define as relações desta entidade com outras entidades
func (Address) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("addresses").
			Field("user_id").
			Unique(),
		edge.To("orders", Order.Type),
	}
} 