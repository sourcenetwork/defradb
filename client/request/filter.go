// Copyright 2022 Democratized Data Foundation
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

const (
	FilterOpOr  = "_or"
	FilterOpAnd = "_and"
	FilterOpNot = "_not"
)

// Filter contains the parsed condition map to be
// run by the Filter Evaluator.
// @todo: Cache filter structure for faster condition
// evaluation.
type Filter struct {
	// parsed filter conditions
	Conditions map[string]any
}

// Filterable is an embeddable struct that hosts a consistent set of properties
// for filtering an aspect of a request.
type Filterable struct {
	// OrderBy is an optional set of conditions used to filter records prior to
	// being processed by the request.
	Filter immutable.Option[Filter]
}

// FieldReference is a filter condition value that names another field of the same document
// instead of a literal value, e.g. `expectedChunks` in `{chunkCount: {_lt: expectedChunks}}`.
//
// It is produced when a filter operator is given a bare, unquoted identifier.
type FieldReference struct {
	// Name is the name of the referenced field.
	Name string
}

// CollectFieldReferences returns the names of the fields referenced by field-to-field
// comparisons in the given filter conditions.
//
// References written inside a relation sub-filter are not returned, as they would name a
// field of a different document.  Those are rejected when the filter is run.
func CollectFieldReferences(conditions map[string]any) []string {
	return collectFieldReferences(conditions, nil)
}

func collectFieldReferences(conditions map[string]any, names []string) []string {
	for key, value := range conditions {
		switch key {
		case FilterOpAnd, FilterOpOr:
			clauses, ok := value.([]any)
			if !ok {
				continue
			}
			for _, clause := range clauses {
				if inner, ok := clause.(map[string]any); ok {
					names = collectFieldReferences(inner, names)
				}
			}

		case FilterOpNot, AliasFieldName:
			if inner, ok := value.(map[string]any); ok {
				names = collectFieldReferences(inner, names)
			}

		default:
			// The value of a field key is a block of operators, e.g. `{_lt: expectedChunks}`.
			// A nested map keyed by field names instead belongs to a related document, and is
			// deliberately not descended into.
			operators, ok := value.(map[string]any)
			if !ok {
				continue
			}
			for _, operand := range operators {
				if reference, ok := operand.(FieldReference); ok {
					names = append(names, reference.Name)
				}
			}
		}
	}
	return names
}
