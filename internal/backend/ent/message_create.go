// Code generated by entc, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/ProtonMail/gluon/internal/backend/ent/message"
	"github.com/ProtonMail/gluon/internal/backend/ent/messageflag"
	"github.com/ProtonMail/gluon/internal/backend/ent/uid"
)

// MessageCreate is the builder for creating a Message entity.
type MessageCreate struct {
	config
	mutation *MessageMutation
	hooks    []Hook
}

// SetMessageID sets the "MessageID" field.
func (mc *MessageCreate) SetMessageID(s string) *MessageCreate {
	mc.mutation.SetMessageID(s)
	return mc
}

// SetInternalID sets the "InternalID" field.
func (mc *MessageCreate) SetInternalID(s string) *MessageCreate {
	mc.mutation.SetInternalID(s)
	return mc
}

// SetDate sets the "Date" field.
func (mc *MessageCreate) SetDate(t time.Time) *MessageCreate {
	mc.mutation.SetDate(t)
	return mc
}

// SetSize sets the "Size" field.
func (mc *MessageCreate) SetSize(i int) *MessageCreate {
	mc.mutation.SetSize(i)
	return mc
}

// SetBody sets the "Body" field.
func (mc *MessageCreate) SetBody(s string) *MessageCreate {
	mc.mutation.SetBody(s)
	return mc
}

// SetBodyStructure sets the "BodyStructure" field.
func (mc *MessageCreate) SetBodyStructure(s string) *MessageCreate {
	mc.mutation.SetBodyStructure(s)
	return mc
}

// SetEnvelope sets the "Envelope" field.
func (mc *MessageCreate) SetEnvelope(s string) *MessageCreate {
	mc.mutation.SetEnvelope(s)
	return mc
}

// SetDeleted sets the "Deleted" field.
func (mc *MessageCreate) SetDeleted(b bool) *MessageCreate {
	mc.mutation.SetDeleted(b)
	return mc
}

// SetNillableDeleted sets the "Deleted" field if the given value is not nil.
func (mc *MessageCreate) SetNillableDeleted(b *bool) *MessageCreate {
	if b != nil {
		mc.SetDeleted(*b)
	}
	return mc
}

// AddFlagIDs adds the "flags" edge to the MessageFlag entity by IDs.
func (mc *MessageCreate) AddFlagIDs(ids ...int) *MessageCreate {
	mc.mutation.AddFlagIDs(ids...)
	return mc
}

// AddFlags adds the "flags" edges to the MessageFlag entity.
func (mc *MessageCreate) AddFlags(m ...*MessageFlag) *MessageCreate {
	ids := make([]int, len(m))
	for i := range m {
		ids[i] = m[i].ID
	}
	return mc.AddFlagIDs(ids...)
}

// AddUIDIDs adds the "UIDs" edge to the UID entity by IDs.
func (mc *MessageCreate) AddUIDIDs(ids ...int) *MessageCreate {
	mc.mutation.AddUIDIDs(ids...)
	return mc
}

// AddUIDs adds the "UIDs" edges to the UID entity.
func (mc *MessageCreate) AddUIDs(u ...*UID) *MessageCreate {
	ids := make([]int, len(u))
	for i := range u {
		ids[i] = u[i].ID
	}
	return mc.AddUIDIDs(ids...)
}

// Mutation returns the MessageMutation object of the builder.
func (mc *MessageCreate) Mutation() *MessageMutation {
	return mc.mutation
}

