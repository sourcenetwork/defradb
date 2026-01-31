// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package request

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

	result = append(result, c.Select.Validate()...)

	return result
}
