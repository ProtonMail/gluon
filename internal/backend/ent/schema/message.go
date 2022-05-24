package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Message holds the schema definition for the Message entity.
type Message struct {
	ent.Schema
}

// Fields of the Message.
func (Message) Fields() []ent.Field {
	return []ent.Field{
		field.String("MessageID").Unique(),
		field.String("InternalID").Unique(),
		field.Time("Date"),
		field.Int("Size"),
		field.String("Body"),
		field.String("BodyStructure"),
		field.String("Envelope"),
	}
}

// Edges of the Message.
func (Message) Edges() []ent.Edge {
	return []ent.Edge{
		// A message has many flags.
		edge.To("flags", MessageFlag.Type).Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),

		// A message has many UIDs.
		edge.From("UIDs", UID.Type).Ref("message"),
	}
}
