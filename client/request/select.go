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

import (
	"github.com/sourcenetwork/immutable"
)

// SelectionType is the type of selection.
type SelectionType int

const (
	ObjectSelection SelectionType = iota
	CommitSelection
)

// Select is a complex Field with strong typing.
// It is used for sub-types in a request.
// Includes fields, and request arguments like filters, limits, etc.
type Select struct {
	Field

	DocKeys immutable.Option[[]string]
	CID     immutable.Option[string]

	// Root is the top level type of parsed request
	Root SelectionType

	Limit   immutable.Option[uint64]
	Offset  immutable.Option[uint64]
	OrderBy immutable.Option[OrderBy]
	GroupBy immutable.Option[GroupBy]
	Filter  immutable.Option[Filter]

	Fields []Selection

	ShowDeleted bool
}

// Validate validates the Select.
func (s *Select) Validate() []error {
	result := []error{}

	result = append(result, s.validateShallow()...)

	for _, childSelection := range s.Fields {
		switch typedChildSelection := childSelection.(type) {
		case *Select:
			result = append(result, typedChildSelection.validateShallow()...)
		default:
			// Do nothing
		}
	}

	return result
}

func (s *Select) validateShallow() []error {
	result := []error{}

	result = append(result, s.validateGroupBy()...)

	return result
}

func (s *Select) validateGroupBy() []error {
	result := []error{}

	if !s.GroupBy.HasValue() {
		return result
	}

	for _, childSelection := range s.Fields {
		switch typedChildSelection := childSelection.(type) {
		case *Field:
			if typedChildSelection.Name == TypeNameFieldName {
				// _typeName is permitted
				continue
			}

			var fieldExistsInGroupBy bool
			var isAliasFieldInGroupBy bool
			for _, groupByField := range s.GroupBy.Value().Fields {
				if typedChildSelection.Name == groupByField {
					fieldExistsInGroupBy = true
					break
				} else if typedChildSelection.Name == groupByField+RelatedObjectID {
					isAliasFieldInGroupBy = true
					break
				}
			}
			if !fieldExistsInGroupBy && !isAliasFieldInGroupBy {
				result = append(result, NewErrSelectOfNonGroupField(typedChildSelection.Name))
			}
		default:
			// Do nothing
		}
	}

	return result
}
