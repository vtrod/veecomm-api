package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"time"
)

// Coupon define o schema da entidade Cupom
type Coupon struct {
	ent.Schema
}

// Fields define os campos da entidade Cupom
func (Coupon) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Unique().
			StorageKey("id").
			Immutable(),
		field.String("code").
			NotEmpty().
			Unique(),
		field.Enum("discount_type").
			Values("percentage", "fixed"),
		field.Float("discount_value").
			Positive(),
		field.Float("min_purchase").
			Default(0),
		field.Time("expires_at").
			Optional(),
		field.Bool("is_active").
			Default(true),
		field.Int("max_uses").
			Optional(),
		field.Int("times_used").
			Default(0),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges define as relações desta entidade com outras entidades
func (Coupon) Edges() []ent.Edge {
	return nil
} 