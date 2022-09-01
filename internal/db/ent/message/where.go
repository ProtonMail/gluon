// Code generated by ent, DO NOT EDIT.

package message

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/db/ent/predicate"
)

// ID filters vertices based on their ID field.
func ID(id int) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldID), id))
	})
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id int) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldID), id))
	})
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id int) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.NEQ(s.C(FieldID), id))
	})
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...int) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		v := make([]interface{}, len(ids))
		for i := range v {
			v[i] = ids[i]
		}
		s.Where(sql.In(s.C(FieldID), v...))
	})
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...int) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		v := make([]interface{}, len(ids))
		for i := range v {
			v[i] = ids[i]
		}
		s.Where(sql.NotIn(s.C(FieldID), v...))
	})
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id int) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.GT(s.C(FieldID), id))
	})
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id int) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.GTE(s.C(FieldID), id))
	})
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id int) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.LT(s.C(FieldID), id))
	})
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id int) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.LTE(s.C(FieldID), id))
	})
}

// MessageID applies equality check predicate on the "MessageID" field. It's identical to MessageIDEQ.
func MessageID(v imap.InternalMessageID) predicate.Message {
	vc := string(v)
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldMessageID), vc))
	})
}

// RemoteID applies equality check predicate on the "RemoteID" field. It's identical to RemoteIDEQ.
func RemoteID(v imap.MessageID) predicate.Message {
	vc := string(v)
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldRemoteID), vc))
	})
}

// Date applies equality check predicate on the "Date" field. It's identical to DateEQ.
func Date(v time.Time) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldDate), v))
	})
}

// Size applies equality check predicate on the "Size" field. It's identical to SizeEQ.
func Size(v int) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldSize), v))
	})
}

// Body applies equality check predicate on the "Body" field. It's identical to BodyEQ.
func Body(v string) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldBody), v))
	})
}

// BodyStructure applies equality check predicate on the "BodyStructure" field. It's identical to BodyStructureEQ.
func BodyStructure(v string) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldBodyStructure), v))
	})
}

// Envelope applies equality check predicate on the "Envelope" field. It's identical to EnvelopeEQ.
func Envelope(v string) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldEnvelope), v))
	})
}

// Deleted applies equality check predicate on the "Deleted" field. It's identical to DeletedEQ.
func Deleted(v bool) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldDeleted), v))
	})
}

// MessageIDEQ applies the EQ predicate on the "MessageID" field.
func MessageIDEQ(v imap.InternalMessageID) predicate.Message {
	vc := string(v)
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldMessageID), vc))
	})
}

// MessageIDNEQ applies the NEQ predicate on the "MessageID" field.
func MessageIDNEQ(v imap.InternalMessageID) predicate.Message {
	vc := string(v)
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.NEQ(s.C(FieldMessageID), vc))
	})
}

// MessageIDIn applies the In predicate on the "MessageID" field.
func MessageIDIn(vs ...imap.InternalMessageID) predicate.Message {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = string(vs[i])
	}
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.In(s.C(FieldMessageID), v...))
	})
}

// MessageIDNotIn applies the NotIn predicate on the "MessageID" field.
func MessageIDNotIn(vs ...imap.InternalMessageID) predicate.Message {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = string(vs[i])
	}
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.NotIn(s.C(FieldMessageID), v...))
	})
}

// MessageIDGT applies the GT predicate on the "MessageID" field.
func MessageIDGT(v imap.InternalMessageID) predicate.Message {
	vc := string(v)
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.GT(s.C(FieldMessageID), vc))
	})
}

// MessageIDGTE applies the GTE predicate on the "MessageID" field.
func MessageIDGTE(v imap.InternalMessageID) predicate.Message {
	vc := string(v)
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.GTE(s.C(FieldMessageID), vc))
	})
}

// MessageIDLT applies the LT predicate on the "MessageID" field.
func MessageIDLT(v imap.InternalMessageID) predicate.Message {
	vc := string(v)
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.LT(s.C(FieldMessageID), vc))
	})
}

