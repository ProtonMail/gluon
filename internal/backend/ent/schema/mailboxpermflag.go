package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// MailboxPermFlag holds the schema definition for the MailboxPermFlag entity.
type MailboxPermFlag struct {
	ent.Schema
}

// Fields of the MailboxPermFlag.
func (MailboxPermFlag) Fields() []ent.Field {
	return []ent.Field{
		field.String("Value"),
	}
}

// Edges of the MailboxPermFlag.
func (MailboxPermFlag) Edges() []ent.Edge {
	return nil
}
