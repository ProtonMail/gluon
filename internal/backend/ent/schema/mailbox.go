package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/ProtonMail/gluon/imap"
)

// Mailbox holds the schema definition for the Mailbox entity.
type Mailbox struct {
	ent.Schema
}

// Fields of the Mailbox.
func (Mailbox) Fields() []ent.Field {
	return []ent.Field{
		field.String("MailboxID").Unique().Immutable().GoType(imap.InternalMailboxID("")),
		field.String("RemoteID").Optional().Unique().GoType(imap.LabelID("")),
		field.String("Name").Unique(),
		field.Int("UIDNext").Default(1),
		field.Int("UIDValidity").Default(1),
		field.Bool("Subscribed").Default(true),
	}
}

// Edges of the Mailbox.
func (Mailbox) Edges() []ent.Edge {
	return []ent.Edge{
		// A mailbox has many UIDs.
		edge.To("UIDs", UID.Type).Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),

		// A mailbox has many flags.
		edge.To("flags", MailboxFlag.Type).Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),

		// A mailbox has many permanent flags.
		edge.To("permanent_flags", MailboxPermFlag.Type).Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),

		// A mailbox has many attributes.
		edge.To("attributes", MailboxAttr.Type).Annotations(entsql.Annotation{OnDelete: entsql.Cascade}),
	}
}
