package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/ProtonMail/gluon/imap"
)

// UID holds the schema definition for the UID entity.
type UID struct {
	ent.Schema
}

// Fields of the UID.
func (UID) Fields() []ent.Field {
	return []ent.Field{
		field.Uint32("UID").GoType(imap.UID(0)),
		field.Bool("Deleted").Default(false),
		field.Bool("Recent").Default(true),
	}
}

// Edges of the UID.
func (UID) Edges() []ent.Edge {
	return []ent.Edge{
		// Apply UID has a single message.
		edge.To("message", Message.Type).Unique(),

		// Apply UID is in a single mailbox.
		edge.From("mailbox", Mailbox.Type).Ref("UIDs").Unique(),
	}
}
