// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"fmt"
	"math"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/ProtonMail/gluon/internal/db/ent/predicate"
	"github.com/ProtonMail/gluon/internal/db/ent/uidvalidity"
)

// UIDValidityQuery is the builder for querying UIDValidity entities.
type UIDValidityQuery struct {
	config
	limit      *int
	offset     *int
	unique     *bool
	order      []OrderFunc
	fields     []string
	predicates []predicate.UIDValidity
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the UIDValidityQuery builder.
func (uvq *UIDValidityQuery) Where(ps ...predicate.UIDValidity) *UIDValidityQuery {
	uvq.predicates = append(uvq.predicates, ps...)
	return uvq
}

// Limit adds a limit step to the query.
func (uvq *UIDValidityQuery) Limit(limit int) *UIDValidityQuery {
	uvq.limit = &limit
	return uvq
}

// Offset adds an offset step to the query.
func (uvq *UIDValidityQuery) Offset(offset int) *UIDValidityQuery {
	uvq.offset = &offset
	return uvq
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (uvq *UIDValidityQuery) Unique(unique bool) *UIDValidityQuery {
	uvq.unique = &unique
	return uvq
}

// Order adds an order step to the query.
func (uvq *UIDValidityQuery) Order(o ...OrderFunc) *UIDValidityQuery {
	uvq.order = append(uvq.order, o...)
	return uvq
}

// First returns the first UIDValidity entity from the query.
// Returns a *NotFoundError when no UIDValidity was found.
func (uvq *UIDValidityQuery) First(ctx context.Context) (*UIDValidity, error) {
	nodes, err := uvq.Limit(1).All(ctx)
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{uidvalidity.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (uvq *UIDValidityQuery) FirstX(ctx context.Context) *UIDValidity {
	node, err := uvq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first UIDValidity ID from the query.
// Returns a *NotFoundError when no UIDValidity ID was found.
func (uvq *UIDValidityQuery) FirstID(ctx context.Context) (id int, err error) {
	var ids []int
	if ids, err = uvq.Limit(1).IDs(ctx); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{uidvalidity.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (uvq *UIDValidityQuery) FirstIDX(ctx context.Context) int {
	id, err := uvq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single UIDValidity entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one UIDValidity entity is found.
// Returns a *NotFoundError when no UIDValidity entities are found.
func (uvq *UIDValidityQuery) Only(ctx context.Context) (*UIDValidity, error) {
	nodes, err := uvq.Limit(2).All(ctx)
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{uidvalidity.Label}
	default:
		return nil, &NotSingularError{uidvalidity.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (uvq *UIDValidityQuery) OnlyX(ctx context.Context) *UIDValidity {
	node, err := uvq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only UIDValidity ID in the query.
// Returns a *NotSingularError when more than one UIDValidity ID is found.
// Returns a *NotFoundError when no entities are found.
func (uvq *UIDValidityQuery) OnlyID(ctx context.Context) (id int, err error) {
	var ids []int
	if ids, err = uvq.Limit(2).IDs(ctx); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{uidvalidity.Label}
	default:
		err = &NotSingularError{uidvalidity.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (uvq *UIDValidityQuery) OnlyIDX(ctx context.Context) int {
	id, err := uvq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of UIDValidities.
func (uvq *UIDValidityQuery) All(ctx context.Context) ([]*UIDValidity, error) {
	if err := uvq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	return uvq.sqlAll(ctx)
}

// AllX is like All, but panics if an error occurs.
func (uvq *UIDValidityQuery) AllX(ctx context.Context) []*UIDValidity {
	nodes, err := uvq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of UIDValidity IDs.
func (uvq *UIDValidityQuery) IDs(ctx context.Context) ([]int, error) {
	var ids []int
	if err := uvq.Select(uidvalidity.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (uvq *UIDValidityQuery) IDsX(ctx context.Context) []int {
	ids, err := uvq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (uvq *UIDValidityQuery) Count(ctx context.Context) (int, error) {
	if err := uvq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return uvq.sqlCount(ctx)
}

// CountX is like Count, but panics if an error occurs.
func (uvq *UIDValidityQuery) CountX(ctx context.Context) int {
	count, err := uvq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (uvq *UIDValidityQuery) Exist(ctx context.Context) (bool, error) {
	if err := uvq.prepareQuery(ctx); err != nil {
		return false, err
	}
	return uvq.sqlExist(ctx)
}

// ExistX is like Exist, but panics if an error occurs.
func (uvq *UIDValidityQuery) ExistX(ctx context.Context) bool {
	exist, err := uvq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the UIDValidityQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (uvq *UIDValidityQuery) Clone() *UIDValidityQuery {
	if uvq == nil {
		return nil
	}
	return &UIDValidityQuery{
		config:     uvq.config,
		limit:      uvq.limit,
		offset:     uvq.offset,
		order:      append([]OrderFunc{}, uvq.order...),
		predicates: append([]predicate.UIDValidity{}, uvq.predicates...),
		// clone intermediate query.
		sql:    uvq.sql.Clone(),
		path:   uvq.path,
		unique: uvq.unique,
	}
}

// GroupBy is used to group vertices by one or more fields/columns.
// It is often used with aggregate functions, like: count, max, mean, min, sum.
//
// Example:
//
//	var v []struct {
//		UIDValidity imap.UID `json:"UIDValidity,omitempty"`
//		Count int `json:"count,omitempty"`
//	}
//
//	client.UIDValidity.Query().
//		GroupBy(uidvalidity.FieldUIDValidity).
//		Aggregate(ent.Count()).
//		Scan(ctx, &v)
func (uvq *UIDValidityQuery) GroupBy(field string, fields ...string) *UIDValidityGroupBy {
	grbuild := &UIDValidityGroupBy{config: uvq.config}
	grbuild.fields = append([]string{field}, fields...)
	grbuild.path = func(ctx context.Context) (prev *sql.Selector, err error) {
		if err := uvq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		return uvq.sqlQuery(ctx), nil
	}
	grbuild.label = uidvalidity.Label
	grbuild.flds, grbuild.scan = &grbuild.fields, grbuild.Scan
	return grbuild
}

// Select allows the selection one or more fields/columns for the given query,
// instead of selecting all fields in the entity.
//
// Example:
//
//	var v []struct {
//		UIDValidity imap.UID `json:"UIDValidity,omitempty"`
//	}
//
//	client.UIDValidity.Query().
//		Select(uidvalidity.FieldUIDValidity).
//		Scan(ctx, &v)
func (uvq *UIDValidityQuery) Select(fields ...string) *UIDValiditySelect {
	uvq.fields = append(uvq.fields, fields...)
	selbuild := &UIDValiditySelect{UIDValidityQuery: uvq}
	selbuild.label = uidvalidity.Label
	selbuild.flds, selbuild.scan = &uvq.fields, selbuild.Scan
	return selbuild
}

func (uvq *UIDValidityQuery) prepareQuery(ctx context.Context) error {
	for _, f := range uvq.fields {
		if !uidvalidity.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
		}
	}
	if uvq.path != nil {
		prev, err := uvq.path(ctx)
		if err != nil {
			return err
		}
		uvq.sql = prev
	}
	return nil
}

func (uvq *UIDValidityQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*UIDValidity, error) {
	var (
		nodes = []*UIDValidity{}
		_spec = uvq.querySpec()
	)
	_spec.ScanValues = func(columns []string) ([]interface{}, error) {
		return (*UIDValidity).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []interface{}) error {
		node := &UIDValidity{config: uvq.config}
		nodes = append(nodes, node)
		return node.assignValues(columns, values)
	}
	for i := range hooks {
		hooks[i](ctx, _spec)
	}
	if err := sqlgraph.QueryNodes(ctx, uvq.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	return nodes, nil
}

func (uvq *UIDValidityQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := uvq.querySpec()
	_spec.Node.Columns = uvq.fields
	if len(uvq.fields) > 0 {
		_spec.Unique = uvq.unique != nil && *uvq.unique
	}
	return sqlgraph.CountNodes(ctx, uvq.driver, _spec)
}

func (uvq *UIDValidityQuery) sqlExist(ctx context.Context) (bool, error) {
	n, err := uvq.sqlCount(ctx)
	if err != nil {
		return false, fmt.Errorf("ent: check existence: %w", err)
	}
	return n > 0, nil
}

func (uvq *UIDValidityQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := &sqlgraph.QuerySpec{
		Node: &sqlgraph.NodeSpec{
			Table:   uidvalidity.Table,
			Columns: uidvalidity.Columns,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeInt,
				Column: uidvalidity.FieldID,
			},
		},
		From:   uvq.sql,
		Unique: true,
	}
	if unique := uvq.unique; unique != nil {
		_spec.Unique = *unique
	}
	if fields := uvq.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, uidvalidity.FieldID)
		for i := range fields {
			if fields[i] != uidvalidity.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
	}
	if ps := uvq.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := uvq.limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := uvq.offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := uvq.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return _spec
}

func (uvq *UIDValidityQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(uvq.driver.Dialect())
	t1 := builder.Table(uidvalidity.Table)
	columns := uvq.fields
	if len(columns) == 0 {
		columns = uidvalidity.Columns
	}
	selector := builder.Select(t1.Columns(columns...)...).From(t1)
	if uvq.sql != nil {
		selector = uvq.sql
		selector.Select(selector.Columns(columns...)...)
	}
	if uvq.unique != nil && *uvq.unique {
		selector.Distinct()
	}
	for _, p := range uvq.predicates {
		p(selector)
	}
	for _, p := range uvq.order {
		p(selector)
	}
	if offset := uvq.offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := uvq.limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// UIDValidityGroupBy is the group-by builder for UIDValidity entities.
type UIDValidityGroupBy struct {
	config
	selector
	fields []string
	fns    []AggregateFunc
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Aggregate adds the given aggregation functions to the group-by query.
func (uvgb *UIDValidityGroupBy) Aggregate(fns ...AggregateFunc) *UIDValidityGroupBy {
	uvgb.fns = append(uvgb.fns, fns...)
	return uvgb
}

// Scan applies the group-by query and scans the result into the given value.
func (uvgb *UIDValidityGroupBy) Scan(ctx context.Context, v interface{}) error {
	query, err := uvgb.path(ctx)
	if err != nil {
		return err
	}
	uvgb.sql = query
	return uvgb.sqlScan(ctx, v)
}

func (uvgb *UIDValidityGroupBy) sqlScan(ctx context.Context, v interface{}) error {
	for _, f := range uvgb.fields {
		if !uidvalidity.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("invalid field %q for group-by", f)}
		}
	}
	selector := uvgb.sqlQuery()
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := uvgb.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

func (uvgb *UIDValidityGroupBy) sqlQuery() *sql.Selector {
	selector := uvgb.sql.Select()
	aggregation := make([]string, 0, len(uvgb.fns))
	for _, fn := range uvgb.fns {
		aggregation = append(aggregation, fn(selector))
	}
	// If no columns were selected in a custom aggregation function, the default
	// selection is the fields used for "group-by", and the aggregation functions.
	if len(selector.SelectedColumns()) == 0 {
		columns := make([]string, 0, len(uvgb.fields)+len(uvgb.fns))
		for _, f := range uvgb.fields {
			columns = append(columns, selector.C(f))
		}
		columns = append(columns, aggregation...)
		selector.Select(columns...)
	}
	return selector.GroupBy(selector.Columns(uvgb.fields...)...)
}

// UIDValiditySelect is the builder for selecting fields of UIDValidity entities.
type UIDValiditySelect struct {
	*UIDValidityQuery
	selector
	// intermediate query (i.e. traversal path).
	sql *sql.Selector
}

// Scan applies the selector query and scans the result into the given value.
func (uvs *UIDValiditySelect) Scan(ctx context.Context, v interface{}) error {
	if err := uvs.prepareQuery(ctx); err != nil {
		return err
	}
	uvs.sql = uvs.UIDValidityQuery.sqlQuery(ctx)
	return uvs.sqlScan(ctx, v)
}

func (uvs *UIDValiditySelect) sqlScan(ctx context.Context, v interface{}) error {
	rows := &sql.Rows{}
	query, args := uvs.sql.Query()
	if err := uvs.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}