// Save creates the Message in the database.
func (mc *MessageCreate) Save(ctx context.Context) (*Message, error) {
	var (
		err  error
		node *Message
	)
	mc.defaults()
	if len(mc.hooks) == 0 {
		if err = mc.check(); err != nil {
			return nil, err
		}
		node, err = mc.sqlSave(ctx)
	} else {
		var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
			mutation, ok := m.(*MessageMutation)
			if !ok {
				return nil, fmt.Errorf("unexpected mutation type %T", m)
			}
			if err = mc.check(); err != nil {
				return nil, err
			}
			mc.mutation = mutation
			if node, err = mc.sqlSave(ctx); err != nil {
				return nil, err
			}
			mutation.id = &node.ID
			mutation.done = true
			return node, err
		})
		for i := len(mc.hooks) - 1; i >= 0; i-- {
			if mc.hooks[i] == nil {
				return nil, fmt.Errorf("ent: uninitialized hook (forgotten import ent/runtime?)")
			}
			mut = mc.hooks[i](mut)
		}
		if _, err := mut.Mutate(ctx, mc.mutation); err != nil {
			return nil, err
		}
	}
	return node, err
}

// SaveX calls Save and panics if Save returns an error.
func (mc *MessageCreate) SaveX(ctx context.Context) *Message {
	v, err := mc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (mc *MessageCreate) Exec(ctx context.Context) error {
	_, err := mc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (mc *MessageCreate) ExecX(ctx context.Context) {
	if err := mc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (mc *MessageCreate) defaults() {
	if _, ok := mc.mutation.Deleted(); !ok {
		v := message.DefaultDeleted
		mc.mutation.SetDeleted(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (mc *MessageCreate) check() error {
	if _, ok := mc.mutation.MessageID(); !ok {
		return &ValidationError{Name: "MessageID", err: errors.New(`ent: missing required field "Message.MessageID"`)}
	}
	if _, ok := mc.mutation.InternalID(); !ok {
		return &ValidationError{Name: "InternalID", err: errors.New(`ent: missing required field "Message.InternalID"`)}
	}
	if _, ok := mc.mutation.Date(); !ok {
		return &ValidationError{Name: "Date", err: errors.New(`ent: missing required field "Message.Date"`)}
	}
	if _, ok := mc.mutation.Size(); !ok {
		return &ValidationError{Name: "Size", err: errors.New(`ent: missing required field "Message.Size"`)}
	}
	if _, ok := mc.mutation.Body(); !ok {
		return &ValidationError{Name: "Body", err: errors.New(`ent: missing required field "Message.Body"`)}
	}
	if _, ok := mc.mutation.BodyStructure(); !ok {
		return &ValidationError{Name: "BodyStructure", err: errors.New(`ent: missing required field "Message.BodyStructure"`)}
	}
	if _, ok := mc.mutation.Envelope(); !ok {
		return &ValidationError{Name: "Envelope", err: errors.New(`ent: missing required field "Message.Envelope"`)}
	}
	if _, ok := mc.mutation.Deleted(); !ok {
		return &ValidationError{Name: "Deleted", err: errors.New(`ent: missing required field "Message.Deleted"`)}
	}
	return nil
}

func (mc *MessageCreate) sqlSave(ctx context.Context) (*Message, error) {
	_node, _spec := mc.createSpec()
	if err := sqlgraph.CreateNode(ctx, mc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{err.Error(), err}
		}
		return nil, err
	}
	id := _spec.ID.Value.(int64)
	_node.ID = int(id)
	return _node, nil
}

func (mc *MessageCreate) createSpec() (*Message, *sqlgraph.CreateSpec) {
	var (
		_node = &Message{config: mc.config}
		_spec = &sqlgraph.CreateSpec{
			Table: message.Table,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeInt,
				Column: message.FieldID,
			},
		}
	)
	if value, ok := mc.mutation.MessageID(); ok {
		_spec.Fields = append(_spec.Fields, &sqlgraph.FieldSpec{
			Type:   field.TypeString,
			Value:  value,
			Column: message.FieldMessageID,
		})
		_node.MessageID = value
	}
	if value, ok := mc.mutation.InternalID(); ok {
		_spec.Fields = append(_spec.Fields, &sqlgraph.FieldSpec{
			Type:   field.TypeString,
			Value:  value,
			Column: message.FieldInternalID,
		})
		_node.InternalID = value
	}
	if value, ok := mc.mutation.Date(); ok {
		_spec.Fields = append(_spec.Fields, &sqlgraph.FieldSpec{
			Type:   field.TypeTime,
			Value:  value,
			Column: message.FieldDate,
		})
		_node.Date = value
	}
	if value, ok := mc.mutation.Size(); ok {
		_spec.Fields = append(_spec.Fields, &sqlgraph.FieldSpec{
			Type:   field.TypeInt,
			Value:  value,
			Column: message.FieldSize,
		})
		_node.Size = value
	}
	if value, ok := mc.mutation.Body(); ok {
		_spec.Fields = append(_spec.Fields, &sqlgraph.FieldSpec{
			Type:   field.TypeString,
			Value:  value,
			Column: message.FieldBody,
		})
		_node.Body = value
	}
	if value, ok := mc.mutation.BodyStructure(); ok {
		_spec.Fields = append(_spec.Fields, &sqlgraph.FieldSpec{
			Type:   field.TypeString,
			Value:  value,
			Column: message.FieldBodyStructure,
		})
		_node.BodyStructure = value
	}
	if value, ok := mc.mutation.Envelope(); ok {
		_spec.Fields = append(_spec.Fields, &sqlgraph.FieldSpec{
			Type:   field.TypeString,
			Value:  value,
			Column: message.FieldEnvelope,
		})
		_node.Envelope = value
	}
	if value, ok := mc.mutation.Deleted(); ok {
		_spec.Fields = append(_spec.Fields, &sqlgraph.FieldSpec{
			Type:   field.TypeBool,
			Value:  value,
			Column: message.FieldDeleted,
		})
		_node.Deleted = value
	}
	if nodes := mc.mutation.FlagsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   message.FlagsTable,
			Columns: []string{message.FlagsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: &sqlgraph.FieldSpec{
					Type:   field.TypeInt,
					Column: messageflag.FieldID,
				},
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := mc.mutation.UIDsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   message.UIDsTable,
			Columns: []string{message.UIDsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: &sqlgraph.FieldSpec{
					Type:   field.TypeInt,
					Column: uid.FieldID,
				},
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// MessageCreateBulk is the builder for creating many Message entities in bulk.
type MessageCreateBulk struct {
	config
	builders []*MessageCreate
}

// Save creates the Message entities in the database.
func (mcb *MessageCreateBulk) Save(ctx context.Context) ([]*Message, error) {
	specs := make([]*sqlgraph.CreateSpec, len(mcb.builders))
	nodes := make([]*Message, len(mcb.builders))
	mutators := make([]Mutator, len(mcb.builders))
	for i := range mcb.builders {
		func(i int, root context.Context) {
			builder := mcb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*MessageMutation)
				if !ok {
					return nil, fmt.Errorf("unexpected mutation type %T", m)
				}
				if err := builder.check(); err != nil {
					return nil, err
				}
				builder.mutation = mutation
				nodes[i], specs[i] = builder.createSpec()
				var err error
				if i < len(mutators)-1 {
					_, err = mutators[i+1].Mutate(root, mcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, mcb.driver, spec); err != nil {
						if sqlgraph.IsConstraintError(err) {
							err = &ConstraintError{err.Error(), err}
						}
					}
				}
				if err != nil {
					return nil, err
				}
				mutation.id = &nodes[i].ID
				mutation.done = true
				if specs[i].ID.Value != nil {
					id := specs[i].ID.Value.(int64)
					nodes[i].ID = int(id)
				}
				return nodes[i], nil
			})
			for i := len(builder.hooks) - 1; i >= 0; i-- {
				mut = builder.hooks[i](mut)
			}
			mutators[i] = mut
		}(i, ctx)
	}
	if len(mutators) > 0 {
		if _, err := mutators[0].Mutate(ctx, mcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (mcb *MessageCreateBulk) SaveX(ctx context.Context) []*Message {
	v, err := mcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (mcb *MessageCreateBulk) Exec(ctx context.Context) error {
	_, err := mcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (mcb *MessageCreateBulk) ExecX(ctx context.Context) {
	if err := mcb.Exec(ctx); err != nil {
		panic(err)
	}
}
