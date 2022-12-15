package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/ProtonMail/gluon/imap"
)

// DeletedSubscription holds the schema definition for the deleted mailbox which may still be subscribed.
type DeletedSubscription struct {
	ent.Schema
}

// Fields of the Mailbox.
func (DeletedSubscription) Fields() []ent.Field {
	return []ent.Field{
		field.String("Name").Unique(),
		field.String("RemoteID").Unique().GoType(imap.MailboxID("")),
	}
}

func (DeletedSubscription) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("RemoteID"),
		index.Fields("Name"),
	}
}