// MessageIDLTE applies the LTE predicate on the "MessageID" field.
func MessageIDLTE(v imap.InternalMessageID) predicate.Message {
	vc := string(v)
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.LTE(s.C(FieldMessageID), vc))
	})
}

// MessageIDContains applies the Contains predicate on the "MessageID" field.
func MessageIDContains(v imap.InternalMessageID) predicate.Message {
	vc := string(v)
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.Contains(s.C(FieldMessageID), vc))
	})
}

// MessageIDHasPrefix applies the HasPrefix predicate on the "MessageID" field.
func MessageIDHasPrefix(v imap.InternalMessageID) predicate.Message {
	vc := string(v)
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.HasPrefix(s.C(FieldMessageID), vc))
	})
}

// MessageIDHasSuffix applies the HasSuffix predicate on the "MessageID" field.
func MessageIDHasSuffix(v imap.InternalMessageID) predicate.Message {
	vc := string(v)
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.HasSuffix(s.C(FieldMessageID), vc))
	})
}

// MessageIDEqualFold applies the EqualFold predicate on the "MessageID" field.
func MessageIDEqualFold(v imap.InternalMessageID) predicate.Message {
	vc := string(v)
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.EqualFold(s.C(FieldMessageID), vc))
	})
}

// MessageIDContainsFold applies the ContainsFold predicate on the "MessageID" field.
func MessageIDContainsFold(v imap.InternalMessageID) predicate.Message {
	vc := string(v)
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.ContainsFold(s.C(FieldMessageID), vc))
	})
}

// RemoteIDEQ applies the EQ predicate on the "RemoteID" field.
func RemoteIDEQ(v imap.MessageID) predicate.Message {
	vc := string(v)
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldRemoteID), vc))
	})
}

// RemoteIDNEQ applies the NEQ predicate on the "RemoteID" field.
func RemoteIDNEQ(v imap.MessageID) predicate.Message {
	vc := string(v)
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.NEQ(s.C(FieldRemoteID), vc))
	})
}

// RemoteIDIn applies the In predicate on the "RemoteID" field.
func RemoteIDIn(vs ...imap.MessageID) predicate.Message {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = string(vs[i])
	}
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.In(s.C(FieldRemoteID), v...))
	})
}

// RemoteIDNotIn applies the NotIn predicate on the "RemoteID" field.
func RemoteIDNotIn(vs ...imap.MessageID) predicate.Message {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = string(vs[i])
	}
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.NotIn(s.C(FieldRemoteID), v...))
	})
}

// RemoteIDGT applies the GT predicate on the "RemoteID" field.
func RemoteIDGT(v imap.MessageID) predicate.Message {
	vc := string(v)
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.GT(s.C(FieldRemoteID), vc))
	})
}

// RemoteIDGTE applies the GTE predicate on the "RemoteID" field.
func RemoteIDGTE(v imap.MessageID) predicate.Message {
	vc := string(v)
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.GTE(s.C(FieldRemoteID), vc))
	})
}

// RemoteIDLT applies the LT predicate on the "RemoteID" field.
func RemoteIDLT(v imap.MessageID) predicate.Message {
	vc := string(v)
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.LT(s.C(FieldRemoteID), vc))
	})
}

// RemoteIDLTE applies the LTE predicate on the "RemoteID" field.
func RemoteIDLTE(v imap.MessageID) predicate.Message {
	vc := string(v)
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.LTE(s.C(FieldRemoteID), vc))
	})
}

// RemoteIDContains applies the Contains predicate on the "RemoteID" field.
func RemoteIDContains(v imap.MessageID) predicate.Message {
	vc := string(v)
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.Contains(s.C(FieldRemoteID), vc))
	})
}

// RemoteIDHasPrefix applies the HasPrefix predicate on the "RemoteID" field.
func RemoteIDHasPrefix(v imap.MessageID) predicate.Message {
	vc := string(v)
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.HasPrefix(s.C(FieldRemoteID), vc))
	})
}

