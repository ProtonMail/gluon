// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"fmt"
	"math"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/db/ent/mailbox"
	"github.com/ProtonMail/gluon/internal/db/ent/message"
	"github.com/ProtonMail/gluon/internal/db/ent/predicate"
	"github.com/ProtonMail/gluon/internal/db/ent/uid"
)

// UIDQuery is the builder for querying UID entities.
type UIDQuery struct {
	config
	limit       *int
	offset      *int
	unique      *bool
	order       []OrderFunc
	fields      []string
	predicates  []predicate.UID
	withMessage *MessageQuery
	withMailbox *MailboxQuery
	withFKs     bool
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the UIDQuery builder.
func (uq *UIDQuery) Where(ps ...predicate.UID) *UIDQuery {
	uq.predicates = append(uq.predicates, ps...)
	return uq
}

// Limit adds a limit step to the query.
func (uq *UIDQuery) Limit(limit int) *UIDQuery {
	uq.limit = &limit
	return uq
}

// Offset adds an offset step to the query.
func (uq *UIDQuery) Offset(offset int) *UIDQuery {
	uq.offset = &offset
	return uq
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (uq *UIDQuery) Unique(unique bool) *UIDQuery {
	uq.unique = &unique
	return uq
}

// Order adds an order step to the query.
func (uq *UIDQuery) Order(o ...OrderFunc) *UIDQuery {
	uq.order = append(uq.order, o...)
	return uq
}

// QueryMessage chains the current query on the "message" edge.
func (uq *UIDQuery) QueryMessage() *MessageQuery {
	query := &MessageQuery{config: uq.config}
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := uq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := uq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(uid.Table, uid.FieldID, selector),
			sqlgraph.To(message.Table, message.FieldID),
			sqlgraph.Edge(sqlgraph.M2O, false, uid.MessageTable, uid.MessageColumn),
		)
		fromU = sqlgraph.SetNeighbors(uq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// QueryMailbox chains the current query on the "mailbox" edge.
func (uq *UIDQuery) QueryMailbox() *MailboxQuery {
	query := &MailboxQuery{config: uq.config}
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := uq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := uq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(uid.Table, uid.FieldID, selector),
			sqlgraph.To(mailbox.Table, mailbox.FieldID),
			sqlgraph.Edge(sqlgraph.M2O, true, uid.MailboxTable, uid.MailboxColumn),
		)
		fromU = sqlgraph.SetNeighbors(uq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// First returns the first UID entity from the query.
// Returns a *NotFoundError when no UID was found.
func (uq *UIDQuery) First(ctx context.Context) (*UID, error) {
	nodes, err := uq.Limit(1).All(ctx)
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{uid.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (uq *UIDQuery) FirstX(ctx context.Context) *UID {
	node, err := uq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first UID ID from the query.
// Returns a *NotFoundError when no UID ID was found.
func (uq *UIDQuery) FirstID(ctx context.Context) (id int, err error) {
	var ids []int
	if ids, err = uq.Limit(1).IDs(ctx); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{uid.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (uq *UIDQuery) FirstIDX(ctx context.Context) int {
	id, err := uq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single UID entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one UID entity is found.
// Returns a *NotFoundError when no UID entities are found.
func (uq *UIDQuery) Only(ctx context.Context) (*UID, error) {
	nodes, err := uq.Limit(2).All(ctx)
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{uid.Label}
	default:
		return nil, &NotSingularError{uid.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (uq *UIDQuery) OnlyX(ctx context.Context) *UID {
	node, err := uq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only UID ID in the query.
// Returns a *NotSingularError when more than one UID ID is found.
// Returns a *NotFoundError when no entities are found.
func (uq *UIDQuery) OnlyID(ctx context.Context) (id int, err error) {
	var ids []int
	if ids, err = uq.Limit(2).IDs(ctx); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{uid.Label}
	default:
		err = &NotSingularError{uid.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (uq *UIDQuery) OnlyIDX(ctx context.Context) int {
	id, err := uq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of UIDs.
func (uq *UIDQuery) All(ctx context.Context) ([]*UID, error) {
	if err := uq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	return uq.sqlAll(ctx)
}

// AllX is like All, but panics if an error occurs.
func (uq *UIDQuery) AllX(ctx context.Context) []*UID {
	nodes, err := uq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of UID IDs.
func (uq *UIDQuery) IDs(ctx context.Context) ([]int, error) {
	var ids []int
	if err := uq.Select(uid.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (uq *UIDQuery) IDsX(ctx context.Context) []int {
	ids, err := uq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (uq *UIDQuery) Count(ctx context.Context) (int, error) {
	if err := uq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return uq.sqlCount(ctx)
}

// CountX is like Count, but panics if an error occurs.
func (uq *UIDQuery) CountX(ctx context.Context) int {
	count, err := uq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (uq *UIDQuery) Exist(ctx context.Context) (bool, error) {
	if err := uq.prepareQuery(ctx); err != nil {
		return false, err
	}
	return uq.sqlExist(ctx)
}

// ExistX is like Exist, but panics if an error occurs.
func (uq *UIDQuery) ExistX(ctx context.Context) bool {
	exist, err := uq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the UIDQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (uq *UIDQuery) Clone() *UIDQuery {
	if uq == nil {
		return nil
	}
	return &UIDQuery{
		config:      uq.config,
		limit:       uq.limit,
		offset:      uq.offset,
		order:       append([]OrderFunc{}, uq.order...),
		predicates:  append([]predicate.UID{}, uq.predicates...),
		withMessage: uq.withMessage.Clone(),
		withMailbox: uq.withMailbox.Clone(),
		// clone intermediate query.
		sql:    uq.sql.Clone(),
		path:   uq.path,
		unique: uq.unique,
	}
}

// WithMessage tells the query-builder to eager-load the nodes that are connected to
// the "message" edge. The optional arguments are used to configure the query builder of the edge.
func (uq *UIDQuery) WithMessage(opts ...func(*MessageQuery)) *UIDQuery {
	query := &MessageQuery{config: uq.config}
	for _, opt := range opts {
		opt(query)
	}
	uq.withMessage = query
	return uq
}

// WithMailbox tells the query-builder to eager-load the nodes that are connected to
// the "mailbox" edge. The optional arguments are used to configure the query builder of the edge.
func (uq *UIDQuery) WithMailbox(opts ...func(*MailboxQuery)) *UIDQuery {
	query := &MailboxQuery{config: uq.config}
	for _, opt := range opts {
		opt(query)
	}
	uq.withMailbox = query
	return uq
}

// GroupBy is used to group vertices by one or more fields/columns.
// It is often used with aggregate functions, like: count, max, mean, min, sum.
//
// Example:
//
//	var v []struct {
//		UID imap.UID `json:"UID,omitempty"`
//		Count int `json:"count,omitempty"`
//	}
//
//	client.UID.Query().
//		GroupBy(uid.FieldUID).
//		Aggregate(ent.Count()).
//		Scan(ctx, &v)
//
func (uq *UIDQuery) GroupBy(field string, fields ...string) *UIDGroupBy {
	grbuild := &UIDGroupBy{config: uq.config}
	grbuild.fields = append([]string{field}, fields...)
	grbuild.path = func(ctx context.Context) (prev *sql.Selector, err error) {
		if err := uq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		return uq.sqlQuery(ctx), nil
	}
	grbuild.label = uid.Label
	grbuild.flds, grbuild.scan = &grbuild.fields, grbuild.Scan
	return grbuild
}

// Select allows the selection one or more fields/columns for the given query,
// instead of selecting all fields in the entity.
//
// Example:
//
//	var v []struct {
//		UID imap.UID `json:"UID,omitempty"`
//	}
//
//	client.UID.Query().
//		Select(uid.FieldUID).
//		Scan(ctx, &v)
//
func (uq *UIDQuery) Select(fields ...string) *UIDSelect {
	uq.fields = append(uq.fields, fields...)
	selbuild := &UIDSelect{UIDQuery: uq}
	selbuild.label = uid.Label
	selbuild.flds, selbuild.scan = &uq.fields, selbuild.Scan
	return selbuild
}

func (uq *UIDQuery) prepareQuery(ctx context.Context) error {
	for _, f := range uq.fields {
		if !uid.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
		}
	}
	if uq.path != nil {
		prev, err := uq.path(ctx)
		if err != nil {
			return err
		}
		uq.sql = prev
	}
	return nil
}

func (uq *UIDQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*UID, error) {
	var (
		nodes       = []*UID{}
		withFKs     = uq.withFKs
		_spec       = uq.querySpec()
		loadedTypes = [2]bool{
			uq.withMessage != nil,
			uq.withMailbox != nil,
		}
	)
	if uq.withMessage != nil || uq.withMailbox != nil {
		withFKs = true
	}
	if withFKs {
		_spec.Node.Columns = append(_spec.Node.Columns, uid.ForeignKeys...)
	}
	_spec.ScanValues = func(columns []string) ([]interface{}, error) {
		return (*UID).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []interface{}) error {
		node := &UID{config: uq.config}
		nodes = append(nodes, node)
		node.Edges.loadedTypes = loadedTypes
		return node.assignValues(columns, values)
	}
	for i := range hooks {
		hooks[i](ctx, _spec)
	}
	if err := sqlgraph.QueryNodes(ctx, uq.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	if query := uq.withMessage; query != nil {
		if err := uq.loadMessage(ctx, query, nodes, nil,
			func(n *UID, e *Message) { n.Edges.Message = e }); err != nil {
			return nil, err
		}
	}
	if query := uq.withMailbox; query != nil {
		if err := uq.loadMailbox(ctx, query, nodes, nil,
			func(n *UID, e *Mailbox) { n.Edges.Mailbox = e }); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

func (uq *UIDQuery) loadMessage(ctx context.Context, query *MessageQuery, nodes []*UID, init func(*UID), assign func(*UID, *Message)) error {
	ids := make([]imap.InternalMessageID, 0, len(nodes))
	nodeids := make(map[imap.InternalMessageID][]*UID)
	for i := range nodes {
		if nodes[i].uid_message == nil {
			continue
		}
		fk := *nodes[i].uid_message
		if _, ok := nodeids[fk]; !ok {
			ids = append(ids, fk)
		}
		nodeids[fk] = append(nodeids[fk], nodes[i])
	}
	query.Where(message.IDIn(ids...))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		nodes, ok := nodeids[n.ID]
		if !ok {
			return fmt.Errorf(`unexpected foreign-key "uid_message" returned %v`, n.ID)
		}
		for i := range nodes {
			assign(nodes[i], n)
		}
	}
	return nil
}
func (uq *UIDQuery) loadMailbox(ctx context.Context, query *MailboxQuery, nodes []*UID, init func(*UID), assign func(*UID, *Mailbox)) error {
	ids := make([]imap.InternalMailboxID, 0, len(nodes))
	nodeids := make(map[imap.InternalMailboxID][]*UID)
	for i := range nodes {
		if nodes[i].mailbox_ui_ds == nil {
			continue
		}
		fk := *nodes[i].mailbox_ui_ds
		if _, ok := nodeids[fk]; !ok {
			ids = append(ids, fk)
		}
		nodeids[fk] = append(nodeids[fk], nodes[i])
	}
	query.Where(mailbox.IDIn(ids...))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		nodes, ok := nodeids[n.ID]
		if !ok {
			return fmt.Errorf(`unexpected foreign-key "mailbox_ui_ds" returned %v`, n.ID)
		}
		for i := range nodes {
			assign(nodes[i], n)
		}
	}
	return nil
}

func (uq *UIDQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := uq.querySpec()
	_spec.Node.Columns = uq.fields
	if len(uq.fields) > 0 {
		_spec.Unique = uq.unique != nil && *uq.unique
	}
	return sqlgraph.CountNodes(ctx, uq.driver, _spec)
}

func (uq *UIDQuery) sqlExist(ctx context.Context) (bool, error) {
	n, err := uq.sqlCount(ctx)
	if err != nil {
		return false, fmt.Errorf("ent: check existence: %w", err)
	}
	return n > 0, nil
}

func (uq *UIDQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := &sqlgraph.QuerySpec{
		Node: &sqlgraph.NodeSpec{
			Table:   uid.Table,
			Columns: uid.Columns,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeInt,
				Column: uid.FieldID,
			},
		},
		From:   uq.sql,
		Unique: true,
	}
	if unique := uq.unique; unique != nil {
		_spec.Unique = *unique
	}
	if fields := uq.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, uid.FieldID)
		for i := range fields {
			if fields[i] != uid.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
	}
	if ps := uq.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := uq.limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := uq.offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := uq.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return _spec
}

func (uq *UIDQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(uq.driver.Dialect())
	t1 := builder.Table(uid.Table)
	columns := uq.fields
	if len(columns) == 0 {
		columns = uid.Columns
	}
	selector := builder.Select(t1.Columns(columns...)...).From(t1)
	if uq.sql != nil {
		selector = uq.sql
		selector.Select(selector.Columns(columns...)...)
	}
	if uq.unique != nil && *uq.unique {
		selector.Distinct()
	}
	for _, p := range uq.predicates {
		p(selector)
	}
	for _, p := range uq.order {
		p(selector)
	}
	if offset := uq.offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := uq.limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// UIDGroupBy is the group-by builder for UID entities.
type UIDGroupBy struct {
	config
	selector
	fields []string
	fns    []AggregateFunc
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Aggregate adds the given aggregation functions to the group-by query.
func (ugb *UIDGroupBy) Aggregate(fns ...AggregateFunc) *UIDGroupBy {
	ugb.fns = append(ugb.fns, fns...)
	return ugb
}

// Scan applies the group-by query and scans the result into the given value.
func (ugb *UIDGroupBy) Scan(ctx context.Context, v interface{}) error {
	query, err := ugb.path(ctx)
	if err != nil {
		return err
	}
	ugb.sql = query
	return ugb.sqlScan(ctx, v)
}

func (ugb *UIDGroupBy) sqlScan(ctx context.Context, v interface{}) error {
	for _, f := range ugb.fields {
		if !uid.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("invalid field %q for group-by", f)}
		}
	}
	selector := ugb.sqlQuery()
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := ugb.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

func (ugb *UIDGroupBy) sqlQuery() *sql.Selector {
	selector := ugb.sql.Select()
	aggregation := make([]string, 0, len(ugb.fns))
	for _, fn := range ugb.fns {
		aggregation = append(aggregation, fn(selector))
	}
	// If no columns were selected in a custom aggregation function, the default
	// selection is the fields used for "group-by", and the aggregation functions.
	if len(selector.SelectedColumns()) == 0 {
		columns := make([]string, 0, len(ugb.fields)+len(ugb.fns))
		for _, f := range ugb.fields {
			columns = append(columns, selector.C(f))
		}
		columns = append(columns, aggregation...)
		selector.Select(columns...)
	}
	return selector.GroupBy(selector.Columns(ugb.fields...)...)
}

// UIDSelect is the builder for selecting fields of UID entities.
type UIDSelect struct {
	*UIDQuery
	selector
	// intermediate query (i.e. traversal path).
	sql *sql.Selector
}

// Scan applies the selector query and scans the result into the given value.
func (us *UIDSelect) Scan(ctx context.Context, v interface{}) error {
	if err := us.prepareQuery(ctx); err != nil {
		return err
	}
	us.sql = us.UIDQuery.sqlQuery(ctx)
	return us.sqlScan(ctx, v)
}

func (us *UIDSelect) sqlScan(ctx context.Context, v interface{}) error {
	rows := &sql.Rows{}
	query, args := us.sql.Query()
	if err := us.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}
