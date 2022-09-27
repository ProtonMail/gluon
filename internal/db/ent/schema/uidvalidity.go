package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/ProtonMail/gluon/imap"
)

// UIDValidity holds the current global UIDVALITY value for creating new
// mailbox.
type UIDValidity struct {
	ent.Schema
}

// Fields of the UIDValidity.
func (UIDValidity) Fields() []ent.Field {
	return []ent.Field{
		field.Uint32("UIDValidity").GoType(imap.UID(0)),
	}
}

// Edges of the User.
func (UIDValidity) Edges() []ent.Edge {
	return nil
}
