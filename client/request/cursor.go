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

// CursorSelect represents a _cursor block containing a single collection query.
type CursorSelect struct {
	Field

	// Select is the inner collection query (e.g., User { name }).
	Select *Select
}

// Validate ensures exactly one collection query is present.
func (c *CursorSelect) Validate() []error {
	result := []error{}

	if c.Select == nil {
		result = append(result, ErrCursorMustContainQuery)
		return result
	}

	result = append(result, c.Select.Validate()...)

	return result
}
