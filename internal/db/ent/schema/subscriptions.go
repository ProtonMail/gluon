package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/ProtonMail/gluon/imap"
)

// Subscription holds the schema definition for the Subscription entity.
type Subscription struct {
	ent.Schema
}

// Fields of the Mailbox.
func (Subscription) Fields() []ent.Field {
	return []ent.Field{
		field.String("Name").Unique(),
		field.Uint64("MailboxID").GoType(imap.InternalMailboxID(0)).Unique(),
		field.String("RemoteID").Optional().Unique().GoType(imap.MailboxID("")),
	}
}

func (Subscription) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("MailboxID"),
		index.Fields("RemoteID"),
		index.Fields("Name"),
	}
}
