// Code generated by ent, DO NOT EDIT.

package internal

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/ProtonMail/gluon/internal/db_impl/ent_db/internal/mailbox"
	"github.com/ProtonMail/gluon/internal/db_impl/ent_db/internal/predicate"
)

// MailboxDelete is the builder for deleting a Mailbox entity.
type MailboxDelete struct {
	config
	hooks    []Hook
	mutation *MailboxMutation
}

// Where appends a list predicates to the MailboxDelete builder.
func (md *MailboxDelete) Where(ps ...predicate.Mailbox) *MailboxDelete {
	md.mutation.Where(ps...)
	return md
}

// Exec executes the deletion query and returns how many vertices were deleted.
func (md *MailboxDelete) Exec(ctx context.Context) (int, error) {
	return withHooks[int, MailboxMutation](ctx, md.sqlExec, md.mutation, md.hooks)
}

// ExecX is like Exec, but panics if an error occurs.
func (md *MailboxDelete) ExecX(ctx context.Context) int {
	n, err := md.Exec(ctx)
	if err != nil {
		panic(err)
	}
	return n
}

func (md *MailboxDelete) sqlExec(ctx context.Context) (int, error) {
	_spec := sqlgraph.NewDeleteSpec(mailbox.Table, sqlgraph.NewFieldSpec(mailbox.FieldID, field.TypeUint64))
	if ps := md.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	affected, err := sqlgraph.DeleteNodes(ctx, md.driver, _spec)
	if err != nil && sqlgraph.IsConstraintError(err) {
		err = &ConstraintError{msg: err.Error(), wrap: err}
	}
	md.mutation.done = true
	return affected, err
}

// MailboxDeleteOne is the builder for deleting a single Mailbox entity.
type MailboxDeleteOne struct {
	md *MailboxDelete
}

// Where appends a list predicates to the MailboxDelete builder.
func (mdo *MailboxDeleteOne) Where(ps ...predicate.Mailbox) *MailboxDeleteOne {
	mdo.md.mutation.Where(ps...)
	return mdo
}

// Exec executes the deletion query.
func (mdo *MailboxDeleteOne) Exec(ctx context.Context) error {
	n, err := mdo.md.Exec(ctx)
	switch {
	case err != nil:
		return err
	case n == 0:
		return &NotFoundError{mailbox.Label}
	default:
		return nil
	}
}

// ExecX is like Exec, but panics if an error occurs.
func (mdo *MailboxDeleteOne) ExecX(ctx context.Context) {
	if err := mdo.Exec(ctx); err != nil {
		panic(err)
	}
}