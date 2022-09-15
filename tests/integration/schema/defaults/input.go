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

var inputObjectAuthorFilterArg = map[string]any{
	"kind": "INPUT_OBJECT",
	"name": "authorFilterArg",
}

func makeOuterTypeImplicit(name, typeName, ofType any) Field {
	return makeOuterType(Yes(name), Yes(typeName), Yes(ofType), No())
}

// The default known input type fields for order.
var orderInputFieldsStatic = []any{
	makeOuterTypeImplicit("_key", "Ordering", nil),

	//	makeOrderingInput("age"),
	//	makeOrderingInput("name"),
	//	makeOrderingInput("verified"),
	//	makeOuterType(Yes("wrote"), Yes("bookOrderArg"), Yes(nil), No()),
	//	makeOrderingInput("wrote_id"),
}

// The default known input type fields for filter
var filterInputFieldsStatic = []any{
	makeOuterTypeImplicit("_and", nil, inputObjectAuthorFilterArg),
	makeOuterTypeImplicit("_key", "IDOperatorBlock", nil),
	makeOuterTypeImplicit("_not", "authorFilterArg", nil),
	makeOuterTypeImplicit("_or", nil, inputObjectAuthorFilterArg),

	makeOuterTypeImplicit("age", "IntOperatorBlock", nil),
	makeOuterTypeImplicit("name", "StringOperatorBlock", nil),
	makeOuterTypeImplicit("verified", "BooleanOperatorBlock", nil),
	makeOuterTypeImplicit("wrote", "bookFilterArg", nil),
	makeOuterTypeImplicit("wrote_id", "IDOperatorBlock", nil),
}

// BuildAllOrderInputFields builds order related dynamic fields with the
// known static / default fields.
func BuildAllOrderInputFields(objectName string, fields []ArgDef) Field {
	// Start from our already known default fields, and keep building.
	allInputFields := orderInputFieldsStatic

	// Add the dynamic fields.
	for _, field := range fields {
		allInputFields = append(
			allInputFields,
			makeOuterType(Yes(field.FieldName), Yes(field.TypeName), Yes(nil), No()),
		)
	}

	return makeOuterType(Yes("order"), Yes(objectName+"OrderArg"), Yes(nil), Yes(allInputFields))
}

// BuildAllFilterInputFields builds all filter related dynamic fields with the
// known static / default fields.
func BuildAllFilterInputFields(objectName string, fields []ArgDef) Field {
	// Start from our already known default fields, and keep building.
	allInputFields := filterInputFieldsStatic

	for _, field := range fields {
		allInputFields = append(
			allInputFields,
			makeOuterType(Yes(field.FieldName), Yes(field.TypeName), Yes(nil), No()),
		)
	}

	return makeOuterType(Yes("filter"), Yes(objectName+"FilterArg"), Yes(nil), Yes(allInputFields))
}
