// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package planner

import (
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"

	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/immutable/enumerable"
)

type number interface {
	int64 | float64
}

func lessN[T number](a T, b T) bool {
	return a < b
}

func lessO[T number](a immutable.Option[T], b immutable.Option[T]) bool {
	if !a.HasValue() {
		return true
	}

	if !b.HasValue() {
		return false
	}

	return a.Value() < b.Value()
}

func reverse[T any](original func(T, T) bool) func(T, T) bool {
	return func(t1, t2 T) bool {
		return !original(t1, t2)
	}
}

// reduces the documents in a slice, skipping over hidden items (a grouping mechanic).
//
// Docs should be reduced with this function to avoid applying offsets twice (once in the
// select, then once here).
func reduceDocs[T any](
	docs []core.Doc,
	initialValue T,
	reduceFunc func(core.Doc, T) T,
) T {
	var value = initialValue
	for _, doc := range docs {
		if !doc.Hidden {
			value = reduceFunc(doc, value)
		}
	}
	return value
}

func reduceItems[T any, V any](
	source []T,
	aggregateTarget *mapper.AggregateTarget,
	less func(T, T) bool,
	initialValue V,
	reduceFunc func(T, V) V,
) (V, error) {
	items := enumerable.New(source)
	if aggregateTarget.Filter != nil {
		items = enumerable.Where(items, func(item T) (bool, error) {
			return mapper.RunFilter(item, aggregateTarget.Filter)
		})
	}

	if aggregateTarget.OrderBy != nil && len(aggregateTarget.OrderBy.Conditions) > 0 {
		if aggregateTarget.OrderBy.Conditions[0].Direction == mapper.ASC {
			items = enumerable.Sort(items, less, len(source))
		} else {
			items = enumerable.Sort(items, reverse(less), len(source))
		}
	}

	if aggregateTarget.Limit != nil {
		items = enumerable.Skip(items, aggregateTarget.Limit.Offset)
		items = enumerable.Take(items, aggregateTarget.Limit.Limit)
	}

	var value = initialValue
	err := enumerable.ForEach(items, func(item T) {
		value = reduceFunc(item, value)
	})
	return value, err
}
