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
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/immutables"
)

type SelectionType int

const (
	ObjectSelection SelectionType = iota
	CommitSelection
)

// Select is a complex Field with strong typing
// It used for sub types in a query. Includes
// fields, and query arguments like filters,
// limits, etc.
type Select struct {
	Field

	DocKeys immutables.Option[[]string]
	CID     immutables.Option[string]

	// Root is the top level query parsed type
	Root SelectionType

	Limit   immutables.Option[uint64]
	Offset  immutables.Option[uint64]
	OrderBy immutables.Option[OrderBy]
	GroupBy immutables.Option[GroupBy]
	Filter  immutables.Option[Filter]

	Fields []Selection
}

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
			for _, groupByField := range s.GroupBy.Value().Fields {
				if typedChildSelection.Name == groupByField {
					fieldExistsInGroupBy = true
					break
				}
			}
			if !fieldExistsInGroupBy {
				result = append(result, client.NewErrSelectOfNonGroupField(typedChildSelection.Name))
			}
		default:
			// Do nothing
		}
	}

	return result
}