// RemoteIDHasSuffix applies the HasSuffix predicate on the "RemoteID" field.
func RemoteIDHasSuffix(v imap.MessageID) predicate.Message {
	vc := string(v)
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.HasSuffix(s.C(FieldRemoteID), vc))
	})
}

// RemoteIDIsNil applies the IsNil predicate on the "RemoteID" field.
func RemoteIDIsNil() predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.IsNull(s.C(FieldRemoteID)))
	})
}

// RemoteIDNotNil applies the NotNil predicate on the "RemoteID" field.
func RemoteIDNotNil() predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.NotNull(s.C(FieldRemoteID)))
	})
}

// RemoteIDEqualFold applies the EqualFold predicate on the "RemoteID" field.
func RemoteIDEqualFold(v imap.MessageID) predicate.Message {
	vc := string(v)
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.EqualFold(s.C(FieldRemoteID), vc))
	})
}

// RemoteIDContainsFold applies the ContainsFold predicate on the "RemoteID" field.
func RemoteIDContainsFold(v imap.MessageID) predicate.Message {
	vc := string(v)
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.ContainsFold(s.C(FieldRemoteID), vc))
	})
}

// DateEQ applies the EQ predicate on the "Date" field.
func DateEQ(v time.Time) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldDate), v))
	})
}

// DateNEQ applies the NEQ predicate on the "Date" field.
func DateNEQ(v time.Time) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.NEQ(s.C(FieldDate), v))
	})
}

// DateIn applies the In predicate on the "Date" field.
func DateIn(vs ...time.Time) predicate.Message {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.In(s.C(FieldDate), v...))
	})
}

// DateNotIn applies the NotIn predicate on the "Date" field.
func DateNotIn(vs ...time.Time) predicate.Message {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.NotIn(s.C(FieldDate), v...))
	})
}

// DateGT applies the GT predicate on the "Date" field.
func DateGT(v time.Time) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.GT(s.C(FieldDate), v))
	})
}

// DateGTE applies the GTE predicate on the "Date" field.
func DateGTE(v time.Time) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.GTE(s.C(FieldDate), v))
	})
}

// DateLT applies the LT predicate on the "Date" field.
func DateLT(v time.Time) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.LT(s.C(FieldDate), v))
	})
}

// DateLTE applies the LTE predicate on the "Date" field.
func DateLTE(v time.Time) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.LTE(s.C(FieldDate), v))
	})
}

// SizeEQ applies the EQ predicate on the "Size" field.
func SizeEQ(v int) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldSize), v))
	})
}

// SizeNEQ applies the NEQ predicate on the "Size" field.
func SizeNEQ(v int) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.NEQ(s.C(FieldSize), v))
	})
}

// SizeIn applies the In predicate on the "Size" field.
func SizeIn(vs ...int) predicate.Message {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.In(s.C(FieldSize), v...))
	})
}

// SizeNotIn applies the NotIn predicate on the "Size" field.
func SizeNotIn(vs ...int) predicate.Message {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.NotIn(s.C(FieldSize), v...))
	})
}

// SizeGT applies the GT predicate on the "Size" field.
func SizeGT(v int) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.GT(s.C(FieldSize), v))
	})
}

// SizeGTE applies the GTE predicate on the "Size" field.
func SizeGTE(v int) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.GTE(s.C(FieldSize), v))
	})
}

// SizeLT applies the LT predicate on the "Size" field.
func SizeLT(v int) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.LT(s.C(FieldSize), v))
	})
}

// SizeLTE applies the LTE predicate on the "Size" field.
func SizeLTE(v int) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.LTE(s.C(FieldSize), v))
	})
}

// BodyEQ applies the EQ predicate on the "Body" field.
func BodyEQ(v string) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldBody), v))
	})
}

// BodyNEQ applies the NEQ predicate on the "Body" field.
func BodyNEQ(v string) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.NEQ(s.C(FieldBody), v))
	})
}

// BodyIn applies the In predicate on the "Body" field.
func BodyIn(vs ...string) predicate.Message {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.In(s.C(FieldBody), v...))
	})
}

