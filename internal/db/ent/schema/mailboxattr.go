package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// MailboxAttr holds the schema definition for the MailboxAttr entity.
type MailboxAttr struct {
	ent.Schema
}

// Fields of the MailboxAttr.
func (MailboxAttr) Fields() []ent.Field {
	return []ent.Field{
		field.String("Value"),
	}
}

// Edges of the MailboxAttr.
func (MailboxAttr) Edges() []ent.Edge {
	return nil
}
