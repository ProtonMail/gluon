// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/ProtonMail/gluon/internal/db/ent/mailboxflag"
	"github.com/ProtonMail/gluon/internal/db/ent/predicate"
)

// MailboxFlagUpdate is the builder for updating MailboxFlag entities.
type MailboxFlagUpdate struct {
	config
	hooks    []Hook
	mutation *MailboxFlagMutation
}

// Where appends a list predicates to the MailboxFlagUpdate builder.
func (mfu *MailboxFlagUpdate) Where(ps ...predicate.MailboxFlag) *MailboxFlagUpdate {
	mfu.mutation.Where(ps...)
	return mfu
}

// SetValue sets the "Value" field.
func (mfu *MailboxFlagUpdate) SetValue(s string) *MailboxFlagUpdate {
	mfu.mutation.SetValue(s)
	return mfu
}

// Mutation returns the MailboxFlagMutation object of the builder.
func (mfu *MailboxFlagUpdate) Mutation() *MailboxFlagMutation {
	return mfu.mutation
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (mfu *MailboxFlagUpdate) Save(ctx context.Context) (int, error) {
	var (
		err      error
		affected int
	)
	if len(mfu.hooks) == 0 {
		affected, err = mfu.sqlSave(ctx)
	} else {
		var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
			mutation, ok := m.(*MailboxFlagMutation)
			if !ok {
				return nil, fmt.Errorf("unexpected mutation type %T", m)
			}
			mfu.mutation = mutation
			affected, err = mfu.sqlSave(ctx)
			mutation.done = true
			return affected, err
		})
		for i := len(mfu.hooks) - 1; i >= 0; i-- {
			if mfu.hooks[i] == nil {
				return 0, fmt.Errorf("ent: uninitialized hook (forgotten import ent/runtime?)")
			}
			mut = mfu.hooks[i](mut)
		}
		if _, err := mut.Mutate(ctx, mfu.mutation); err != nil {
			return 0, err
		}
	}
	return affected, err
}

// SaveX is like Save, but panics if an error occurs.
func (mfu *MailboxFlagUpdate) SaveX(ctx context.Context) int {
	affected, err := mfu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (mfu *MailboxFlagUpdate) Exec(ctx context.Context) error {
	_, err := mfu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (mfu *MailboxFlagUpdate) ExecX(ctx context.Context) {
	if err := mfu.Exec(ctx); err != nil {
		panic(err)
	}
}

func (mfu *MailboxFlagUpdate) sqlSave(ctx context.Context) (n int, err error) {
	_spec := &sqlgraph.UpdateSpec{
		Node: &sqlgraph.NodeSpec{
			Table:   mailboxflag.Table,
			Columns: mailboxflag.Columns,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeInt,
				Column: mailboxflag.FieldID,
			},
		},
	}
	if ps := mfu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := mfu.mutation.Value(); ok {
		_spec.Fields.Set = append(_spec.Fields.Set, &sqlgraph.FieldSpec{
			Type:   field.TypeString,
			Value:  value,
			Column: mailboxflag.FieldValue,
		})
	}
	if n, err = sqlgraph.UpdateNodes(ctx, mfu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{mailboxflag.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	return n, nil
}

// MailboxFlagUpdateOne is the builder for updating a single MailboxFlag entity.
type MailboxFlagUpdateOne struct {
	config
	fields   []string
	hooks    []Hook
	mutation *MailboxFlagMutation
}

// SetValue sets the "Value" field.
func (mfuo *MailboxFlagUpdateOne) SetValue(s string) *MailboxFlagUpdateOne {
	mfuo.mutation.SetValue(s)
	return mfuo
}

// Mutation returns the MailboxFlagMutation object of the builder.
func (mfuo *MailboxFlagUpdateOne) Mutation() *MailboxFlagMutation {
	return mfuo.mutation
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (mfuo *MailboxFlagUpdateOne) Select(field string, fields ...string) *MailboxFlagUpdateOne {
	mfuo.fields = append([]string{field}, fields...)
	return mfuo
}

// Save executes the query and returns the updated MailboxFlag entity.
func (mfuo *MailboxFlagUpdateOne) Save(ctx context.Context) (*MailboxFlag, error) {
	var (
		err  error
		node *MailboxFlag
	)
	if len(mfuo.hooks) == 0 {
		node, err = mfuo.sqlSave(ctx)
	} else {
		var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
			mutation, ok := m.(*MailboxFlagMutation)
			if !ok {
				return nil, fmt.Errorf("unexpected mutation type %T", m)
			}
			mfuo.mutation = mutation
			node, err = mfuo.sqlSave(ctx)
			mutation.done = true
			return node, err
		})
		for i := len(mfuo.hooks) - 1; i >= 0; i-- {
			if mfuo.hooks[i] == nil {
				return nil, fmt.Errorf("ent: uninitialized hook (forgotten import ent/runtime?)")
			}
			mut = mfuo.hooks[i](mut)
		}
		v, err := mut.Mutate(ctx, mfuo.mutation)
		if err != nil {
			return nil, err
		}
		nv, ok := v.(*MailboxFlag)
		if !ok {
			return nil, fmt.Errorf("unexpected node type %T returned from MailboxFlagMutation", v)
		}
		node = nv
	}
	return node, err
}

// SaveX is like Save, but panics if an error occurs.
func (mfuo *MailboxFlagUpdateOne) SaveX(ctx context.Context) *MailboxFlag {
	node, err := mfuo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (mfuo *MailboxFlagUpdateOne) Exec(ctx context.Context) error {
	_, err := mfuo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (mfuo *MailboxFlagUpdateOne) ExecX(ctx context.Context) {
	if err := mfuo.Exec(ctx); err != nil {
		panic(err)
	}
}

func (mfuo *MailboxFlagUpdateOne) sqlSave(ctx context.Context) (_node *MailboxFlag, err error) {
	_spec := &sqlgraph.UpdateSpec{
		Node: &sqlgraph.NodeSpec{
			Table:   mailboxflag.Table,
			Columns: mailboxflag.Columns,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeInt,
				Column: mailboxflag.FieldID,
			},
		},
	}
	id, ok := mfuo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "MailboxFlag.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := mfuo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, mailboxflag.FieldID)
		for _, f := range fields {
			if !mailboxflag.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != mailboxflag.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := mfuo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := mfuo.mutation.Value(); ok {
		_spec.Fields.Set = append(_spec.Fields.Set, &sqlgraph.FieldSpec{
			Type:   field.TypeString,
			Value:  value,
			Column: mailboxflag.FieldValue,
		})
	}
	_node = &MailboxFlag{config: mfuo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, mfuo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{mailboxflag.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	return _node, nil
}