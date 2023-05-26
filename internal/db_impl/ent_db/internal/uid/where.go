// Code generated by ent, DO NOT EDIT.

package uid

import (
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/db_impl/ent_db/internal/predicate"
)

// ID filters vertices based on their ID field.
func ID(id int) predicate.UID {
	return predicate.UID(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id int) predicate.UID {
	return predicate.UID(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id int) predicate.UID {
	return predicate.UID(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...int) predicate.UID {
	return predicate.UID(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...int) predicate.UID {
	return predicate.UID(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id int) predicate.UID {
	return predicate.UID(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id int) predicate.UID {
	return predicate.UID(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id int) predicate.UID {
	return predicate.UID(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id int) predicate.UID {
	return predicate.UID(sql.FieldLTE(FieldID, id))
}

// UID applies equality check predicate on the "UID" field. It's identical to UIDEQ.
func UID(v imap.UID) predicate.UID {
	vc := uint32(v)
	return predicate.UID(sql.FieldEQ(FieldUID, vc))
}

// Deleted applies equality check predicate on the "Deleted" field. It's identical to DeletedEQ.
func Deleted(v bool) predicate.UID {
	return predicate.UID(sql.FieldEQ(FieldDeleted, v))
}

// Recent applies equality check predicate on the "Recent" field. It's identical to RecentEQ.
func Recent(v bool) predicate.UID {
	return predicate.UID(sql.FieldEQ(FieldRecent, v))
}

// UIDEQ applies the EQ predicate on the "UID" field.
func UIDEQ(v imap.UID) predicate.UID {
	vc := uint32(v)
	return predicate.UID(sql.FieldEQ(FieldUID, vc))
}

// UIDNEQ applies the NEQ predicate on the "UID" field.
func UIDNEQ(v imap.UID) predicate.UID {
	vc := uint32(v)
	return predicate.UID(sql.FieldNEQ(FieldUID, vc))
}

// UIDIn applies the In predicate on the "UID" field.
func UIDIn(vs ...imap.UID) predicate.UID {
	v := make([]any, len(vs))
	for i := range v {
		v[i] = uint32(vs[i])
	}
	return predicate.UID(sql.FieldIn(FieldUID, v...))
}

// UIDNotIn applies the NotIn predicate on the "UID" field.
func UIDNotIn(vs ...imap.UID) predicate.UID {
	v := make([]any, len(vs))
	for i := range v {
		v[i] = uint32(vs[i])
	}
	return predicate.UID(sql.FieldNotIn(FieldUID, v...))
}

// UIDGT applies the GT predicate on the "UID" field.
func UIDGT(v imap.UID) predicate.UID {
	vc := uint32(v)
	return predicate.UID(sql.FieldGT(FieldUID, vc))
}

// UIDGTE applies the GTE predicate on the "UID" field.
func UIDGTE(v imap.UID) predicate.UID {
	vc := uint32(v)
	return predicate.UID(sql.FieldGTE(FieldUID, vc))
}

// UIDLT applies the LT predicate on the "UID" field.
func UIDLT(v imap.UID) predicate.UID {
	vc := uint32(v)
	return predicate.UID(sql.FieldLT(FieldUID, vc))
}

// UIDLTE applies the LTE predicate on the "UID" field.
func UIDLTE(v imap.UID) predicate.UID {
	vc := uint32(v)
	return predicate.UID(sql.FieldLTE(FieldUID, vc))
}

// DeletedEQ applies the EQ predicate on the "Deleted" field.
func DeletedEQ(v bool) predicate.UID {
	return predicate.UID(sql.FieldEQ(FieldDeleted, v))
}

// DeletedNEQ applies the NEQ predicate on the "Deleted" field.
func DeletedNEQ(v bool) predicate.UID {
	return predicate.UID(sql.FieldNEQ(FieldDeleted, v))
}

// RecentEQ applies the EQ predicate on the "Recent" field.
func RecentEQ(v bool) predicate.UID {
	return predicate.UID(sql.FieldEQ(FieldRecent, v))
}

// RecentNEQ applies the NEQ predicate on the "Recent" field.
func RecentNEQ(v bool) predicate.UID {
	return predicate.UID(sql.FieldNEQ(FieldRecent, v))
}

// HasMessage applies the HasEdge predicate on the "message" edge.
func HasMessage() predicate.UID {
	return predicate.UID(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, false, MessageTable, MessageColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasMessageWith applies the HasEdge predicate on the "message" edge with a given conditions (other predicates).
func HasMessageWith(preds ...predicate.Message) predicate.UID {
	return predicate.UID(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.To(MessageInverseTable, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, false, MessageTable, MessageColumn),
		)
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasMailbox applies the HasEdge predicate on the "mailbox" edge.
func HasMailbox() predicate.UID {
	return predicate.UID(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, true, MailboxTable, MailboxColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasMailboxWith applies the HasEdge predicate on the "mailbox" edge with a given conditions (other predicates).
func HasMailboxWith(preds ...predicate.Mailbox) predicate.UID {
	return predicate.UID(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.To(MailboxInverseTable, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, true, MailboxTable, MailboxColumn),
		)
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.UID) predicate.UID {
	return predicate.UID(func(s *sql.Selector) {
		s1 := s.Clone().SetP(nil)
		for _, p := range predicates {
			p(s1)
		}
		s.Where(s1.P())
	})
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.UID) predicate.UID {
	return predicate.UID(func(s *sql.Selector) {
		s1 := s.Clone().SetP(nil)
		for i, p := range predicates {
			if i > 0 {
				s1.Or()
			}
			p(s1)
		}
		s.Where(s1.P())
	})
}

// Not applies the not operator on the given predicate.
func Not(p predicate.UID) predicate.UID {
	return predicate.UID(func(s *sql.Selector) {
		p(s.Not())
	})
}