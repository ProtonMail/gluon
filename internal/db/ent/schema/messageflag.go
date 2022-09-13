package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// MessageFlag holds the schema definition for the MessageFlag entity.
type MessageFlag struct {
	ent.Schema
}

// Fields of the Flag.
func (MessageFlag) Fields() []ent.Field {
	return []ent.Field{
		field.String("Value"),
	}
}

// Edges of the Flag.
func (MessageFlag) Edges() []ent.Edge {
	return []ent.Edge{edge.From("messages", Message.Type).Ref("flags").Unique()}
}
