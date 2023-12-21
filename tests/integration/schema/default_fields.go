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

type Field = map[string]any
type fields []Field

func concat(fieldSets ...fields) fields {
	result := fields{}
	for _, fieldSet := range fieldSets {
		result = append(result, fieldSet...)
	}
	return result
}

// Append appends the given Field onto a shallow clone
// of the given fieldset.
func (fieldSet fields) Append(field Field) fields {
	result := make(fields, len(fieldSet)+1)
	copy(result, fieldSet)

	result[len(result)-1] = field
	return result
}

// Tidy sorts and casts the given fieldset into a format suitable
// for comparing against introspection result fields.
func (fieldSet fields) Tidy() []any {
	return fieldSet.sort().array()
}

func (fieldSet fields) sort() fields {
	sort.Slice(fieldSet, func(i, j int) bool {
		return fieldSet[i]["name"].(string) < fieldSet[j]["name"].(string)
	})
	return fieldSet
}

func (fieldSet fields) array() []any {
	result := make([]any, len(fieldSet))
	for i, v := range fieldSet {
		result[i] = v
	}
	return result
}

// DefaultFields contains the list of fields every
// defra schema-object should have.
var DefaultFields = concat(
	fields{
		keyField,
		versionField,
		groupField,
		deletedField,
	},
	aggregateFields,
)

// DefaultEmbeddedObjFields contains the list of fields every
// defra embedded-object should have.
var DefaultEmbeddedObjFields = concat(
	fields{
		groupField,
	},
	aggregateFields,
)

var keyField = Field{
	"name": "_key",
	"type": map[string]any{
		"kind": "SCALAR",
		"name": "ID",
	},
}

var deletedField = Field{
	"name": "_deleted",
	"type": map[string]any{
		"kind": "SCALAR",
		"name": "Boolean",
	},
}

var versionField = Field{
	"name": "_version",
	"type": map[string]any{
		"kind": "LIST",
		"name": nil,
	},
}

var groupField = Field{
	"name": "_group",
	"type": map[string]any{
		"kind": "LIST",
		"name": nil,
	},
}

var aggregateFields = fields{
	map[string]any{
		"name": "_avg",
		"type": map[string]any{
			"kind": "SCALAR",
			"name": "Float",
		},
	},
	map[string]any{
		"name": "_count",
		"type": map[string]any{
			"kind": "SCALAR",
			"name": "Int",
		},
	},
	map[string]any{
		"name": "_sum",
		"type": map[string]any{
			"kind": "SCALAR",
			"name": "Float",
		},
	},
}

var cidArg = Field{
	"name": "cid",
	"type": map[string]any{
		"name":        "String",
		"inputFields": nil,
	},
}
var dockeyArg = Field{
	"name": "dockey",
	"type": map[string]any{
		"name":        "String",
		"inputFields": nil,
	},
}
var dockeysArg = Field{
	"name": "dockeys",
	"type": map[string]any{
		"name":        nil,
		"inputFields": nil,
		"ofType": map[string]any{
			"kind": "NON_NULL",
			"name": nil,
		},
	},
}
var showDeletedArg = Field{
	"name": "showDeleted",
	"type": map[string]any{
		"name":        "Boolean",
		"inputFields": nil,
	},
}

var groupByArg = Field{
	"name": "groupBy",
	"type": map[string]any{
		"name":        nil,
		"inputFields": nil,
		"ofType": map[string]any{
			"kind": "NON_NULL",
			"name": nil,
		},
	},
}

var limitArg = Field{
	"name": "limit",
	"type": map[string]any{
		"name":        "Int",
		"inputFields": nil,
		"ofType":      nil,
	},
}

var offsetArg = Field{
	"name": "offset",
	"type": map[string]any{
		"name":        "Int",
		"inputFields": nil,
		"ofType":      nil,
	},
}

type argDef struct {
	fieldName string
	typeName  string
}

func buildOrderArg(objectName string, fields []argDef) Field {
	inputFields := []any{
		makeInputObject("_key", "Ordering", nil),
	}

	for _, field := range fields {
		inputFields = append(inputFields, makeInputObject(field.fieldName, field.typeName, nil))
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

func buildFilterArg(objectName string, fields []argDef) Field {
	filterArgName := objectName + "FilterArg"

	inputFields := []any{
		makeInputObject("_and", nil, map[string]any{
			"kind": "INPUT_OBJECT",
			"name": filterArgName,
		}),
		makeInputObject("_key", "IDOperatorBlock", nil),
		makeInputObject("_not", filterArgName, nil),
		makeInputObject("_or", nil, map[string]any{
			"kind": "INPUT_OBJECT",
			"name": filterArgName,
		}),
	}

	for _, field := range fields {
		inputFields = append(inputFields, makeInputObject(field.fieldName, field.typeName, nil))
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

// trimField creates a new object using the provided defaults, but only containing
// the provided properties. Function is recursive and will respect inner properties.
func trimField(fullDefault Field, properties map[string]any) Field {
	result := Field{}
	for key, children := range properties {
		switch childProps := children.(type) {
		case map[string]any:
			fullValue := fullDefault[key]
			var value any
			if fullValue == nil {
				value = nil
			} else if fullField, isField := fullValue.(Field); isField {
				value = trimField(fullField, childProps)
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

// trimFields creates a new slice of new objects using the provided defaults, but only containing
// the provided properties. Function is recursive and will respect inner prop properties.
func trimFields(fullDefaultFields fields, properties map[string]any) fields {
	result := fields{}
	for _, field := range fullDefaultFields {
		result = append(result, trimField(field, properties))
	}
	return result
}

// makeInputObject retrned a properly made input field type
// using name (outer), name of type (inner), and types ofType.
func makeInputObject(
	name string,
	typeName any,
	ofType any,
) map[string]any {
	return map[string]any{
		"name": name,
		"type": map[string]any{
			"name":   typeName,
			"ofType": ofType,
		},
	}
}
