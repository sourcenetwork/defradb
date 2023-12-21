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
	"encoding/json"

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

// selectJson is a private object used for handling json deserialization
// of `Select` objects.
type selectJson struct {
	Field
	DocKeys     immutable.Option[[]string]
	CID         immutable.Option[string]
	Root        SelectionType
	Limit       immutable.Option[uint64]
	Offset      immutable.Option[uint64]
	OrderBy     immutable.Option[OrderBy]
	GroupBy     immutable.Option[GroupBy]
	Filter      immutable.Option[Filter]
	ShowDeleted bool

	// Properties above this line match the `Select` object and
	// are deserialized using the normal/default logic.
	// Properties below this line require custom logic in `UnmarshalJSON`
	// in order to be deserialized correctly.

	Fields []map[string]json.RawMessage
}

func (s *Select) UnmarshalJSON(bytes []byte) error {
	var selectMap selectJson
	err := json.Unmarshal(bytes, &selectMap)
	if err != nil {
		return err
	}

	s.Field = selectMap.Field
	s.DocKeys = selectMap.DocKeys
	s.CID = selectMap.CID
	s.Root = selectMap.Root
	s.Limit = selectMap.Limit
	s.Offset = selectMap.Offset
	s.OrderBy = selectMap.OrderBy
	s.GroupBy = selectMap.GroupBy
	s.Filter = selectMap.Filter
	s.ShowDeleted = selectMap.ShowDeleted
	s.Fields = make([]Selection, len(selectMap.Fields))

	for i, field := range selectMap.Fields {
		fieldJson, err := json.Marshal(field)
		if err != nil {
			return err
		}

		var fieldValue Selection
		// We detect which concrete type each `Selection` object is by detecting
		// non-nillable fields, if the key is present it must be of that type.
		// They must be non-nillable as nil values may have their keys omitted from
		// the json. This also relies on the fields being unique.  We may wish to change
		// this later to custom-serialize with a `_type` property.
		if _, ok := field["Root"]; ok {
			// This must be a Select, as only the `Select` type has a `Root` field
			var fieldSelect Select
			err := json.Unmarshal(fieldJson, &fieldSelect)
			if err != nil {
				return err
			}
			fieldValue = &fieldSelect
		} else if _, ok := field["Targets"]; ok {
			// This must be an Aggregate, as only the `Aggregate` type has a `Targets` field
			var fieldAggregate Aggregate
			err := json.Unmarshal(fieldJson, &fieldAggregate)
			if err != nil {
				return err
			}
			fieldValue = &fieldAggregate
		} else {
			// This must be a Field
			var fieldField Field
			err := json.Unmarshal(fieldJson, &fieldField)
			if err != nil {
				return err
			}
			fieldValue = &fieldField
		}

		s.Fields[i] = fieldValue
	}

	return nil
}
