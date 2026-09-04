// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package request

import "github.com/sourcenetwork/immutable"

// PageInfoSelect represents the _pageInfo selection within a cursor query.
type PageInfoSelect struct {
	HasNext     bool
	HasPrev     bool
	StartCursor bool
	EndCursor   bool
}

// CursorSelect represents a _cursor block containing a single collection query.
type CursorSelect struct {
	Field

	// Select is the inner collection query (e.g., User { name }).
	Select *Select

	Firstable
	Afterable
	Lastable
	Beforeable

	// PageInfoSelect captures which _pageInfo fields were selected (nil if not selected).
	PageInfoSelect *PageInfoSelect
}

// Validate ensures exactly one collection query is present and validates parameters.
func (c *CursorSelect) Validate() []error {
	result := []error{}

	if c.Select == nil {
		result = append(result, ErrCursorMustContainQuery)
		return result
	}

	if c.After.HasValue() && c.After.Value() == "" {
		result = append(result, ErrInvalidCursor)
	}

	if c.Before.HasValue() && c.Before.Value() == "" {
		result = append(result, ErrInvalidCursor)
	}

	hasForward := c.First.HasValue() || c.After.HasValue()
	hasBackward := c.Last.HasValue() || c.Before.HasValue()
	if hasForward && hasBackward {
		result = append(result, ErrForwardBackwardConflict)
	}

	result = append(result, c.Select.Validate()...)

	return result
}

// Firstable is an embeddable struct for forward cursor-based pagination.
type Firstable struct {
	First immutable.Option[uint64]
}

// Afterable is an embeddable struct for forward cursor-based pagination.
type Afterable struct {
	After immutable.Option[string]
}

// Lastable is an embeddable struct for backward cursor-based pagination.
type Lastable struct {
	Last immutable.Option[uint64]
}

// Beforeable is an embeddable struct for backward cursor-based pagination.
type Beforeable struct {
	Before immutable.Option[string]
}