// BodyNotIn applies the NotIn predicate on the "Body" field.
func BodyNotIn(vs ...string) predicate.Message {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.NotIn(s.C(FieldBody), v...))
	})
}

// BodyGT applies the GT predicate on the "Body" field.
func BodyGT(v string) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.GT(s.C(FieldBody), v))
	})
}

// BodyGTE applies the GTE predicate on the "Body" field.
func BodyGTE(v string) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.GTE(s.C(FieldBody), v))
	})
}

// BodyLT applies the LT predicate on the "Body" field.
func BodyLT(v string) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.LT(s.C(FieldBody), v))
	})
}

// BodyLTE applies the LTE predicate on the "Body" field.
func BodyLTE(v string) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.LTE(s.C(FieldBody), v))
	})
}

// BodyContains applies the Contains predicate on the "Body" field.
func BodyContains(v string) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.Contains(s.C(FieldBody), v))
	})
}

// BodyHasPrefix applies the HasPrefix predicate on the "Body" field.
func BodyHasPrefix(v string) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.HasPrefix(s.C(FieldBody), v))
	})
}

// BodyHasSuffix applies the HasSuffix predicate on the "Body" field.
func BodyHasSuffix(v string) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.HasSuffix(s.C(FieldBody), v))
	})
}

// BodyEqualFold applies the EqualFold predicate on the "Body" field.
func BodyEqualFold(v string) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.EqualFold(s.C(FieldBody), v))
	})
}

// BodyContainsFold applies the ContainsFold predicate on the "Body" field.
func BodyContainsFold(v string) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.ContainsFold(s.C(FieldBody), v))
	})
}

// BodyStructureEQ applies the EQ predicate on the "BodyStructure" field.
func BodyStructureEQ(v string) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldBodyStructure), v))
	})
}

// BodyStructureNEQ applies the NEQ predicate on the "BodyStructure" field.
func BodyStructureNEQ(v string) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.NEQ(s.C(FieldBodyStructure), v))
	})
}

// BodyStructureIn applies the In predicate on the "BodyStructure" field.
func BodyStructureIn(vs ...string) predicate.Message {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.In(s.C(FieldBodyStructure), v...))
	})
}

// BodyStructureNotIn applies the NotIn predicate on the "BodyStructure" field.
func BodyStructureNotIn(vs ...string) predicate.Message {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.NotIn(s.C(FieldBodyStructure), v...))
	})
}

// BodyStructureGT applies the GT predicate on the "BodyStructure" field.
func BodyStructureGT(v string) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.GT(s.C(FieldBodyStructure), v))
	})
}

// BodyStructureGTE applies the GTE predicate on the "BodyStructure" field.
func BodyStructureGTE(v string) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.GTE(s.C(FieldBodyStructure), v))
	})
}

// BodyStructureLT applies the LT predicate on the "BodyStructure" field.
func BodyStructureLT(v string) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.LT(s.C(FieldBodyStructure), v))
	})
}

// BodyStructureLTE applies the LTE predicate on the "BodyStructure" field.
func BodyStructureLTE(v string) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.LTE(s.C(FieldBodyStructure), v))
	})
}

// BodyStructureContains applies the Contains predicate on the "BodyStructure" field.
func BodyStructureContains(v string) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.Contains(s.C(FieldBodyStructure), v))
	})
}

// BodyStructureHasPrefix applies the HasPrefix predicate on the "BodyStructure" field.
func BodyStructureHasPrefix(v string) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.HasPrefix(s.C(FieldBodyStructure), v))
	})
}

// BodyStructureHasSuffix applies the HasSuffix predicate on the "BodyStructure" field.
func BodyStructureHasSuffix(v string) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.HasSuffix(s.C(FieldBodyStructure), v))
	})
}

// BodyStructureEqualFold applies the EqualFold predicate on the "BodyStructure" field.
func BodyStructureEqualFold(v string) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.EqualFold(s.C(FieldBodyStructure), v))
	})
}

