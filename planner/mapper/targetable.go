// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package mapper

import (
	"github.com/sourcenetwork/defradb/connor"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/immutables"
)

var (
	_ connor.FilterKey = (*PropertyIndex)(nil)
	_ connor.FilterKey = (*Operator)(nil)
)

// PropertyIndex is a FilterKey that represents a property in a document.
type PropertyIndex struct {
	// The index at which the target property can be found on its parent.
	Index int
}

func (k *PropertyIndex) GetProp(data any) any {
	if data == nil {
		return nil
	}

	return data.(core.Doc).Fields[k.Index]
}

func (k *PropertyIndex) GetOperatorOrDefault(defaultOp string) string {
	return defaultOp
}

func (k *PropertyIndex) Equal(other connor.FilterKey) bool {
	if otherKey, isOk := other.(*PropertyIndex); isOk && *k == *otherKey {
		return true
	}
	return false
}

// Operator is a FilterKey that represents a filter operator.
type Operator struct {
	// The filter operation string that this Operator represents.
	//
	// E.g. "_eq", or "_and".
	Operation string
}

func (k *Operator) GetProp(data any) any {
	return data
}

func (k *Operator) GetOperatorOrDefault(defaultOp string) string {
	return k.Operation
}

func (k *Operator) Equal(other connor.FilterKey) bool {
	if otherKey, isOk := other.(*Operator); isOk && *k == *otherKey {
		return true
	}
	return false
}

// Filter represents a series of conditions that may reduce the number of
// records that a query returns.
type Filter struct {
	// The filter conditions that must pass in order for a record to be returned.
	Conditions map[connor.FilterKey]any

	// The filter conditions in human-readable form.
	ExternalConditions map[string]any
}

func NewFilter() *Filter {
	return &Filter{
		Conditions: map[connor.FilterKey]any{},
	}
}

// Limit represents a limit-offset pairing that controls how many
// and which records will be returned from a query.
type Limit struct {
	// The maximum number of records that can be returned from a query.
	Limit uint64

	// The offset from which counting towards the Limit will begin.
	// Before records before the Offset will not be returned.
	Offset uint64
}

// GroupBy represents a grouping instruction on a query.
type GroupBy struct {
	// The indexes of fields by which documents should be grouped. Ordered.
	Fields []Field
}

type SortDirection string

const (
	ASC  SortDirection = "ASC"
	DESC SortDirection = "DESC"
)

// OrderCondition represents a single property by which query results should
// be ordered, and the direction in which they should be ordered.
type OrderCondition struct {
	// A chain of field indexes by which the property to sort by may be found.
	// This is relative to the host/defining object and may traverse through
	// multiple object layers.
	FieldIndexes []int

	// The direction in which the sort should be applied.
	Direction SortDirection
}

type OrderBy struct {
	Conditions []OrderCondition
}

// Targetable represents a targetable property.
type Targetable struct {
	// The basic field information of this property.
	Field

	// A optional collection of docKeys that can be specified to restrict results
	// to belonging to this set.
	DocKeys immutables.Option[[]string]

	// An optional filter, that can be specified to restrict results to documents
	// that satisfies all of its conditions.
	Filter *Filter

	// An optional limit, that can be specified to restrict the number and location
	// of documents returned.
	Limit *Limit

	// An optional grouping clause, that can be specified to group results by property
	// value.
	GroupBy *GroupBy

	// An optional order clause, that can be specified to order results by property
	// value
	OrderBy *OrderBy
}

func (t *Targetable) cloneTo(index int) *Targetable {
	return &Targetable{
		Field:   *t.Field.cloneTo(index),
		DocKeys: t.DocKeys,
		Filter:  t.Filter,
		Limit:   t.Limit,
		GroupBy: t.GroupBy,
		OrderBy: t.OrderBy,
	}
}

func (t *Targetable) AsTargetable() (*Targetable, bool) {
	return t, true
}
