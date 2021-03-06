// Code generated by entc, DO NOT EDIT.

package ent

import (
	"fmt"
	"strings"

	"entgo.io/ent/dialect/sql"
	"github.com/ProtonMail/gluon/internal/backend/ent/messageflag"
)

// MessageFlag is the model entity for the MessageFlag schema.
type MessageFlag struct {
	config `json:"-"`
	// ID of the ent.
	ID int `json:"id,omitempty"`
	// Value holds the value of the "Value" field.
	Value         string `json:"Value,omitempty"`
	message_flags *int
}

// scanValues returns the types for scanning values from sql.Rows.
func (*MessageFlag) scanValues(columns []string) ([]interface{}, error) {
	values := make([]interface{}, len(columns))
	for i := range columns {
		switch columns[i] {
		case messageflag.FieldID:
			values[i] = new(sql.NullInt64)
		case messageflag.FieldValue:
			values[i] = new(sql.NullString)
		case messageflag.ForeignKeys[0]: // message_flags
			values[i] = new(sql.NullInt64)
		default:
			return nil, fmt.Errorf("unexpected column %q for type MessageFlag", columns[i])
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the MessageFlag fields.
func (mf *MessageFlag) assignValues(columns []string, values []interface{}) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case messageflag.FieldID:
			value, ok := values[i].(*sql.NullInt64)
			if !ok {
				return fmt.Errorf("unexpected type %T for field id", value)
			}
			mf.ID = int(value.Int64)
		case messageflag.FieldValue:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field Value", values[i])
			} else if value.Valid {
				mf.Value = value.String
			}
		case messageflag.ForeignKeys[0]:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for edge-field message_flags", value)
			} else if value.Valid {
				mf.message_flags = new(int)
				*mf.message_flags = int(value.Int64)
			}
		}
	}
	return nil
}

// Update returns a builder for updating this MessageFlag.
// Note that you need to call MessageFlag.Unwrap() before calling this method if this MessageFlag
// was returned from a transaction, and the transaction was committed or rolled back.
func (mf *MessageFlag) Update() *MessageFlagUpdateOne {
	return (&MessageFlagClient{config: mf.config}).UpdateOne(mf)
}

// Unwrap unwraps the MessageFlag entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (mf *MessageFlag) Unwrap() *MessageFlag {
	tx, ok := mf.config.driver.(*txDriver)
	if !ok {
		panic("ent: MessageFlag is not a transactional entity")
	}
	mf.config.driver = tx.drv
	return mf
}

// String implements the fmt.Stringer.
func (mf *MessageFlag) String() string {
	var builder strings.Builder
	builder.WriteString("MessageFlag(")
	builder.WriteString(fmt.Sprintf("id=%v", mf.ID))
	builder.WriteString(", Value=")
	builder.WriteString(mf.Value)
	builder.WriteByte(')')
	return builder.String()
}

// MessageFlags is a parsable slice of MessageFlag.
type MessageFlags []*MessageFlag

func (mf MessageFlags) config(cfg config) {
	for _i := range mf {
		mf[_i].config = cfg
	}
}
