// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package defaults

import "sort"

type Field = map[string]interface{}
type Fields []Field

func Concat(fieldSets ...Fields) Fields {
	result := Fields{}
	for _, fieldSet := range fieldSets {
		result = append(result, fieldSet...)
	}
	return result
}

// Append appends the given field onto a shallow clone
// of the given fieldset.
func (fieldSet Fields) Append(field Field) Fields {
	result := make(Fields, len(fieldSet)+1)
	copy(result, fieldSet)

	result[len(result)-1] = field
	return result
}

// Tidy sorts and casts the given fieldset into a format suitable
// for comparing against introspection result fields.
func (fieldSet Fields) Tidy() []interface{} {
	return fieldSet.Sort().Array()
}

func (fieldSet Fields) Sort() Fields {
	sort.Slice(fieldSet, func(i, j int) bool {
		return fieldSet[i]["name"].(string) < fieldSet[j]["name"].(string)
	})
	return fieldSet
}

func (fieldSet Fields) Array() []interface{} {
	result := make([]interface{}, len(fieldSet))
	for i, v := range fieldSet {
		result[i] = v
	}
	return result
}

// DefaultFields contains the list of fields every
// defra schema-object should have.
var DefaultFields = Concat(
	Fields{
		keyField,
		versionField,
		groupField,
	},
	aggregateFields,
)

var keyField = Field{
	"name": "_key",
	"type": map[string]interface{}{
		"kind": "SCALAR",
		"name": "ID",
	},
}

var versionField = Field{
	"name": "_version",
	"type": map[string]interface{}{
		"kind": "LIST",
		"name": nil,
	},
}

var groupField = Field{
	"name": "_group",
	"type": map[string]interface{}{
		"kind": "LIST",
		"name": nil,
	},
}

var aggregateFields = Fields{
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

type ArgDef struct {
	FieldName string
	TypeName  string
}

func BuildOrderArg(objectName string, fields []ArgDef) Field {
	inputFields := []any{
		makeInputObject("_key", "Ordering", nil),
	}

	for _, field := range fields {
		inputFields = append(inputFields, makeInputObject(field.FieldName, field.TypeName, nil))
	}

	return Field{
		"name": "order",
		"type": Field{
			"name":        objectName + "OrderArg",
			"ofType":      nil,
			"inputFields": inputFields,
		},
	}
}

func BuildFilterArg(objectName string, fields []ArgDef) Field {
	filterArgName := objectName + "FilterArg"

	inputFields := []any{
		makeInputObject("_and", nil, map[string]any{
			"kind": "INPUT_OBJECT",
			"name": filterArgName,
		}),
		makeInputObject("_key", "IDOperatorBlock", nil),
		makeInputObject("_not", "authorFilterArg", nil),
		makeInputObject("_or", nil, map[string]any{
			"kind": "INPUT_OBJECT",
			"name": filterArgName,
		}),
	}

	for _, field := range fields {
		inputFields = append(inputFields, makeInputObject(field.FieldName, field.TypeName, nil))
	}

	return Field{
		"name": "filter",
		"type": Field{
			"name":        filterArgName,
			"ofType":      nil,
			"inputFields": inputFields,
		},
	}
}

// TrimField creates a new object using the provided defaults, but only containing
// the provided properties. Function is recursive and will respect inner properties.
func TrimField(fullDefault Field, properties map[string]any) Field {
	result := Field{}
	for key, children := range properties {
		switch childProps := children.(type) {
		case map[string]any:
			fullValue := fullDefault[key]
			var value any
			if fullValue == nil {
				value = nil
			} else if fullField, isField := fullValue.(Field); isField {
				value = TrimField(fullField, childProps)
			} else {
				value = fullValue
			}
			result[key] = value

		default:
			result[key] = fullDefault[key]
		}
	}
	return result
}

// TrimFields creates a new slice of new objects using the provided defaults, but only containing
// the provided properties. Function is recursive and will respect inner prop properties.
func TrimFields(fullDefaultFields Fields, properties map[string]any) Fields {
	result := Fields{}
	for _, field := range fullDefaultFields {
		result = append(result, TrimField(field, properties))
	}
	return result
}
