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

import "github.com/sourcenetwork/defradb/client"

func BuildFilterArg(objectName string, fields []ArgDef) Field {
	filterArgName := objectName + "FilterArg"

	inputFields := []any{
		makeInputObject(Yes("_and"), Yes(nil), Yes(map[string]any{
			"kind": "INPUT_OBJECT",
			"name": filterArgName,
		}), No()),
		makeInputObject(Yes("_key"), Yes("IDOperatorBlock"), Yes(nil), No()),
		makeInputObject(Yes("_not"), Yes("authorFilterArg"), Yes(nil), No()),
		makeInputObject(Yes("_or"), Yes(nil), Yes(map[string]any{
			"kind": "INPUT_OBJECT",
			"name": filterArgName,
		}), No()),
	}

	for _, field := range fields {
		inputFields = append(
			inputFields,
			makeInputObject(Yes(field.FieldName), Yes(field.TypeName), Yes(nil), No()),
		)
	}

	return makeInputObject(Yes("filter"), Yes(filterArgName), Yes(nil), Yes(inputFields))
}

func makeAuthorOrdering(name any) Field {
	return makeInputObject(Yes(name), Yes("Ordering"), Yes(nil), No())
}

type Opt = client.Option[any]

func Yes(yes any) Opt {
	return client.Some(yes)
}

func No() Opt {
	return client.None[any]()
}

// makeInputObject retrned a properly made input field type
// using name (outer), name of type (inner), and types ofType.
func makeInputObject(name Opt, typeName Opt, ofType Opt, inputFields Opt) Field {
	result := Field{}

	if name.HasValue() {
		result["name"] = name.Value()
	}

	typeValue := Field{}
	if typeName.HasValue() {
		typeValue["name"] = typeName.Value()
	}
	if ofType.HasValue() {
		typeValue["ofType"] = ofType.Value()
	}
	if inputFields.HasValue() {
		typeValue["inputFields"] = inputFields.Value()
	}
	result["type"] = typeValue

	return result
}

func BuildOrderArg(objectName string, fields []ArgDef) Field {
	inputFields := []any{
		makeInputObject(Yes("_key"), Yes("Ordering"), Yes(nil), No()),
	}

	for _, field := range fields {
		inputFields = append(
			inputFields,
			makeInputObject(Yes(field.FieldName), Yes(field.TypeName), Yes(nil), No()),
		)
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

func MakeDefaultGroupArgsWithout(skipThese []string) []any {
	defaultsWithSomeSkipped := make(
		[]any,
		0,
		len(allDefaultGroupArgs)-len(skipThese),
	)

ArgLoop:
	for _, argAny := range allDefaultGroupArgs {
		arg, ok := argAny.(map[string]any)
		if !ok {
			panic("wrong usage of testing schema factory method.")
		}

		argName, found := arg["name"]
		if !found {
			panic("`name` not found in the default group arg.")
		}

		for _, toSkip := range skipThese {
			if toSkip == argName {
				continue ArgLoop
			}
		}

		// If we make it till here we, don't want to skip this arg.
		defaultsWithSomeSkipped = append(defaultsWithSomeSkipped, argAny)
	}

	return defaultsWithSomeSkipped
}
