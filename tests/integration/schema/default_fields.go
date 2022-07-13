// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package schema

import "sort"

type field = map[string]interface{}
type fields []field

func concat(fieldSets ...fields) fields {
	result := fields{}
	for _, fieldSet := range fieldSets {
		result = append(result, fieldSet...)
	}
	return result
}

// append appends the given field onto a shallow clone
// of the given fieldset.
func (fieldSet fields) append(field field) fields {
	result := make(fields, len(fieldSet)+1)
	copy(result, fieldSet)

	result[len(result)-1] = field
	return result
}

// tidy sorts and casts the given fieldset into a format suitable
// for comparing against introspection result fields.
func (fieldSet fields) tidy() []interface{} {
	return fieldSet.sort().array()
}

func (fieldSet fields) sort() fields {
	sort.Slice(fieldSet, func(i, j int) bool {
		return fieldSet[i]["name"].(string) < fieldSet[j]["name"].(string)
	})
	return fieldSet
}

func (fieldSet fields) array() []interface{} {
	result := make([]interface{}, len(fieldSet))
	for i, v := range fieldSet {
		result[i] = v
	}
	return result
}

// defaultFields contains the list of fields every
// defra schema-object should have.
var defaultFields = concat(
	fields{
		keyField,
		versionField,
		groupField,
	},
	aggregateFields,
)

var keyField = field{
	"name": "_key",
	"type": map[string]interface{}{
		"kind": "SCALAR",
		"name": "ID",
	},
}

var versionField = field{
	"name": "_version",
	"type": map[string]interface{}{
		"kind": "LIST",
		"name": nil,
	},
}

var groupField = field{
	"name": "_group",
	"type": map[string]interface{}{
		"kind": "LIST",
		"name": nil,
	},
}

var aggregateFields = fields{
	map[string]interface{}{
		"name": "_avg",
		"type": map[string]interface{}{
			"kind": "SCALAR",
			"name": "Float",
		},
	},
	map[string]interface{}{
		"name": "_count",
		"type": map[string]interface{}{
			"kind": "SCALAR",
			"name": "Int",
		},
	},
	map[string]interface{}{
		"name": "_sum",
		"type": map[string]interface{}{
			"kind": "SCALAR",
			"name": "Float",
		},
	},
}