// BodyStructureContainsFold applies the ContainsFold predicate on the "BodyStructure" field.
func BodyStructureContainsFold(v string) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.ContainsFold(s.C(FieldBodyStructure), v))
	})
}

// EnvelopeEQ applies the EQ predicate on the "Envelope" field.
func EnvelopeEQ(v string) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldEnvelope), v))
	})
}

// EnvelopeNEQ applies the NEQ predicate on the "Envelope" field.
func EnvelopeNEQ(v string) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.NEQ(s.C(FieldEnvelope), v))
	})
}

// EnvelopeIn applies the In predicate on the "Envelope" field.
func EnvelopeIn(vs ...string) predicate.Message {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.In(s.C(FieldEnvelope), v...))
	})
}

// EnvelopeNotIn applies the NotIn predicate on the "Envelope" field.
func EnvelopeNotIn(vs ...string) predicate.Message {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.NotIn(s.C(FieldEnvelope), v...))
	})
}

// EnvelopeGT applies the GT predicate on the "Envelope" field.
func EnvelopeGT(v string) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.GT(s.C(FieldEnvelope), v))
	})
}

// EnvelopeGTE applies the GTE predicate on the "Envelope" field.
func EnvelopeGTE(v string) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.GTE(s.C(FieldEnvelope), v))
	})
}

// EnvelopeLT applies the LT predicate on the "Envelope" field.
func EnvelopeLT(v string) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.LT(s.C(FieldEnvelope), v))
	})
}

// EnvelopeLTE applies the LTE predicate on the "Envelope" field.
func EnvelopeLTE(v string) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.LTE(s.C(FieldEnvelope), v))
	})
}

// EnvelopeContains applies the Contains predicate on the "Envelope" field.
func EnvelopeContains(v string) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.Contains(s.C(FieldEnvelope), v))
	})
}

// EnvelopeHasPrefix applies the HasPrefix predicate on the "Envelope" field.
func EnvelopeHasPrefix(v string) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.HasPrefix(s.C(FieldEnvelope), v))
	})
}

// EnvelopeHasSuffix applies the HasSuffix predicate on the "Envelope" field.
func EnvelopeHasSuffix(v string) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.HasSuffix(s.C(FieldEnvelope), v))
	})
}

// EnvelopeEqualFold applies the EqualFold predicate on the "Envelope" field.
func EnvelopeEqualFold(v string) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.EqualFold(s.C(FieldEnvelope), v))
	})
}

// EnvelopeContainsFold applies the ContainsFold predicate on the "Envelope" field.
func EnvelopeContainsFold(v string) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.ContainsFold(s.C(FieldEnvelope), v))
	})
}

// DeletedEQ applies the EQ predicate on the "Deleted" field.
func DeletedEQ(v bool) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldDeleted), v))
	})
}

// DeletedNEQ applies the NEQ predicate on the "Deleted" field.
func DeletedNEQ(v bool) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s.Where(sql.NEQ(s.C(FieldDeleted), v))
	})
}

// HasFlags applies the HasEdge predicate on the "flags" edge.
func HasFlags() predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.To(FlagsTable, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, FlagsTable, FlagsColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasFlagsWith applies the HasEdge predicate on the "flags" edge with a given conditions (other predicates).
func HasFlagsWith(preds ...predicate.MessageFlag) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.To(FlagsInverseTable, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, FlagsTable, FlagsColumn),
		)
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasUIDs applies the HasEdge predicate on the "UIDs" edge.
func HasUIDs() predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.To(UIDsTable, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, true, UIDsTable, UIDsColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasUIDsWith applies the HasEdge predicate on the "UIDs" edge with a given conditions (other predicates).
func HasUIDsWith(preds ...predicate.UID) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.To(UIDsInverseTable, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, true, UIDsTable, UIDsColumn),
		)
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.Message) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		s1 := s.Clone().SetP(nil)
		for _, p := range predicates {
			p(s1)
		}
		s.Where(s1.P())
	})
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.Message) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
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
func Not(p predicate.Message) predicate.Message {
	return predicate.Message(func(s *sql.Selector) {
		p(s.Not())
	})
}