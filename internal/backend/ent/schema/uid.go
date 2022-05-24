package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// UID holds the schema definition for the UID entity.
type UID struct {
	ent.Schema
}

// Fields of the UID.
func (UID) Fields() []ent.Field {
	return []ent.Field{
		field.Int("UID"),
		field.Bool("Deleted").Default(false),
		field.Bool("Recent").Default(true),
		field.Bool("InDeletionPool").Default(false),
	}
}

// Edges of the UID.
func (UID) Edges() []ent.Edge {
	return []ent.Edge{
		// A UID has a single message.
		edge.To("message", Message.Type).Unique(),

		// A UID is in a single mailbox.
		edge.From("mailbox", Mailbox.Type).Ref("UIDs").Unique(),
	}
}
