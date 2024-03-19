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
)

// Select is a complex Field with strong typing.
// It is used for sub-types in a request.
// Includes fields, and request arguments like filters, limits, etc.
type Select struct {
	Field
	ChildSelect

	Limitable
	Offsetable
	Orderable
	Filterable
	DocIDsFilter
	CIDFilter
	Groupable

	// ShowDeleted will return deleted documents along with non-deleted ones
	// if set to true.
	ShowDeleted bool
}

// ChildSelect represents a type with selectable child properties.
//
// At least one child must be selected.
type ChildSelect struct {
	// Fields contains the set of child properties to return.
	//
	// At least one child propertt must be selected.
	Fields []Selection
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
// of [Select] objects.
//
// It contains everything minus the [ChildSelect], which uses a custom UnmarshalJSON
// and is skipped over when embedding due to the way the std lib json pkg works.
type selectJson struct {
	Field
	Limitable
	Offsetable
	Orderable
	Filterable
	DocIDsFilter
	CIDFilter
	Groupable
	ShowDeleted bool
}

func (s *Select) UnmarshalJSON(bytes []byte) error {
	var selectMap selectJson
	err := json.Unmarshal(bytes, &selectMap)
	if err != nil {
		return err
	}

	s.Field = selectMap.Field
	s.DocIDs = selectMap.DocIDs
	s.CID = selectMap.CID
	s.Limitable = selectMap.Limitable
	s.Offsetable = selectMap.Offsetable
	s.Orderable = selectMap.Orderable
	s.Groupable = selectMap.Groupable
	s.Filterable = selectMap.Filterable
	s.ShowDeleted = selectMap.ShowDeleted

	var childSelect ChildSelect
	err = json.Unmarshal(bytes, &childSelect)
	if err != nil {
		return err
	}

	s.ChildSelect = childSelect

	return nil
}

// childSelectJson is a private object used for handling json deserialization
// of [ChildSelect] objects.
type childSelectJson struct {
	Fields []map[string]json.RawMessage
}

func (s *ChildSelect) UnmarshalJSON(bytes []byte) error {
	var selectMap childSelectJson
	err := json.Unmarshal(bytes, &selectMap)
	if err != nil {
		return err
	}

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
		if _, ok := field["Fields"]; ok {
			// This must be a Select, as only the `Select` type has a `Fields` field
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
