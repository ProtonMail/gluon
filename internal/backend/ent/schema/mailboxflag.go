package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// MailboxFlag holds the schema definition for the MailboxFlag entity.
type MailboxFlag struct {
	ent.Schema
}

// Fields of the MailboxFlag.
func (MailboxFlag) Fields() []ent.Field {
	return []ent.Field{
		field.String("Value"),
	}
}

// Edges of the MailboxFlag.
func (MailboxFlag) Edges() []ent.Edge {
	return nil
}
