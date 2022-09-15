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
		makeOuterType(Yes("_and"), Yes(nil), Yes(map[string]any{
			"kind": "INPUT_OBJECT",
			"name": filterArgName,
		}), No()),
		makeOuterType(Yes("_key"), Yes("IDOperatorBlock"), Yes(nil), No()),
		makeOuterType(Yes("_not"), Yes("authorFilterArg"), Yes(nil), No()),
		makeOuterType(Yes("_or"), Yes(nil), Yes(map[string]any{
			"kind": "INPUT_OBJECT",
			"name": filterArgName,
		}), No()),
	}

	for _, field := range fields {
		inputFields = append(
			inputFields,
			makeOuterType(Yes(field.FieldName), Yes(field.TypeName), Yes(nil), No()),
		)
	}

	return makeOuterType(Yes("filter"), Yes(filterArgName), Yes(nil), Yes(inputFields))
}

type Opt = client.Option[any]

func Yes(yes any) Opt {
	return client.Some(yes)
}

func No() Opt {
	return client.None[any]()
}

// makeType returns the smallest inner type with all option keys that have value,
// the options that don't have values are skipped from the returned type.
func makeType(name, ofType, inputFields Opt) Field {
	inner := Field{}

	if name.HasValue() {
		inner["name"] = name.Value()
	}

	if ofType.HasValue() {
		inner["ofType"] = ofType.Value()
	}

	if inputFields.HasValue() {
		inner["inputFields"] = inputFields.Value()
	}

	return inner
}

// makeNameType returns a pair of name and pair type if both inputs have values,
// otherwise options that are invalid (i.e. `No()`) are ignored.
func makeNameType(name, innerType Opt) Field {
	result := Field{}

	if name.HasValue() {
		result["name"] = name.Value()
	}

	if innerType.HasValue() {
		result["type"] = innerType.Value()
	}

	return result
}

// makeOuterType returns the outer object type where all options that HasValue() == true
// are populated, and others are skipped.
//
// The retuned obeject always contains the "type" key (AKA the inner-type).
func makeOuterType(name, typeName, ofType, inputFields Opt) Field {
	return makeNameType(name, Yes(makeType(typeName, ofType, inputFields)))
}

func BuildOrderArg(objectName string, fields []ArgDef) Field {
	inputFields := []any{
		makeOuterType(Yes("_key"), Yes("Ordering"), Yes(nil), No()),
	}

	for _, field := range fields {
		inputFields = append(
			inputFields,
			makeOuterType(Yes(field.FieldName), Yes(field.TypeName), Yes(nil), No()),
		)
	}

	return makeOuterType(Yes("order"), Yes(objectName+"OrderArg"), Yes(nil), Yes(inputFields))
}

func MakeDefaultsWithout(skipThese []string) []any {
	defaultsWithSomeSkipped := make(
		[]any,
		0,
		len(allDefaultGroupArgs)-len(skipThese),
	)

ArgLoop:
	for _, arg := range allDefaultGroupArgs {
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
		defaultsWithSomeSkipped = append(defaultsWithSomeSkipped, arg)
	}

	return defaultsWithSomeSkipped
}
